# 阶段施工包

阶段状态由 `progress/current.md` 决定。每个阶段目录固定包含：

- `plan.md`：目标、边界、施工顺序、代码清单与验收。
- `contracts.md`：本阶段对外或跨包稳定的命名与行为约束。
- `checklist.md`：可勾选的完成项。
- `outcome.md`：阶段结束后填写的实际结果与证据。

M0 已完成并冻结实际结果。下一次施工从 M1 的 `plan.md`、`contracts.md`、`checklist.md` 和 `outcome.md` 开始；在施工包经项目所有者审阅前，不写 M1 业务源码。M2–M8 继续只保留入口。
