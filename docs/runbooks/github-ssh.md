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

## 推送

```bash
git status --short
git remote -v
git push
```

推送前，`git status --short` 应为空；`git remote -v` 应显示预期的 GitHub SSH 地址。首次向某个远程分支推送时，按 Git 输出使用 `git push -u origin <branch>`。

## 常见失败处理

`Permission denied (publickey)` 通常表示当前终端没有启动 agent、私钥未加入 agent，或 GitHub 账号尚未添加对应公钥。重新执行“恢复 SSH agent”中的三条命令；不要改用明文密码、上传私钥，或关闭 SSH 主机验证。

## 会话结束与密钥处置

新的 SSH 终端不会自动继承旧终端的 `SSH_AUTH_SOCK`，因此下次推送前重复本 runbook。若私钥疑似泄露，立即从 GitHub 删除对应公钥、停止使用该密钥并生成新密钥对。
