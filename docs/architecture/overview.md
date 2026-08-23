# 当前架构

当前是一个模块化单体 Go 服务，二进制入口为 `cmd/apiserver`。

```text
LYAPUS_HTTP_ADDR ─────┐
                     ├──→ internal/platform/config ──→ cmd/apiserver
LYAPUS_DATABASE_URL ─┘                                  ├──→ JSON slog ──→ stdout
                                                        ├──→ pgxpool ──→ PostgreSQL
                                                        └──→ HTTP transport
                                                               ├──→ request-ID + completion log
                                                               ├──→ /livez
                                                               ├──→ /readyz ──→ bounded pool.Ping
                                                               ├──→ /v1/teams POST/GET, /v1/teams/{team_id} GET/PATCH/DELETE ──→ TeamService
                                                               └──→ /v1/services POST/GET, /v1/services/{service_id} GET ──→ ServiceService

internal/catalog domain ←── PostgreSQL repository adapter ←── sqlcgen
        （Team CRUD 与 Service Create/Get/List、稳定分页、事务创建和业务服务已实现）
```

- 默认监听地址是 `127.0.0.1:8080`；可用 `LYAPUS_HTTP_ADDR` 覆盖，非法地址会在启动前失败。
- `LYAPUS_DATABASE_URL` 必填；进程启动时以五秒有界 context 创建并 Ping `pgxpool`，失败则不启动 HTTP 服务，退出时先关闭 HTTP 再释放连接池。
- `/livez` 只表示进程存活，不访问数据库；`/readyz` 以一秒有界 context Ping PostgreSQL，成功返回 `200`，失败返回 `503`。
- HTTP server 为每个请求生成 request ID，并输出含 request ID、方法、路径、状态、耗时的 JSON 完成日志；收到 `SIGINT` 或 `SIGTERM` 后以 10 秒超时执行优雅关闭。
- `internal/catalog` 已有 Team、Service 与 Environment domain model。Team 已具备完整 repository/service/HTTP CRUD；Service 已具备 Create/Get/List repository 与 service、基于 `(created_at, id)` 的全局及 `team_id` 过滤分页，以及 POST/GET HTTP 路由。创建 Service 与可选初始 Environment 使用一个显式 pgx 事务；Service 单项响应展开 Environment，集合响应不展开。chi、HTTP DTO 与 sqlc 类型均未进入 domain。

当前仍未实现 Service PATCH/DELETE、Environment 独立 repository/service/HTTP CRUD、Compose、Kafka、OpenTelemetry、Kubernetes、前端或 AI 组件；这些内容不能画入已实现数据流。

## 与原始目录示意的映射

v3.1 方案书的 M0–M2 目录树是职责与演化方向示意，不是要求第一天逐字建立的固定路径；同一节同时要求“不提前建立空目录”。当前映射如下：

| 原始示意 | 当前实现 | 说明 |
| --- | --- | --- |
| `internal/catalog` | `internal/catalog` | 保存 domain model、repository 接口、业务服务与 catalog 专属 transport。 |
| `internal/storage` | `internal/catalog/postgres` | PostgreSQL adapter 目前只服务 catalog，先放在业务边界内；出现跨业务共享存储基础设施时再复评是否上移。 |
| `internal/transport` | `internal/platform/transport/http` 与 `internal/catalog/transport/http` | 前者负责 server/mux/日志等平台装配，后者负责 catalog 路由、DTO 与错误呈现，避免业务 HTTP 细节进入平台包。 |
| `migrations` | `db/schema.sql`、`db/migrations`、`db/queries` | 将声明式 schema、版本迁移和 sqlc 查询集中在同一数据库目录；职责未减少。 |
| `tests/integration` | 各包内 `*_integration_test.go` | 测试与被测 adapter 同包放置，仍通过 `integration` build tag 和独立 `_test` 数据库运行。 |
| `docs/adr` | `docs/architecture/decisions` | 使用完整名称区分 accepted decision 与 proposal。 |
| `deploy/compose`、`configs`、`cmd/worker`、`internal/auth` | 尚未建立 | 对应能力尚未进入当前实现；遵守“不提前建立空目录”。Compose 会在 M1 交付施工包中建立，worker/auth 分别等待后续真实需求。 |

这些差异保持方案书的结构规则：单一 Go module、`cmd/` 只装配、业务默认进入 `internal/`、数据库类型不泄漏到 API、没有真实第二用例前不抽象通用层。
