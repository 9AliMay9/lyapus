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

本清单的全部项目均已完成。原始 v3.1 方案书中更宽口径的工程证据（全新机器复现与 CI 分支保护门禁等）不属于本施工包的初始验收，收口状态以 `../../knowledge/go/m0-engineering-baseline.md` 和 `../../progress/current.md` 为准。

## 原始 v3.1 宽口径复核

- [x] Makefile 提供 `fmt`、`lint`、`test`、`race`、`integration`、`vuln`、`run`、`verify`。
- [x] 本地与 CI 运行普通测试、race detector、静态检查和固定版本漏洞扫描。
- [x] GitHub-hosted clean Ubuntu runner 启动服务并检查 `/livez`、`/readyz`。
- [x] `protect-main` ruleset 要求 PR、最新分支、`verify` 与 `smoke`；失败 CI 已实测阻止合并。
- [x] 保存脱敏的空闲资源基线，并在公开 README 记录资源、内存和磁盘控制口径。
- [x] ADR-0002 记录四条跨阶段执行护栏；ADR-0003 单独记录 PostgreSQL 主库决策。
- [x] `.gitignore`、公开 `.env.example`、提交前凭据检查和公开/私有信息边界已建立。
- [ ] 在另一台全新 Ubuntu 主机上按 README 手工计时并在 15 分钟内启动。

最后一项由项目所有者决定延期：当前主机真实运行、新 shell 验收和 GitHub clean-runner smoke 已提供足够的 M1 前置信心，但不能据此声称手工全新机 15 分钟目标已经实测完成。
