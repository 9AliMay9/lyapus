# sqlc 生成检查与 Team 施工包终局审计

日期：2026-08-05
分支：`chore/use-sqlc-diff`

## 触发原因

Team CRUD/分页施工中，`make generate-check` 只有在先暂存生成代码后才能通过。这与项目“先完成生成和验证，再构造候选提交”的常规顺序冲突，也暴露出施工包契约只写“检查工作树无差异”、没有说明 Git index 前置条件的问题。

## 根因与决策

- 原目标以 `generate-check: generate` 先更新磁盘，再通过 `git diff --quiet` 比较 worktree 与 index；因此正确但未暂存的生成代码也会被判为失败。
- Git 检查同时混合了生成新鲜度与候选提交范围，且不能单独证明输入 SQL 和输出代码都已进入同一候选提交。
- 固定版本 sqlc 1.31.1 原生提供 `sqlc diff`，能够比较当前 schema/query 的预期生成结果与磁盘代码。
- `make generate-check` 改用 `sqlc diff --no-remote`，继续固定版本、保持 Community/本地边界，并从 Git 暂存状态中解耦。

## 验证

- 在未暂存 Makefile 修改的工作区运行 `make generate-check`，命令无输出并成功退出。
- 随后的 `git status --short` 只显示 Makefile 修改，没有生成文件被改写。
- PR 合入前仍由 required `verify` 在 clean checkout 中执行同一目标，验证它与完整门禁组合后的行为。

## 一致性审计

- 对照原始 v3.1 方案书，Team repository CRUD/分页符合 M1 的 PostgreSQL CRUD、分页与测试方向，但不能替代仍未完成的 Service/Environment、事务/并发、Compose 和查询优化证据。
- 活跃架构、项目上下文、阶段入口、当前进度、M1 计划/清单/结果已同步 PR #13 与 `f4df01f` 的事实。
- 公开 HTTP cursor 编解码和 1–100 limit 契约保持未完成；repository 内部结构化 cursor 不能被误写成公共 API 已完成。
- 旧会话中的 Team Create/Get 描述保留为当时历史，不回写成今天的状态。

## 下一步

- 通过本 PR 的 required checks 并合入后，从 `main` 开始 Team 业务服务层施工包。
