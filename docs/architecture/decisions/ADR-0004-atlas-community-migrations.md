# ADR-0004：采用 Atlas Community 管理 M1 versioned migrations

- 状态：accepted
- 日期：2026-07-29
- 决策范围：M1 PostgreSQL schema、migration 目录、生成/校验与 CI

## 背景

M1 需要同时保留可审阅的 versioned SQL 与 `db/schema.sql` 期望状态，并且不能依赖 Atlas Cloud、token、试用或付费能力。候选 Atlas Community 必须先通过 P-0001；否则使用 `golang-migrate` 和手写 SQL。

## 决策

M1 采用 Atlas Community `v1.2.0`：

- `db/schema.sql` 是 Atlas diff 与 sqlc 的唯一 schema 真源；`db/migrations/` 保存不可原地修改的 versioned SQL 与 `atlas.sum`。
- `scripts/install-atlas-community.sh` 从固定源码 tag `v1.2.0` 与 commit `47daa88aea519f7f4c4aab5adfde2beab9b10b13` 构建 Community CLI 到被忽略的 `.tools/bin/atlas`；本机和 GitHub Actions 使用同一脚本。
- migration 由显式部署/运维步骤执行，`apiserver` 启动时绝不自动迁移。
- 修改 schema 时先运行 `migrate diff`，人工审阅新增 SQL 与 `atlas.sum`，再对已验证目标运行 `migrate apply` 和 `migrate status`。
- `protect-main` 要求 `Verify / atlas-community`、`Verify / verify` 与 `Verify / smoke` 都成功。

实际命令与安全边界见 [数据库 migration](../../runbooks/database-migrations.md)。

## 证据

P-0001 在 PostgreSQL 16.14 的可丢弃 `_test` 数据库上验证了：首版空库 apply、status、重复 apply、无变更 diff、索引变更的前滚，以及 `atlas.sum` 对已应用 migration 篡改的拦截。同一 `db/schema.sql` 已由固定 sqlc 1.31.1 生成 pgx/v5 代码，`go test ./...` 通过。

PR #8 的 GitHub-hosted `Verify / atlas-community` 使用同一固定 Community 构建脚本，在空 PostgreSQL 16.14 上成功执行无变更 diff、apply 与 status；无需 Cloud token。apply/status 会校验既有 `atlas.sum`，而不是重算它。

## 后果

- 获得声明式 schema 与可审阅 SQL 历史，但开发者必须理解生成 SQL、锁与迁移顺序；生成器不证明正确性。
- 已在共享历史中应用的 migration 永不修改；新变化必须新增 migration。发现历史被改动时先调查，不能直接执行 `migrate hash` 覆盖完整性证据。
- Community 不提供的 lint、drift detection、Cloud Registry 或 Pro 能力不是 M1 的承诺；数据库测试和人工审阅补足本阶段边界内的控制。
- 固定源码构建比直接下载 latest 多一步，但避免 CLI 版本漂移，也不依赖当前主机临时 `/tmp` 二进制。

## 备选方案

`golang-migrate` 仍是回退方案：若未来 Atlas Community 的固定构建、核心 diff/apply/status、SQL 可解释性或维护成本不再满足本 ADR，就新建 ADR 记录证据后切换；不原地篡改历史迁移。

## 关联

- [P-0001：Atlas Community migration 工作流](../proposals/P-0001-atlas-migration-workflow.md)
- [数据库 migration runbook](../../runbooks/database-migrations.md)
