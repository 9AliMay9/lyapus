# M1 实际实施结果

状态：M1 尚未完成；数据库基础设施和 Team repository 施工包已完成，以下只记录已发生的事实。

## 实际完成

- `LYAPUS_DATABASE_URL` 成为必填配置；仅接受 `postgres` 或 `postgresql` URL，配置与日志不输出连接串。
- `internal/platform/database.Open` 创建 `pgxpool.Pool` 后显式 Ping；Ping 失败时关闭 pool，调用方以五秒启动 context 约束等待时间。
- HTTP 进程在收到 `SIGINT` 或 `SIGTERM` 后先停止接收 HTTP 流量，再关闭连接池。
- `/livez` 不访问数据库；`/readyz` 以一秒有界 context 执行 Ping，成功返回 200，数据库不可用时返回 503。
- Team 的 Create/Get/List/Update/Delete 查询由 sqlc 生成 pgx 调用；PostgreSQL adapter 将生成行映射为 domain `catalog.Team`。列表以 `(created_at DESC, id DESC)` 排序，以“多取一行”计算下一页游标；生成类型仍不越过 adapter。
- PostgreSQL adapter 将 no rows 映射为稳定的 `ErrNotFound`，将 unique violation 与 foreign-key violation 映射为 `ErrConflict`；非法 repository 分页 limit 映射为 `ErrInvalidArgument`。

## 验证

- `go test ./...`：2026-07-30 本地通过。
- 使用一次性 PostgreSQL 16.14 容器：应用启动时 Ping 成功，`/livez` 与 `/readyz` 均返回 200。
- 保持 API 进程运行时停止数据库：`/livez` 仍返回 200，`/readyz` 返回 503；重新建立同配置数据库后，未重启 API 的 `/readyz` 恢复 200。
- API 进程收到 `Ctrl-C` 后记录关闭信号与 HTTP server 停止；随后一次性数据库容器已停止并自动删除。
- PR #10 的 `verify`、数据库感知 `smoke` 与 `atlas-community` required checks 均通过；smoke 在 PostgreSQL 16.14 service 上启动 API，Atlas job 在真实 SQL 就绪确认后验证 migration。
- 在可丢弃 `_test` PostgreSQL 16.14 数据库中，显式 apply 两份 migration 后，Team repository integration test 已验证 Create、Get、List 分页、Update、Delete、唯一冲突、外键引用删除冲突与 not-found 映射；普通测试与 race 检测均通过。
- 2026-08-05，本地 `make verify` 通过：生成检查、`go vet`、普通测试、race、真实 PostgreSQL integration test 与漏洞扫描均为成功。PR #13 的 `verify`、`smoke` 与 `atlas-community` required checks 全部通过；clean-runner `verify` 在独立 PostgreSQL 16.14 service 上执行了完整 Team repository integration 路径。

## 与计划的偏差

- 尚未引入 Compose；本次使用一次性容器仅作为运行证据，不能替代最终 Compose 空环境验收。
- `/readyz` 暂时返回最小探针体 `{"status":"not_ready"}`。请求 ID 与统一错误信封将在 HTTP transport 施工包实现后统一纳入，不能提前宣称该公共契约已完成。
- Team repository 已完成 CRUD 与游标分页；业务校验、HTTP 资源契约、HTTP cursor 编解码与公开错误映射仍未实现。

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
