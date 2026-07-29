# Docker daemon 代理与镜像拉取

## 用途

当交互 shell 能下载 Go 依赖或通过代理访问上游，但 `docker pull` 在 Docker Hub 认证、TLS 握手或镜像层下载处超时/EOF 时，检查 Docker daemon 与其独立的 containerd 服务是否都获得了主机私有代理配置。

本 runbook 已在 M1 Atlas Community 实验中验证。它只描述公开、可复现的操作结构；真实代理 URL、端口、隧道和环境变量值只保留在主机私有配置中。

## 前置条件与风险

- 已有可用的私有 HTTP/HTTPS 代理；本 runbook 不负责建立隧道或 VPN。
- Docker 和 containerd 使用 systemd 管理；先确认实际 unit 名称。
- 重启 `containerd.service` 与 `docker.service` 会中断正在运行的容器。执行前必须检查并协调工作负载。
- 不把代理值写入仓库、终端录屏、会话记录或 issue。

## 诊断

确认服务状态，并以脱敏方式确认 Docker daemon 是否读取代理：

```bash
systemctl is-active docker.service
systemctl is-active containerd.service
docker info --format 'http_proxy={{if .HTTPProxy}}set{{else}}unset{{end}} https_proxy={{if .HTTPSProxy}}set{{else}}unset{{end}} no_proxy={{if .NoProxy}}set{{else}}unset{{end}}'
```

检查是否有需要中断的容器：

```bash
docker ps --format '{{.ID}} {{.Names}} {{.Status}}'
```

如果 Docker 已有代理而 `containerd.service` 没有，镜像拉取仍可能在 registry token 或 layer 阶段失败。不要仅根据 shell 的 `HTTP_PROXY` 推断 daemon 已配置。

## 配置

### 用正确的编辑入口

`systemctl edit` 不是直接编辑 unit 原文件：它会创建或更新该 unit 的 systemd drop-in，因此应使用它保存 Docker/containerd 的本机覆盖配置，而不是手工改发行版提供的 unit 文件。若当前用户的 `~/.bashrc` 已将 `EDITOR`、`VISUAL` 和 `SUDO_EDITOR` 设为 nvim，普通受保护文件可用 `sudoedit /etc/…` 编辑；但 `systemctl edit` 应显式设置它读取的 `SYSTEMD_EDITOR`：

```bash
sudo env SYSTEMD_EDITOR="$HOME/.local/bin/nvim" systemctl edit docker.service
```

这只影响本次编辑，不让 root 或 systemd 的全局默认编辑器依赖某个用户目录。未安装用户级 nvim 时，改用已验证存在的编辑器路径。

为 Docker daemon 建立私有 drop-in：

```bash
sudo env SYSTEMD_EDITOR="$HOME/.local/bin/nvim" systemctl edit docker.service
```

填入下列结构，并将占位符替换为只存在于主机私有配置中的代理 URL：

```ini
[Service]
Environment="HTTP_PROXY=<private-http-proxy-url>"
Environment="HTTPS_PROXY=<private-http-proxy-url>"
Environment="NO_PROXY=127.0.0.1,localhost"
```

若 `containerd.service` 是独立、活跃的服务，也为它建立相同的私有 drop-in：

```bash
sudo env SYSTEMD_EDITOR="$HOME/.local/bin/nvim" systemctl edit containerd.service
```

保存后重载并按依赖顺序重启。仅在已确认没有需保留的容器时执行：

```bash
sudo systemctl daemon-reload
sudo systemctl restart containerd.service
sudo systemctl restart docker.service
```

## 验证

确认镜像能从固定 tag 拉取，并记录镜像摘要而不是私有网络细节：

```bash
docker pull postgres:16.14
docker image inspect postgres:16.14 --format '{{.Id}}'
```

M1 实验实际得到的 PostgreSQL 16.14 manifest digest 是：

```text
sha256:33f923b05f64ca54ac4401c01126a6b92afe839a0aa0a52bc5aeb5cc958e5f20
```

若仍失败，先将错误归类为 DNS、proxy CONNECT/TLS、registry token、镜像层或认证问题；不通过反复更改无关 Docker 参数掩盖根因。

## 回滚

不再需要私有代理且确认没有依赖它的 Docker 工作负载时，删除各自的 systemd drop-in 并重启服务：

```bash
sudo systemctl revert docker.service
sudo systemctl revert containerd.service
sudo systemctl daemon-reload
sudo systemctl restart containerd.service
sudo systemctl restart docker.service
```

回滚同样会中断容器；它只删除 systemd override，不会删除镜像、卷或容器数据。

## 依据

- Docker daemon proxy configuration：https://docs.docker.com/engine/daemon/proxy/
