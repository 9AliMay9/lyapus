# M0：从空仓库到受控的 Go 服务工程

## 一句话模型

M0 的价值不是“启动了一个 HTTP 端口”，而是建立一条可信的工程控制链：先固定契约，再实现可控进程，用本地测试、CI、真实运行、资源证据和合并门禁证明它能被长期维护。

## 终局判断

当前实现完整满足仓库内 M0 施工包，并基本符合原始 v3.1 方案书的工程基线预期，可以进入 M1。它不是生产系统，也没有逐字完成原方案的每项证据：另一台全新 Ubuntu 主机上的手工 15 分钟计时复现被明确延期。

M0 没有“近期最小版/深度增强”拆分；这种 release 分层从 M1 开始。M0 本身是后续所有阶段共用的工程底座，因此本次补齐了 CI、race、漏洞扫描、资源基线、ADR 和合并门禁，而没有给服务添加超出阶段的业务功能。

## 原始方案逐项对照

| 原始 M0 要求 | 当前事实 | 结论 |
| --- | --- | --- |
| README、路线、贡献约定、ADR 模板 | 根 README、`docs/stages/`、协作规范、ADR 规则和模板均存在 | 已完成 |
| `fmt`、`lint`、`test`、`integration`、`run`、`verify` | Makefile 提供全部入口，并额外提供 `race`、`vuln` | 已完成；integration 只有预留入口 |
| 配置、结构化日志、优雅退出、健康检查 | `config`、`health`、`transport/http`、`cmd/apiserver` 已实现并做真实进程验证 | 已完成 |
| 单元测试、race、静态检查、漏洞扫描、CI | 普通测试、`go vet`、race、固定 `govulncheck@v1.6.0`、GitHub Actions `verify`/`smoke` 均有成功证据 | 已完成 |
| `.gitignore`、公开 `.env.example`、secret scan 和私密边界 | 忽略规则、示例配置、公开/私有边界与提交前凭据检查已建立 | 基本完成；未引入独立自动化 secret scanner |
| 模块化单体与 PostgreSQL 决策 | ADR-0001 记录模块化单体；为不改写已接受 ADR，ADR-0003 单独记录 PostgreSQL 主库 | 已完成，编号形式与原文不同 |
| 公开资源口径与 M0–M3 磁盘控制 | 根 README 记录通用规格、容器内存目标、磁盘停止线和阶段数据边界 | 已完成规划口径；容量仍待各阶段实测 |
| ADR-0002 四条执行护栏 | 双压测拓扑、公开/私有边界、M2-min 声明、Compose/k3s 状态隔离均已记录 | 已完成 |
| 空闲资源基线 | `free`、`vmstat`、memory PSI、`ps`、磁盘与 `docker stats` 已脱敏保存 | 已完成单次基线 |
| 一条命令校验 | `make verify` 覆盖格式、vet、普通测试、race、integration 入口和漏洞扫描 | 已完成 |
| CI 失败阻止合并 | `protect-main` 要求 PR、最新分支、`verify`、`smoke`；故意失败 PR 已被阻止，修复后才可 squash merge | 已完成 |
| 全新机器 15 分钟启动 | 现有主机、新 shell 和 GitHub clean runner 均成功；没有另一台全新主机的手工计时 | 延期，不能写成已完成 |

## 工程控制链

```text
契约与规范
    ↓
源码与单元测试
    ↓
make verify
    ↓
GitHub Actions verify
    ↓
clean-runner smoke
    ↓
protect-main required checks
    ↓
PR 合并后的 Git 真源
```

每一层解决的问题不同：

- 契约防止名称和语义随实现漂移。
- 单元测试证明包内行为。
- `make verify` 统一开发者与 CI 的校验入口。
- clean-runner smoke 证明服务能在仓库外的干净 runner 上构建、启动并响应健康检查。
- ruleset 把“建议运行 CI”变成“失败就不能合并”的工程约束。
- 文档只记录已验证事实和明确边界，不能替代源码与 Git commit。

## 代码结构与 Go 认识

### 配置边界

`config.Load()` 只读取 `LYAPUS_HTTP_ADDR`，缺省为 `127.0.0.1:8080`。它用 `net.SplitHostPort` 和端口范围检查让非法配置在监听前失败，并用 `%w` 保留底层错误链。

重要认识：配置解析成功不等于服务一定能监听成功，但配置错误应尽早、明确地失败。环境变量不能散落在业务包中，否则测试和未来迁移都会变困难。

### 健康语义

`health.Handler` 把健康语义从 HTTP server 装配中分离。`Livez` 表示进程是否存活；`Readyz` 表示实例是否能接收流量。M0 没有外部依赖，所以二者现在返回相同的 `200` 与 `{"status":"ok"}`，但契约已经不同。

`httptest.NewRequest` 与 `httptest.NewRecorder` 可以直接验证 handler，不占真实端口。`json.Encoder.Encode` 会追加换行，因此测试期望包含 `\n`。

### HTTP 装配

`transport/http.NewServer` 组合 `config.Config`、健康 handler、标准库 `http.ServeMux` 和请求日志中间件。M0 只有两个静态端点，标准库足够；此时引入 Web 框架只会增加依赖和隐式行为。

`ReadHeaderTimeout` 限制读取请求头的等待时间。logger 通过 `*slog.Logger` 注入而非使用全局变量，便于测试和未来装配。当前请求日志只记录 method、path、duration，没有状态码、响应大小或 request ID；这是一条明确边界，不是遗漏的生产日志平台。

### 进程生命周期

`cmd/apiserver` 只负责依赖装配和生命周期：

```text
加载配置 → 组装 server → goroutine 启动监听
                         ↓
               监听失败或收到退出信号
                         ↓
                10 秒限时 Shutdown
                         ↓
                 等待 server 返回结果
```

带缓冲的 `serverErr` channel 把监听结果传回主 goroutine；`signal.NotifyContext` 把 `SIGINT`/`SIGTERM` 转成取消；`http.ErrServerClosed` 是正常 Shutdown 的结果，不能记录成启动故障。

优雅关闭不是立即杀掉进程，而是停止接受新连接、等待在途请求，并为等待设置上限。当前关闭路径由真实 `Ctrl-C` 实验验证，尚没有 `main` 包自动化测试。

## 测试与 CI 的准确含义

- `go test ./...`：执行当前单元测试；通过不代表真实监听和信号路径已经覆盖。
- `go test -race ./...`：在测试执行路径中检测数据竞争；不能证明未执行路径无竞争。
- `go vet ./...`：发现一类静态可疑模式；不是完整形式化验证。
- `go test -tags=integration -count=1 ./...`：保留 M1 的 integration 入口并禁用测试缓存；M0 没有真正跨外部依赖的测试。
- `govulncheck ./...`：按可达调用关系检查已知 Go 漏洞；“No vulnerabilities found”只代表当时工具、数据库和源码范围内未发现。
- `smoke`：在 GitHub-hosted clean Ubuntu runner 构建并启动进程，轮询 `/livez` 后检查两个端点；它不验证手工安装速度，也不验证 `Ctrl-C` 日志。

CI 页面最初只有一个绿色 `verify`，因为 job 才是检查单位，job 内多个 step 不会各自显示为合并门禁。增加独立 `smoke` job 后，ruleset 才要求两个明确检查名。

## 两次值得保留的工程教训

### 工具链错误不等于项目错误

最初 race 失败是 Go 二进制与标准库元数据不匹配。重新安装校验过摘要的官方 Go 1.26.5 后，`go list -race runtime/race` 和全仓 race 测试通过。

排障顺序应是：读取原始错误 → 判断项目、工具链还是网络层 → 修复最小根因 → 重跑同一验收。没有成功重跑就不能把文档改成“已完成”。

### 失败门禁必须做负向实验

仅看到 CI 绿色不能证明失败会阻止合并。本项目在临时分支加入一个故意 `t.Fatal` 的测试：`verify` 失败、`smoke` 因依赖跳过、GitHub 明确显示 merge requirements 未满足；删除探针并让两个检查转绿后，PR 才能 squash merge。

负向实验完成后，临时故障代码不进入 `main`，但文档和 GitHub PR 保留门禁证据。

## 容易混淆或踩坑

- `/livez` 与 `/readyz` 当前响应相同，不代表语义相同；未来不能让 liveness 依赖数据库。
- `ListenAndServe` 在正常 Shutdown 后返回 `http.ErrServerClosed`。
- `make verify` 当前包含会写回格式的 `gofmt`；提交前仍要检查 diff，避免忽略格式化产生的变更。
- CI 的“step”与 ruleset 要求的“job/check”不是同一层。
- GitHub clean runner 是很强的可复现证据，但不是“另一台新服务器上由人按 README 15 分钟完成”的同义词。
- M0 只证明工程外壳可信，不证明 API 容量、数据库正确性、生产可用性或高可用。
- 公开资源数字必须附带环境和限制；私有实例、IP、账单、代理和密钥不能进入 Git。

## 最小复现实验

```bash
go install golang.org/x/vuln/cmd/govulncheck@v1.6.0
make verify
go run ./cmd/apiserver
curl -i http://127.0.0.1:8080/livez
curl -i http://127.0.0.1:8080/readyz
```

确认两个端点返回 `200` 和 `{"status":"ok"}`，再向前台进程发送 `Ctrl-C`，观察 `shutdown_signal_received` 与 `http_server_stopped`。

仍未完成的实验：在另一台全新 Ubuntu 主机上从获取仓库开始计时，按 README 在 15 分钟内完成工具准备、统一校验和服务启动。

## 带入 M1 的稳定边界

- 保留模块化单体，不为展示微服务提前拆分。
- `cmd/` 继续只做装配；业务进入新的明确包边界。
- PostgreSQL 是 M1 主库，但版本、迁移工具、测试隔离和数据模型必须在 M1 施工包中先锁定。
- M1 开始区分“近期最小 release”和“深度增强”；鉴权、复杂 RBAC、审计和性能增强不能反向阻塞最小版。
- M1 引入数据库与 Compose 后，必须重复空闲/运行资源采样，并第一次形成真实 integration test。

## 面试表达

可以这样概括 M0：我从空仓库建立了一个模块化单体 Go 服务外壳，把配置、健康语义、HTTP 装配和进程生命周期拆成清晰边界；使用标准库 `slog`、`net/http`、`signal.NotifyContext` 和 `Server.Shutdown`，并用单元测试、vet、race、固定版本漏洞扫描、clean-runner smoke 和真实进程实验保存证据。我还通过一次故意失败的 PR 验证 required checks 确实阻止合并。对于没有做过的全新机 15 分钟复现、真实集成测试和生产容量，我在文档中明确保留边界，而不把工程骨架包装成生产系统。

## 延伸阅读

- [M0 施工计划](../../stages/m0-engineering-baseline/plan.md)
- [M0 实际结果](../../stages/m0-engineering-baseline/outcome.md)
- [M0 空闲资源基线](../../benchmarks/m0-idle-resource-baseline.md)
- [ADR-0001：模块化单体](../../architecture/decisions/ADR-0001-modular-monolith.md)
- [ADR-0002：执行护栏](../../architecture/decisions/ADR-0002-execution-guardrails.md)
- [ADR-0003：PostgreSQL 主库](../../architecture/decisions/ADR-0003-postgresql-primary-database.md)
- [当前架构](../../architecture/overview.md)
- [Go 实现规范](../../standards/go.md)
