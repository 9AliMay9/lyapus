# 工作会话：2026-07-28 - M1 施工包设计

## 目标

对照仓库现状、原始 v3.1 方案书和目标岗位材料，建立可边做边学、范围受控且有明确证据要求的 M1 施工包。

## 实际完成

- 将 M1 v0.1 固定为原始方案七项最小交付：核心模型、CRUD/校验/错误/分页、PostgreSQL migration/约束/事务/并发、测试/race、Compose、查询计划优化记录和 README 演示。
- 初稿选择标准库路由与全手写 scan；复核细分方向后改为 chi/v5 + sqlc 1.31.1 + pgx/v5：SQL 仍手写，chi 保持标准 handler，sqlc 类型限制在 repository adapter 内。
- 设计 Team、Service、Environment 数据与 HTTP 契约，明确 Service + 初始 Environment 的事务，以及同 Team 同 slug 的并发冲突场景。
- 建立 Atlas Community 提案：先验证声明式期望 schema 生成 versioned SQL 的免费核心路径；若登录/付费、稳定性、可解释性或工具成本触发退出条件，就回退 `golang-migrate`。
- 明确鉴权/RBAC、审计、复杂幂等、Gin、ORM、k6、pprof、OpenTelemetry、Kafka、Kubernetes 等不阻塞 M1 v0.1。
- 项目所有者确认施工包，并补充技术栈决策优先服务 Go 可观测性、可靠性、平台工程和基础架构方向；协作助手必须先查已有文档与原始方案，不要求项目所有者逐行代审。
- 更新阶段导航和当前交接页；业务源码仍未开始。

## 修改的文件

- `docs/stages/m1-go-backend/README.md`
- `docs/stages/m1-go-backend/plan.md`
- `docs/stages/m1-go-backend/contracts.md`
- `docs/stages/m1-go-backend/checklist.md`
- `docs/stages/m1-go-backend/outcome.md`
- `docs/stages/README.md`
- `docs/architecture/proposals/README.md`
- `docs/architecture/proposals/P-0001-atlas-migration-workflow.md`
- `docs/progress/current.md`
- `docs/progress/sessions/2026-07-28-m1-construction-package.md`
- `docs/README.md`
- `docs/project-context.md`
- `docs/standards/collaboration.md`
- `docs/runbooks/github-ssh.md`

最后四项是本会话前半段已经审阅的协作边界和 GitHub CLI 合并后验证更新。整个分支不包含业务源码。

## 验证证据

```text
命令：git diff --check
结果：通过，无空白错误。

命令：检查阶段文档相对链接与 M1 状态用语
结果：施工包导航存在；计划、提案、当前进度和 outcome 均明确业务尚未实施，Atlas 仍是候选。
```

## 偏差、风险或待确认事项

- Atlas 仍是候选而不是 accepted 决策；不能在小实验前把它写成项目既定事实。
- Migration 与 Compose 的具体命令尚未实测，因此没有提前写入 runbook。
- chi 与 sqlc 增加了工具面；通过标准 handler、domain repository 接口、生成包隔离、人工阅读生成代码和 `generate-check` 控制锁定与理解风险。
- M0 延期的“另一台全新 Ubuntu 主机手工 15 分钟复现”仍未完成，不阻塞 M1，但不能写成已有证据。

## 下一次从这里继续

- 具体文件：`docs/stages/m1-go-backend/plan.md`、`contracts.md`、`checklist.md`、`docs/architecture/proposals/P-0001-atlas-migration-workflow.md`。
- 具体任务：先合并施工包文档 PR；再建立实验分支，执行 Atlas Community 小实验并形成 migration ADR。
- 验收命令：先以提案中的七步实验为准，实际命令验证后再建立 runbook。
- 不要做：不要在迁移决策门前写 catalog 业务源码；不要使用 Atlas Cloud/Pro 或把深度增强塞入 v0.1。
