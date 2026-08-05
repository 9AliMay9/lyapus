# Team repository CRUD 与分页

日期：2026-08-05
分支：`feat/team-crud-pagination`

## 完成事实

- 扩展 `db/queries/teams.sql`，由 sqlc 生成 Team 的 List、Update、Delete 调用。
- PostgreSQL repository 实现 Team List/Update/Delete；列表使用 `created_at DESC, id DESC` 的复合稳定顺序，并以多取一行生成下一页游标。
- domain 明确 `TeamCursor`、分页输入/结果、更新输入及 `ErrInvalidArgument`；pgx/sqlc 类型继续限制在 PostgreSQL adapter 内。
- 将 `23505`（唯一冲突）与 `23503`（外键冲突）映射为 `ErrConflict`，将 `pgx.ErrNoRows` 映射为 `ErrNotFound`。
- 扩展单元测试与真实 PostgreSQL integration test，覆盖 limit 边界、分页、更新、删除、唯一冲突、外键引用删除冲突和 not-found。

## 验证证据

- 在可丢弃的 `lyapus_team_crud_pagination_test` PostgreSQL 16.14 容器中，对两份既有 migration 完成 dry-run、apply 和 status；状态为 latest、2 个 migration 已执行。
- `make generate-check`、`go test ./...`、`go test -race ./...` 均通过。
- 设置隔离的 `LYAPUS_TEST_DATABASE_URL` 后，`make integration` 通过。
- 2026-08-05 最终 `make verify` 通过，包含 `go vet`、生成检查、普通测试、race、真实 PostgreSQL integration test 与漏洞扫描；测试容器随后停止并由 `--rm` 删除。

## 过程经验

- `make generate-check` 当前是候选提交检查：它先执行生成，再要求 `sqlcgen` 没有未跟踪文件且工作区与 Git 暂存区一致。变更查询后应先 `make generate`，再显式暂存生成文件后运行该检查；其命名和是否拆分为纯新鲜度检查留待独立工作流决策，不在本施工包顺带修改。
- repository 层分页使用结构化 cursor；对外 HTTP 的不透明、URL-safe cursor 编解码属于后续 transport 契约，不能提前宣称完成。

## 下一步

- 实现 Team 业务服务层校验与边界，再将其接入 chi HTTP transport。
