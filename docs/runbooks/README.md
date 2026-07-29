# Runbook

每份 runbook 是经过验证的操作手册，至少包含前置条件、风险、命令、预期结果、失败处理和回滚步骤。未实际验证的命令只能放在阶段计划中，不得标为 runbook。

- [GitHub SSH 推送](github-ssh.md)：在新的 SSH 终端中恢复 GitHub 身份并安全推送。
- [交互 shell 的临时代理下载](temporary-proxy-downloads.md)：对单条 Go/curl 等下载命令使用脱敏的私有代理前缀，并区分 Docker 与 SSH 路径。
- [Docker daemon 代理与镜像拉取](docker-daemon-proxy.md)：脱敏配置 Docker/containerd 的私有代理、验证镜像拉取并安全回滚。
