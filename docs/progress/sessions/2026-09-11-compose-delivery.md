# 工作会话：2026-09-11 - Compose 本地交付复盘

## 范围与审阅

项目所有者确认模型配置足够后进行收口复盘。只读核对 Dockerfile、忽略规则、Compose、Makefile、CI、启动/测试辅助代码、ADR-0004、M1 施工包和原始 v3.1 方案；助手只维护文档，不替所有者执行容器操作、测试或 Git 写操作。当前分支 `feat/m1-compose-delivery`；运行证据来自所有者提供的终端输出。

## 已实现且有证据

- Go 1.26.6 builder、CGO 关闭、scratch 运行镜像、非 root 用户及 CA 证书；本地构建导出 `lyapus-apiserver:m1-dev` 成功。
- 代理按 daemon、客户端认证、RUN 下载分层处理；临时环境、代理 build args 与构建 host 网络组合成功。真实私有网络配置不入库。
- Compose 固定 PostgreSQL 16.14。新 `lyapus-dev` project/卷启动，宿主机 Atlas 完成两份 migration dry-run/apply/status，版本 `20260729030502`、pending 0，重复 apply 无待执行文件。
- API 从本地镜像启动，两个健康端点 200；Team 创建 201、列表 200；Service 携带两个初始 Environment 创建 201，Service 详情、过滤列表和 Environment 过滤列表均 200。
- 数据库停止后 livez 200、readyz 503；恢复数据库而不重启 API，readyz 回到 200，演示数据保留。仅证明本次停止/恢复，不证明所有断连或 IP 变化场景。
- 独立 `lyapus-integration` project、新卷、55433 端口、`lyapus_integration_test` 库完成迁移；integration race 通过（4.562s），`make verify` 通过，漏洞扫描无发现。测试后开发演示数据保留。
- 已核对测试卷 project 标签；测试容器、网络、卷删除完成，测试变量 unset。删除后开发 API 仍返回原 Service 与两个 Environment；开发 project/卷有意保留。

## 有意边界与问题

- Atlas 在宿主机作为显式部署步骤执行，不嵌入应用启动或增加未经验证的迁移镜像。这符合现有 ADR；不是单条 `compose up` 自动建表。
- API 无镜像内健康检查，使用宿主机 HTTP 探针验收；数据库健康不代表 migration 完成。
- dev/test 卷独立但没有严格权限隔离。相同演示凭据、开发库 `_test` 后缀使误配 URL 仍有破坏风险；操作文档显式区分端口/库/project，未擅自修改既有测试保护策略。
- 容器名称、卷名不硬编码全局名称，保持 project 隔离。host 网络仅作为本机临时构建手段，不进入运行服务。
- 原始最小版仍要求查询优化证据；本轮不做 M2、鉴权、压测或生产就绪声明。

## 尚未完成的验收

- 原 CI smoke 继续使用 `go run`；已在同一 workflow 新增独立 `compose` job，覆盖镜像构建、新数据库迁移、探针、Team 创建/回读、诊断和清理。新增 job 已通过静态复查及 PR #27 首轮实跑；2026-09-12 的 `gh pr checks 27 --required` 输出确认它与原三项共同成为 required checks，且四项均成功。
- 2026-09-13 已在独立 project 复走 README 业务演示（替代端口、已有镜像）；五分钟是已准备好工具/镜像后的演示路径，不是全新机器下载与启动计时承诺。
- Compose API 停止验收已补齐：`docker compose -p lyapus-dev stop apiserver` 后 `status=exited exit=0 oom=false`，日志包含 `shutdown_signal_received` 和 `http_server_stopped`；重新启动后 readiness 200。证明本次停止/恢复，不证明有在途长请求时的排空行为。
- M1 查询计划实验、最终资源记录、最终验收及 release 未完成。当前审阅和文档同步完成不等于交付包可立即合并。

## 学习要点

### 2026-09-13 独立演示复走与清理

使用 `lyapus-readme-check` project、新卷、`lyapus_readme_test` 库、数据库端口 55434 和 API 端口 8081，复用已有镜像，避免影响原开发环境。实际库名与两份迁移 dry-run 输出已核对；本轮 apply/status 完成由项目所有者确认，不将缺少的终端输出重构为独立记录。两个健康端点 200，Team 与带两个 Environment 的 Service 创建 201；名称转写错误经 PATCH 修正为 200，Team 列表、Service 过滤列表/详情、Environment 过滤列表均 200。证明已有工具与镜像条件下的业务演示复走，不是全新机器构建或五分钟计时证明。

清理前核对了两个容器和卷 project 标签；`down --volumes` 实际删除本轮两个容器、网络和卷，专用连接变量已 unset，之后容器及卷列表为空。演示卷数据已删除不保留；原开发 API 8080 仍返回 200 和 2026-09-11 的原 Service、两个 Environment，开发 project 与镜像有意保留。

### 2026-09-12 CI 与门禁补证

PR #27 首次提交 `7cdac0c` 的 [Actions run 34615373246](https://github.com/9AliMay9/lyapus/actions/runs/34615373246) 四项均成功：atlas-community 35s、compose 1m16s、smoke 52s、verify 2m41s。项目所有者提供 required-only 输出，确认四项都属于合并门禁。它证明本次 clean-runner 构建、显式迁移、健康检查和 Team 写入回读，不证明长期稳定性、完整 Compose CRUD 或 dev/test 隔离的 CI 覆盖；后两者相应依赖既有 smoke 和本地证据。最终 README 业务演示复走已完成，文档提交后最新 CI 仍待完成，不提前记录合并。

服务名是内部 DNS 与地址解耦，不是隐藏宿主机 IP；容器 loopback 与宿主机 loopback 不同；镜像可以共享而卷不共享；`LYAPUS_TEST_DATABASE_URL` 不会重配已运行 API。URL 中的 `&` 必须用引号保护，否则 shell 会后台执行前半段命令。
