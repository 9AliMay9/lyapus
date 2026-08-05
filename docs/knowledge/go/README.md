# Go

建议按顺序建立：语言与运行时、并发、`context` 与错误、测试、pprof/trace。每项都链接到项目中的实际使用位置。

## 项目演进笔记

- [M0：从空仓库到可控的 Go 服务进程](m0-engineering-baseline.md)：沿配置、HTTP 边界、日志、进程生命周期和验证证据复盘第一个可运行版本。
- [生成代码检查为什么不应依赖 Git 暂存区](generated-code-and-git-index.md)：区分生成、验证、跟踪与候选提交，并记录 sqlc diff 的本地/CI 落点。
