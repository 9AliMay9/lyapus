# 当前架构

当前是一个模块化单体 Go 服务，二进制入口为 `cmd/apiserver`。

```text
LYAPUS_HTTP_ADDR
        ↓
internal/platform/config
        ↓
cmd/apiserver ── JSON slog ──→ stdout
        ↓
internal/platform/transport/http
        ↓
/livez、/readyz ──→ internal/platform/health
```

- 默认监听地址是 `127.0.0.1:8080`；可用 `LYAPUS_HTTP_ADDR` 覆盖，非法地址会在启动前失败。
- `/livez` 表示进程存活，`/readyz` 表示服务可以接收流量；M0 中两者均返回 `200` 与 `{"status":"ok"}`。
- HTTP server 对每个请求输出 JSON 结构化完成日志；收到 `SIGINT` 或 `SIGTERM` 后以 10 秒超时执行优雅关闭。

不要在这里提前写入 PostgreSQL、Kafka、OpenTelemetry、Kubernetes 或 AI 组件；它们尚未进入当前实现。
