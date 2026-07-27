# GitHub SSH 推送

## 用途

在新的 SSH 终端中恢复本项目的 GitHub SSH 身份，然后手动推送已审核的 Git commit。

## 前置条件

- 当前用户拥有私钥 `~/.ssh/id_ed25519_github_lyapus`，其对应公钥已添加到 GitHub 账号。
- 仓库远程地址使用 SSH，例如 `git@github.com:9AliMay9/lyapus.git`。
- 待推送内容已完成项目规定的隐私检查与 Git 复核。

私钥和 passphrase 都不得复制、打印、提交或发送到聊天记录。

## 在新终端恢复 SSH agent

```bash
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/id_ed25519_github_lyapus
ssh -T git@github.com
```

`ssh-add` 会要求输入私钥 passphrase，输入时终端不会回显。验证成功时，GitHub 会显示已认证的账户名，并说明不提供 shell access；该命令的退出状态不必为零。

## 推送分支

```bash
git status --short
git remote -v
git push
```

推送前，`git status --short` 应为空；`git remote -v` 应显示预期的 GitHub SSH 地址。首次向某个远程分支推送时，使用：

```bash
git push -u origin <branch>
```

不要在受保护的 `main` 上直接开发或推送。先从最新 `main` 创建短生命周期分支，在该分支完成校验、提交并推送。

## 受保护 main 的 Pull Request 流程

`protect-main` ruleset 要求变更经 Pull Request 合并，并要求目标分支最新、`verify` 和 `smoke` 都通过。推送分支后：

1. 在 GitHub 页面创建以 `main` 为 base 的 Pull Request。
2. 确认 Files changed 只包含本次预期范围；不要把临时探针、凭据或无关格式化混入。
3. 等待 `verify` 和 `smoke` 成功。若失败，修复同一分支后再次推送；不要绕过 required checks。
4. 选择适合本次变更的合并方式；短生命周期、多个临时 commit 的分支通常使用 squash merge。
5. 合并后在本地执行：

```bash
git switch main
git pull --ff-only
git status --short
```

确认工作区干净后，再删除已完成的本地与远程分支。若是 squash merge，本地 Git 可能不会把原分支识别为已合并；确认 PR 已合并且分支只含该任务后，才使用 `git branch -D <branch>` 清理本地引用。

## 常见失败处理

`Permission denied (publickey)` 通常表示当前终端没有启动 agent、私钥未加入 agent，或 GitHub 账号尚未添加对应公钥。重新执行“恢复 SSH agent”中的三条命令；不要改用明文密码、上传私钥，或关闭 SSH 主机验证。若网页已删除远程分支而本地仍显示 `remotes/origin/<branch>`，恢复 SSH 身份后执行 `git fetch --prune` 清理过期远程引用。

## 会话结束与密钥处置

新的 SSH 终端不会自动继承旧终端的 `SSH_AUTH_SOCK`，因此下次推送前重复本 runbook。若私钥疑似泄露，立即从 GitHub 删除对应公钥、停止使用该密钥并生成新密钥对。
