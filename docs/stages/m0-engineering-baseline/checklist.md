# M0 清单

- [x] 建立文档导航、进度记录和工程规范。
- [x] 创建根目录 README、`.gitignore`、`.env.example`、Makefile。
- [x] 创建 `cmd/apiserver/main.go`。
- [x] 实现配置加载与校验。
- [x] 实现结构化日志。
- [x] 实现 HTTP server、`/livez`、`/readyz`。
- [x] 实现优雅关闭。
- [x] 编写配置与健康检查单元测试。
- [x] 写 ADR-0001。
- [x] 在全新 shell 中执行全部验收命令。
- [x] 更新 `progress/current.md` 与会话记录。
- [x] 创建 M0 第一个可运行 commit。

本清单的全部项目均已完成。原始 v3.1 方案书中更宽口径的工程证据（CI 漏洞扫描、全新机器复现、资源基线与 CI 分支保护门禁等）不属于本施工包的初始验收，收口状态以 `../../knowledge/go/m0-engineering-baseline.md` 和 `../../progress/current.md` 为准。
