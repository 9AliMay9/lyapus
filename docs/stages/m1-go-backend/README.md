# M1：常规 Go 后端

状态：Atlas migration、schema/sqlc 基线、数据库基础设施，以及 Team、Service、Environment 三类资源的完整 HTTP CRUD 已完成；Environment 纵切面已由 PR #23 以 squash commit `4b311b9` 合入 `main`。数据库约束实证已完成本地验证、收口审阅与 required CI，并由 PR #25 以 `51fd1d2` 合入 `main`；Compose 空环境交付、查询计划实验和根 README 演示文档仍待完成，M1 v0.1 尚未完成。

施工前依次阅读：

1. `plan.md`：目标、范围、顺序和验收证据。
2. `contracts.md`：数据模型、API、错误、数据库和测试契约。
3. `checklist.md`：当前唯一的阶段完成清单。
4. `../../architecture/decisions/ADR-0004-atlas-community-migrations.md` 与 `../../runbooks/database-migrations.md`：已接受的 migration 决策和经过验证的操作路径。

`outcome.md` 只记录实际结果；约束测试的本地、CI 与合并证据已记录；尚未发生的 Compose、查询计划实验和最终 M1 验收不得写成完成事实。
