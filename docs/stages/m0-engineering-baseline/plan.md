# M0：工程基线

## 目标

建立一个小而可运行的 Go 服务骨架，使后续 M1 的业务代码有固定的目录、配置、日志、健康检查、测试和执行入口。

## 完成定义

- 应用能从本地配置启动，并能优雅响应退出信号。
- `/livez` 与 `/readyz` 可访问，且语义不同。
- `make fmt`、`make test`、`make verify` 有明确、可重复的行为。
- README 让陌生人在不接触私有配置的前提下完成最小验证。

## 本阶段必须完成

1. 建立根目录 README、`.gitignore`、`.env.example`、Makefile。
2. 建立 `cmd/apiserver` 启动入口。
3. 建立 `internal/platform/config`、`internal/platform/health` 与 `internal/platform/transport/http`。
4. 实现结构化日志、配置校验、优雅关闭和两个健康检查端点。
5. 为健康检查和配置校验写单元测试。
6. 写 ADR-0001：为何从模块化单体开始。

## 明确不做

- PostgreSQL、迁移、鉴权、Service Catalog 业务模型。
- Docker Compose、Kafka、Redis、OpenTelemetry、Kubernetes、前端。
- 性能数字、压测或生产可用性声明。

## 推荐施工顺序

1. 先创建 README、Makefile 和 `.gitignore`，定义本阶段验收命令。
2. 写最小 `main.go`，仅能启动 HTTP server。
3. 加配置与结构化日志；配置不合法时快速失败。
4. 加 `/livez`、`/readyz` 与优雅关闭。
5. 补单元测试与 `make verify`。
6. 更新进度记录，形成第一个 Git commit。

## 代码施工清单

后续写代码时，以以下路径和名称为准；如需改变，先更新 `contracts.md`。

```text
cmd/apiserver/main.go
internal/platform/config/config.go
internal/platform/health/handler.go
internal/platform/transport/http/server.go
internal/platform/transport/http/server_test.go
Makefile
README.md
.env.example
.gitignore
```

完整代码写在上述源码文件中。完成一次可运行版本后，在本文件记录对应 commit SHA；不在文档复制整份代码。

## 验收命令

```bash
make fmt
make test
make verify
go run ./cmd/apiserver
curl -fsS http://127.0.0.1:8080/livez
curl -fsS http://127.0.0.1:8080/readyz
```

## 交付证据

- Git commit：`fd9faf4`（`feat: establish M0 engineering baseline`）
- 会话记录：`../../progress/sessions/2026-07-21-m0-engineering-baseline.md`
- ADR：`../../architecture/decisions/ADR-0001-modular-monolith.md`
