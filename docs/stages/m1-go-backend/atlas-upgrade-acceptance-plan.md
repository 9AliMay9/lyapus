# Atlas 1.3.0 考察与 M1 联合验收计划

状态（2026-10-02）：补丁版已切换为本地默认 Atlas，升级专项、应用回归和新卷 API 演示通过；资源已清理并确认，PR #30 首轮 CI 已通过，最新门禁、合并与 release 待完成。基线 PR #29 squash `fdf8f25`，当前分支 `chore/m1-final-acceptance`。完整证据见 [收口复盘](../../progress/sessions/2026-10-01-m1-final-acceptance.md)。

## 考察结论

通过 GitHub 官方 compare API 检查 `v1.2.0...v1.3.0`，共12个提交、40个变更文件。另以官方 tag ref API 核对：

- 旧版：v1.2.0 / `47daa88aea519f7f4c4aab5adfde2beab9b10b13`。
- 候选：v1.3.0 / `9a6bc601212130aaaefcbc8dd36c710baf9716ff`，tag ref 指向 commit。
- `cmd/atlas/go.mod` 的 Go 要求从1.25.0升至1.26.4；根模块也要求1.26.4。项目 go.mod 与 Dockerfile 声明1.26.6，满足声明要求；实际主机版本、下载与构建仍待确认。
- CLI 依赖更新包括 x/crypto 0.46.0→0.52.0、x/net 0.48.0→0.55.0、x/sys 0.42.0→0.45.0，以及 x/mod、x/sync、x/text。提交7021f1694e49117bef93d05aeb45108cbcb4ea09明确以修复 CVE 为目的。这是升级收益，不是本项目实际漏洞可达性的证明。
- 完整文件列表未列出 `sql/postgres` 或 `cmd/atlas/internal/cmdapi` 的修改；`sql/migrate/dir.go`、`lex.go` 的变化是 lint 注释，不是执行逻辑。不能因此断言依赖更新毫无行为风险。
- 新增 `atlasexec` 的 Script / Cloud repo 调用包装，属于 SDK；不等于源码构建的 Community CLI 新增相应付费/Cloud 能力，本项目也不引入该 SDK。
- 根 LICENSE 仍为 Apache-2.0，发布页区分默认发行二进制与 Community；继续现有固定源码构建路线，不切换 curl 安装 latest 或扩展版，不引入 token/Cloud。

推荐从“无已知收益则延期”调整为“在 M1 最终验收前验证升级”，因为现在发现了依赖安全维护的具体收益且差异范围较小。尚未做完整依赖许可证或漏洞审计；本项目 make verify 扫描应用模块，不自动扫描另行构建的 Atlas。

来源（2026-09-29读取）：[官方 compare API](https://api.github.com/repos/ariga/atlas/compare/v1.2.0...v1.3.0)、[tag ref](https://api.github.com/repos/ariga/atlas/git/ref/tags/v1.3.0)、[候选 CLI go.mod](https://github.com/ariga/atlas/blob/9a6bc601212130aaaefcbc8dd36c710baf9716ff/cmd/atlas/go.mod)、[安全维护提交](https://github.com/ariga/atlas/commit/7021f1694e49117bef93d05aeb45108cbcb4ea09)、[发布说明](https://github.com/ariga/atlas/releases/tag/v1.3.0)。

## 最自然的交付单位

修正此前“必须独立升级 PR”的建议：本次以“升级后的 M1 交付验收”为同一目标，在当前分支用独立提交隔离工具升级，再提交验收证据与学习总结，用一个 PR 收口。不是一般性要求把所有升级混在功能 PR 中。

理由：查询优化已经合并，当前没有并行业务功能改动；候选未发现迁移 CLI/PG 驱动源码行为调整。拆两个 PR 会重复整套四项门禁。若后续发现功能/授权改变、需要改历史 migration、需要大幅适配或更换工具，则停止联合流程，单独讨论，不强行升级。

本轮范围：固定工具版本与 commit、必要安装可靠性、升级专项验证、最终空环境演示、文档/学习总结。禁止顺带升级 PostgreSQL、pgx、sqlc、Go、扩展 Atlas 功能或新增服务组件。实际源码仍由所有者手敲、助手复查、所有者运行。

## 覆盖与复用矩阵

截至2026-10-01，前八项本地验证已执行，具体覆盖与例外见收口复盘；第九项已由PR #30首轮CI通过，第十项最新门禁待完成。下表保留设计目标，不代替执行证据。

| 证据目标 | 唯一主验证位置 | 通过条件 | 后续如何复用 |
| --- | --- | --- | --- |
| 候选构建与身份 | 本地先保留旧工具，单独产出候选 | 版本/源码commit/Go构建信息可追溯 | 同候选进入后续验证，不反复下载重建 |
| 工具依赖安全 | 对候选二进制做独立 govulncheck binary 扫描 | 可识别结果；命中需评估或阻断，不能跳过 | 应用扫描与工具扫描分开记录，binary扫描限制注明 |
| 新版完整空库迁移 | 一次性 integration project | dry-run/apply/status、版本及pending、重复apply | 同库用于全量回归，不再另开第三个“空库验证库” |
| 旧版→新版接续 | 同测试PostgreSQL内单独兼容性库，不与清表测试并行 | 旧版仅应用第一份；写入哨兵；新版接续第二份；哨兵及revision正常 | 只做这一轮升级专属路径，不为每类API重做 |
| 历史完整性拒绝 | migration目录的独立临时副本 | 修改副本且不重算checksum，新版拒绝；真源不变 | 不改共享历史，不能用migrate hash把失败“修好” |
| 无变更 diff | 本地一次 + required atlas-community job | 无新增迁移，schema和atlas.sum未变 | 同时覆盖新工具schema解析及回放 |
| 应用回归 | 上述integration库 | make verify一次 + adapter integration race一次 | 不再先逐包普通/race再立即完整verify；需要排错时才分段 |
| 新版最终交付 | 独立空Compose验收project/卷，当前源码新镜像 | 完整迁移、三类API最短CRUD/分页链、探针中断恢复、优雅退出 | 此次就是M1最终本地验收，不再命名另一轮重复同样操作 |
| clean runner门禁 | 本PR现有四项required jobs | 对候选提交全部成功 | 本地和CI是不同环境证据，不互相替代 |
| 文档补交后的最终门禁 | 同PR最终提交 | 四项required再通过 | 必须重跑自动门禁，但纯文档无需手工重走数据库/API |

顺序：先工具静态复查与候选构建，后升级专项数据库验证，再切换项目固定工具并完成全量回归，最后当前源码镜像的空环境演示；每步失败即停并保留诊断。升级 commit 不应夹入业务逻辑变化。

项目当前四项 CI 已覆盖无变更diff、空库apply/status、compose重复apply、生成/应用测试和HTTP冒烟。没有覆盖旧版→新版接续或篡改副本负例，不把它们说成现有CI能力；这两项本轮本地专项补证据即可，不为一次升级另造大框架。

## 资源、回退和证据规则

- 项目旧 dev、query-plan、integration 卷已经按历史记录删除。任何新资源创建前重新检查名字、端口、卷；计划本身不代表端口仍空闲。
- 集成库和API演示库隔离；前者允许清表，后者保存演示数据。连接变量按用途分开，执行up时重复显式环境前缀。
- 建议兼容性库与integration库共用一次性测试PostgreSQL实例，但为两个数据库；串行验证结束后才运行清表测试。无需模拟生产集群或复制生产数据。
- 构建候选不直接覆盖唯一旧二进制。保留旧固定版本构建路径和身份信息；验证前不要改历史ADR为“已升级”。安装脚本后续可采用临时输出、成功确认后替换，具体源码另行交付。
- 工具回退不等于数据库回滚；若新工具已改变revision状态，必须核查旧工具兼容性，不能盲目换回。本次数据库均可丢弃，失败可保留现场后重建；不承诺生产恢复能力。
- 证据记录候选commit、工具完整版本/构建信息、镜像ID、迁移版本、操作命令、结果、清理确认及CI run。待测试完成后才更新阶段清单。
- 只有文档变化可以复用本地运行证据；改SQL、依赖、Dockerfile、配置或生成代码会使相关证据失效，应重跑受影响项。最终CI始终绑定最新提交。
- 不重跑PR #29的10万行性能实验：本轮不改查询、索引或PostgreSQL，复用其归档；若迁移结果或数据库版本改变再重新评估。
- 本地与CI通过后按同PR流程补证据、等待最新门禁、合并、记录清理。然后对最终main做版本核对和release，不为单记“已合并”循环新增PR。

## 已执行及待办

最新进展见文首及收口复盘。以下均为09-30调查过程的历史记录；其中“未切换”“待验证”“下一步”描述只代表当时状态，不能覆盖当前接续点。

2026-09-30 经所有者授权，助手在 `/tmp/lyapus-atlas-audit.VmTq68/source` 重跑候选CLI主包源码扫描；执行前确认commit为9a6bc601212130aaaefcbc8dd36c710baf9716ff、工作树干净，Go1.26.6。沙箱内首次因代理socket权限失败（退出1），获准联网重跑后完成：govulncheck v1.6.0、漏洞库更新时间2026-09-28 16:43:40 UTC、退出3。完整626行报告保存在 `/tmp/lyapus-atlas-audit.VmTq68/source-scan-network.txt`（已归档至 [扫描证据](../../progress/evidence/m1-atlas-upgrade/README.md)）。源码扫描符号级命中GO-2026-6348、6061、5970，共3项；另有包级2项、模块级10项。扫描后源码工作树仍干净，未修改依赖或替换二进制，未连接数据库。GO-2026-6443未在源码符号级列出，不等于依赖已修复或自动豁免。x/text报告包含migrateApplyCmd→EnvByName→parseConfig→HCL/cty→norm.Form.LastBoundary的静态示例路径；gRPC示例涉及云依赖及接口调用。静态路径不是实际攻击复现，下一步核对配置分支、输入来源及修复取舍，不依据命中数量放行。

所有者已核验主机Go1.26.6、当前Atlas Community v1.2.0，并保存 `.tools/bin/atlas-v1.2.0`。两文件SHA-256均为 `250490dc947630fde59ea7d90a3cd01be58ec76dde3321a1376e8b2df86d31b1`。旧二进制构建信息为Go1.26.5、CGO_ENABLED=1、linux/amd64、vcs.revision=47daa88aea519f7f4c4aab5adfde2beab9b10b13、vcs.modified=false。主机工具链升级不会自动重编译既有工具；旧文件仅是兼容性验证/回退候选，不代表已确认安全。

2026-09-30 所有者提供候选构建证据：安装脚本已手敲并通过静态及Shell语法检查，独立输出 `.tools/bin/atlas-v1.3.0`；版本为Community v1.3.0，Go1.26.6、CGO_ENABLED=1、vcs.revision=9a6bc601212130aaaefcbc8dd36c710baf9716ff、vcs.modified=false。SHA-256为 `1a4f77eaa3c4da2e141515222d34ddb2ec49d1f87d68002b693e0fdca048c9a8`；现用Atlas与旧版备份哈希仍一致，未替换现用工具。

同日所有者执行 `govulncheck -mode=binary -show=version .tools/bin/atlas-v1.3.0`：扫描器v1.6.0，数据库https://vuln.go.dev，更新时间2026-09-28 16:43:40 UTC，退出码3。符号级命中GO-2026-6443、GO-2026-6348、GO-2026-6061（grpc v1.79.3）及GO-2026-5970（x/text v0.37.0）；另外报告1项包级及10项模块级结果，尚未展开评估。不是网络失败，也不是安全验收通过。

官方依据：[6443](https://pkg.go.dev/vuln/GO-2026-6443)、[6348](https://pkg.go.dev/vuln/GO-2026-6348)、[6061](https://pkg.go.dev/vuln/GO-2026-6061)、[5970](https://pkg.go.dev/vuln/GO-2026-5970)、[扫描边界](https://pkg.go.dev/golang.org/x/vuln/cmd/govulncheck)。二进制符号命中不提供完整调用链，不证明本项目迁移命令实际可利用，也不能据此直接豁免；grpc修复区间存在分支差异，不能只取扫描摘要中最大版本。不擅自修改上游依赖制作未标识的补丁版。当前未创建本轮数据库，迁移兼容性和M1最终验收仍未开始。

旧版对照扫描亦由所有者于2026-09-30执行，扫描器与数据库更新时间和候选相同，退出3：符号级11项，另有包级3项、模块级27项。与候选共有GO-2026-6443、6348、6061、5970；旧版额外报告GO-2026-6218、6091、6090、6088、5972、5942、5026，涉及Go1.26.5标准库及x/net。两二进制的Go工具链和依赖版本同时不同，不能把7项差异全部归因于Atlas源码升级，也不能按漏洞数量推算实际风险。旧版并非已通过安全验证的退路。

后续分诊：2026-09-30查阅官方发布列表仍标记v1.3.0为Latest；同日master的CLI go.mod仍含x/text v0.37.0，grpc为v1.83.1，不能把跟随master作为完整修复方案。来源：[发布列表](https://github.com/ariga/atlas/releases)、[master模块声明（动态页面，仅代表查阅时状态）](https://raw.githubusercontent.com/ariga/atlas/master/cmd/atlas/go.mod)。下一步在隔离源码目录、固定候选commit和本机Go1.26.6下对CLI主包做源码调用链扫描；扫描命中仍需结合实际迁移参数和输入信任边界评估。不自动豁免，也不直接更换迁移工具。
