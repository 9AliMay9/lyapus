# Compose 本地交付

## 范围与前置条件

2026-09-11 在 Linux 本机 Docker 默认 driver 上验证：构建镜像、新 project/新卷启动 PostgreSQL、宿主机 Atlas 显式迁移、API 创建/读取、数据库中断恢复、独立测试和清理。不是全新机器计时验收，也不是生产部署。

从仓库根目录操作。需要 Docker Compose（支持 `up --wait`）、curl、Go（见 `go.mod`）及按 `scripts/install-atlas-community.sh` 构建的 `.tools/bin/atlas`；完整测试还需 Make、固定 sqlc 和 govulncheck（见根 README）。下列凭据只用于本地演示，不得用于生产。

首次运行前确认未设置会覆盖默认值的 `LYAPUS_POSTGRES_DB`、`LYAPUS_POSTGRES_PORT`、`LYAPUS_HTTP_PORT`，并检查本地 `.env`。改变数据库名不会重建已有卷中的数据库；不要对旧卷盲目更换初始化变量。

## 开发环境：构建、数据库、迁移、API

检查目标没有已有容器、卷和端口占用；如有则先核对用途，不删除覆盖：

```bash
docker compose -p lyapus-dev config --quiet
docker compose -p lyapus-dev ps -a
docker volume ls --filter name=lyapus-dev_postgres_data
ss -ltn 'sport = :55432'
ss -ltn 'sport = :8080'
```

构建镜像：

```bash
docker build --progress=plain -t lyapus-apiserver:m1-dev .
```

本机受限网络下普通构建曾失败；实际成功方式见 [分层代理与临时 host 构建](docker-daemon-proxy.md#构建时的三层代理2026-09-11-已验证)。不要把私有代理或 host 网络写入通用 Compose。

```bash
docker compose -p lyapus-dev up -d --wait --wait-timeout 120 postgres
export LYAPUS_DATABASE_URL='postgres://lyapus:lyapus@127.0.0.1:55432/lyapus_dev_test?sslmode=disable'
./.tools/bin/atlas migrate apply --dry-run --dir 'file://db/migrations' --url "$LYAPUS_DATABASE_URL"
```

审阅 SQL 后执行；预期版本 `20260729030502`、2 个文件、pending 0，最后重复 apply 提示无待执行迁移：

```bash
./.tools/bin/atlas migrate apply --dir 'file://db/migrations' --url "$LYAPUS_DATABASE_URL"
./.tools/bin/atlas migrate status --dir 'file://db/migrations' --url "$LYAPUS_DATABASE_URL"
./.tools/bin/atlas migrate apply --dir 'file://db/migrations' --url "$LYAPUS_DATABASE_URL"
docker compose -p lyapus-dev up -d --no-build --pull never apiserver
docker compose -p lyapus-dev ps
curl --noproxy '*' --silent --show-error --include --max-time 10 http://127.0.0.1:8080/livez
curl --noproxy '*' --silent --show-error --include --max-time 10 http://127.0.0.1:8080/readyz
```

两端点应为 200。API 使用 `postgres:5432` 访问容器内数据库，宿主机 Atlas 使用发布的 `55432`。`depends_on` 只等待数据库健康，不证明表已迁移；不能在空库上用一次无前置步骤的 `compose up` 代替上述流程。API 使用 scratch，无 shell/curl 内置探针，`Up` 不等于 API 已就绪。业务验证见根 README。

## 数据库中断与恢复

只针对可中断的本地开发环境；保留 API 运行，不删除卷：

```bash
docker compose -p lyapus-dev stop postgres
curl --noproxy '*' --silent --show-error --include --max-time 10 http://127.0.0.1:8080/livez
curl --noproxy '*' --silent --show-error --include --max-time 10 http://127.0.0.1:8080/readyz
docker compose -p lyapus-dev up -d --no-build --pull never --wait --wait-timeout 120 postgres
curl --noproxy '*' --silent --show-error --include --max-time 10 http://127.0.0.1:8080/readyz
```

中断时依次应为 200、503；恢复后不重启 API 即回到 200。无论探针结果如何，均执行数据库恢复；再次读取演示 Service，确认数据仍在。

## 独立 integration project

两个 project 使用相同镜像，但有独立容器、网络、卷和数据库文件，不共享业务数据。卷位于宿主机磁盘并不意味着是同一目录。隔离不是权限屏障：开发库也以 `_test` 结尾，测试辅助代码的后缀保护不能识别误连开发库。

先确认 `lyapus-integration` 容器和卷不存在、55433 空闲：

```bash
docker compose -p lyapus-integration ps -a
docker volume ls --filter name=lyapus-integration_postgres_data
ss -ltn 'sport = :55433'
env LYAPUS_POSTGRES_DB=lyapus_integration_test LYAPUS_POSTGRES_PORT=55433 \
  docker compose -p lyapus-integration up -d --no-build --pull never --wait --wait-timeout 120 postgres
docker compose -p lyapus-integration exec -T -e PGPASSWORD=lyapus postgres \
  psql -h 127.0.0.1 -U lyapus -d lyapus_integration_test -v ON_ERROR_STOP=1 -c 'SELECT current_database();'
```

预期实际库名 `lyapus_integration_test`。后续若再次执行该 project 的 `up`，必须重复上述 DB/PORT 前缀；它们没有永久写入 shell。`ps`、`exec` 不重建容器。

测试会清表。只使用下面的独立测试 URL，不从开发变量复制：

```bash
export LYAPUS_TEST_DATABASE_URL='postgres://lyapus:lyapus@127.0.0.1:55433/lyapus_integration_test?sslmode=disable'
./.tools/bin/atlas migrate apply --dry-run --dir 'file://db/migrations' --url "$LYAPUS_TEST_DATABASE_URL"
```

审阅后执行：

```bash
./.tools/bin/atlas migrate apply --dir 'file://db/migrations' --url "$LYAPUS_TEST_DATABASE_URL"
./.tools/bin/atlas migrate status --dir 'file://db/migrations' --url "$LYAPUS_TEST_DATABASE_URL"
go test -tags=integration -race -count=1 ./internal/catalog/postgres
make verify
```

均应通过；再次通过开发 API 读取演示数据，确认未被测试清空。这仍是查询开发库，不是在查询 integration。

## 清理与失败处理

本次已验证的测试清理命令如下。先核对容器与卷标签，确认只含可丢弃测试数据，再执行 `down --volumes`；它删除该 project 的容器、网络和卷，卷数据不再保留，不删除镜像或开发 project：

```bash
docker compose -p lyapus-integration ps -a
docker volume inspect lyapus-integration_postgres_data --format '{{.Name}} {{index .Labels "com.docker.compose.project"}}'
docker compose -p lyapus-integration down --volumes
unset LYAPUS_TEST_DATABASE_URL
docker compose -p lyapus-integration ps -a
docker volume ls --filter name=lyapus-integration_postgres_data
```

预期容器和卷为空，开发 API 仍能读取原数据。本次开发 project 和开发卷有意保留，不声明已清理，也不对其执行 `down --volumes`。

失败时保留最小日志检查，勿重算历史 migration checksum、关闭 TLS 校验或清卷重试来掩盖原因。诊断入口：`docker compose -p <目标project> logs --tail=80 <服务名>`。数据库停止后的恢复命令见上节。最终没有运行新提交的 CI，不把本地成功写成 clean-runner 成功。

依据：[Compose 网络](https://docs.docker.com/compose/how-tos/networking/)、[项目名称与隔离](https://docs.docker.com/compose/how-tos/project-name/)。
