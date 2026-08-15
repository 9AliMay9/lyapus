# M1 实际实施结果

状态：M1 尚未完成；数据库基础设施、Team repository、Team 业务服务与完整 Team HTTP CRUD 纵切面已完成，以下只记录已发生的事实。

## 实际完成

- `LYAPUS_DATABASE_URL` 成为必填配置；仅接受 `postgres` 或 `postgresql` URL，配置与日志不输出连接串。
- `internal/platform/database.Open` 创建 `pgxpool.Pool` 后显式 Ping；Ping 失败时关闭 pool，调用方以五秒启动 context 约束等待时间。
- HTTP 进程在收到 `SIGINT` 或 `SIGTERM` 后先停止接收 HTTP 流量，再关闭连接池。
- `/livez` 不访问数据库；`/readyz` 以一秒有界 context 执行 Ping，成功返回 200，数据库不可用时返回 503。
- Team 的 Create/Get/List/Update/Delete 查询由 sqlc 生成 pgx 调用；PostgreSQL adapter 将生成行映射为 domain `catalog.Team`。列表以 `(created_at DESC, id DESC)` 排序，以“多取一行”计算下一页游标；生成类型仍不越过 adapter。
- PostgreSQL adapter 将 no rows 映射为稳定的 `ErrNotFound`，将 unique violation 与 foreign-key violation 映射为 `ErrConflict`；非法 repository 分页 limit 映射为 `ErrInvalidArgument`。
- `catalog.TeamService` 集中校验 Team ID、slug、name、更新输入与分页输入；它只依赖 domain repository 接口，不暴露 HTTP、pgx 或 sqlc 类型。
- `POST /v1/teams`、`GET /v1/teams/{team_id}` 与 `GET /v1/teams` 已通过 chi 接到真实 Team service/repository；请求体使用 1 MiB 上限、拒绝未知字段和多个 JSON 值，并将 catalog 的参数、not-found、冲突与未知错误映射为稳定 JSON 错误信封。集合读取将 URL-safe base64 cursor 解码为 domain `(created_at, id)`，并将下一页 cursor 编码为不透明 token。仅未提供 `limit` 时使用默认值；显式空值、零值或重复的 `limit`/`cursor` 都映射为 `400 invalid_argument`。
- request-ID 已提升至 `internal/platform/requestid`，包围整个服务 mux；健康检查与 `/v1` 都生成新的 ID，并在响应头、catalog 错误体和完成日志中复用。完成日志同时记录 method、path、status 和 duration。

## 验证

- `go test ./...`：2026-07-30 本地通过。
- 使用一次性 PostgreSQL 16.14 容器：应用启动时 Ping 成功，`/livez` 与 `/readyz` 均返回 200。
- 保持 API 进程运行时停止数据库：`/livez` 仍返回 200，`/readyz` 返回 503；重新建立同配置数据库后，未重启 API 的 `/readyz` 恢复 200。
- API 进程收到 `Ctrl-C` 后记录关闭信号与 HTTP server 停止；随后一次性数据库容器已停止并自动删除。
- PR #10 的 `verify`、数据库感知 `smoke` 与 `atlas-community` required checks 均通过；smoke 在 PostgreSQL 16.14 service 上启动 API，Atlas job 在真实 SQL 就绪确认后验证 migration。
- 在可丢弃 `_test` PostgreSQL 16.14 数据库中，显式 apply 两份 migration 后，Team repository integration test 已验证 Create、Get、List 分页、Update、Delete、唯一冲突、外键引用删除冲突与 not-found 映射；普通测试与 race 检测均通过。
- 2026-08-05，本地 `make verify` 通过：生成检查、`go vet`、普通测试、race、真实 PostgreSQL integration test 与漏洞扫描均为成功。PR #13 的 `verify`、`smoke` 与 `atlas-community` required checks 全部通过；clean-runner `verify` 在独立 PostgreSQL 16.14 service 上执行了完整 Team repository integration 路径。
- 2026-08-10，Team service、catalog HTTP、共享 request-ID 与 platform server 的普通测试及 `go test -race ./...` 通过。对可丢弃的本地 PostgreSQL 16.14 开发容器执行 migration 后，`/livez`、`/readyz`、Team 创建、非法 slug 与唯一冲突分别实测为 200、200、201、400、409；每条响应/错误与完成日志的 request ID 一致。验证后 API 进程和一次性容器均已停止。
- 2026-08-11，在新建的可丢弃 `_test` PostgreSQL 16.14 数据库完成 migration dry-run、apply、status 后，PR #15 合并前的本地 `make verify` 通过：格式化、生成新鲜度、`go vet`、普通测试、race、真实 PostgreSQL integration test 与漏洞扫描均成功。CI smoke 已扩展为以同一版本化 migration 准备临时数据库后验证 `/livez`、`/readyz` 及 Team 的 400、201、409 路径；PR #15 的 `verify`、`atlas-community` 与该 smoke 均在 clean runner 通过，squash commit `8e84c20` 已合入 `main`。
- 2026-08-13，在可丢弃 PostgreSQL 16.14 开发库完成 migration dry-run、apply、status 后，实测创建两个 Team、单项读取、`limit=1` 第一页和携带 `next_cursor` 的第二页；结果按 `(created_at DESC, id DESC)` 返回，cursor 分页无重复或遗漏。随后在独立 `_test` 库完成 migration dry-run、apply、status，并以该库运行 `make verify`；`go vet`、普通测试、race、真实 PostgreSQL integration test 与漏洞扫描均成功。待 PR 的 clean-runner 验收更新后的 smoke 断言。
- 2026-08-14，`govulncheck` 识别出 Go 1.26.5 标准库中四个代码路径可达漏洞；本机工具链与 `go.mod` 均升级到 Go 1.26.6（扫描报告给出的修复版本）后，以同一独立 `_test` PostgreSQL 16.14 数据库重跑 `make verify`，普通/竞态/真实 integration test 通过，漏洞扫描恢复为 `No vulnerabilities found.`。
- 2026-08-14，PR #17 的 required `verify`、`smoke` 与 `atlas-community` checks 全部通过；其中 clean-runner smoke 在版本化 migration 后验证 Team 创建、单项读取、`limit=1` 列表读取及既有错误路径。
- 2026-08-15，Team HTTP PATCH/DELETE 的普通与竞态测试通过；在可丢弃 PostgreSQL 16.14 开发库完成 migration dry-run、apply、status 后，实测 PATCH 成功、空更新 400、唯一 slug 冲突 409、DELETE 空 204 与删除后查询 404，响应与完成日志的 request ID 一致。独立 `_test` 库完成 migration dry-run、apply、status；`make verify` 通过。更新后的 CI smoke 待本施工包 PR 的 clean-runner 验收。

## 与计划的偏差

- 尚未引入 Compose；本次使用一次性容器仅作为运行证据，不能替代最终 Compose 空环境验收。
- `/readyz` 暂时返回最小探针体 `{"status":"not_ready"}`；它已有全局 request-ID，但尚未改成 catalog 的错误信封，仍应保持健康探针的最小契约。
- Team repository、业务校验和完整 Team HTTP CRUD 已完成；Service/Environment HTTP 资源契约仍未实现。

## 证据

- Git commit / release：数据库基础设施已由 `df0154d`（PR #10）合入；Team Create/Get repository 已由 `57f19d4`（PR #11）合入；Team CRUD/稳定分页已由 `f4df01f`（PR #13）合入；M1 release 待完成。
- Migration ADR 与 runbook：ADR-0004 与 migration runbook 已完成。
- 查询计划 benchmark：待完成。
- 会话与学习记录：数据库基础设施会话已记录；M1 总结待完成。

## 施工包进入 M2 的条件

- [ ] `checklist.md` 中 M1 v0.1 必需项全部完成。
- [ ] 原始方案七项最小交付全部有可重现证据。
- [ ] 没有把 M1 深度增强、性能实验或后续阶段写成已完成。

结论：M1 当前未完成，不能进入 M2。
