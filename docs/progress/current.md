# 当前施工状态

## 当前落点

- 当前阶段：M0 已完成终局审计；M1 已完成数据库基础设施和完整 Team HTTP CRUD 纵切面。Team CRUD/稳定分页施工包已由 PR #13 以 squash commit `f4df01f` 合入 `main`；Team service、创建 API、全局 request ID 与 API smoke 已由 PR #15 以 squash commit `8e84c20` 合入 `main`；Team HTTP read 与 cursor 分页已由 PR #17 以 squash commit `284021e` 合入 `main`；Team HTTP mutate 已由 PR #18 以 squash commit `de7e2d3` 合入 `main`；Service Create/Get/List、初始 Environment 事务创建和并发冲突已由 PR #20 以 squash commit `43c627d` 合入 `main`。当前 `feat/service-http-mutate` 分支已完成 Service PATCH/Delete 的本地实现、测试、真实 PostgreSQL 与 HTTP 验证；PR #21 的 required `verify`、`smoke` 与 `atlas-community` 已在 clean runner 通过，待提交最终文档证据并等待最新门禁后 squash merge。
- M0 结论：真实实现完整满足仓库内施工包，并基本符合原始 v3.1 工程基线预期，可以进入 M1。
- M1 状态：施工包、Atlas 决策、schema、两份 versioned migration、sqlc 基线与 CI 门禁已完成；`LYAPUS_DATABASE_URL`、`pgxpool` 启动 Ping/关闭路径及数据库感知 `/readyz` 已完成并作真实运行验证。Team 已具备完整 repository/service/HTTP CRUD 和 clean-runner 证据。Service 已具备完整 repository/service/HTTP CRUD、全局及 `team_id` 过滤的 `(created_at, id)` 稳定游标分页；创建 Service 与可选初始 Environment 使用显式 pgx 事务，缺失 Team、同 Team slug 冲突和初始 Environment 冲突映射为稳定错误，并以两个独立 repository 并发创建验证恰好一个成功、一个冲突。PATCH 区分 description 未提供、字符串和显式 `null`；DELETE 在仍有 Environment 时返回冲突，无子资源时返回空 `204`。创建与单项 GET 展开 Environment，集合与 PATCH 不展开。PR #21 clean-runner smoke 验证 PATCH 字段保留/显式清空、引用删除 409、无子资源删除 204 与删除后 404。严格 JSON、415 媒体类型、统一错误、全局 request ID 与不透明 cursor 保持既有边界。Service mutation 待最终文档门禁与 merge；独立 Environment CRUD、Compose 与查询计划实验仍待完成。
- M1 最小范围：Team、Service、Environment CRUD，PostgreSQL migration/约束/事务/并发正确性，单元与真实数据库测试，Compose 空环境复现，以及一份查询计划优化记录。
- M1 默认实现：Go 1.26.6、PostgreSQL 16.14、`pgx/v5` + `pgxpool`、chi/v5、sqlc 1.31.1、手写 SQL + repository adapter、identity bigint 和不透明游标。chi 保持标准 HTTP handler；sqlc 生成类型不越过 PostgreSQL adapter。选择理由与适用边界见施工包。
- Migration 已由 ADR-0004 最终确定：P-0001 完成固定 Atlas Community v1.2.0 的两次 migration（空库 apply、已有库前滚、重复 apply、status 与完整性篡改拦截）、同一 `db/schema.sql` 的 sqlc 1.31.1 解析/生成、本机可复现 Community 构建，以及 PR #8 中 required `atlas-community` CI 实跑。未来触发退出条件时才以新 ADR 记录并回退 `golang-migrate`。
- 本次终局审计所依据的已合并门禁基线：commit `6306d4b`（`docs: verify M0 merge gate (#1)`）。
- 仓库标识修正：commit `2b3764b`（`chore: normalize module path (#5)`）已将 `go.mod` 与 5 处内部 import 统一为远端规范地址 `github.com/9AliMay9/lyapus`；commit `d96946d`（`docs: record module path merge (#6)`）完成合并事实记录。

## M0 已验证事实

- 服务实现：配置加载与校验、JSON 结构化日志、HTTP server、`/livez`、`/readyz`、`SIGINT`/`SIGTERM` 优雅关闭。
- 本地校验：`gofmt`、`go vet ./...`、`go test ./...`、`go test -race ./...`、`go test -tags=integration -count=1 ./...` 和固定版本 `govulncheck@v1.6.0`。
- 真实运行：两个健康端点均返回 `200` 与 `{"status":"ok"}`；`Ctrl-C` 记录关闭信号和 server 停止日志。
- CI：GitHub Actions 的 `verify` 与 clean-runner `smoke` 已通过；smoke 在干净 Ubuntu runner 启动服务并检查两个健康端点。
- 合并门禁：M0 曾以 `verify` 与 `smoke` 实测失败阻止合并；当前 `protect-main` 进一步要求 PR、最新分支以及 `verify`、`smoke`、`atlas-community` 三项检查。
- 工程证据：脱敏的 M0 空闲资源基线、公开/私有信息边界、ADR-0001/0002/0003 和系统化 M0 学习笔记已经建立。
- 协作工具：当前远程开发主机已配置可选的 GitHub CLI；Git 继续使用 SSH，CLI API 登录与最小命令见 `../runbooks/github-ssh.md`。

## 明确延期与边界

- 没有在另一台全新 Ubuntu 主机上手工计时完成“按 README 15 分钟启动”。当前主机、新 shell 和 GitHub clean-runner smoke 提供了近似证据，但不能写成该项已经实测完成。
- M0 验收时 `make integration` 只有测试入口；当时没有数据库、容器或跨服务依赖，因此没有真实集成测试内容。M1 现已加入 Team repository 的真实 PostgreSQL integration test。
- M0 没有 PostgreSQL、Kafka、OpenTelemetry、Kubernetes、前端或 AI 功能，也不声明生产容量、高可用或生产就绪。

## 下一次从这里开始

1. 提交 PR #21 clean-runner 证据，等待该最终文档 commit 的 `verify`、`smoke` 与 `atlas-community` 重新全绿后 squash merge。
2. 随后按同一分层实现 Environment 独立 CRUD 与 `service_id` 过滤。

不要在应用启动路径自动执行 migration；不要让 Atlas Cloud/Pro、鉴权、RBAC、k6、OpenTelemetry 或其他后续增强进入 M1 v0.1 的阻塞路径。

## 协作审阅约定

- 日常学习、施工和测试解释使用适合连续协作的模型配置；阶段设计、高影响判断和收口一致性审阅使用当时可用的高能力配置。
- 功能性 PR、独立施工包或阶段达到收口点时，协作助手先提醒并等待项目所有者明确确认模型/参数已经调整，再开始核对实现、施工包、原始方案、证据和文档一致性。
- 完整触发条件、审阅范围和例外以 `../standards/collaboration.md` 的“高能力收口审阅门禁”为准；该约定是协作流程，不是项目运行时依赖或对外承诺。
