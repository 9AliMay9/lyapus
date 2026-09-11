# lyapus

一个以 Go 为主语言、以 OpenTelemetry 为遥测标准、以 SLO 与故障闭环为可靠性核心，并逐步演化为平台工程与 AI 可观测性实践的项目。

M0（工程基线）已经完成。M1 正在施工：当前已有 PostgreSQL migration、`pgxpool` 启动与 readiness、sqlc 基线，以及 Team、Service、Environment 三类资源的完整 HTTP CRUD。Environment 独立 CRUD 与 `service_id` 过滤已由 PR #23 以 squash commit `4b311b9` 合入 `main`。数据库约束测试已通过本地验证、收口审阅及三项 required CI，并由 PR #25 以 squash commit `51fd1d2` 合入 `main`。Compose 本地新 project/新卷交付、隔离测试和中断恢复已验证，README 演示路径已补齐；本次容器交付的 clean-runner 验收、查询计划实验及 M1 最终验收仍待完成。当前也没有消息队列、OpenTelemetry、前端或 AI 功能。

## 前置条件

- Go（版本以 `go.mod` 为准）
- GNU Make
- Git、curl 与 Docker
- `sqlc@v1.31.1`
- `govulncheck@v1.6.0`
- 可丢弃的 PostgreSQL 16.14 dev/test 数据库

## 快速开始

获取仓库并安装固定工具：

```bash
git clone https://github.com/9AliMay9/lyapus.git
cd lyapus
go install golang.org/x/vuln/cmd/govulncheck@v1.6.0
go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
./scripts/install-atlas-community.sh
```

容器交付按 [Compose runbook](docs/runbooks/compose-delivery.md) 依次完成镜像构建、开发数据库启动、Atlas 显式迁移和 API 启动。该路径已在本机新 project/新卷验证，不是自动迁移的一键 `up`；本机代理环境见 [构建代理排障](docs/runbooks/docker-daemon-proxy.md)。运行完整校验前须创建独立测试 project，不能将测试 URL 指向开发库。

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

## 数据模型与 API

```text
Team 1 ── N Service 1 ── N Environment
```

三表使用 bigint identity 主键及创建/更新时间。Team slug 全局唯一，Service slug 在 Team 内唯一，Environment slug 在 Service 内唯一；外键限制删除仍有子资源的父资源。Service 另有可空 description。详细校验与字段契约见 [M1 契约](docs/stages/m1-go-backend/contracts.md)。

三类资源分别使用 `/v1/teams`、`/v1/services`、`/v1/environments`：集合支持 POST/GET，`/{id}` 支持 GET/PATCH/DELETE。列表接受 `limit`、`cursor`，Service 可按 `team_id` 过滤，Environment 可按 `service_id` 过滤。返回 `items` 与 `next_cursor`，按 `(created_at DESC, id DESC)` 排序。创建/单项读取 Service 展开 Environment，列表/PATCH 不展开。

## 五分钟演示路径（依赖与镜像已准备）

先按 Compose runbook 完成数据库迁移与 API 启动。这是演示步骤，不承诺首次下载、构建或全新主机在五分钟内完成；最终文档尚待复走。以下创建会保留演示数据，重复执行可能返回 409，不要为重跑清空未知数据库。

```bash
curl --noproxy '*' --silent --show-error --include --max-time 10 http://127.0.0.1:8080/livez
curl --noproxy '*' --silent --show-error --include --max-time 10 http://127.0.0.1:8080/readyz
curl --noproxy '*' --silent --show-error --include \
  -X POST http://127.0.0.1:8080/v1/teams \
  -H 'Content-Type: application/json' \
  --data '{"slug":"compose-demo","name":"Compose Demo"}'
```

预期探针 200、创建 201。记录返回的 Team ID；下例的 `1` 只适用于本次新库演示，其他运行替换为实际 ID：

```bash
curl --noproxy '*' --silent --show-error --include \
  -X POST http://127.0.0.1:8080/v1/services \
  -H 'Content-Type: application/json' \
  --data '{"team_id":1,"slug":"compose-api","name":"Compose API","description":"Compose delivery demo","environments":[{"slug":"staging","name":"Staging"},{"slug":"production","name":"Production"}]}'
```

预期 201，两个 Environment 的 `service_id` 等于返回的 Service ID。将下列 Service ID 和 Team ID 替换为实际值；URL 引号不可省略，否则 `&` 会被 shell 当作后台运行符：

```bash
curl --noproxy '*' --silent --show-error --include 'http://127.0.0.1:8080/v1/teams?limit=10'
curl --noproxy '*' --silent --show-error --include 'http://127.0.0.1:8080/v1/services?team_id=1&limit=10'
curl --noproxy '*' --silent --show-error --include 'http://127.0.0.1:8080/v1/services/1'
curl --noproxy '*' --silent --show-error --include 'http://127.0.0.1:8080/v1/environments?service_id=1&limit=10'
```

预期均为 200。该演示证明成功写入与回读，不替代失败回滚、并发和完整 CRUD 测试。数据库中断/恢复、测试隔离与清理步骤见 Compose runbook。

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
- M0 不产生持久业务数据；M1 当前只引入必要的 PostgreSQL dev/test 数据，Compose 本地路径已验证，clean-runner 验收仍待完成；M2 增加受控故障工作负载与 Kafka 数据；M3 通过短保留期和 TTL 控制遥测数据。每阶段都要重新记录实际峰值，不能把预算当成已验证容量。

详细的脱敏测量见 [M0 空闲资源基线](docs/benchmarks/m0-idle-resource-baseline.md)。实例身份、真实 IP、账单、租期、代理端口和私有运维配置不进入公开仓库。

## 贡献与合并

对 `main` 的改动使用短生命周期分支和 Pull Request。`protect-main` ruleset 要求分支与目标分支保持最新，并要求 `verify`、`smoke`、`atlas-community` 三个检查通过；不要绕过检查或直接向 `main` 推送。完整操作见 [GitHub SSH 与 PR runbook](docs/runbooks/github-ssh.md)。

## 文档

施工入口是 [docs/progress/current.md](docs/progress/current.md)。开始实现前，先阅读当前进度、当前阶段施工包和工程规范。

完整可执行代码以源码和 Git commit 为准；`docs/progress/handoffs/` 只在协作式手敲代码时临时保存尚未落地的短小代码接力稿。
