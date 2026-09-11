# 当前施工状态

## 当前落点

- 当前阶段：M0 已完成终局审计；M1 已完成数据库基础设施和 Team、Service、Environment 三类资源的完整 HTTP CRUD 纵切面。Team CRUD/稳定分页施工包已由 PR #13 以 squash commit `f4df01f` 合入 `main`；Team service、创建 API、全局 request ID 与 API smoke 已由 PR #15 以 squash commit `8e84c20` 合入 `main`；Team HTTP read 与 cursor 分页已由 PR #17 以 squash commit `284021e` 合入 `main`；Team HTTP mutate 已由 PR #18 以 squash commit `de7e2d3` 合入 `main`；Service Create/Get/List、初始 Environment 事务创建和并发冲突已由 PR #20 以 squash commit `43c627d` 合入 `main`；Service PATCH/Delete 已由 PR #21 以 squash commit `41f2651` 合入 `main`；Environment 独立 CRUD 与 `service_id` 过滤已由 PR #23 以 squash commit `4b311b9` 合入 `main`。
- M0 结论：真实实现完整满足仓库内施工包，并基本符合原始 v3.1 工程基线预期，可以进入 M1。
- M1 状态：施工包、Atlas 决策、schema、两份 versioned migration、sqlc 基线与 CI 门禁已完成；`LYAPUS_DATABASE_URL`、`pgxpool` 启动 Ping/关闭路径及数据库感知 `/readyz` 已完成并作真实运行验证。三类资源均已具备 repository/service/HTTP CRUD、本地与 required clean-runner 证据。Service 支持全局及 `team_id` 过滤的稳定游标分页，并以显式事务原子创建 Service 与可选初始 Environment；两个独立 repository 并发创建同一 Service 已验证恰好一个成功、一个冲突。Environment 支持全局及 `service_id` 过滤的稳定游标分页。严格 JSON、415 媒体类型、统一错误、全局 request ID 与不透明 cursor 保持既有边界。直接 SQL 约束测试与收口审阅已完成：17 个函数、100 个子测试分段通过，integration race 与 `make verify` 通过；PR #25 的三项 required CI 已通过，并以 `51fd1d2` 合入 `main`。Compose 本地新 project/卷路径已验证，README 模型/API/演示文档已补齐；容器交付 clean-runner 验收、查询计划实验和 M1 总验收仍待完成。
- M1 最小范围：Team、Service、Environment CRUD，PostgreSQL migration/约束/事务/并发正确性，单元与真实数据库测试，Compose 空环境复现，以及一份查询计划优化记录。
- M1 默认实现：Go 1.26.6、PostgreSQL 16.14、`pgx/v5` + `pgxpool`、chi/v5、sqlc 1.31.1、手写 SQL + repository adapter、identity bigint 和不透明游标。chi 保持标准 HTTP handler；sqlc 生成类型不越过 PostgreSQL adapter。选择理由与适用边界见施工包。
- Migration 已由 ADR-0004 最终确定：P-0001 完成固定 Atlas Community v1.2.0 的两次 migration（空库 apply、已有库前滚、重复 apply、status 与完整性篡改拦截）、同一 `db/schema.sql` 的 sqlc 1.31.1 解析/生成、本机可复现 Community 构建，以及 PR #8 中 required `atlas-community` CI 实跑。未来触发退出条件时才以新 ADR 记录并回退 `golang-migrate`。
- 本次终局审计所依据的已合并门禁基线：commit `6306d4b`（`docs: verify M0 merge gate (#1)`）。
- 仓库标识修正：commit `2b3764b`（`chore: normalize module path (#5)`）已将 `go.mod` 与 5 处内部 import 统一为远端规范地址 `github.com/9AliMay9/lyapus`；commit `d96946d`（`docs: record module path merge (#6)`）完成合并事实记录。

## M0 已验证事实

- 服务实现：配置加载与校验、JSON 结构化日志、HTTP server、`/livez`、`/readyz`、`SIGINT`/`SIGTERM` 优雅关闭。
- 本地校验：`gofmt`、`go vet ./...`、`go test ./...`、`go test -race ./...`、`go test -tags=integration -count=1 ./...` 和固定版本 `govulncheck@v1.6.0`。
- 真实运行：两个健康端点均返回 `200` 与 `{"status":"ok"}`；`Ctrl-C` 记录关闭信号和 server 停止日志。
- CI：GitHub Actions 的 `verify` 与 clean-runner `smoke` 已通过；smoke 在干净 Ubuntu runner 启动服务并检查两个健康端点。
- 合并门禁：M0 曾以 `verify` 与 `smoke` 实测失败阻止合并；当前 `protect-main` 进一步要求 PR、最新分支以及 `verify`、`smoke`、`atlas-community` 三项检查。
- 工程证据：脱敏的 M0 空闲资源基线、公开/私有信息边界、ADR-0001/0002/0003 和系统化 M0 学习笔记已经建立。
- 协作工具：当前远程开发主机已配置可选的 GitHub CLI；Git 继续使用 SSH，CLI API 登录与最小命令见 `../runbooks/github-ssh.md`。

## 明确延期与边界

- 没有在另一台全新 Ubuntu 主机上手工计时完成“按 README 15 分钟启动”。当前主机、新 shell 和 GitHub clean-runner smoke 提供了近似证据，但不能写成该项已经实测完成。
- M0 验收时 `make integration` 只有测试入口；当时没有数据库、容器或跨服务依赖，因此没有真实集成测试内容。M1 现已加入 Team repository 的真实 PostgreSQL integration test。
- M0 没有 PostgreSQL、Kafka、OpenTelemetry、Kubernetes、前端或 AI 功能，也不声明生产容量、高可用或生产就绪。

## 与原始 v3.1 M1 最小版的差距

- 已成立：Service/Environment/归属模型，三类资源 CRUD、校验、统一错误、分页/过滤，关键单元测试、真实数据库 repository 测试和 race，以及 Service + 初始 Environment 事务与并发唯一冲突。
- 约束实证已成立：PR #25 已补齐本次清单要求的字段 CHECK 插入边界、唯一性范围、外键和引用删除测试；结合既有 migration、事务与并发证据，原始最小项 3 的本地与 CI 证据已具备，已由 PR #25 合入 `main`。未穷尽 NULL、UPDATE 或全部 Unicode 空白行为。
- 尚未成立：本次容器交付的 clean-runner 验收、最终 README 复走；带数据量、SQL、参数及 `EXPLAIN (ANALYZE, BUFFERS)` 的索引前后查询计划记录。Compose 本地新 project/空卷已验证，README 模型与演示路径已补齐。
- 不阻塞 v0.1：API Token、最小 RBAC、审计日志、复杂幂等/乐观并发、更丰富过滤排序、HTTP E2E、k6 与 pprof 均属于原始“深度增强”，当前未实现也不得写成最小版缺陷。

## 下一次从这里开始

### 2026-09-11 Compose 复盘后接续点（优先于下方历史交接）

- PR #26 已由项目所有者提供合并输出，基线 `cd1d382`；当前 `feat/m1-compose-delivery` 已有 Dockerfile、Compose 和文档未提交变更，保留现有修改。
- 本机新开发 project/卷完成构建、宿主机 Atlas 显式迁移、API 创建/回读、数据库中断恢复。独立 integration project/卷完成 migration、integration race（4.562s）和 `make verify`，测试资源已删除，开发 API/卷和演示数据保留。
- 本轮配置确认后的复盘与文档同步已完成，见 [Compose 复盘](sessions/2026-09-11-compose-delivery.md) 和 [操作手册](../runbooks/compose-delivery.md)。README 已补模型和演示路径，不声明全新主机五分钟验收。
- 新增独立 `compose` job 已通过静态复查，等待首次 PR 实跑；项目所有者同意加入 required checks，但规则实际更新尚待确认。API 停止已验证 exit 0、非 OOM、关闭日志及重启后 readiness 200。下一段准备首次 PR/CI，再在同一 PR 补运行证据与最终 README 复走；最新检查通过且门禁配置确认后才合并。
- M1 查询计划实验、最终资源记录及 release 仍未完成，不进入 M2。下方是之前的交接历史，不再重新建立已存在分支。

1. PR #25 合并后已确认干净、同步的 `main...origin/main`，基线为 `51fd1d2 test: verify catalog database constraints (#25)`。本次合并事实回填文档另走文档 PR；若其已合并，以该文档 PR 的 squash commit 为最新基线。始终先检查实际 Git 状态并保留未提交修改。
2. 首读 `docs/project-context.md`、`docs/standards/README.md`、本文件、`docs/stages/m1-go-backend/{README,plan,contracts,checklist,outcome}.md`、ADR-0004、数据库 migration/integration runbook，以及 `docs/progress/sessions/2026-09-10-catalog-db-constraints.md`。
3. 约束测试施工包已由 PR #25 完成本地、required CI 和合并收口，原本地及远端分支已删除。2026-09-11 项目所有者已停止临时测试容器，确认容器列表为空，并 unset 当前 shell 的测试连接变量。当前 `docs/post-pr25-sync` 汇总合并事实、资源清理和新 Git 流程，作为一次过渡文档 PR 完成提交、CI 和合并；不再为确认此文档 PR 合并而新增分支。纯文档同步不递归触发收口审阅。
4. 随后建立 `feat/m1-compose-delivery`：增加 Dockerfile、固定 PostgreSQL 16.14 的 Compose、显式 migration 步骤、API 启动/健康顺序、dev/test 隔离和空 project/空卷验收；migration 不进入应用启动路径。同步补齐根 README 的数据模型、API 示例和五分钟演示。
5. 再以独立实验完成 Service 按 Team 游标列表的查询计划与索引前后证据；最后执行 M1 v0.1 总验收与 release 收口。不要提前进入 M1 深度增强或 M2。

不要在应用启动路径自动执行 migration；不要让 Atlas Cloud/Pro、鉴权、RBAC、k6、OpenTelemetry 或其他后续增强进入 M1 v0.1 的阻塞路径。

## 协作审阅约定

- 2026-09-11 起，默认同一分支、同一 PR 完成实现与文档收口：首次 CI 通过后补充证据、审阅和文档，再次 push 更新原 PR，最新提交 CI 通过后统一合并、删除分支；最终 CI 与合并状态由 PR 页面保存，不递归新增记录提交。具体规则见 `../standards/collaboration.md`。
- 2026-09-10 起，所有手敲代码按“手敲完成 → 静态复查与必要修正 → 运行验证 → 下一段”推进；静态复查也检查助手交付代码本身的正确性，不能仅核对转写。具体规则以 `../standards/collaboration.md` 为准，替代此前静态复查通过即交付下一段的顺序。
- 日常学习、施工和测试解释使用适合连续协作的模型配置；阶段设计、高影响判断和收口一致性审阅使用当时可用的高能力配置。
- 功能性 PR、独立施工包或阶段达到收口点时，协作助手先提醒并等待项目所有者明确确认模型/参数已经调整，再开始核对实现、施工包、原始方案、证据和文档一致性。
- 完整触发条件、审阅范围和例外以 `../standards/collaboration.md` 的“高能力收口审阅门禁”为准；该约定是协作流程，不是项目运行时依赖或对外承诺。
