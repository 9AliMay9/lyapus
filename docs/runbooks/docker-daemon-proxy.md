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

## 构建时的三层代理（2026-09-11 已验证）

daemon 代理不是 Docker 的全局代理。按失败所在步骤区分作用域：

| 网络请求 | 配置入口 | 本次证据 |
| --- | --- | --- |
| 镜像拉取服务侧请求 | Docker daemon；按实际部署检查独立 containerd | daemon 已读取私有代理 |
| BuildKit 客户端侧认证令牌请求 | 启动 `docker build` 的临时代理环境 | 补充后 `load metadata` 通过，基础镜像下载完成 |
| Dockerfile `RUN go mod download` | 代理 build args 与构建网络 | 再补充后下载、编译与镜像导出全部成功 |

先用 `curl --proxy` 对失败的 HTTPS 端点验证连通性；不要输出认证令牌正文。`docker info` 的 set/unset 只能证明已设置，实际地址和 `NO_PROXY` 是否匹配需在私有终端核对，不粘贴到公开记录。代理 curl 成功不等于客户端或构建容器已使用代理。

### 本机回环代理的临时构建

前提：本次已验证的 Linux 本机 Docker、默认 `docker` driver，以及宿主机回环地址上的可用 HTTP 代理。远程 builder、其他 driver 或其他平台不能直接套用；其中的 host 指 builder 所在主机，不一定是 CLI 所在主机。

在项目根目录执行，将占位符换成私有配置中的地址：

```bash
env \
  HTTP_PROXY='http://<private-local-proxy>' \
  HTTPS_PROXY='http://<private-local-proxy>' \
  http_proxy='http://<private-local-proxy>' \
  https_proxy='http://<private-local-proxy>' \
  NO_PROXY='127.0.0.1,localhost' \
  no_proxy='127.0.0.1,localhost' \
  docker build --progress=plain \
    --network=host \
    --build-arg HTTP_PROXY \
    --build-arg HTTPS_PROXY \
    --build-arg NO_PROXY \
    -t lyapus-apiserver:m1-dev .
```

- `env` 配置本次客户端进程；无值的 `--build-arg` 从该进程环境读取同名变量，传给构建步骤。预定义代理参数不需要在 Dockerfile 添加 `ARG` 或 `ENV`。
- 默认隔离网络中，构建容器的 `127.0.0.1` 是它自己；`--network=host` 让本次构建的 `RUN` 使用宿主机网络，因此能访问宿主机回环代理。它不是代理设置本身，不能替代 build args，也不负责修复客户端认证请求。
- host 网络放宽了构建步骤对宿主机网络服务的访问边界，只用于可信构建的本地临时排障；不是 `--privileged`，也不改变最终应用容器的运行网络。不要为此给 Compose 服务添加 `network_mode: host`，或把代理监听扩大到公网。
- 不把真实代理写进 Dockerfile、Compose、CI 或 Git。预定义代理参数默认不进入 `docker history` 和缓存键；不要在 Dockerfile 显式引用或打印它们，也不要把这个机制当作通用 secret 传递方式。

预期：依赖下载、Go 编译通过，最终出现镜像命名与导出成功。2026-09-11 项目所有者提供的实际输出显示 `go mod download` 用时 99.6s、`go build` 用时 12.9s，成功生成 `lyapus-apiserver:m1-dev`。这是本地镜像构建证据，不代表容器启动、Compose、migration 或 CI 已验证。

如果失败，按上述三层记录首个失败步骤；其他 builder 若拒绝 host 网络权限，应先核对其权限策略，不自动开启广泛特权。没有修改 systemd 或 Dockerfile，无需重启服务来回滚；后续命令去掉临时参数即可。已生成镜像与缓存仍保留，不自动清理。

### 与早期 Atlas 实验的关系

早期本机 Atlas Community 实验已记录 Docker/containerd 代理与镜像拉取问题；不能据此认定 GitHub-hosted CI 当时遇到完全相同的客户端或构建容器问题。`scripts/install-atlas-community.sh` 在调用环境中执行 `git clone` 和 `go build`，产物位于 `.tools/bin/atlas`，脚本本身没有硬编码代理。`.tools` 是本地产物目录，不是代理配置真源。本次补齐的是 Docker 构建另外两层的实验证据。

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
- Docker build proxy arguments：https://docs.docker.com/build/building/variables/#proxy-arguments
- Docker build network：https://docs.docker.com/reference/cli/docker/buildx/build/#network
- BuildKit client auth provider：https://github.com/moby/buildkit/blob/master/session/auth/authprovider/authprovider.go
