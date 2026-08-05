# M1 契约与命名

本文件是 M1 实现的稳定约束。计划与代码冲突时先判断是否需要改变公共行为；若需要，先更新本文件并说明迁移影响。

## 服务与配置

- 二进制仍为 `apiserver`，默认 HTTP 地址仍为 `127.0.0.1:8080`。
- 新增必填环境变量 `LYAPUS_DATABASE_URL`；公开示例只能使用本地占位凭据。
- PostgreSQL 主版本固定为 16，当前 Compose patch 固定为 16.14。
- 使用 `pgx/v5` 的 `pgxpool.Pool`；应用启动必须在有超时的 context 内 `Ping`，不能把“创建了 pool”误当作连接成功。
- 应用退出时先停止接收 HTTP 流量，再关闭数据库连接池。
- migration 是显式部署步骤，不由 `apiserver` 启动路径隐式执行。

## 包边界

```text
cmd/apiserver
  ├── internal/platform/config
  ├── internal/platform/database
  ├── internal/platform/health
  ├── internal/platform/transport/http
  └── internal/catalog
        ├── postgres
        └── transport/http
```

- `internal/catalog` 定义模型、业务错误、repository 所需接口和业务服务。
- `internal/catalog/postgres/sqlcgen` 只保存 sqlc 生成代码；生成类型不得出现在业务服务或 HTTP 公共契约中。
- `internal/catalog/postgres` 用 adapter 实现 catalog repository，只向上返回 domain model、稳定业务错误或带上下文的未知错误。
- `internal/catalog/transport/http` 负责 HTTP 解析与呈现，不包含 SQL。
- `cmd/apiserver` 只装配依赖、启动、等待信号和关闭，不放业务规则。
- 不建立全局数据库连接、全局 service locator 或通用 CRUD/repository 泛型。

## SQL 与代码生成

- `db/schema.sql` 同时是 Atlas diff 与 sqlc 分析的 schema 真源；不得另建一份只为生成代码服务的漂移 schema。
- 业务查询由项目所有者手写在 `db/queries/*.sql`，使用 sqlc 命名注解生成 `pgx/v5` 调用。
- sqlc 版本固定为 1.31.1，配置位于根目录 `sqlc.yaml`，生成目标为 `internal/catalog/postgres/sqlcgen`。
- 生成代码提交到 Git，使普通构建不依赖本机已安装 sqlc；任何人不得手工编辑生成文件。
- `make generate` 更新生成代码；`make generate-check` 使用固定版本的本地 `sqlc diff --no-remote` 比较当前 schema/query 的预期生成结果与磁盘生成代码，防止漏生成或手改生成文件。该检查不读取 Git 暂存区，required CI 必须执行。
- repository adapter 负责 sqlc 参数/行类型与 domain model 的映射、SQLSTATE 分类和事务边界；sqlc 不替代这些职责。
- 初次生成每类查询时必须阅读对应生成函数，确认参数化 SQL、返回行和 `Scan` 顺序；“由工具生成”不是跳过理解的理由。

## 数据模型

三个主表都使用 `bigint generated always as identity` 主键和 `timestamptz` 时间；时间由数据库写入并以 UTC 对外呈现。

### Team

| 字段 | 约束 |
| --- | --- |
| `id` | 主键；创建后不可变。 |
| `slug` | 1–63 个字符；小写字母开头，只含小写字母、数字或连字符；全局唯一。 |
| `name` | trim 后 1–100 个字符。 |
| `created_at`、`updated_at` | 非空；`updated_at` 不早于 `created_at`。 |

### Service

| 字段 | 约束 |
| --- | --- |
| `id` | 主键；创建后不可变。 |
| `team_id` | 非空外键指向 Team；创建后不可变。 |
| `slug` | 与 Team slug 同一格式；在同一 `team_id` 内唯一。 |
| `name` | trim 后 1–100 个字符。 |
| `description` | 可为空；最多 500 个字符。 |
| `created_at`、`updated_at` | 非空；`updated_at` 不早于 `created_at`。 |

### Environment

| 字段 | 约束 |
| --- | --- |
| `id` | 主键；创建后不可变。 |
| `service_id` | 非空外键指向 Service；创建后不可变。 |
| `slug` | 与 Team slug 同一格式；在同一 `service_id` 内唯一。 |
| `name` | trim 后 1–100 个字符。 |
| `created_at`、`updated_at` | 非空；`updated_at` 不早于 `created_at`。 |

归属关系：

- 一个 Team 可以拥有多个 Service；一个 Service 只属于一个 Team。
- 一个 Service 可以拥有多个 Environment；一个 Environment 只属于一个 Service。
- 删除仍有 Service 的 Team 返回冲突；删除仍有 Environment 的 Service 返回冲突。M1 不做隐式级联删除或软删除。
- 数据库中的外键、非空、检查和唯一约束是最终正确性边界；Go 校验用于更早、更清晰地返回错误，但不能替代数据库约束。

## HTTP 通用契约

- API 前缀为 `/v1`，Content-Type 为 `application/json`。
- 请求体最大 1 MiB；解码器拒绝未知字段、多个 JSON 值和类型不匹配。
- ID 是十进制正整数；非法 ID 返回 `400 invalid_argument`。
- 时间使用 UTC RFC 3339 字符串。
- 创建成功返回 `201` 与资源；查询/更新成功返回 `200`；删除成功返回 `204` 且无响应体。
- 每个请求生成新的请求 ID，在 `X-Request-ID` 响应头、错误体和完成日志中使用同一值。M1 不信任或透传外部传入的请求 ID。
- 所有 handler 接收并向 repository 传递 `r.Context()`；不把 context 存入结构体。

错误响应固定为：

```json
{
  "code": "invalid_argument",
  "message": "name must not be empty",
  "request_id": "..."
}
```

| HTTP 状态 | `code` | 语义 |
| --- | --- | --- |
| 400 | `invalid_argument` | JSON、路径、游标或业务参数不合法。 |
| 404 | `not_found` | 目标资源或必需的父资源不存在。 |
| 409 | `conflict` | 唯一性冲突或资源仍被子资源引用。 |
| 500 | `internal` | 未分类内部错误；响应不泄漏 SQL、连接串或内部路径。 |
| 503 | `not_ready` | `/readyz` 的数据库检查失败。 |

PostgreSQL 可识别的 unique violation、foreign key violation 等通过 SQLSTATE 分类；不得用错误字符串匹配。

## HTTP 资源契约

所有资源提供创建、列表、单项查询、部分更新和删除：

| 资源 | 集合路径 | 单项路径 | 可选列表过滤 |
| --- | --- | --- | --- |
| Team | `/v1/teams` | `/v1/teams/{team_id}` | 无 |
| Service | `/v1/services` | `/v1/services/{service_id}` | `team_id` |
| Environment | `/v1/environments` | `/v1/environments/{environment_id}` | `service_id` |

集合支持 `POST`、`GET`，单项支持 `GET`、`PATCH`、`DELETE`。不实现 `PUT`。

- `POST /v1/services` 可携带 `environments` 数组；Service 和初始 Environment 必须在一个数据库事务中全部成功或全部失败。
- `PATCH` 只允许可变字段，至少提供一个字段；显式空字符串仍要经过校验。
- `team_id` 与 `service_id` 是不可变归属。M1 不支持跨 Team 或跨 Service 搬迁资源。
- Service 单项响应包含其 Environment 列表；Service 集合响应不展开 Environment，避免列表放大。

## 分页与排序

- 列表参数为 `limit` 和 `cursor`；`limit` 默认 20，范围 1–100。
- 顺序固定为 `created_at DESC, id DESC`。
- cursor 是不透明的 URL-safe base64 token，内部绑定上一页最后一项的 `created_at` 与 `id`；调用方不得依赖其内部格式。
- 非法 cursor 返回 `400 invalid_argument`，空结果返回空 `items`。
- 列表响应固定为：

```json
{
  "items": [],
  "next_cursor": ""
}
```

- Service 的 `team_id` 和 Environment 的 `service_id` 是等值过滤；M1 不实现关键字搜索、任意排序或过滤表达式。

## 事务与并发正确性

- `CreateService` 始终以一个显式 pgx 事务写入 Service 与可选初始 Environment，并通过 sqlc 生成的 `Queries.WithTx` 在同一事务上执行查询。
- 任一步失败必须 rollback；只有全部写入成功才能 commit。commit 失败也视为整体失败，不能返回创建成功。
- 同一 Team 下 `(team_id, slug)` 由数据库唯一约束保证。并发测试至少启动两个独立连接/事务同时创建同一 slug，并断言恰好一个成功、一个返回稳定 `conflict`。
- Go 进程内 mutex 不是该场景的正确性机制，因为它不能覆盖多进程或多实例。

## 健康检查

| 路径 | 成功 | 失败 | 语义 |
| --- | --- | --- | --- |
| `/livez` | `200 {"status":"ok"}` | 仅进程无法服务时失败 | 不检查数据库。 |
| `/readyz` | `200 {"status":"ok"}` | `503` 统一错误 | 在短超时 context 内检查 PostgreSQL。 |

数据库暂时不可用不能使 `/livez` 失败。ready 检查不能无限等待或创建无界连接。

## Migration 契约

- `db/schema.sql` 表达期望 schema，`db/migrations/` 保存按序、不可原地修改的 versioned SQL。
- 通过评审的 migration 与完整性校验文件一同提交；生成后的 SQL 必须人工解释锁、表重写、约束、数据回填和失败恢复影响。
- 空数据库只通过 versioned migration 建立；不能以直接 `schema apply` 代替可审计迁移历史。
- 已在共享历史中应用的 migration 不修改；变化通过新 migration 前滚。
- M1 的恢复范围是可丢弃的 dev/test 数据库重建。未演练备份恢复前，不声明生产 rollback 能力。

最终工具命令以 P-0001 验证后建立的 ADR 与 `docs/runbooks/database-migrations.md` 为准。

## 测试契约

- 单元测试不连接数据库，覆盖参数校验、业务错误、事务调用边界、JSON、错误映射、分页和 handler。
- 单元测试依赖 catalog repository 接口，不直接 mock 或暴露 sqlc 生成类型。
- integration test 使用真实 PostgreSQL 16.14 与真实 migrations，不以 mock 或 SQLite 替代。
- 集成测试数据库名必须以 `_test` 结尾；测试辅助代码在清库或迁移前主动校验，防止误操作开发库。
- 本地集成测试使用独立 Compose project 和一次性卷；测试结束清理该 project，不操作日常开发卷。
- 并发测试使用独立数据库连接并记录同步起点，不能用串行循环冒充并发。
- `make verify` 在本地和 required CI 中都必须真实执行 integration tests 与 race detector；缺少数据库依赖时应失败并给出明确说明，不能静默跳过。
