# lyapus

一个以 Go 为主语言、以 OpenTelemetry 为遥测标准、以 SLO 与故障闭环为可靠性核心，并逐步演化为平台工程与 AI 可观测性实践的项目。

M0（工程基线）已经完成。M1 正在施工：当前已有 PostgreSQL migration、`pgxpool` 启动与 readiness、sqlc 基线，以及 Team repository 的 CRUD、稳定游标分页和真实 PostgreSQL 集成测试。Team 业务服务与 HTTP API、Service/Environment、Compose、查询计划实验仍未完成；当前也没有消息队列、OpenTelemetry、前端或 AI 功能。

## 前置条件

- Go（版本以 `go.mod` 为准）
- GNU Make
- Git、curl 与 Docker
- `sqlc@v1.31.1`
- `govulncheck@v1.6.0`
- 可丢弃的 PostgreSQL 16.14 dev/test 数据库

## 快速开始

当前 M1 尚未提供 Compose 一键启动路径。获取仓库并安装固定工具后，先按 migration 与 integration runbook 建立可丢弃数据库：

```bash
git clone https://github.com/9AliMay9/lyapus.git
cd lyapus
go install golang.org/x/vuln/cmd/govulncheck@v1.6.0
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
./scripts/install-atlas-community.sh
```

将已迁移的 `_test` 数据库 URL 只设置在私有 shell 环境中，再执行完整校验：

```bash
export LYAPUS_TEST_DATABASE_URL='<postgres-test-url>'
make verify
```

运行服务前，另行设置已迁移的开发数据库 URL：

```bash
export LYAPUS_DATABASE_URL='<postgres-dev-url>'
make run
```

具体 migration 与测试数据库步骤分别见 [数据库 migration](docs/runbooks/database-migrations.md) 和 [PostgreSQL integration tests](docs/runbooks/postgresql-integration-tests.md)。`make run` 会在前台启动服务；健康检查命令需在另一个 shell 中执行。占位 URL 不能原样使用，真实连接串也不能提交到 Git。

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

可用 `LYAPUS_HTTP_ADDR` 覆盖监听地址；`LYAPUS_DATABASE_URL` 是必填配置，应用启动时会在有界时间内连接并 Ping PostgreSQL，但不会自动执行 migration。

## 公开资源口径

当前实验环境的通用口径是 4 vCPU、8 GB 内存、约 180 GB SSD 和 Ubuntu 24.04 LTS；M0 基线采样时的内核为 Linux 6.8.0-124-generic。该口径只用于解释和复现实验，不代表生产容量。

- 后续容器常态总内存目标不高于约 6 GiB，为操作系统、云端 Agent、页缓存和峰值留出余量；M0 尚无应用容器，因此该目标尚未形成容量结论。
- 磁盘使用达到 70% 时预警并停止扩张实验数据；达到 80% 时停止新的数据写入实验，先保存证据并清理。
- M0 不产生持久业务数据；M1 当前只引入必要的 PostgreSQL dev/test 数据，Compose 交付仍待完成；M2 增加受控故障工作负载与 Kafka 数据；M3 通过短保留期和 TTL 控制遥测数据。每阶段都要重新记录实际峰值，不能把预算当成已验证容量。

详细的脱敏测量见 [M0 空闲资源基线](docs/benchmarks/m0-idle-resource-baseline.md)。实例身份、真实 IP、账单、租期、代理端口和私有运维配置不进入公开仓库。

## 贡献与合并

对 `main` 的改动使用短生命周期分支和 Pull Request。`protect-main` ruleset 要求分支与目标分支保持最新，并要求 `verify`、`smoke`、`atlas-community` 三个检查通过；不要绕过检查或直接向 `main` 推送。完整操作见 [GitHub SSH 与 PR runbook](docs/runbooks/github-ssh.md)。

## 文档

施工入口是 [docs/progress/current.md](docs/progress/current.md)。开始实现前，先阅读当前进度、当前阶段施工包和工程规范。

完整可执行代码以源码和 Git commit 为准；`docs/progress/handoffs/` 只在协作式手敲代码时临时保存尚未落地的短小代码接力稿。
