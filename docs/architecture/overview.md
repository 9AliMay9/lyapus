# 当前架构

当前是一个模块化单体 Go 服务，二进制入口为 `cmd/apiserver`。

```text
LYAPUS_HTTP_ADDR ─────┐
                     ├──→ internal/platform/config ──→ cmd/apiserver
LYAPUS_DATABASE_URL ─┘                                  ├──→ JSON slog ──→ stdout
                                                        ├──→ pgxpool ──→ PostgreSQL
                                                        └──→ HTTP transport
                                                               ├──→ request-ID + completion log
                                                               ├──→ /livez
                                                               ├──→ /readyz ──→ bounded pool.Ping
                                                               └──→ /v1/teams POST ──→ TeamService

internal/catalog domain ←── PostgreSQL repository adapter ←── sqlcgen
        （Team repository CRUD/稳定分页与业务服务已实现；仅创建端点已装配到 HTTP）
```

- 默认监听地址是 `127.0.0.1:8080`；可用 `LYAPUS_HTTP_ADDR` 覆盖，非法地址会在启动前失败。
- `LYAPUS_DATABASE_URL` 必填；进程启动时以五秒有界 context 创建并 Ping `pgxpool`，失败则不启动 HTTP 服务，退出时先关闭 HTTP 再释放连接池。
- `/livez` 只表示进程存活，不访问数据库；`/readyz` 以一秒有界 context Ping PostgreSQL，成功返回 `200`，失败返回 `503`。
- HTTP server 为每个请求生成 request ID，并输出含 request ID、方法、路径、状态、耗时的 JSON 完成日志；收到 `SIGINT` 或 `SIGTERM` 后以 10 秒超时执行优雅关闭。
- `internal/catalog` 已有 Team domain、完整 repository 接口、PostgreSQL adapter 与 TeamService；当前已实现 CRUD、基于 `(created_at, id)` 的稳定分页和 `POST /v1/teams`。chi 与 HTTP DTO 留在 catalog transport，不进入 domain。

当前仍未实现 Team 的其余 HTTP CRUD 与 cursor、Service/Environment repository 和 HTTP API、Compose、Kafka、OpenTelemetry、Kubernetes、前端或 AI 组件；这些内容不能画入已实现数据流。
