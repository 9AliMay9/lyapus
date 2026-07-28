# 当前施工状态

## 当前落点

- 当前阶段：M0 已完成终局审计，下一次从 M1 施工包设计开始。
- M0 结论：真实实现完整满足仓库内施工包，并基本符合原始 v3.1 工程基线预期，可以进入 M1。
- M1 状态：业务源码尚未开始；进入前必须先建立并审阅 `plan.md`、`contracts.md`、`checklist.md` 和 `outcome.md`。
- 本次终局审计所依据的已合并门禁基线：commit `6306d4b`（`docs: verify M0 merge gate (#1)`）。
- 仓库标识待校正：远端规范地址为 `github.com/9AliMay9/lyapus`，而当前 `go.mod` 和 4 处内部 import 仍使用 `github.com/9Alimay/lyapus`。这是已确认的一致性缺陷，必须在 M1 施工包前用独立小 PR 更正并运行 `make verify`。

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

1. 建立短生命周期分支，统一 `go.mod`、全部内部 import 与文档中的仓库地址为 `github.com/9AliMay9/lyapus`；运行 `make verify`，再按 PR 门禁合并。
2. 阅读本页、`../knowledge/go/m0-engineering-baseline.md`、ADR-0002 和 ADR-0003。
3. 回读原始 v3.1 的 M1 近期最小版与深度增强边界。
4. 从模板建立 M1 的 `plan.md`、`contracts.md`、`checklist.md` 和 `outcome.md`，先确定数据模型、PostgreSQL/迁移版本、API 契约与验收证据。
5. 由项目所有者审阅 M1 施工包；确认后再开始手敲业务源码。

不要在施工包确认前提前实现 PostgreSQL、CRUD、鉴权、Compose 或其他 M1 功能。

## 协作审阅约定

- 日常的学习、施工、测试解释和文档维护默认使用 Terra + medium。
- 阶段开始时，先用 Sol 完成原始方案边界、施工包和验收标准的概览审阅；施工包经项目所有者确认后，切回 Terra + medium 执行。
- 阶段验收前，再用 Sol 独立核对实现、测试证据、文档事实、风险和原始方案的一致性。安全、数据迁移/删除、公共 API 或数据模型等高影响判断也可临时使用 Sol。
- 此约定是当前协作偏好，不是项目运行时依赖、架构决策或对外承诺；模型可用性变化时以实际界面为准。
