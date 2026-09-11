# Runbook

- [Compose 本地交付](compose-delivery.md)：本地新卷迁移、API、故障恢复、测试隔离和清理；明确代理、数据与 CI 边界。

每份 runbook 是经过验证的操作手册，至少包含前置条件、风险、命令、预期结果、失败处理和回滚步骤。未实际验证的命令只能放在阶段计划中，不得标为 runbook。

- [GitHub SSH 推送](github-ssh.md)：在新的 SSH 终端中恢复 GitHub 身份并安全推送。
- [交互 shell 的临时代理下载](temporary-proxy-downloads.md)：对单条 Go/curl 等下载命令使用脱敏的私有代理前缀，并区分 Docker 与 SSH 路径。
- [Docker daemon 代理与镜像拉取](docker-daemon-proxy.md)：脱敏配置 Docker/containerd 代理；区分构建客户端认证、镜像拉取与 RUN 下载，记录临时 host 网络构建及其边界。
- [数据库 migration](database-migrations.md)：用固定 Atlas Community 生成、审阅、apply、核验 versioned PostgreSQL migration。
- [PostgreSQL integration tests](postgresql-integration-tests.md)：在受保护的可丢弃 `_test` 数据库上显式迁移并运行真实 repository integration tests。
