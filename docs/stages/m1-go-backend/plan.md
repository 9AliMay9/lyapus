# M1：常规 Go 后端

## 目标与完成定义

实现一个不依赖后续 SRE、可观测性或分布式阶段也能独立成立的 Service Catalog API，集中练习 Go HTTP 服务、PostgreSQL 数据建模、约束、事务、并发正确性、测试和可复现交付。

满足以下条件才算完成 M1 v0.1：

- Team、Service、Environment 三类资源及归属关系可以通过 `/v1` HTTP API 管理。
- 参数校验、统一错误、游标分页和基本归属过滤具有稳定契约。
- PostgreSQL migration、外键和唯一约束可以从空库重复建立。
- 一个跨多条 SQL 的写操作具备事务原子性；同一团队下并发创建同名 Service 时，数据库保证最多一个成功，失败方稳定映射为 `409 conflict`。
- 关键业务规则有单元测试；repository 有真实 PostgreSQL 集成测试；`go test -race ./...` 通过。
- Docker Compose 可以从空环境启动数据库、执行 migration 并运行 API。
- 保存一份带数据量、SQL、`EXPLAIN (ANALYZE, BUFFERS)` 和索引前后对比的查询优化记录。
- 根 README 包含数据模型、API 示例和五分钟演示路径。

## 本阶段必须完成

1. 先完成 Atlas Community 可行性实验，并将通过后的选择或回退结果写成 ADR。
2. 固定 PostgreSQL 16.14、连接驱动、连接池、迁移目录、测试库隔离方式和显式迁移入口。
3. 建立 Team、Service、Environment 模型、数据库约束和 versioned migrations。
4. 实现 repository、业务服务和 HTTP transport；依赖方向保持业务层不依赖 HTTP 或 PostgreSQL 细节。
5. 实现请求 ID、严格 JSON 解码、统一错误映射、游标分页和基础请求日志。
6. 让 `/livez` 保持进程存活语义，让 `/readyz` 真实检查数据库可用性。
7. 实现单元测试、真实 PostgreSQL 集成测试和并发冲突测试，并接入现有 `make verify` 与 required checks。
8. 建立 Dockerfile、Compose 空环境路径、migration runbook 和 API 演示文档。
9. 完成一次可复现查询计划实验，并如实记录未证明的性能边界。

## 技术默认项

| 问题 | M1 默认项 | 原因与边界 |
| --- | --- | --- |
| HTTP 路由 | `chi/v5`，handler 保持标准 `net/http` | M1 有三类 CRUD、分组路由和 middleware 链；chi 小而可组合，仍完全兼容 `net/http.Handler`，不会把业务层绑定到框架 context。 |
| PostgreSQL | 16.14 | 当前仍受上游支持，并与开发主机已有 `psql` 主版本一致；Compose 固定具体版本，不使用 `latest`。 |
| 驱动与连接池 | `pgx/v5`、`pgxpool` | PostgreSQL 原生、支持 context、事务和 SQLSTATE；实际 patch 版本由 `go.mod` 固定。 |
| 数据访问 | `sqlc` 1.31.1 + 手写 SQL + repository adapter | SQL 仍由项目所有者编写和解释；sqlc 生成 pgx/v5 类型安全调用，repository 负责事务、业务类型映射和错误分类。生成代码不越过 adapter。 |
| 标识符 | PostgreSQL `bigint generated always as identity` | 简单、无额外生成依赖，便于解释索引与游标；顺序 ID 不是授权边界，未来事件 ID 可独立设计。 |
| 迁移 | Atlas Community 1.2.0：声明式期望 schema 生成 versioned SQL | P-0001 已通过并由 ADR-0004 接受；同时保留 schema-as-code 和可审阅迁移，不在应用启动时自动迁移，不依赖 Atlas Cloud/Pro。 |
| 测试数据库 | Compose 独立 project + 名称带 `_test` 的一次性数据库 | 与开发数据隔离；migration 前重建，测试后删除卷。测试代码必须拒绝非测试库。 |

这些是当前阶段的最优默认项，不代表工具流行度或生产能力已经被本项目证明。

### 主要替代与验证方式

- PostgreSQL 不追逐当时最新主版本 18：16.14 仍在官方支持期，版本更成熟且与当前主机客户端一致；M1 会通过真实 migration 和查询计划验证所需能力，而不是用版本号证明先进性。
- `pgxpool` 相对标准库 `database/sql` 更直接暴露 PostgreSQL、SQLSTATE 和连接池能力；若未来需要多数据库驱动兼容才复评，M1 不为假设需求付出抽象成本。
- chi 相对直接使用 ServeMux 多一项依赖，但更清晰地表达当前三类资源、子路由和 middleware；它仍使用标准 handler，因此可以低成本换回 ServeMux。Gin 的自定义 context 和更完整框架抽象对 M1 没有额外收益。
- sqlc 相对全手写 `QueryRow`/`Scan` 少了机械样板并增加 schema/query 类型检查，但不生成业务规则或隐藏 SQL。项目所有者仍手写查询、阅读生成代码并解释执行计划；repository adapter 防止生成类型扩散。ORM 不符合本阶段直接学习 SQL 的目标。
- identity bigint 相对 UUID 更简单且便于观察 B-tree 与复合游标；顺序 ID 不承担保密或授权职责，若 M2 的跨系统事件需要全局生成，再为事件单独选择标识符。
- Compose 隔离相对一开始引入 testcontainers-go 更少工具层；先证明真实数据库测试生命周期，若并行测试或 CI 隔离变得脆弱，再用 testcontainers-go 解决已经出现的问题。

上游事实入口：

- PostgreSQL versioning policy：https://www.postgresql.org/support/versioning/
- pgxpool 文档：https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool
- chi：https://github.com/go-chi/chi
- sqlc：https://docs.sqlc.dev/en/stable/

## 明确不做

- API Token、RBAC、审计日志、复杂幂等键、乐观锁。
- ORM、通用 repository 框架或为未来数据库替换设计的抽象。
- Gin、Echo 等完整 HTTP 框架；M1 只使用 chi core 和确有需要的 middleware。
- Kafka、Redis、OpenTelemetry、Kubernetes、前端或 AI 功能。
- k6 容量报告、pprof 深度优化、高可用、备份恢复或零停机迁移声明。
- 由应用进程在启动时自动执行 migration。
- Atlas Cloud、付费 lint/drift 功能或需要 token 才能复现的工作流。

以上可作为 M1 v0.1.x 或后续阶段增强，但不得阻塞首次可投递 release。

## 实施顺序

每一段先阅读相应契约，由项目所有者手敲核心代码，完成本段验证后再进入下一段。

1. **迁移工具小实验（已完成）**：P-0001 已在空 PostgreSQL 16.14 上完成 diff、人工审阅、apply、status、重复执行和完整性校验，并由 ADR-0004 接受 Atlas Community。
2. **数据库基础设施（已完成）**：扩展配置，建立 `pgxpool`，启动时显式连接检查，关闭时释放连接；把 `/readyz` 接到有超时的数据库 ping。
3. **schema 与 migrations（基线已完成）**：期望 schema、两份 versioned migration、空库与前滚路径已建立；具体约束成功/失败行为仍须由真实 PostgreSQL 集成测试验证。
4. **queries 与 repository（进行中）**：Team Create/Get/List/Update/Delete、错误分类、基于 `(created_at, id)` 的稳定游标分页和真实数据库集成测试已完成；下一步实现 Team 业务服务，再按同一路径完成 Service 和 Environment。
5. **业务服务**：集中放置校验、归属规则和“Service + 初始 Environments”事务；完成并发唯一性测试。
6. **HTTP transport**：实现请求 ID、JSON/错误工具、`/v1` 路由、CRUD、游标分页和归属过滤；先单元测试 handler，再接真实 repository。
7. **交付路径**：建立 Dockerfile、Compose、migration runbook，确保空卷可以按顺序完成 migration、启动和 API 演示。
8. **查询计划实验**：构造明确数据规模，对 Service 按 Team 的游标列表查询保存索引前后证据。
9. **收口**：更新根 README、架构图、知识笔记、`outcome.md` 和当前进度；在 clean runner 与空 Compose project 上完成最终验收。

## 代码施工清单

路径是施工目标；细小辅助文件可在不改变分层的前提下增加，公共命名变化必须先更新 `contracts.md`。

```text
cmd/apiserver/main.go
internal/platform/config/config.go
internal/platform/config/config_test.go
internal/platform/database/postgres.go
internal/platform/health/handler.go
internal/platform/health/handler_test.go
internal/platform/transport/http/server.go
internal/platform/transport/http/server_test.go
internal/platform/transport/http/middleware.go
internal/platform/transport/http/response.go

internal/catalog/model.go
internal/catalog/errors.go
internal/catalog/repository.go
internal/catalog/service.go
internal/catalog/service_test.go
internal/catalog/postgres/repository.go
internal/catalog/postgres/repository_integration_test.go
internal/catalog/postgres/sqlcgen/
internal/catalog/transport/http/handler.go
internal/catalog/transport/http/handler_test.go

db/schema.sql
db/migrations/
db/queries/teams.sql
db/queries/services.sql
db/queries/environments.sql
atlas.hcl
sqlc.yaml
compose.yaml
Dockerfile
.dockerignore
.env.example
Makefile
.github/workflows/verify.yml
README.md

docs/runbooks/database-migrations.md
docs/benchmarks/m1-service-list-query-plan.md
```

完整实现只存在于源码、SQL 和 Git commit 中；文档只保留契约、操作步骤和实验证据。

## 验收命令

Migration 命令以 ADR-0004 与已实测 runbook 为准；Compose 命令仍须在对应交付路径完成后实测。最终至少能够从仓库根目录安全执行：

```bash
make fmt
make generate
make generate-check
make test
make race
make integration
make verify
docker compose config
docker compose up --build
```

最终验收还必须包含 API CRUD 示例、数据库不可用时 `/readyz` 失败、并发冲突、空卷迁移和查询计划实验；命令与结果记录在 `outcome.md` 及对应证据文档。

## 交付证据

- M1 v0.1 Git commit 与 release：待完成。
- 迁移工具 ADR：ADR-0004 已 accepted。
- 数据库 migration runbook：核心 diff/apply/status 路径已在本机与 required CI 实测。
- PostgreSQL 集成测试：Team Create/Get/List/Update/Delete、唯一冲突、外键引用删除冲突、not-found 映射和游标分页已在本地与 PR #13 clean-runner 验证；其余约束、事务与并发场景待实现。
- Compose 空环境记录：待验证。
- 查询计划对比：待实验。
- M1 学习总结和会话记录：待完成。

## 下一阶段入口

只有 `checklist.md` 的 v0.1 必需项全部完成、`outcome.md` 有真实证据且原始方案七项最小交付均成立后，才能发布 M1 v0.1 并进入 M2。M1 深度增强不阻塞该入口。
