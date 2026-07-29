# 当前施工状态

## 当前落点

- 当前阶段：M0 已完成终局审计；M1 施工包已合并，当前正在 `spike/atlas-community-workflow` 执行 P-0001。
- M0 结论：真实实现完整满足仓库内施工包，并基本符合原始 v3.1 工程基线预期，可以进入 M1。
- M1 状态：`../stages/m1-go-backend/` 已建立 `plan.md`、`contracts.md`、`checklist.md` 和 `outcome.md`；业务源码尚未开始。
- M1 最小范围：Team、Service、Environment CRUD，PostgreSQL migration/约束/事务/并发正确性，单元与真实数据库测试，Compose 空环境复现，以及一份查询计划优化记录。
- M1 默认实现：PostgreSQL 16.14、`pgx/v5` + `pgxpool`、chi/v5、sqlc 1.31.1、手写 SQL + repository adapter、identity bigint 和不透明游标。chi 保持标准 HTTP handler；sqlc 生成类型不越过 PostgreSQL adapter。选择理由与适用边界见施工包。
- Migration 尚未最终决策：P-0001 已完成固定 Atlas Community v1.2.0 的两次 migration（空库 apply、已有库前滚、重复 apply、status 与完整性篡改拦截）、同一 `db/schema.sql` 的 sqlc 1.31.1 解析/生成及本机可复现 Community 构建；CI 中的同一路径尚待本 PR 实跑。CI 成功后写 ADR，任何退出条件触发则回退 `golang-migrate`。
- 本次终局审计所依据的已合并门禁基线：commit `6306d4b`（`docs: verify M0 merge gate (#1)`）。
- 仓库标识修正：commit `2b3764b`（`chore: normalize module path (#5)`）已将 `go.mod` 与 5 处内部 import 统一为远端规范地址 `github.com/9AliMay9/lyapus`；commit `d96946d`（`docs: record module path merge (#6)`）完成合并事实记录。

## M0 已验证事实

- 服务实现：配置加载与校验、JSON 结构化日志、HTTP server、`/livez`、`/readyz`、`SIGINT`/`SIGTERM` 优雅关闭。
- 本地校验：`gofmt`、`go vet ./...`、`go test ./...`、`go test -race ./...`、`go test -tags=integration -count=1 ./...` 和固定版本 `govulncheck@v1.6.0`。
- 真实运行：两个健康端点均返回 `200` 与 `{"status":"ok"}`；`Ctrl-C` 记录关闭信号和 server 停止日志。
- CI：GitHub Actions 的 `verify` 与 clean-runner `smoke` 已通过；smoke 在干净 Ubuntu runner 启动服务并检查两个健康端点。
- 合并门禁：`protect-main` ruleset 要求 PR、最新分支、`verify` 与 `smoke`；故意失败的测试曾使 PR 明确不可合并，修复后两个检查通过并以 squash merge 进入 `main`。
- 工程证据：脱敏的 M0 空闲资源基线、公开/私有信息边界、ADR-0001/0002/0003 和系统化 M0 学习笔记已经建立。
- 协作工具：当前远程开发主机已配置可选的 GitHub CLI；Git 继续使用 SSH，CLI API 登录与最小命令见 `../runbooks/github-ssh.md`。

## 明确延期与边界

- 没有在另一台全新 Ubuntu 主机上手工计时完成“按 README 15 分钟启动”。当前主机、新 shell 和 GitHub clean-runner smoke 提供了近似证据，但不能写成该项已经实测完成。
- `make integration` 目前只有测试入口；M0 没有数据库、容器或跨服务依赖，因此没有真实集成测试内容。
- M0 没有 PostgreSQL、Kafka、OpenTelemetry、Kubernetes、前端或 AI 功能，也不声明生产容量、高可用或生产就绪。

## 下一次从这里开始

1. 审阅并提交 `spike/atlas-community-workflow` 的 schema、migration、sqlc、固定 Atlas 构建、CI 与脱敏网络文档；在 PR 上实跑 `atlas-community` job。
2. 仅在该 CI job 成功后，根据 P-0001 建立 accepted ADR 与实测 migration runbook；若失败，记录证据并回退 `golang-migrate`。
3. 决策落定后，再按 `plan.md` 顺序，由数据库配置、连接池和 readiness 开始逐段手敲核心代码。

不要把 Atlas 实验与未合并的施工包混在同一分支；不要让 Atlas Cloud/Pro、鉴权、RBAC、k6、OpenTelemetry 或其他后续增强进入 M1 v0.1 的阻塞路径。

## 协作审阅约定

- 日常的学习、施工、测试解释和文档维护默认使用 Terra + medium。
- 阶段开始时，先用 Sol 完成原始方案边界、施工包和验收标准的概览审阅；施工包经项目所有者确认后，切回 Terra + medium 执行。
- 阶段验收前，再用 Sol 独立核对实现、测试证据、文档事实、风险和原始方案的一致性。安全、数据迁移/删除、公共 API 或数据模型等高影响判断也可临时使用 Sol。
- 此约定是当前协作偏好，不是项目运行时依赖、架构决策或对外承诺；模型可用性变化时以实际界面为准。
