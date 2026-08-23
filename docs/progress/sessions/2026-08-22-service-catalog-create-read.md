# Service Catalog 创建与读取纵切面

日期：2026-08-22

## 范围

- 在 `internal/catalog` 增加 Service、Environment、ServiceDetail、Service 游标/分页输入输出，以及只依赖 domain repository 接口的 ServiceService。
- 手写 `db/queries/services.sql`，由 sqlc 生成 pgx/v5 调用；在 PostgreSQL adapter 中实现 Service Create/Get/List、`team_id` 过滤和稳定游标分页。
- `CreateService` 使用显式 pgx 事务创建 Service 与可选初始 Environment；缺失父 Team、唯一冲突和事务中途失败映射为稳定业务错误。
- 增加 `POST /v1/services`、`GET /v1/services/{service_id}` 与 `GET /v1/services`。单项响应展开 Environment，集合响应不展开；列表支持 `team_id`、`limit` 与不透明 cursor。
- 将 catalog 写请求的媒体类型边界固定为 `application/json`，不合规请求返回 `415 unsupported_media_type`。

## 验证

- `go test ./...` 与 `go test -race ./...` 通过。
- 真实 PostgreSQL integration test 覆盖创建/回读、初始 Environment、事务回滚、父 Team 缺失、唯一冲突、全局及 Team 过滤分页，以及并发重复创建时恰好一个成功、一个冲突。
- 可丢弃开发库完成 Atlas dry-run、apply、status；真实 HTTP 实测 Team 创建、Service 创建、Service 单项读取和两页列表，响应形状、Environment 展开边界、排序、cursor 与 request ID 符合契约。API 经 `Ctrl-C` 优雅关闭，容器已停止并自动删除。
- 独立 `_test` 库完成 Atlas dry-run、apply、status；`make verify` 的 vet、生成检查、普通测试、race、真实 integration test 与漏洞扫描全部通过，结果为 `No vulnerabilities found.`。测试容器已停止并自动删除。

## 与施工包和原始方案的对齐

- 当前目录保持模块化单体边界：domain/service 在 `internal/catalog`，PostgreSQL/sqlc adapter 在 `internal/catalog/postgres`，HTTP DTO/cursor/handler 在 `internal/catalog/transport/http`，`cmd/apiserver` 只装配依赖。它是 v3.1 目录职责示意的具体化，没有引入跨层类型泄漏或提前抽象通用 storage。
- Service、Environment 归属模型、Service CRUD 方向、事务、并发正确性、PostgreSQL、分页过滤和测试都来自 M1 原始范围。本施工包先完成 Service Create/Get/List 和“Service + 初始 Environment”，属于既定施工顺序的一个纵切面，不删除后续 mutation 或 Environment 范围。
- M1 仍未完成：Service PATCH/Delete、Environment 独立 CRUD、Compose 空环境、查询计划实验、README 五分钟演示与最终 release 均无完整证据。

## 合并前剩余

- 扩展 GitHub Actions smoke：创建父 Team，创建带初始 Environment 的 Service，再验证 Service 单项读取及带 `team_id`、`limit=1` 的列表读取；保留既有 Team CRUD smoke。
- 提交并创建 PR，等待 required `verify`、`smoke` 与 `atlas-community` 在 clean runner 全部通过，再执行合并前收口复核。
