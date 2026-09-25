# Copilot 审阅后续修复（进行中）

## 范围

基于 PR #23–#27 的审阅意见及源码核对，在 `fix/review-followups` 收敛五项采纳建议；基线为已合并的 Compose 交付 PR #27（`8126a09`）。不将本轮修复视为 M1 总验收完成。

## 修改

- 开发数据库默认名改为 `lyapus_dev`；集成测试在建立连接前校验 pgx 解析的实际数据库名，要求 `_test` 后缀且明确拒绝历史开发库 `lyapus_dev_test`，覆盖查询参数覆盖和编码名称。
- Team、Service、Environment smoke 的数字 ID 子串匹配及相关提取改为 JSON 类型/值断言，保留状态码、request ID 和失败诊断。
- 初始 Environment 使用 Environment 的 slug/name 规范化函数。
- README 明确 Compose API 与宿主机 `make run` 是替代启动路径；plan 修正 Compose 尚待实测的过时表述。
- Compose runbook 区分新默认库与旧卷：修改初始化变量不会重命名已有数据库，旧卷迁移或重建未执行。

## 已有证据

- 项目所有者提供：`go vet ./...`、`go test -count=1 ./...`、`go test -race -count=1 ./...` 全部通过。
- `go test -tags=integration -count=1 -v -run '^TestValidate(Parsed)?IntegrationDatabaseName$' ./internal/catalog/postgres`：两个测试函数、11 个子测试通过，不连接数据库。
- 之前针对旧库和查询参数覆盖旧库的真实测试入口调用，在连接前按预期拒绝。
- 助手仅运行无外部服务访问的分段模拟与静态检查：有效响应被接受，错误类型/ID/字段等被拒绝；完整 YAML、15 个 Shell 块和19 段 Python 语法检查通过，诊断文件名和 shell 变量名扫描通过。
- 项目所有者确认本轮开始前 integration project 无容器、目标卷不存在、55433 无监听；随后建立独立 project/卷，PostgreSQL Healthy，查询实际库名为 `lyapus_integration_test`。
- dry-run 后正式应用两份迁移，status 为 `20260729030502`、executed 2、pending 0；integration race 通过（4.435s），`make verify` 通过（全量 integration 中 postgres 包 3.209s），漏洞扫描报告 `No vulnerabilities found.`。
- `docker ps --filter publish=55432` 确认开发 PostgreSQL 仍为 `lyapus-dev-postgres-1`，映射 `127.0.0.1:55432->5432/tcp`。2026-09-25 开发 API 回读 Service 1 返回 200，原有 `compose-api`、staging/production 及 2026-09-11 创建时间均保留。这验证旧开发实例数据未被本轮测试清空，不证明它运行了本轮新代码。
- 项目所有者核对 integration project 和卷归属后执行 `down --volumes`，容器、网络、卷均 Removed；取消测试 URL 后再次查询容器和卷均为空。测试数据随卷删除，不再保留；未清理开发项目。

## 收口审阅结论

项目所有者确认模型配置足够后完成本轮本地收口审阅：核对当前改动、阶段 plan/contracts/checklist/outcome、运行证据以及外层原始 v3.1 的 M1 七项最小交付。

- 未发现本轮需要新增核心代码修改的阻塞缺陷。schema、历史 migration、公共 API 和包依赖方向未变；新增 pgx 配置解析仅位于 PostgreSQL 测试代码，业务层仍不依赖 pgx。
- 名称校验使用与建池相同的解析配置，并在连接及清表前执行；只降低误连已知开发库的风险，不提供任意目标数据库的安全证明。
- 初始 Environment 使用自身规范化入口，当前底层规则与之前相同，不改变现有业务契约。
- CI 数字 ID 断言检查类型与值，保留成功、错误、删除及分页链路。分段模拟和静态检查不替代本轮真实 clean-runner HTTP 验证。
- 已修正文档中的启动方式混淆和过时 Compose 待验证描述，补齐新默认数据库与旧卷兼容提醒。原始方案要求的查询计划实验和 M1 总验收仍未完成，不进入 M2。
- 选择继续保留旧开发卷，不在本次修复中新增数据库迁移/重命名操作；重复的内联 smoke 断言可在后续出现维护需求时抽取，本轮不扩展成测试框架重构。
- 本地审阅允许进入 PR/CI 阶段，不代表允许立即合并。收到第一轮 CI 证据后，在同一分支补充本记录，最终以最新提交四项 required checks 为门禁。

## PR #28 首轮 CI 证据

项目所有者提供提交 `c8a2ad7` 的推送、[PR #28](https://github.com/9AliMay9/lyapus/pull/28) 创建与 `gh pr checks --required --watch` 输出：四项成功，0 failing、0 skipped、0 pending。

- [verify](https://github.com/9AliMay9/lyapus/actions/runs/36146668446/job/108109442084)：4m12s。
- [smoke](https://github.com/9AliMay9/lyapus/actions/runs/36146668446/job/108111018688)：59s，本轮 JSON 精确断言已在真实 HTTP 链路执行通过。
- [atlas-community](https://github.com/9AliMay9/lyapus/actions/runs/36146668446/job/108109442475)：2m9s。
- [compose](https://github.com/9AliMay9/lyapus/actions/runs/36146668446/job/108109442452)：2m16s。

本次文档补交仍在原分支、原 PR；提交后须等待新一轮 required checks。此记录只证明上述提交的该次运行，不保证长期稳定性，不声明 PR 已合并。最终门禁及合并状态以 PR 页面为准，不为反复记录最终 CI 再新增提交。

## 待完成与边界

- CI 证据文档补交后，最新提交的四项 required checks 及同一 PR 合并；不能以首轮或历史 PR #27 的成功替代。
- 保留现有开发数据，旧 dev 卷尚未迁移或重建。数据库名称防护不替代操作者确认与权限控制。
- Python `assert` 用于测试预期；执行环境不能开启 `-O` / `PYTHONOPTIMIZE`，生产安全校验不得依赖可被禁用的断言。
