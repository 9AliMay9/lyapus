# 工作会话：2026-07-28 - M1 前上下文交接

## 目标

记录 M0 关闭后的协作工具状态、模型使用约定，以及开始 M1 前发现的仓库标识一致性缺陷，使下一会话可以从确定入口继续。

## 实际完成

- 确认远端 GitHub SSH 地址为 `git@github.com:9AliMay9/lyapus.git`，因此公开仓库的规范 Go module/import 路径应为 `github.com/9AliMay9/lyapus`。
- 发现当前 `go.mod` 和 5 处内部 Go import 仍写作 `github.com/9Alimay/lyapus`；README clone 地址已使用远端规范拼写。该缺陷由 Copilot review 指出并经本地远端配置与源码检索确认。
- 将该修复列为 M1 施工包前的独立小 PR，不与业务实现混合。
- 记录当前协作模型约定：日常使用 Terra + medium；阶段施工包确认前和阶段验收前使用 Sol 做范围与证据审阅，高影响决策按需临时使用 Sol。

## 修改的文件

- `docs/project-context.md`
- `docs/progress/current.md`
- `docs/progress/sessions/2026-07-28-pre-m1-context-handoff.md`

## 验证证据

```text
命令：git remote -v
结果：origin 指向 git@github.com:9AliMay9/lyapus.git

命令：rg -n 'github\\.com/9Ali(?:may|May)9?/lyapus' --glob '*.go' --glob 'go.mod' .
结果：go.mod 与 5 处内部 import 使用 9Alimay；README clone 地址使用 9AliMay9
```

## 偏差、风险或待确认事项

- 本次只更新交接文档，不修改 `go.mod` 或 Go 源码；下一次必须先完成路径规范化 PR。
- GitHub 的仓库访问通常不区分该大小写，但 Go module path 是公开标识，应与规范仓库地址保持完全一致，避免未来消费者、文档和 import 语义混乱。

## 下一次从这里继续

- 具体文件：`go.mod`、`cmd/apiserver/main.go`、`internal/platform/transport/http/server.go`、`internal/platform/transport/http/server_test.go`、`docs/project-context.md`。
- 具体任务：建立并合并仅包含 module/import 路径规范化的 PR；完成后再创建 M1 施工包。
- 验收命令：`make verify`、`git diff --check`、PR required checks（`verify` 与 `smoke`）。
- 不要做：不要在这次小修复中引入 M1 的 PostgreSQL、CRUD、鉴权、Compose 或其他业务实现。
