# M1：常规 Go 后端

状态：M0 已完成；M1 的三类资源 CRUD、数据库约束/事务/并发、Compose 交付和查询优化已合入 main（PR #29：`fdf8f25`）。2026-10-01，本分支 Atlas `v1.3.0-lyapus.1` 升级专项、应用回归及新卷 API 演示已在本地通过。2026-10-02临时资源已清理；本分支四项 required CI、合并和 release 仍待完成；不声明 M1 已发布或生产就绪。

施工前依次阅读：

2026-10-01接续：[联合验收计划](atlas-upgrade-acceptance-plan.md)的本地验证已完成，详见[收口复盘](../../progress/sessions/2026-10-01-m1-final-acceptance.md)。当前默认 Atlas 已切换到补丁版，尚待本分支 CI 与合并。

1. `plan.md`：目标、范围、顺序和验收证据。
2. `contracts.md`：数据模型、API、错误、数据库和测试契约。
3. `checklist.md`：当前唯一的阶段完成清单。
4. `../../architecture/decisions/ADR-0004-atlas-community-migrations.md` 与 `../../runbooks/database-migrations.md`：已接受的 migration 决策和经过验证的操作路径。

`outcome.md` 只记录实际结果；本地、历史 CI 与本分支 CI 分开记载，不以过去的成功替代最新提交门禁。
