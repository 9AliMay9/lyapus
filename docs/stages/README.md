# 阶段施工包

阶段状态由 `progress/current.md` 决定。每个阶段目录固定包含：

- `plan.md`：目标、边界、施工顺序、代码清单与验收。
- `contracts.md`：本阶段对外或跨包稳定的命名与行为约束。
- `checklist.md`：可勾选的完成项。
- `outcome.md`：阶段结束后填写的实际结果与证据。

M0 已完成并冻结实际结果。M1 已完成施工包、Atlas migration 决策、schema/sqlc 基线、数据库基础设施，以及 Team、Service、Environment 三类资源的完整 HTTP CRUD；Environment 纵切面已由 PR #23 以 squash commit `4b311b9` 合入 `main`。数据库约束实证已在当前分支完成本地验证与收口审阅，尚待 PR/CI 和合并；M1 v0.1 尚缺 Compose 空环境交付、查询计划实验和根 README 演示文档。状态与下一步以 `progress/current.md` 为准。M2–M8 继续只保留入口。
