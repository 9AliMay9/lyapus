# 阶段施工包

阶段状态由 `progress/current.md` 决定。每个阶段目录固定包含：

- `plan.md`：目标、边界、施工顺序、代码清单与验收。
- `contracts.md`：本阶段对外或跨包稳定的命名与行为约束。
- `checklist.md`：可勾选的完成项。
- `outcome.md`：阶段结束后填写的实际结果与证据。

M0 已完成并冻结实际结果。M1 已完成施工包、Atlas migration 决策、schema/sqlc 基线、数据库基础设施和 Team Create/Get repository 纵切面，下一步补齐 Team CRUD 与游标分页；状态与下一步以 `progress/current.md` 为准。M2–M8 继续只保留入口。
