# 架构文档

`overview.md` 只描述当前已实现架构。设计中的内容放入对应阶段 `plan.md`。

重要技术选型遵循以下路径：

```text
proposals/（问题、想法、备选方案、待验证假设）
    → decisions/（已接受或拒绝的 ADR）
    → stages/、benchmarks/、postmortems/（实现和证据）
```

`proposals/` 可以修改，适合尚在思考的选择；`decisions/` 保存 ADR。ADR 一经接受不改写历史；后续改变用新的 ADR 取代或修订它。
