# lyapus

一个以 Go 为主语言、以 OpenTelemetry 为遥测标准、以 SLO 与故障闭环为可靠性核心，并逐步演化为平台工程与 AI 可观测性实践的项目。

当前处于 M0（工程基线）：正在建立可运行的 Go 服务骨架。此阶段只实现配置、结构化日志、HTTP 健康检查、优雅关闭和测试；不引入数据库、消息队列、容器编排、OpenTelemetry、前端或 AI 功能。

## 前置条件

- Go（版本以 `go.mod` 为准）
- GNU Make

## 开发命令

M0 完成后，以下命令应可在仓库根目录执行：

```bash
make fmt
make test
make verify
go run ./cmd/apiserver
```

服务默认监听 `127.0.0.1:8080`。验证健康检查：

```bash
curl -fsS http://127.0.0.1:8080/livez
curl -fsS http://127.0.0.1:8080/readyz
```

可用 `LYAPUS_HTTP_ADDR` 覆盖监听地址。

## 文档

施工入口是 [docs/progress/current.md](docs/progress/current.md)。开始实现前，先阅读当前进度、M0 施工包和工程规范。

完整可执行代码以源码和 Git commit 为准；`docs/progress/handoffs/` 只在协作式手敲代码时临时保存尚未落地的短小代码接力稿。
