# 生成代码检查为什么不应依赖 Git 暂存区

## 一句话模型

生成器负责从真源产生派生代码，生成检查负责比较“预期输出”和“磁盘输出”，Git 暂存区只负责描述下一次候选提交；三者不应互相冒充。

## 必须掌握

Git 中有三个相关状态：

- `HEAD` 是上一次提交的快照。
- index（暂存区）是下一次候选提交的快照。
- worktree（工作区）是磁盘上的当前文件。

`git diff` 默认比较 worktree 与 index，`git diff --cached` 比较 index 与 `HEAD`。tracked 表示 Git 已知道该路径；对 tracked 文件执行 `git add` 是更新 index 中的内容，不是第一次“开启跟踪”。

生成代码同样有三个独立动作：

1. `make generate` 根据 schema/query 改写磁盘生成代码。
2. `make generate-check` 验证磁盘生成代码是否等于当前真源的预期结果。
3. `git add` 把已经审阅和验证的源码、真源与生成代码纳入候选提交。

## 容易混淆或踩坑

原实现以 `generate-check: generate` 先改写磁盘，再用 `git diff --quiet` 比较 worktree 与 index。因此，新生成但尚未暂存的正确代码也会失败；暂存后又会通过。这能在干净 CI checkout 中发现漂移，但把“生成是否新鲜”和“是否进入候选提交”混成了一件事。

这种检查还不能单独证明候选提交内部一致：输入 SQL 可能仍在工作区，而输出代码已经进入 index。最终提交范围仍必须用 `git status`、`git diff --cached --check` 和 `git diff --cached --stat` 审阅。

## 在本项目中的落点

项目固定 sqlc 1.31.1，并提交 `internal/catalog/postgres/sqlcgen`。该版本原生提供 `sqlc diff`，可比较当前 schema/query 的预期生成结果与磁盘文件。因此：

```make
generate-check:
	@# 省略固定版本与安装检查
	"$$sqlc" diff --no-remote
```

`--no-remote` 明确保持 Community、本地、无 Cloud token 的验证边界。`make verify` 可以在 `git add` 之前运行，CI 继续在 clean checkout 中执行同一命令。

## 最小验证实验

在生成代码与真源一致时：

```bash
make generate-check
```

命令应无 diff 并成功退出，且 `git status --short` 不应出现生成文件改动。修改 query 后若尚未执行 `make generate`，同一检查应输出预期与磁盘代码的差异并失败；执行生成后再次检查应通过。

## 面试表达

“项目提交 sqlc 生成代码，但生成新鲜度不依赖 Git 暂存状态。CI 和本地都使用固定版本的 `sqlc diff` 比较预期输出与磁盘结果；Git index 只承担候选提交审阅。这样既能发现漏生成和手改生成文件，又允许在暂存前完成完整验证。”

## 延伸阅读

- sqlc：Using sqlc in CI/CD：https://docs.sqlc.dev/en/stable/howto/ci-cd.html
- Go Blog：Generating code：https://go.dev/blog/generate
