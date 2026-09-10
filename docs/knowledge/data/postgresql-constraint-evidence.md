# PostgreSQL 约束的直接行为证据

## 一句话模型

Go 校验让错误更早、更清晰地返回；直接 SQL 测试证明数据到达 PostgreSQL 后，数据库仍会执行约定的约束。

## 必须掌握

- 正向用例证明合法输入能进入数据库，反向用例证明指定非法输入被拒绝。只检查任意错误会把 SQL 拼写、参数数量或缺失表误判为约束生效。
- 本项目通过 `errors.As` 取得 `*pgconn.PgError`，同时断言 SQLSTATE 和约束名：CHECK 为 `23514`，unique 为 `23505`，foreign key 为 `23503`。
- 联合唯一性需要同时验证“同父资源重复失败”和“不同父资源同 slug 成功”，才能说明唯一范围。
- 外键有两个方向的使用场景：子行不能引用缺失父行；仍被引用的父行不能删除。本项目还检查删除失败后父子行均保留。
- 中文 100/101、500/501 字符用例验证字符长度边界，不能用字节长度替代；时间用固定 UTC 值和微秒差值测试相等边界。

## 容易混淆或踩坑

- `gofmt` 不检查 SQL 字符串，Go 编译也不校验其中的表名与占位符数量。
- `-run` 拼错可能输出 `PASS [no tests to run]`，必须确认预期测试实际运行。
- Atlas 的 `--url` 不会设置 Go 测试需要的 `LYAPUS_TEST_DATABASE_URL`；环境变量需在运行测试的 shell 设置。
- 复查顺序：手敲完成、静态复查与修正、运行验证、下一段。静态一致性检查也要审查助手示例本身。
- 共享测试库使用 TRUNCATE 重置，不能并行运行多个清表测试进程，也不使用 `t.Parallel`。`_test` 后缀是保护条件，操作者仍需确认数据库可丢弃。
- 当前 name 矩阵覆盖普通空格，不声称穷尽所有 Unicode 空白；CHECK 插入边界也不等于穷尽 NOT NULL 或 UPDATE 行为。

## 在本项目中的落点

`internal/catalog/postgres/constraints_integration_test.go` 使用真实迁移后的 PostgreSQL，与 repository 集成测试分工：前者断言数据库原始错误，后者验证 adapter 映射、事务和业务读写。生产层不因测试增加依赖或改变接口。

## 最小验证实验

先按 [integration runbook](../../runbooks/postgresql-integration-tests.md) 准备独立已迁移的测试库并配置环境变量；手敲变更先静态复查。再定向运行对应测试，整体完成后运行带 integration 标签的 race 和 `make verify`。

本次 17 个函数、100 个子测试的覆盖与运行证据见 [收口记录](../../progress/sessions/2026-09-10-catalog-db-constraints.md)。本地通过与 CI/合并状态分开记录。

## 面试表达

我用绕过应用校验的 SQL 用例验证约束，既检查错误码和约束名，也检查成功边界及拒绝删除后的数据状态。这样可以区分数据库约束、应用错误映射和测试自身的 SQL 错误，并明确哪些场景已经验证。

## 延伸阅读

- [M1 契约](../../stages/m1-go-backend/contracts.md)
- [Atlas migration 决策](../../architecture/decisions/ADR-0004-atlas-community-migrations.md)
