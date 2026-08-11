# Team service 与首个 HTTP 创建端点

日期：2026-08-10
分支：`feat/team-service`

## 完成事实

- 在 `internal/catalog` 增加 `TeamService`：集中校验 Team ID、slug、name、更新输入和列表分页输入；service 只依赖 domain repository 接口。
- 引入 chi/v5，并在 catalog HTTP transport 中实现严格 JSON 解码、统一 catalog 错误信封、`POST /v1/teams` 与 handler 单元测试。
- `cmd/apiserver` 装配 PostgreSQL Team repository、Team service 与 catalog HTTP handler；`/v1/teams` 已走真实数据库路径。
- 新增平台级 `internal/platform/requestid`。它包围整个 mux，因此 `/livez`、`/readyz` 与 `/v1` 都获得新的 request ID；响应头、catalog 错误体和完成日志使用同一 ID。
- platform 完成日志补充 request ID 与最终 HTTP 状态；避免只记录处理耗时而无法关联一次请求。

## 验证证据

- `go test ./...` 与 `go test -race ./...` 在 request-ID 提升到 platform 后通过。
- 用一次性 PostgreSQL 16.14 开发容器完成 migration dry-run、apply 与 status；服务连接该库启动成功。
- 实测 `/livez`、`/readyz` 返回 200 并带 request ID；`POST /v1/teams` 分别实测 201、非法 slug 的 400 和重复 slug 的 409。每次响应与日志均可通过 request ID 对应。
- 验证结束后以 `Ctrl-C` 停止 API，再停止带 `--rm` 的开发容器，未保留测试进程或容器。

## 边界与下一步

- 只实现了 Team 创建端点；Team 的 list/get/PATCH/DELETE、HTTP cursor 编解码以及 Service/Environment 仍未实现。
- 健康检查复用全局 request ID，但保留其既有最小响应体；不把健康探针误写成 catalog 资源错误。
- 2026-08-11 已在新建的可丢弃 `_test` PostgreSQL 16.14 数据库上完成 migration dry-run、apply、status 和完整 `make verify`；PR #15 的 required `verify`、`atlas-community` 与 clean-runner API smoke 随后全部通过，施工包以 squash commit `8e84c20` 合入 `main`。

## 与原始方案和施工包的对齐

- 原始 v3.1 方案书已在技术栈表中明确选择 `net/http + chi`，理由是保留标准库基础并使用轻量路由；M1 施工包在 2026-07-28 又从标准库路由与 Gin 等替代方案中复核并确认 chi/v5。因此 chi 不是本次实现时临时加入的框架。
- 原始方案的 M1 最小交付要求 Service Catalog、CRUD、校验、统一错误、分页、PostgreSQL、测试、Compose 和查询优化；其实施日程明确出现 Team/Service/Environment。当前先完成 Team 纵切面，是在既定模型内先建立可复用模板的顺序调整，不改变 M1 范围。
- request ID 没有作为原始方案七项最小交付的独立条目，但从最初的 M1 施工包起就是 HTTP 稳定契约和日志关联要求，不是本次会话临时扩张的产品功能。
- request-ID 最初可局部放在 catalog HTTP middleware；实际装配后发现这只覆盖 `/v1`，不满足施工包中“每个请求”和完成日志关联的字面契约。将其提升为 `internal/platform/requestid`、包围整个 mux 并删除 catalog 局部实现，是施工过程中的自然架构优化：它纠正作用域、消除重复点，但没有改变对外 API 或扩大阶段目标，因此无需单独 ADR。

## 尚未满足的 M1 原始交付

- Team HTTP 的 list/get/PATCH/DELETE 和不透明 cursor 尚未实现。
- Service/Environment CRUD、归属过滤、“Service + 初始 Environment”事务与并发唯一性场景尚未实现。
- Compose 空环境复现、完整约束行为验证、查询计划优化记录、README 五分钟演示和 M1 v0.1 release 尚未完成。
- 当前实现已通过完整 `make verify`、手工 HTTP 纵切面验证和 PR clean-runner API smoke；下一施工包从 Team 的 HTTP 查询、列表、PATCH、DELETE 与不透明 cursor 开始。
