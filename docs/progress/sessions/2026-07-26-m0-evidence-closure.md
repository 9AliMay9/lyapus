# 工作会话：2026-07-26 - M0 工程证据收口

## 目标

补齐原始 v3.1 方案书中 M0 宽口径工程证据，并如实记录尚未完成的事项。

## 实际完成

- 以官方 Go 1.26.5 发行包恢复匹配的本机工具链；`go list -race runtime/race` 和 `go test -race ./...` 均已通过。
- 扩展 Makefile：`verify` 统一执行格式化、静态检查、普通测试、race 测试、预留集成测试和漏洞扫描。
- 添加 GitHub Actions `Verify` job；commit `9af1a16` 的 run 30190218910 成功，随后 commit `fe1ab2e` 的漏洞扫描版 Verify job 也成功。
- 固定使用 `govulncheck@v1.6.0`，本机和 CI 扫描均返回 `No vulnerabilities found.`。
- 记录 M0 空闲资源基线，详见 `../../benchmarks/m0-idle-resource-baseline.md`。
- 添加 GitHub-hosted clean Ubuntu 的服务启动与健康检查 smoke job；commit `a05ff0f` 的 `verify`（18 秒）与 `smoke`（10 秒）均已成功。

## 验证证据

```text
命令：go test -race ./...
结果：通过；config、health、transport/http 测试包成功，cmd/apiserver 无测试文件

命令：make verify
结果：通过；包含 gofmt、go vet、普通测试、race 测试、预留 integration 测试和 govulncheck

命令：govulncheck ./...
结果：No vulnerabilities found.

命令：GitHub Actions Verify（commit 9af1a16）
结果：成功，57 秒

命令：GitHub Actions Verify（commit fe1ab2e）
结果：成功，26 秒

命令：GitHub Actions verify 与 smoke（commit a05ff0f）
结果：均成功，总计 28 秒；smoke 在 clean Ubuntu runner 启动服务并检查 /livez、/readyz
```

## 偏差、风险或待确认事项

- GitHub-hosted smoke job 仅验证干净 runner 上的构建、启动和两个健康端点；它不等同于手工全新 Ubuntu 的 15 分钟复现。
- 尚未配置或验证“失败 CI 阻止合并”的 GitHub 分支保护门禁。
- 不要将上述两项待确认事项写成已完成。

## 下一次从这里继续

- 具体文件：`docs/progress/current.md` 和 `docs/knowledge/go/m0-engineering-baseline.md`。
- 具体任务：进行手工全新机器复现，或配置并验证 GitHub 分支保护门禁。
- 验收命令：以选择的收口任务为准；提交前执行 `make verify` 与 `git diff --check`。
- 不要做：不要把 clean-runner smoke 等同于手工全新 Ubuntu 的 15 分钟复现，也不要在未验证前勾选分支保护门禁。
