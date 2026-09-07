# 阶段施工包

阶段状态由 `progress/current.md` 决定。每个阶段目录固定包含：

- `plan.md`：目标、边界、施工顺序、代码清单与验收。
- `contracts.md`：本阶段对外或跨包稳定的命名与行为约束。
- `checklist.md`：可勾选的完成项。
- `outcome.md`：阶段结束后填写的实际结果与证据。

M0 已完成并冻结实际结果。M1 已完成施工包、Atlas migration 决策、schema/sqlc 基线、数据库基础设施、完整 Team HTTP CRUD，以及由 PR #20、#21 分段合入的完整 Service HTTP CRUD、`team_id` 过滤、初始 Environment 事务创建和并发冲突测试；Environment 独立 CRUD 与 `service_id` 过滤已在 PR #23 完成本地及 required clean-runner 验证，等待合并。状态与下一步以 `progress/current.md` 为准。M2–M8 继续只保留入口。
