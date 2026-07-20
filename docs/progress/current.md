# 当前施工状态

- 当前阶段：M0 工程基线
- 状态：M0 工程基线已完成并已推送到 GitHub；M1 尚未开始
- 已完成：初始化 Go module；建立文档结构、规范、知识笔记地图、M0 施工包、技术提案/ADR 流程和阅读提醒规则；实现 `internal/platform/config` 的 HTTP 地址加载与校验；实现 `internal/platform/health` 的 `/livez`、`/readyz` 处理逻辑；实现 `internal/platform/transport/http` 的路由注册与请求完成日志；实现 `cmd/apiserver` 的 JSON 日志初始化、启动和信号关闭
- 已验证：`make fmt`、`make test`、`make verify` 通过；`go run ./cmd/apiserver` 启动成功；`curl -i http://127.0.0.1:8080/livez` 与 `curl -i http://127.0.0.1:8080/readyz` 均返回 `200` 和 `{"status":"ok"}`；`Ctrl-C` 输出关闭信号与 server 已停止的 JSON 日志；提交前扫描未发现强凭据特征，公开地址仅为 `127.0.0.1` 本地回环地址
- 当前阻塞：无
- 下一步：依次阅读本文件、`../stages/m1-go-backend/README.md` 和相关 M1 施工包；确认 M1 的具体范围后开始下一次实现会话
- 本阶段不要做：PostgreSQL、Kafka、Kubernetes、OpenTelemetry、前端或 AI 功能
