# M0 实际实施结果

> 在 M0 验收结束时填写。计划不等于结果；未完成项必须如实保留。

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

## 进入 M1 前仍需满足

- M0 `checklist.md` 的所有必需项完成。
- 新 shell 下可重复通过验收命令。
