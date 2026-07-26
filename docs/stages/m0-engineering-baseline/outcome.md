# M0 实际实施结果

> 本文件记录仓库内 M0 施工包的实际结果。原始 v3.1 方案书的更宽工程证据另见 `../../knowledge/go/m0-engineering-baseline.md`。

## 实际完成

- 建立 `cmd/apiserver` 启动入口，以及 `config`、`health`、`transport/http` 平台包。
- 实现 `LYAPUS_HTTP_ADDR` 加载与启动前校验；默认地址为 `127.0.0.1:8080`。
- 实现 `/livez`、`/readyz` JSON 健康检查、JSON 结构化请求日志和 `SIGINT`/`SIGTERM` 优雅关闭。
- 为配置、健康检查和 HTTP server 编写单元测试。
- 建立 README、示例环境配置、忽略规则、Makefile、协作式手敲代码规范和 ADR-0001。

## 验证

- `make fmt` 通过。
- `make test` 通过。
- `make verify` 通过。
- `go run ./cmd/apiserver` 启动成功；`/livez` 与 `/readyz` 均返回 `200` 和 `{"status":"ok"}`。
- 启动后按 `Ctrl-C`，日志依次记录关闭信号与 HTTP server 停止。

## 与计划的偏差

- 无实现偏差。首个 Git commit 由项目所有者在完成提交前的隐私检查后手动创建。

## 证据

- Git commit：`fd9faf4`（`feat: establish M0 engineering baseline`）
- 会话记录：`../../progress/sessions/2026-07-21-m0-engineering-baseline.md`

## 施工包进入 M1 的条件

- [x] M0 `checklist.md` 的所有必需项完成。
- [x] 新 shell 下可重复通过验收命令。

这些条件仅表示当前施工包已验收；全新机器复现和 CI 分支保护门禁仍在 M0 宽口径收口中，不得提前写成已完成。`go test -race ./...` 已在匹配的 Go 1.26.5 工具链上通过，`govulncheck@v1.6.0` 在本地与 GitHub Actions `Verify` 中均未发现漏洞；空闲资源基线见 `../../benchmarks/m0-idle-resource-baseline.md`。
