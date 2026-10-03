# M1：常规 Go 后端

状态：M0 已完成；M1 最小交付已通过本地验收、PR #30 最终四项 required CI 和合并后 main CI，并发布 [v0.1.0](https://github.com/9AliMay9/lyapus/releases/tag/v0.1.0)，固定提交 `872d4a679539febd9349899ce7127f71db15d2f3`。包含三类资源 CRUD、数据库约束/事务/并发、Compose 交付、查询优化及固定 Atlas 依赖补丁；这是工程基线，不是生产就绪声明。M2 尚未开始。

施工前依次阅读：

2026-10-03接续：M1已发布，验收结果见[收口复盘](../../progress/sessions/2026-10-01-m1-final-acceptance.md)，发布证据与当前学习/求职接续见[状态同步](../../progress/sessions/2026-10-03-m1-release-sync.md)。不重跑M1验收，暂不扩大M1。

1. `plan.md`：目标、范围、顺序和验收证据。
2. `contracts.md`：数据模型、API、错误、数据库和测试契约。
3. `checklist.md`：当前唯一的阶段完成清单。
4. `../../architecture/decisions/ADR-0004-atlas-community-migrations.md` 与 `../../runbooks/database-migrations.md`：已接受的 migration 决策和经过验证的操作路径。

`outcome.md` 只记录实际结果；本地、历史 CI 与本分支 CI 分开记载，不以过去的成功替代最新提交门禁。
