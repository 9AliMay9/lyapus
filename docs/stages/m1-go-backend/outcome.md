# M1 实际实施结果

状态：M1 尚未完成；数据库基础设施施工包已完成，以下只记录已发生的事实。

## 实际完成

- `LYAPUS_DATABASE_URL` 成为必填配置；仅接受 `postgres` 或 `postgresql` URL，配置与日志不输出连接串。
- `internal/platform/database.Open` 创建 `pgxpool.Pool` 后显式 Ping；Ping 失败时关闭 pool，调用方以五秒启动 context 约束等待时间。
- HTTP 进程在收到 `SIGINT` 或 `SIGTERM` 后先停止接收 HTTP 流量，再关闭连接池。
- `/livez` 不访问数据库；`/readyz` 以一秒有界 context 执行 Ping，成功返回 200，数据库不可用时返回 503。

## 验证

- `go test ./...`：2026-07-30 本地通过。
- 使用一次性 PostgreSQL 16.14 容器：应用启动时 Ping 成功，`/livez` 与 `/readyz` 均返回 200。
- 保持 API 进程运行时停止数据库：`/livez` 仍返回 200，`/readyz` 返回 503；重新建立同配置数据库后，未重启 API 的 `/readyz` 恢复 200。
- API 进程收到 `Ctrl-C` 后记录关闭信号与 HTTP server 停止；随后一次性数据库容器已停止并自动删除。
- PR #10 的 `verify`、数据库感知 `smoke` 与 `atlas-community` required checks 均通过；smoke 在 PostgreSQL 16.14 service 上启动 API，Atlas job 在真实 SQL 就绪确认后验证 migration。

## 与计划的偏差

- 尚未引入 Compose；本次使用一次性容器仅作为运行证据，不能替代最终 Compose 空环境验收。
- `/readyz` 暂时返回最小探针体 `{"status":"not_ready"}`。请求 ID 与统一错误信封将在 HTTP transport 施工包实现后统一纳入，不能提前宣称该公共契约已完成。

## 证据

- Git commit / release：当前施工包待提交；M1 release 待完成。
- Migration ADR 与 runbook：ADR-0004 与 migration runbook 已完成。
- 查询计划 benchmark：待完成。
- 会话与学习记录：数据库基础设施会话已记录；M1 总结待完成。

## 施工包进入 M2 的条件

- [ ] `checklist.md` 中 M1 v0.1 必需项全部完成。
- [ ] 原始方案七项最小交付全部有可重现证据。
- [ ] 没有把 M1 深度增强、性能实验或后续阶段写成已完成。

结论：M1 当前未完成，不能进入 M2。
