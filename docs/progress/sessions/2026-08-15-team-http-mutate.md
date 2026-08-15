# Team HTTP PATCH 与 DELETE

日期：2026-08-15

## 范围

- 将既有 TeamService 的 `UpdateTeam`、`DeleteTeam` 接入 chi 的 `PATCH /v1/teams/{team_id}` 与 `DELETE /v1/teams/{team_id}`。
- PATCH 使用字段出现状态表示可选字符串：未提供字段不更新；空字符串交由业务层校验；`null`、未知字段、类型不匹配和多个 JSON 值返回统一 `400 invalid_argument`。
- 扩展 clean-runner smoke 为创建、读取、列表、更新、唯一冲突、删除和删除后 404 的单条 Team HTTP CRUD 链路。

## 验证

- catalog HTTP 单元测试、全量 `go test ./...` 与 `go test -race ./...` 通过。
- 可丢弃 PostgreSQL 16.14 开发库完成 Atlas dry-run、apply、status；实测 Team 创建 201、PATCH 200、空 PATCH 400、唯一 slug 冲突 409、DELETE 空 204 与删除后 GET 404。响应 request ID 与完成日志一致；API 进程和容器均已停止。
- 独立、名称以 `_test` 结尾的 PostgreSQL 16.14 数据库完成 dry-run、apply、status；`make verify` 通过 vet、普通测试、race、真实 integration test 与漏洞扫描。

## 边界与后续

- 本施工包不改动 Team schema、migration、sqlc 查询、repository 或 service 的既有 Update/Delete 语义，只补齐 HTTP transport 与验证链路。
- PATCH 与 DELETE 的更新 smoke 尚待本施工包 PR 的 clean-runner 运行确认；在成功前不将该 CI 证据写为已完成。
- Team 已形成完整的 HTTP CRUD 模板；下一步按同一分层实现 Service，并在其创建路径集中处理“Service + 初始 Environment”的事务原子性与并发唯一冲突。
