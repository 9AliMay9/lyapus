# M1 清单

状态：按实际完成证据持续维护；勾选只代表对应条目本身已验证，不代表 M1 v0.1 已完成。

## 施工包与决策门

- [x] 对照原始 v3.1 方案书确定 M1 v0.1 七项最小交付。
- [x] 建立 `plan.md`、`contracts.md`、`checklist.md` 和 `outcome.md`。
- [x] 项目所有者审阅并确认施工包边界。
- [x] 完成 P-0001 Atlas Community 小实验。
- [x] 将 migration 选择写成 accepted ADR；若实验失败，记录原因并采用 fallback。

## PostgreSQL 与 migration

- [ ] Compose 固定 PostgreSQL 16.14，不使用 `latest`。
- [x] 配置并校验 `LYAPUS_DATABASE_URL`，且不在日志输出连接串。
- [x] 建立 `pgxpool`，启动 Ping、关闭和错误路径均有测试或运行证据。
- [x] 固定 sqlc 1.31.1，`sqlc.yaml` 使用 `pgx/v5` 并只生成到 `sqlcgen`。
- [x] `db/schema.sql` 同时供 Atlas 与 sqlc 使用，没有第二份漂移 schema。
- [x] `make generate` 与 `make generate-check` 完成并进入 required CI。
- [x] `/livez` 保持进程语义；`/readyz` 真实反映数据库可用性和超时。
- [x] 建立 `db/schema.sql`、首个 versioned migration 和完整性校验文件。
- [x] 空库 apply、重复 apply 与 status 验证通过。
- [ ] 在真实 PostgreSQL 上验证外键、唯一、格式与时间检查约束的成功/失败行为。
- [x] migration runbook 已按实际工具命令验证。

## 数据模型与业务

- [x] Team、Service、Environment 表、关系和声明式约束与 `contracts.md` 一致。
- [x] Team repository CRUD、稳定游标分页与 PostgreSQL integration test 完成。
- [ ] Service CRUD 与 `team_id` 过滤完成。
- [ ] Environment CRUD 与 `service_id` 过滤完成。
- [ ] Service + 初始 Environment 在一个事务内原子创建。
- [ ] 父资源不存在、唯一冲突和仍被引用的删除均映射为稳定业务错误。
- [ ] 并发创建相同 `(team_id, service.slug)` 恰好一个成功、一个冲突。

## HTTP 与工程质量

- [ ] 使用 chi/v5 组织路由和 middleware，handler 保持标准 `net/http` 签名。
- [ ] chi 与 sqlc 类型都没有进入 catalog domain 或公共 API 契约。
- [ ] `/v1`、严格 JSON、1 MiB 上限、请求 ID 和统一错误完成。
- [x] repository 层 cursor 分页、稳定排序与 limit 边界完成。
- [ ] HTTP 日志包含同一 request ID、方法、路由/路径、状态和耗时，不泄漏敏感值。
- [ ] 单元测试覆盖关键校验、业务服务和 handler。
- [x] PostgreSQL integration tests 使用独立 `_test` 数据库且不会静默跳过。
- [x] 本地 `make verify`（含生成检查、普通/竞态/真实 PostgreSQL integration test 与漏洞扫描）通过。
- [ ] required `verify`、`smoke` 与 `atlas-community` checks 覆盖 M1 的数据库依赖和 API 最小路径。

## 可复现交付与证据

- [ ] Dockerfile 与 Compose 可以从空 project/空卷建立数据库、迁移并启动 API。
- [x] 停止数据库后 `/livez` 仍成功而 `/readyz` 返回 503；恢复后 readiness 恢复。
- [ ] 根 README 包含数据模型、API 示例和五分钟演示路径。
- [ ] Service 按 Team 列表查询有明确数据量与索引前后的 `EXPLAIN (ANALYZE, BUFFERS)` 证据。
- [ ] 查询优化记录注明环境、数据量、SQL、参数、结果和限制。
- [ ] `outcome.md`、当前架构、学习笔记、进度和会话记录均按实际结果更新。
- [ ] clean runner 与空 Compose project 完成最终验收。
- [ ] 原始方案七项最小交付全部成立并发布 M1 v0.1。

## 不阻塞 v0.1 的增强

- [ ] API Token 与最小 RBAC。
- [ ] 审计日志。
- [ ] 复杂幂等或乐观并发控制。
- [ ] 更丰富的过滤/排序与 HTTP E2E。
- [ ] k6 双拓扑报告与 pprof 深度优化。

这些增强项默认留在未勾选状态，不影响 M1 v0.1 完成判定。
