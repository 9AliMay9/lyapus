# Team HTTP read 与 cursor 分页施工

日期：2026-08-13

## 范围

- 将 platform request-ID middleware 放到 completion logger 外层，使 logger 从 request context 读取稳定 ID；响应头只保留防御性 fallback。
- 实现 `GET /v1/teams/{team_id}` 与 `GET /v1/teams`，复用既有 TeamService 和 PostgreSQL repository。
- 在 catalog HTTP transport 内实现不透明的 raw URL-safe base64 cursor 编解码；payload 为 UTC RFC 3339 纳秒时间和 Team ID，不向 repository 泄漏 HTTP token。
- 将 clean-runner smoke 扩展为创建后读取单项和 `limit=1` 列表。

## 验证

- catalog HTTP 与 platform HTTP 单元测试通过；全量 `go test ./...` 和 `go test -race ./...` 通过。
- 可丢弃 PostgreSQL 16.14 开发库完成 Atlas dry-run、apply、status；创建 `platform` 和 `observability` 后，单项 GET 返回正确资源，`limit=1` 第一页返回较新的 Team 与非空 cursor，第二页返回较早 Team 与空 cursor。
- 独立、名称以 `_test` 结尾的 PostgreSQL 16.14 数据库完成 dry-run、apply、status；`make verify` 通过 vet、普通测试、race、真实 integration test 与漏洞扫描。
- `govulncheck` 随后发现 Go 1.26.5 标准库中四个代码路径可达漏洞，均已由 1.26.6 修复。升级本机工具链和 `go.mod` 的 `go` 指令到 1.26.6 后，在同一 `_test` 数据库重新运行 `make verify`，结果为 `No vulnerabilities found.`。

## 边界与后续

- Team HTTP PATCH、DELETE 尚未实现；本施工包没有改变 Team repository、service 或 SQL 的既有 CRUD 语义。
- 审计发现 `r.URL.Query().Get(...)` 无法区分未提供参数与显式空值：`limit=0` 会进入 service 的默认值路径，`limit=` 与 `cursor=` 会被当作未提供。这与公共契约“未提供 limit 时默认 20、显式 limit 范围 1–100、非法 cursor 返回 400”不完全一致。已在本 PR 内改为按 query key presence 判断，并覆盖零值、空值和重复参数；修复后的 catalog HTTP 包测试通过。
- 原始方案书 Day 4 以 Service 创建/查询/更新为示例，但 Day 3 与 M1 最小交付同时要求 Team/Service/Environment 关系及完整 CRUD。先完成 Team 纵切面是施工顺序细化，用于稳定分层、错误和分页模板；没有删除或替代 Service/Environment 范围。
- `statusRecorder` 通过 `Unwrap` 支持 `http.NewResponseController`。如未来引入 SSE、WebSocket、反向代理或其他依赖 optional `ResponseWriter` 接口的 handler，须先专项审查 `Flusher`、`Hijacker` 等接口透明性并补真实行为测试；当前 JSON REST API 不提前实现这些透传。
- 更新后的 smoke 尚待此施工包 PR 的 clean-runner 运行确认；在成功前不声明该 CI 证据已完成。
