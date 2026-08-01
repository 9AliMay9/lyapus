# PostgreSQL integration tests

## 用途与边界

本 runbook 用真实 PostgreSQL 验证 repository 的 SQL、sqlc adapter、数据库约束与错误分类。它只适用于可丢弃的测试数据库；测试代码会拒绝数据库名不以 `_test` 结尾的 URL，但操作者仍须在运行前确认目标。

应用启动不执行 migration；测试数据库必须先按 [数据库 migration](database-migrations.md) 显式建立。一次性本地容器是当前开发验证路径，不替代尚未完成的 Compose 空环境交付。

## 本地运行

1. 创建或确认一个隔离的 PostgreSQL 16.14 数据库，名称以 `_test` 结尾。
2. 将目标 URL 设置在当前 shell 的私有环境中；不要把真实 URL、密码或网络地址写入 Git、文档或命令记录。
3. 对该数据库执行 Atlas `migrate apply` 和 `migrate status`，确认无 pending migration。
4. 在同一个 shell 执行：

```bash
LYAPUS_TEST_DATABASE_URL="$LYAPUS_TEST_DATABASE_URL" make integration
```

预期 `go test -tags=integration -count=1 ./...` 实际运行。若变量缺失，`make integration` 必须以明确错误退出；若数据库名不以 `_test` 结尾，测试辅助代码必须拒绝执行。

## 清理与失败处理

- 测试开始时会通过 `TRUNCATE ... RESTART IDENTITY CASCADE` 清理 catalog 表；只可对已通过 `_test` 保护的可丢弃数据库运行。
- migration 失败时先运行 `migrate status`，检查 migration 历史与 `atlas.sum`；不要修改已应用 migration 来“修复”测试库。
- 测试结束后停止一次性容器会自动删除其数据。对持久测试数据库的清理必须先确认名称与连接目标。

## CI 边界

required `verify` job 为 integration tests 提供独立 PostgreSQL service，并在运行 `make verify` 前显式 apply/status migration。该 CI 路径与 smoke、Atlas migration job 使用不同数据库和 runner，不共享容器或数据。
