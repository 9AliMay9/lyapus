# M0 实际实施结果

> 本文件记录仓库内 M0 施工包的实际结果。原始 v3.1 方案书的更宽工程证据另见 `../../knowledge/go/m0-engineering-baseline.md`。

## 实际完成

- 建立 `cmd/apiserver` 启动入口，以及 `config`、`health`、`transport/http` 平台包。
- 实现 `LYAPUS_HTTP_ADDR` 加载与启动前校验；默认地址为 `127.0.0.1:8080`。
- 实现 `/livez`、`/readyz` JSON 健康检查、JSON 结构化请求日志和 `SIGINT`/`SIGTERM` 优雅关闭。
- 为配置、健康检查和 HTTP server 编写单元测试。
- 建立 README、示例环境配置、忽略规则、Makefile、协作式手敲代码规范和 ADR-0001。
- 扩展统一校验：普通测试、vet、race、integration 入口和固定版本漏洞扫描均进入 Makefile 与 CI。
- 添加 clean-runner 启动 smoke、`protect-main` required checks、脱敏空闲资源基线、ADR-0002 执行护栏和 ADR-0003 PostgreSQL 主库决策。

## 验证

- `make fmt` 通过。
- `make test` 通过。
- `make verify` 通过；覆盖 gofmt、vet、普通测试、race、integration 入口和 `govulncheck@v1.6.0`。
- `go run ./cmd/apiserver` 启动成功；`/livez` 与 `/readyz` 均返回 `200` 和 `{"status":"ok"}`。
- 启动后按 `Ctrl-C`，日志依次记录关闭信号与 HTTP server 停止。
- GitHub Actions `verify` 与 `smoke` 通过；故意失败的 `verify` PR 被 ruleset 阻止合并，修复后才允许 squash merge。

## 与计划的偏差

- 服务骨架没有阻断性实现偏差；首个 Git commit 由项目所有者在完成提交前的隐私检查后手动创建。
- 原始方案把模块化单体与 PostgreSQL 主库理由合并在第一份 ADR 中。ADR-0001 已接受后不原地改写，因此 PostgreSQL 决策按文档规范独立记录为 ADR-0003。
- `make integration` 当前验证的是预留入口；M0 没有数据库、容器或跨服务边界，因此没有真实 integration test。
- 已做忽略规则、公开配置边界和提交前凭据特征检查；尚未引入独立的自动化 secret-scanning 产品。
- 没有另租主机执行手工全新 Ubuntu 15 分钟计时复现；现有主机、新 shell 与 GitHub clean runner 的证据不能替代该项。

## 证据

- Git commit：`fd9faf4`（`feat: establish M0 engineering baseline`）
- 工程证据提交：`9af1a16`、`fe1ab2e`、`a05ff0f`
- 门禁验证合并：`6306d4b`（`docs: verify M0 merge gate (#1)`）
- 会话记录：`../../progress/sessions/2026-07-21-m0-engineering-baseline.md`
- 终局审计记录：`../../progress/sessions/2026-07-27-m0-final-audit.md`

## 施工包进入 M1 的条件

- [x] M0 `checklist.md` 的所有必需项完成。
- [x] 新 shell 下可重复通过验收命令。
- [x] clean runner 启动 smoke、漏洞扫描和失败 CI 合并门禁已验证。

终局结论：当前实现完整满足仓库内 M0 施工包，并基本符合原始 v3.1 的 M0 工程基线预期，可以进入 M1。手工全新机器 15 分钟复现被明确延期，不得写成已完成，也不阻塞 M1；其余宽口径差异已经完成或形成可追踪的边界说明。空闲资源基线见 `../../benchmarks/m0-idle-resource-baseline.md`。
