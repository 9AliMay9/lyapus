# lyapus

一个以 Go 为主语言、以 OpenTelemetry 为遥测标准、以 SLO 与故障闭环为可靠性核心，并逐步演化为平台工程与 AI 可观测性实践的项目。

M0（工程基线）已经完成：当前仓库包含可运行、可测试的 Go 服务骨架，以及配置、结构化日志、HTTP 健康检查、优雅关闭、统一校验、漏洞扫描、GitHub Actions 和受保护的 PR 合并门禁。M1 业务后端尚未开始；当前没有数据库、消息队列、容器编排、OpenTelemetry、前端或 AI 功能。

## 前置条件

- Go（版本以 `go.mod` 为准）
- GNU Make
- Git 与 curl
- `govulncheck@v1.6.0`

## 快速开始

获取仓库并安装固定版本的漏洞扫描器：

```bash
git clone https://github.com/9AliMay9/lyapus.git
cd lyapus
go install golang.org/x/vuln/cmd/govulncheck@v1.6.0
make verify
make run
```

`make run` 会在前台启动服务；健康检查命令需在另一个 shell 中执行。

## 开发命令

以下命令可在仓库根目录执行：

```bash
make fmt
make lint
make test
make race
make integration
make vuln
make verify
make run
```

服务默认监听 `127.0.0.1:8080`。验证健康检查：

```bash
curl -fsS http://127.0.0.1:8080/livez
curl -fsS http://127.0.0.1:8080/readyz
```

可用 `LYAPUS_HTTP_ADDR` 覆盖监听地址。

## 公开资源口径

当前实验环境的通用口径是 4 vCPU、8 GB 内存、约 180 GB SSD 和 Ubuntu 24.04 LTS；M0 基线采样时的内核为 Linux 6.8.0-124-generic。该口径只用于解释和复现实验，不代表生产容量。

- 后续容器常态总内存目标不高于约 6 GiB，为操作系统、云端 Agent、页缓存和峰值留出余量；M0 尚无应用容器，因此该目标尚未形成容量结论。
- 磁盘使用达到 70% 时预警并停止扩张实验数据；达到 80% 时停止新的数据写入实验，先保存证据并清理。
- M0 不产生持久业务数据；M1 只引入必要的 PostgreSQL/Compose 数据；M2 增加受控故障工作负载与 Kafka 数据；M3 通过短保留期和 TTL 控制遥测数据。每阶段都要重新记录实际峰值，不能把预算当成已验证容量。

详细的脱敏测量见 [M0 空闲资源基线](docs/benchmarks/m0-idle-resource-baseline.md)。实例身份、真实 IP、账单、租期、代理端口和私有运维配置不进入公开仓库。

## 贡献与合并

对 `main` 的改动使用短生命周期分支和 Pull Request。`protect-main` ruleset 要求分支与目标分支保持最新，并要求 `verify`、`smoke`、`atlas-community` 三个检查通过；不要绕过检查或直接向 `main` 推送。完整操作见 [GitHub SSH 与 PR runbook](docs/runbooks/github-ssh.md)。

## 文档

施工入口是 [docs/progress/current.md](docs/progress/current.md)。开始实现前，先阅读当前进度、当前阶段施工包和工程规范。

完整可执行代码以源码和 Git commit 为准；`docs/progress/handoffs/` 只在协作式手敲代码时临时保存尚未落地的短小代码接力稿。
