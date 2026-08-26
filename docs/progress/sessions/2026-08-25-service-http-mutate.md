# Service HTTP 更新与删除纵切面

日期：2026-08-25

## 范围

- 在 catalog model、repository 接口与 ServiceService 增加 Service Update/Delete；`team_id` 保持不可变。
- Service PATCH 支持更新 slug、name 与 description；description 明确区分未提供、字符串和显式 `null`，空更新与非法字段在进入 repository 前失败。
- 手写 Update/Delete SQL 并由 sqlc 生成 pgx/v5 调用；PostgreSQL adapter 将缺失资源、同 Team slug 冲突和仍被 Environment 引用的删除映射为稳定业务错误。
- 增加 `PATCH /v1/services/{service_id}` 与 `DELETE /v1/services/{service_id}`。PATCH 返回核心 Service 字段而不额外展开 Environment；DELETE 成功返回空 `204`。

## 已有验证

- catalog、PostgreSQL adapter 与 HTTP transport 的普通测试和 race 测试通过。
- 真实 PostgreSQL integration test 验证字段保留、description 显式清空、唯一冲突、not-found、引用删除冲突、成功删除与重复删除；该包的 integration race 测试通过。
- 在独立 `_test` PostgreSQL 16.14 数据库完成 migration dry-run、apply、status 后运行 `make verify`；vet、生成检查、普通测试、race、真实 integration test 与漏洞扫描全部通过，结果为 `No vulnerabilities found.`。
- 在可丢弃开发库实测 Service 创建、名称 PATCH、description 显式 `null`、被 Environment 引用时 DELETE 409、无子资源时 DELETE 204 和删除后 GET 404；响应状态、错误信封与 request ID 日志符合契约。API 经 `Ctrl-C` 优雅停止，开发与测试容器均已停止并自动删除。

## 收口审阅

- 实现保持既有模块化单体边界：domain/service 位于 `internal/catalog`，SQL/sqlc 与 pgx 错误分类留在 PostgreSQL adapter，HTTP 三态解码和 wire response 留在 catalog transport；没有跨层类型泄漏或新目录偏差。
- 本纵切面符合 M1 施工包以及 v3.1 对 Service Catalog CRUD、参数校验、统一错误、数据库约束和真实测试的要求，没有扩大到 RBAC、乐观锁或 Environment 独立 CRUD。
- Service 创建与单项 GET 使用 detail representation 并展开 Environment；PATCH 使用核心 Service representation，避免一次无必要的补查。该边界已明确写入阶段契约。
- 本地实现和验证已完成；PR #21 的 required `verify`、`smoke` 与 `atlas-community` 已在 clean runner 通过。最终文档 commit 的门禁与 merge 尚未发生，不能提前写成已合并。

## 下一步

- 提交本次 CI 证据文档并等待最新 required checks 全绿，再 squash merge。
