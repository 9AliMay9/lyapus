# Environment Catalog 独立 CRUD 纵切面

日期：2026-09-07

## 范围

- 为 Environment 增加独立 domain 输入、repository 接口与业务服务；保留 Service 创建时“初始 Environment”的独立输入语义。
- 将 Environment 查询从 Service 查询文件拆入 `db/queries/environments.sql`，由 sqlc 生成 pgx/v5 调用；adapter 支持 CRUD、全局/`service_id` 过滤和稳定游标分页。
- 增加 `/v1/environments` 集合 POST/GET 与 `/v1/environments/{environment_id}` GET/PATCH/DELETE，并接入应用装配。
- 复用 catalog 的严格 JSON、媒体类型、统一错误、request ID、完成日志与不透明 cursor 契约；Environment 的 `service_id` 创建后不可经 PATCH 修改。

## 已有验证

- domain/service、PostgreSQL adapter 与 HTTP transport 的普通测试和 race 测试通过。
- 真实 PostgreSQL integration test 覆盖 Create/Get/Update/Delete、缺失父 Service、同 Service slug 冲突、not-found、全局与 `service_id` 过滤的两页稳定游标分页。
- 独立 `_test` PostgreSQL 16.14 数据库 migration status 为最新；`make verify` 的 vet、生成检查、普通测试、race、真实 integration test 与漏洞扫描全部通过，结果为 `No vulnerabilities found.`。
- 可丢弃开发库完成 migration dry-run、apply、status；真实 HTTP 实测三层资源创建、Environment 单项读取、过滤分页、PATCH、冲突、DELETE 204 与删除后 404。响应、错误信封及完成日志 request ID 符合契约；开发与测试容器均已停止并自动删除。

## 收口审阅

- 实现保持模块化单体边界：domain/service 位于 `internal/catalog`，SQL/sqlc 与 pgx 错误分类留在 PostgreSQL adapter，HTTP DTO/cursor/handler 留在 catalog transport，`cmd/apiserver` 只负责装配。
- 独立 `environments.sql` 与 HTTP 文件按资源职责拆分；Service detail 仍复用 Environment 查询，不引入跨层生成类型泄漏或无真实复用需求的通用抽象。
- 本纵切面符合施工包与 v3.1 M1 最小版对 Service/Environment 模型、CRUD、校验、错误、分页/过滤和真实数据库测试的要求；没有扩大到 RBAC、审计、复杂幂等、k6 或 pprof 深度项。
- 当前仅能判定本地实现与验证完成；`.github/workflows/verify.yml` 尚未 smoke Environment 路由，PR、required clean-runner checks 与合并均未发生。

## 下一步

- 补齐 Environment clean-runner smoke，复跑本地检查后提交 PR。
- required checks 全部通过后回填 CI 与合并事实；随后进入 Compose 空环境交付，查询计划实验仍是 M1 v0.1 的独立阻塞项。
