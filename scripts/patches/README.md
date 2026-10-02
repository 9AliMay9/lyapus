# Atlas Community 依赖补丁

`atlas-v1.3.0-dependencies.patch` 仅修改上游 `cmd/atlas/go.mod` 与 `cmd/atlas/go.sum`，基于 v1.3.0 / `9a6bc601212130aaaefcbc8dd36c710baf9716ff`。本地产物标识为 `v1.3.0-lyapus.1`，不是上游发布版本；Atlas 自动生成的同名 GitHub release 链接不代表该 release 存在。

## 来源与边界

- 2026-09-30 在隔离源码副本、Go1.26.6、`GOTOOLCHAIN=local` 下执行 `go get google.golang.org/grpc@v1.83.2 golang.org/x/text@v0.41.0`，共更新25个模块声明；x/text v0.41.0同时是该gRPC版本的依赖要求。
- 补丁由实际diff生成，不手写校验值；不修改Atlas业务源码、项目应用依赖或数据库迁移。
- 安装应固定基础commit、补丁哈希和本地版本，先校验再应用；不能在安装时运行动态依赖升级。
- SHA-256：`e9b0ec5264bfd9e33fa0f31d691aacbb53c2ecb59b1385cb011bf19295142638`。这是内容身份核验，不是独立的信任签名。

## 实验证据与尚未验证

临时补丁候选编译成功；govulncheck v1.6.0、漏洞库更新时间2026-09-28 16:43:40 UTC下，源码和二进制扫描均退出0，符号级和包级无命中，仍有7项模块级提示，不能称为所有依赖无漏洞。构建信息保留基础commit及 `vcs.modified=true`。

正式补丁文件已对干净基础源码通过 `git apply --check`。所有者随后通过正式安装脚本重建 `.tools/bin/atlas-v1.3.0-lyapus.1`，Go1.26.6、基础commit不变、`vcs.modified=true`；SHA-256为 `aa7ce349a974cddf071c8d8f5b164a9e535798af1cb65265ca7388c428276295`，与临时实验产物一致。正式产物在上述扫描器/漏洞库下二进制复扫退出0，符号级和包级无命中，仍有7项模块级提示。2026-10-01 默认工具已切换为同哈希补丁版；空库迁移、旧版接续、checksum拒绝、无变更diff、应用回归和新卷API演示已通过。PR #30首轮CI已通过，最新门禁仍待验证，不把本地验收等同于发布通过。扫描原始输出见 [归档](../../docs/progress/evidence/m1-atlas-upgrade/README.md)，专项结果见 [复盘](../../docs/progress/sessions/2026-10-01-m1-final-acceptance.md)。

## 维护与退出

项目维护者负责固定补丁及扫描、兼容性验证。上游发布覆盖相关修复的Community版本后重新比较并验收，优先撤掉本地补丁；不得直接把补丁套到新版本。基础commit、补丁或工具链改变时重做相关验证。修复依据与过程见 [联合验收计划](../../docs/stages/m1-go-backend/atlas-upgrade-acceptance-plan.md)。
