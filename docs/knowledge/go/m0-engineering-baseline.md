# M0：从空仓库到可控的 Go 服务进程

## 一句话模型

M0 的价值不是“启动了一个 HTTP 端口”，而是把一个进程变成可配置、可诊断、可测试、能安全退出且有证据可复现的服务，为 M1 业务代码提供稳定外壳。

## 对照原始方案书的结论

当前实现完整满足仓库内 M0 施工包的核心契约，并覆盖原始 v3.1 方案书最重要的服务骨架。若按原方案书的宽口径衡量，工程证据仍有缺口，因此准确结论是“核心实现基本符合预期”，不是“原方案 M0 全部完成”。

| 原始方案要求 | 当前证据 | 结论 |
| --- | --- | --- |
| 配置、结构化日志、优雅退出、`/livez`、`/readyz` | `config`、`health`、`transport/http`、`cmd/apiserver` 已实现并完成真实启动验证 | 已完成 |
| README、`.gitignore`、公开 `.env.example`、ADR | 工程文件、公开/私有边界和 ADR-0001 已提交 | 核心完成；PostgreSQL 理由与 ADR-0002 尚未进入当前施工包 |
| 单元测试、静态检查、一键校验 | 配置、健康检查、HTTP server 单测；`make verify` 执行格式化、`go vet`、普通测试、race、integration 入口与漏洞扫描验证 | 已完成当前施工包要求 |
| race detector、漏洞扫描、CI | 重装匹配的官方 Go 1.26.5 工具链后，`go test -race ./...` 已通过；GitHub Actions `Verify` 在 `9af1a16` 上成功运行；`govulncheck@v1.6.0` 已通过 `make vuln` 扫描源码且未发现漏洞，并在 `fe1ab2e` 的 CI 中成功运行；`a05ff0f` 的 `verify` 与 clean-runner `smoke` 均成功；`protect-main` ruleset 已要求 PR、`verify`、`smoke` 和最新分支，故意失败的 `verify` PR 已被 GitHub 阻止合并 | 已完成 |
| `fmt`、`lint`、`test`、`integration`、`run`、`verify` 入口 | Makefile 已提供全部入口；M0 的 `integration` 仅是为 M1 预留的 tag 测试入口，尚无外部依赖测试 | 入口已完成；集成测试内容待 M1 |
| 全新机器 15 分钟启动、CI 阻止合并 | 已在现有 Linux 主机和新 shell 验证运行；`a05ff0f` 的 GitHub-hosted clean Ubuntu `smoke` job 已成功启动服务并检查两个健康端点；`protect-main` ruleset 已实测阻止失败 CI 合并；尚未验证手工全新机器复现 | 部分完成 |
| 空闲资源基线与公开资源口径 | 已记录脱敏的 [M0 空闲资源基线](../../benchmarks/m0-idle-resource-baseline.md)：7.5 GiB 内存、178 GiB 根盘、无 Swap 使用且 memory PSI 为 0 | 已完成单次空闲基线；后续阶段须重复采样 |

仓库内阶段施工包是日常执行真源；原方案书用于检查路线有没有被过度缩减。二者不一致时，应明确写成“当前阶段已验收、原方案证据待补”，而不是互相覆盖。

## 项目如何演化到当前形态

### 1. 先固定契约，再写实现

M0 先固定了命令名、默认地址、环境变量、构造函数和健康检查响应。这样做把“调用方可以依赖什么”与“内部如何实现”分开：内部可以重构，`config.Load()`、`http.NewServer(...)` 和端点语义不能随手漂移。

对应位置：M0 的 [契约](../../stages/m0-engineering-baseline/contracts.md) 与 [ADR-0001](../../architecture/decisions/ADR-0001-modular-monolith.md)。

### 2. 配置是进程启动的第一道边界

`config.Load()` 把环境变量转换成经过校验的 `Config`。缺省值保证本地可直接运行；`net.SplitHostPort` 和端口范围检查让非法配置在监听端口前失败；`fmt.Errorf(... %w ...)` 保留底层原因并补充调用上下文。

关键认识：配置解析成功不等于服务启动成功，但配置错误应尽可能早、尽可能明确地失败。把原始环境变量散落在业务包中，会让测试、错误定位和未来配置迁移都变困难。

### 3. 健康语义独立于 HTTP 组装

`health.Handler` 只处理健康语义和 JSON 响应，路由注册留给 transport 包。`Livez` 与 `Readyz` 当前共用 `respondOK`，但名称分别对应两个已经不同的契约：liveness 只证明进程存活；readiness 证明当前可接收流量。M0 没有外部依赖，所以两者结果相同。

单元测试使用 `httptest` 直接调用 handler，不需要占用真实端口。这证明业务边界可以在进程外壳之外独立验证。

### 4. transport 层负责协议装配

`transport/http.NewServer` 组合配置、健康处理器、标准库 `ServeMux` 和请求日志中间件。选择标准库而非 Web 框架，是因为 M0 只有两个端点，没有路由参数、复杂中间件或框架生态需求。

`ReadHeaderTimeout` 限制读取请求头的等待时间，避免一个连接无限占用 server 资源。请求日志通过注入的 `*slog.Logger` 输出稳定字段，避免业务包依赖全局 logger。

当前日志仍有明确边界：只记录 method、path 和 duration，没有捕获响应状态码、响应大小或 request ID；这些应由真实诊断需求驱动，而不是在 M0 提前造完整日志框架。

### 5. `main` 只负责进程生命周期

`cmd/apiserver` 完成依赖装配，不承载健康或传输逻辑。启动错误通过带缓冲的 channel 返回主 goroutine；`signal.NotifyContext` 把 `SIGINT`/`SIGTERM` 转换为取消信号；`Server.Shutdown` 给已有请求最多 10 秒完成时间；`http.ErrServerClosed` 被识别为正常关闭结果。

这条生命周期可以压缩为：

```text
加载配置 → 组装依赖 → 启动监听
                    ↓
          监听失败或退出信号
                    ↓
         限时 Shutdown → 返回结果
```

关键认识：优雅关闭不是“收到信号后立即退出”，而是停止接收新连接、等待在途请求，并为等待设置明确上限。

### 6. 证据把“能运行”变成“可信”

M0 已形成三层证据：单元测试验证纯行为；`make verify` 验证格式、全部测试、静态检查、race 和漏洞扫描；真实进程实验验证两个 HTTP 端点、JSON 日志和 `Ctrl-C` 关闭路径。提交前又检查了暂存范围、空白错误和敏感信息，并用 Git commit 固化结果。

最初的 `go test -race ./...` 因 Go 安装的二进制与标准库元数据不匹配而失败；重装匹配的官方 Go 1.26.5 工具链后已通过。这个过程说明：工具命令失败时应先区分项目错误与工具链错误；只有重跑成功后才能把 race detector 写为已通过。

## 容易混淆或踩坑

- `/livez` 与 `/readyz` 当前响应相同，不代表语义相同；以后也不能让 liveness 依赖数据库等外部系统。
- `json.Encoder.Encode` 会在 JSON 后追加换行，所以测试期望值是 `{"status":"ok"}\n`。
- `ListenAndServe` 在正常 `Shutdown` 后返回 `http.ErrServerClosed`，不能把它记录成启动故障。
- `go test` 通过只覆盖测试实际执行的路径；真实监听、信号和 shell 命令仍需单独验收。
- 结构化日志是字段稳定、机器可解析，不只是把普通文本包进 JSON。
- 文档中的“已完成”必须区分当前施工包验收与原始方案书的更宽要求。

## 最小复现实验

```bash
make verify
go run ./cmd/apiserver
curl -i http://127.0.0.1:8080/livez
curl -i http://127.0.0.1:8080/readyz
```

确认两个端点返回 `200` 和 `{"status":"ok"}`，再向前台进程发送 `Ctrl-C`，观察 `shutdown_signal_received` 与 `http_server_stopped` 日志。

待补实验：在全新 Ubuntu 环境按 README 计时复现。

## 面试表达

可以这样概括 M0：我先用模块化单体建立 Go 服务的工程外壳，把配置校验、健康语义、HTTP 组装和进程生命周期拆到明确边界；使用标准库 `slog`、`net/http`、`signal.NotifyContext` 和 `Server.Shutdown`，并用单元测试、`go vet`、race detector、固定版本漏洞扫描、GitHub Actions、真实 curl 与信号实验保存证据。`protect-main` ruleset 要求 PR、通过 `verify` 与 `smoke`，且已实测拦截失败 CI；我仍明确保留尚未完成的手工全新机器复现，避免把最小骨架描述成生产就绪系统。

## 延伸阅读

- [M0 施工计划](../../stages/m0-engineering-baseline/plan.md)
- [M0 实际结果](../../stages/m0-engineering-baseline/outcome.md)
- [当前架构](../../architecture/overview.md)
- [Go 实现规范](../../standards/go.md)
