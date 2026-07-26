# 当前施工状态

- 当前阶段：M0 宽口径工程证据收口
- 状态：仓库内 M0 施工包已完成并推送到 GitHub；M0 学习笔记已由项目所有者审阅。现在补齐原始 v3.1 方案书的工程证据；M1 尚未开始。
- 已完成：初始化 Go module；建立文档结构、规范、知识笔记地图、M0 施工包、技术提案/ADR 流程和阅读提醒规则；实现 `internal/platform/config` 的 HTTP 地址加载与校验；实现 `internal/platform/health` 的 `/livez`、`/readyz` 处理逻辑；实现 `internal/platform/transport/http` 的路由注册与请求完成日志；实现 `cmd/apiserver` 的 JSON 日志初始化、启动和信号关闭；对照原方案书生成 `../knowledge/go/m0-engineering-baseline.md` 学习笔记
- 已验证：`make verify` 通过；它执行 `gofmt`、`go vet ./...`、`go test ./...`、`go test -race ./...` 和 `go test -tags=integration -count=1 ./...`。后者目前只验证预留入口，M0 尚无数据库或容器集成测试。`go run ./cmd/apiserver` 启动成功；`curl -i http://127.0.0.1:8080/livez` 与 `curl -i http://127.0.0.1:8080/readyz` 均返回 `200` 和 `{"status":"ok"}`；`Ctrl-C` 输出关闭信号与 server 已停止的 JSON 日志；提交前扫描未发现强凭据特征，公开地址仅为 `127.0.0.1` 本地回环地址；重装匹配的官方 Go 1.26.5 工具链后，`go test -race ./...` 已通过。
- 当前阻塞：CI、漏洞扫描、全新机器复现和资源基线尚未建立。
- 下一步：按边界明确的批次补齐 M0 宽口径证据。先建立 CI，再建立漏洞扫描和资源/复现证据；每批完成后更新本页和 M0 学习笔记。源码由项目所有者手敲，协作助手提供完整代码块和验收命令。
- 本阶段不要做：PostgreSQL、Kafka、Kubernetes、OpenTelemetry、前端或 AI 功能
