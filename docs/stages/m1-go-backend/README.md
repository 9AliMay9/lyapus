# M1：常规 Go 后端

状态：Atlas migration、schema/sqlc 基线、数据库基础设施、Team repository CRUD（含稳定游标分页）、Team 业务服务和 `POST /v1/teams` 初始 HTTP 纵切面已完成；Team 其余 HTTP CRUD、Service/Environment、Compose 与查询计划实验待完成，M1 v0.1 尚未完成。

施工前依次阅读：

1. `plan.md`：目标、范围、顺序和验收证据。
2. `contracts.md`：数据模型、API、错误、数据库和测试契约。
3. `checklist.md`：当前唯一的阶段完成清单。
4. `../../architecture/decisions/ADR-0004-atlas-community-migrations.md` 与 `../../runbooks/database-migrations.md`：已接受的 migration 决策和经过验证的操作路径。

`outcome.md` 只记录实际结果；未完成的 Team HTTP list/get/update/delete 与 cursor、Service/Environment、Compose 和查询计划实验不得写成完成事实。
