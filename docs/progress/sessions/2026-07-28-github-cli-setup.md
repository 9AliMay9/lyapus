# 工作会话：2026-07-28 - GitHub CLI 协作工具

## 目标

在远程开发主机安装官方 GitHub CLI，使 PR、required checks 和 GitHub Actions 日志可在终端查看，同时保持现有 SSH Git 推送方式不变。

## 实际完成

- 验证官方 GitHub CLI APT key 的 PGP 指纹后，安装官方仓库提供的 `gh`。
- 安装版本：`gh 2.96.0`。
- 使用浏览器一次性验证码完成 GitHub CLI API 登录；Git protocol 继续为 SSH，未上传新 SSH key，也未执行 `gh auth setup-git`。
- `gh auth status`、`gh pr status` 和 `gh run list --limit 3` 均已成功运行。
- 更新 GitHub runbook，记录受保护分支下的 CLI 使用边界与常用命令。

## 验证证据

```text
命令：gh --version
结果：gh version 2.96.0

命令：gh auth status
结果：github.com 登录成功；Git operations protocol 为 ssh；CLI token 已遮蔽显示

命令：gh pr status
结果：当前 main 没有关联 PR，账户没有待处理 PR

命令：gh run list --limit 3
结果：可读取近期 Verify workflow 运行状态
```

## 安全与边界

- `gh` 是远程开发主机的可选协作工具，不是仓库依赖或应用运行时组件。
- API token 不进入 Git、`.env`、shell 配置、聊天记录或项目文档；`gh auth token` 不作为日常命令。
- Git 继续走现有 SSH remote；CLI 的 API token 不替代 SSH key。
- 如不再需要 CLI，先执行 `gh auth logout`，再在 GitHub 账号设置中撤销 GitHub CLI 授权。

## 下一次从这里继续

- 具体文件：`docs/progress/current.md`、`docs/stages/m1-go-backend/README.md`。
- 具体任务：建立 M1 最小版施工包；需要创建 PR 或检查 CI 时查阅 `docs/runbooks/github-ssh.md`。
- 验收命令：按 M1 施工包确定；GitHub 协作可用 `gh pr checks --required --watch`。
- 不要做：不要把 CLI token 当作项目配置，也不要用 `gh auth setup-git` 改写已验证的 SSH Git 流程。
