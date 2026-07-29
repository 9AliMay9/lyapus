# 交互 shell 的临时代理下载

## 用途

在本开发主机上，Go module、`curl` 或其他 HTTPS 下载需要经本机私有代理访问上游时，为**单条命令**设置代理环境。真实代理 URL、端口和隧道拓扑属于主机私有配置，不能写入仓库。

这不是 Docker daemon 或 Git SSH 的配置手册：它们分别见 [Docker daemon 代理与镜像拉取](docker-daemon-proxy.md) 与 [GitHub SSH 推送](github-ssh.md)。

## 安全使用方式

将占位符替换为当前主机私有配置中的代理地址；变量只对子进程及其子进程生效，不会污染当前 shell：

```bash
env \
  HTTP_PROXY='http://<private-local-proxy>' \
  HTTPS_PROXY='http://<private-local-proxy>' \
  http_proxy='http://<private-local-proxy>' \
  https_proxy='http://<private-local-proxy>' \
  NO_PROXY='127.0.0.1,localhost' \
  no_proxy='127.0.0.1,localhost' \
  <network-command>
```

例如，固定 Go 工具版本时将 `<network-command>` 替换为 `go install <module>@<version>`。版本查询、读取本地文件和已缓存工具运行通常不需要这段前缀。

不要把这些变量永久 `export` 到 `~/.bashrc`，也不要把真实值粘贴到 Git、终端录屏、issue 或会话文档。临时前缀能缩小凭据/网络配置作用域，也让“某个命令是否依赖代理”更可观察。

## 排障边界

- `go install`、`curl`、包管理器或其他由当前 shell 启动的 HTTPS 下载：使用本 runbook 的临时环境。
- `docker pull`：由常驻的 Docker/containerd systemd 服务发起；shell 前缀不会配置它们，应使用 Docker daemon runbook。
- `git@github.com:...`：使用 SSH 身份与 ssh-agent；HTTP 代理变量不能替代 SSH key。

若仍失败，记录脱敏错误并区分 DNS、连接本地代理、CONNECT/TLS、上游认证和实际依赖版本，不要同时改动无关网络配置。
