# 工作会话：2026-07-29 - Atlas Community migration 工作流

## 目标

执行 P-0001，决定 M1 是否采用免费、固定版本、可审阅且 CI 可复现的 Atlas Community migration 工作流。

## 实际完成

- 用 PostgreSQL 16.14 可丢弃 `_test` 数据库验证空库 apply、status、重复 apply、无变更 diff，以及索引前滚 migration。
- 生成并审阅两份 versioned SQL 与 `atlas.sum`；故意编辑首份已应用 migration 后，Atlas 以 checksum mismatch 拒绝状态检查，恢复文件后状态正常。
- `db/schema.sql` 同时被 sqlc 1.31.1 解析，生成 pgx/v5 代码；`make generate-check` 纳入 `make verify`，并在 CI 安装固定 sqlc。
- 固定 Atlas Community v1.2.0 源码 tag/commit 的项目脚本已在本机成功构建；PR #8 的 `Verify / atlas-community` 在 GitHub-hosted runner 成功执行 diff、空库 apply 和 status，后两者校验既有 migration 完整性。
- `protect-main` 新增 `Verify / atlas-community` required check；现有 `verify`、`smoke` 仍保留。
- 发现间接 `golang.org/x/text v0.29.0` 的不可达漏洞公告（GO-2026-5970），升级到修复版 v0.39.0 后完整验证返回 `No vulnerabilities found`。

## 决策

采用 Atlas Community v1.2.0，见 [ADR-0004](../../architecture/decisions/ADR-0004-atlas-community-migrations.md)。不使用 Cloud/token/Pro；若未来固定构建或核心路径不再满足要求，按 ADR 记录证据后回退 `golang-migrate`。

## 下一次从这里继续

按 M1 施工包进入数据库基础设施：配置 `LYAPUS_DATABASE_URL`、建立 `pgxpool`，并让 `/readyz` 使用有超时的 Ping。不要在应用启动路径自动执行 migration。
