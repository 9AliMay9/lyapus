# PostgreSQL integration tests

## 用途与边界

本 runbook 用真实 PostgreSQL 验证 repository 的 SQL、sqlc adapter、数据库约束与错误分类。它只适用于可丢弃的测试数据库；测试代码会拒绝数据库名不以 `_test` 结尾的 URL，但操作者仍须在运行前确认目标。

应用启动不执行 migration；测试数据库必须先按 [数据库 migration](database-migrations.md) 显式建立。独立 Compose project 的本地已验证流程见 [Compose 交付](compose-delivery.md)。

## 本地运行

1. 创建或确认一个隔离的 PostgreSQL 16.14 数据库，名称以 `_test` 结尾。
2. 将目标 URL 设置在当前 shell 的私有环境中；不要把真实 URL、密码或网络地址写入 Git、文档或命令记录。
3. 对该数据库执行 Atlas `migrate apply` 和 `migrate status`，确认无 pending migration。
4. 在同一个 shell 执行：

```bash
LYAPUS_TEST_DATABASE_URL="$LYAPUS_TEST_DATABASE_URL" make integration
```

预期 `go test -tags=integration -count=1 ./...` 实际运行。若变量缺失，`make integration` 必须以明确错误退出；若数据库名不以 `_test` 结尾，测试辅助代码必须拒绝执行。

## 约束测试与整体回归

新增 `constraints_integration_test.go` 直接执行 SQL，检查三表字段 CHECK 插入边界、唯一性范围、外键和引用删除。代码手敲完成后先静态复查，修正并复查通过后再运行。已配置上述独立测试库时，可依次执行：

```bash
go test -tags=integration -count=1 -v -run '^TestCatalogConstraintsIntegration' ./internal/catalog/postgres
go test -tags=integration -race -count=1 ./internal/catalog/postgres
make verify
```

当前约束矩阵应运行 17 个函数、100 个子测试；`no tests to run` 不算通过。后两条命令用于施工包整体回归，不必在每次手敲一个函数后重复运行。`make verify` 的普通 race 不含 integration 标签，因此整体回归单独执行 integration race。不要同时启动这些命令，它们共享会被清空的测试表。

本地通过记录及证据边界见 [约束测试收口复盘](../progress/sessions/2026-09-10-catalog-db-constraints.md)。

## 清理与失败处理

- 测试开始时会通过 `TRUNCATE ... RESTART IDENTITY CASCADE` 清理 catalog 表；只可对已通过 `_test` 保护的可丢弃数据库运行。
- migration 失败时先运行 `migrate status`，检查 migration 历史与 `atlas.sum`；不要修改已应用 migration 来“修复”测试库。
- 只有使用 `--rm` 且没有持久挂载的一次性容器才会在停止后自动删除其数据。Compose 命名卷不随容器停止自动删除；删除测试卷前必须确认 project、卷标签与目标。开发库也以 `_test` 结尾，后缀检查不能防止误连开发库。

## CI 边界

required `verify` job 为 integration tests 提供独立 PostgreSQL service，并在运行 `make verify` 前显式 apply/status migration。该 CI 路径与 smoke、Atlas migration job 使用不同数据库和 runner，不共享容器或数据。
