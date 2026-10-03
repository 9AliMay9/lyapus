# 阶段施工包

阶段状态由 `progress/current.md` 决定。每个阶段目录固定包含：

- `plan.md`：目标、边界、施工顺序、代码清单与验收。
- `contracts.md`：本阶段对外或跨包稳定的命名与行为约束。
- `checklist.md`：可勾选的完成项。
- `outcome.md`：阶段结束后填写的实际结果与证据。

M0 已完成；M1 最小交付已通过本地验收、PR #30 最终四项 required CI 和合并后 main CI，并发布 [v0.1.0](https://github.com/9AliMay9/lyapus/releases/tag/v0.1.0)，固定提交 `872d4a679539febd9349899ce7127f71db15d2f3`。包含三类资源 CRUD、数据库约束/事务/并发、Compose 交付、查询优化及固定 Atlas 依赖补丁；这是工程基线，不是生产就绪声明。M2 尚未开始。
