# 依赖下载为何会在 Docker daemon 失败

## 一句话模型

网络配置属于进程，不属于“整台机器”。交互 shell、Docker daemon、containerd 是不同进程环境；shell 能联网，不等于由 daemon 发起的镜像拉取也能联网。

## 必须掌握

- `HTTP_PROXY`/`HTTPS_PROXY` 只影响继承它们的进程。长期运行的 systemd 服务通常不会继承 SSH、tmux 或当前 shell 的环境。
- Docker Engine 的镜像链路可能经过独立的 `containerd.service`。Docker daemon 已显示代理配置，不足以证明 containerd 也拥有它。
- 把问题拆成可观察层：DNS、到代理端口的连接、proxy CONNECT/TLS、registry token、manifest、镜像 layer、应用自身下载。
- 同一上游可以用不同客户端验证。M1 实验中 curl 与 Go `net/http` 经私有代理可获得 Docker Hub token，而 Docker/containerd 拉取失败；检查发现 containerd 没有代理环境。

## 容易混淆或踩坑

- SSH agent 只处理 Git SSH 身份，不会修复 curl、Go module 或 Docker Registry 的 HTTPS 连接。
- 为当前 shell 临时 `export` 代理，不能配置已启动的 Docker/containerd service。
- Docker Hub token 的 EOF、超时不等于镜像 tag、Atlas 或 PostgreSQL 本身有问题。
- 真实代理 URL、端口、隧道拓扑和 VPN 配置属于私有运维信息，不能进入 Git。
- 重启 Docker/containerd 前必须先看 `docker ps`；网络修复可能中断工作负载。

## 在本项目中的落点

M1 的 Atlas dev database、PostgreSQL Compose、integration tests 和后续 CI 本地复现都依赖 Docker 拉取镜像。项目在 Atlas Community 实验中为 Docker 与独立 containerd systemd service 配置私有代理后，成功拉取固定的 `postgres:16.14` 镜像。

可执行步骤见 [Docker daemon 代理与镜像拉取](../../runbooks/docker-daemon-proxy.md)。这证明当前主机可获得 M1 所需镜像，不证明公网网络质量、Docker Hub 可用性或生产部署网络已经被全面验证。

交互 shell 下载使用的临时环境前缀见 [交互 shell 的临时代理下载](../../runbooks/temporary-proxy-downloads.md)。它故意不记录真实代理地址，并且不能替代 daemon 或 SSH 配置。

## 最小验证实验

1. 先确认 shell 或 Go 工具对目标上游的访问结果；不要输出代理变量值。
2. 用 `docker info` 的布尔模板确认 Docker daemon 是否已读取代理。
3. 查询 `containerd.service` 是否独立运行且带有相同环境。
4. 在无运行容器时重启两个服务并执行 `docker pull postgres:16.14`。
5. 用 `docker image inspect` 保存 image digest；失败时按网络层级记录错误，而不是假定应用代码有问题。

## 面试表达

“我排查过一个依赖下载路径不一致的问题：Go 工具可访问上游，但 Docker 拉取镜像在 registry token 阶段失败。关键不是反复换镜像或改项目代码，而是区分交互 shell、dockerd 和 containerd 的进程环境。给两个 systemd 服务配置同一条私有代理、在无工作负载时重启后，固定 PostgreSQL 镜像拉取成功；同时把真实网络拓扑留在主机私有配置，只在仓库保留脱敏 runbook 与证据。”

## 延伸阅读

- Docker daemon proxy configuration：https://docs.docker.com/engine/daemon/proxy/
- M1 Atlas 提案：`../../architecture/proposals/P-0001-atlas-migration-workflow.md`
