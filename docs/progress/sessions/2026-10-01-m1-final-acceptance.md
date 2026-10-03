# 工作会话：2026-10-01 - Atlas 升级与 M1 本地最终验收

发布后补记（2026-10-03）：PR #30已合并，最终PR门禁及main CI通过，v0.1.0已发布。以下按当时记录保留的待办已被[发布同步记录](2026-10-03-m1-release-sync.md)取代，不代表当前未完成。

## 目标

在 `chore/m1-final-acceptance`（基线 PR #29 squash `fdf8f25`）验证Atlas依赖安全补丁，复用有效证据完成M1本地收口，不重复查询性能实验，不扩展业务能力。本地通过不等于本分支CI、合并或release已完成。

## 实际完成

- 所有者手敲安装脚本并运行构建、迁移、回归、镜像和API验证；助手负责审阅、此前获授权的隔离依赖实验、补丁生成及文档。下列运行事实依据所有者提供的终端输出；扫描原始文件另行归档。
- 安装脚本锁定上游v1.3.0 / `9a6bc601212130aaaefcbc8dd36c710baf9716ff`，校验固定依赖补丁后只读模块构建，核验版本再在同文件系统替换目标。`GOTOOLCHAIN=local`不自动升级工具链；拒绝相对输出路径、目录或符号链接目标。
- 本地版 `v1.3.0-lyapus.1` 只修改上游CLI的go.mod/go.sum（25项模块声明），不是应用依赖升级或业务兼容补丁，也不是上游发行版。自动打印的同名release链接不是上游发布证据。
- 正式构建与实验产物哈希一致，源码与二进制扫描符号/包级无命中，但模块级仍有7项提示。旧版扫描并未通过，不能以退回旧版替代安全评估。
- 所有者已把验收候选复制到默认 `.tools/bin/atlas`；本轮静态复核两者哈希一致。没有修改schema、历史migration、业务Go或SQL。

## 身份与验证证据

| 对象 | 身份 |
| --- | --- |
| 主机 / 新工具构建 | Go1.26.6，linux/amd64；Atlas CGO_ENABLED=1 |
| 旧工具v1.2.0 | commit47daa88aea519f7f4c4aab5adfde2beab9b10b13；Go1.26.5；SHA-256 `250490dc947630fde59ea7d90a3cd01be58ec76dde3321a1376e8b2df86d31b1` |
| 补丁 | SHA-256 `e9b0ec5264bfd9e33fa0f31d691aacbb53c2ecb59b1385cb011bf19295142638` |
| 默认/候选Atlas | v1.3.0-lyapus.1；基础commit9a6bc601212130aaaefcbc8dd36c710baf9716ff，vcs.modified=true；SHA-256 `aa7ce349a974cddf071c8d8f5b164a9e535798af1cb65265ca7388c428276295` |
| API镜像 | `lyapus-apiserver:m1-dev`；inspect ID `sha256:959f1a4e0f359117593d879c13a04d5eccf2a6ad1a567cbabe1e009d9fcc4206` |
| 数据库 / migration | PostgreSQL16.14；最终版本20260729030502，2 executed / 0 pending |

Atlas工具构建与API镜像不同：后者为Go1.26.6构建的CGO_ENABLED=0运行时产物，不包含Atlas。镜像构建复用未变化的缓存层并重新编译应用，不宣称全程无缓存构建。

| 验证 | 目标、结果与边界 |
| --- | --- |
| 新版空库 | lyapus-integration / lyapus_integration_test / 55433：dry-run、2份apply、status和重复apply均通过 |
| 旧版接续 | 同PG中的lyapus_atlas_upgrade_test：v1.2.0仅apply第一份，写Team哨兵，新版dry-run及apply第二份，status2/0；哨兵id1、slug/name及创建/更新时间2026-09-30 14:28:26.796857+00均保留 |
| 完整性拒绝 | 独立临时migration副本：baseline validate退出0；仅修改第一份SQL、不改atlas.sum，validate报checksum mismatch退出1；真源未改 |
| 无变更diff | 另一完整副本，to=db/schema.sql，dev-url使用PostgreSQL16.14/public：无新增migration、退出0；`diff -ru`与真源一致、退出0 |
| 应用回归 | 独立integration库：`go test -tags=integration -race -count=1 ./internal/catalog/postgres`通过4.496s；`make verify`通过，integration3.243s，应用扫描无发现。普通/race部分显示cached，不写成全部无缓存 |
| 新卷交付 | lyapus-acceptance / lyapus_acceptance_test / 55434，API8081：预检无资源，新卷建库确认实际库名，默认新版Atlas迁移2/0，再启动新镜像；livez/readyz200 |
| 写入与读取 | Team201/id1，Service及初始Environment201/id1和2；Team/Service列表、单项GET符合响应契约；三类名称PATCH200并回读，未改字段及关联保留 |
| 分页 | 两个Environment创建时间相同；limit1第一页production/id2，携实际cursor第二页staging/id1，next_cursor为空，无重复 |
| 依赖故障 | 停PG后livez200、readyz503；仅恢复PG，API不重启，readyz200且Service及更新过的Environment数据保留 |
| 引用与删除 | 有子资源时Team和Service DELETE409，回读子资源仍在；依次删除两个Environment、Service、Team均204，对Environment1/Service1/Team1回读404 |
| 空闲退出 | stop apiserver后status=exited、exit=0、oom=false；2026-10-01T11:57:53Z日志依次出现shutdown_signal_received、http_server_stopped |

扫描对象、版本、数据库时点与原始输出见 [扫描归档](../evidence/m1-atlas-upgrade/README.md)。安装脚本本轮 `bash -n`、补丁哈希和干净基础源码 `git apply --check` 通过；未在复盘中再次执行数据库、构建或全量测试。

本轮文档复核：23份新增/修改Markdown的相对文件链接目标均存在，`git diff --check`通过；归档三份报告的SHA-256与临时原件一致。代码抽查确认Service写事务绑定同一sqlc WithTx、显式Commit和Rollback，测试在创建连接前检查pgx解析后的有效库名。上述静态检查不替代CI；首轮CI结果见下方补记。

## 对原始方案的复核

### 2026-10-02 PR #30 首轮 CI 补记

工具升级 `e6cf98e` 与验收文档 `8ec1040` 已推送并创建 [PR #30](https://github.com/9AliMay9/lyapus/pull/30)。依据所有者提供的 `gh pr checks --required --watch` 输出，[run36942161208](https://github.com/9AliMay9/lyapus/actions/runs/36942161208) 四项成功，0失败/取消/跳过/等待。本条未另行下载日志，不声称独立审计每一步输出。

| Required job | 用时 | 证据 |
| --- | --- | --- |
| atlas-community | 1m8s | [job110635936472](https://github.com/9AliMay9/lyapus/actions/runs/36942161208/job/110635936472) |
| compose | 1m31s | [job110635936734](https://github.com/9AliMay9/lyapus/actions/runs/36942161208/job/110635936734) |
| smoke | 1m11s | [job110636990725](https://github.com/9AliMay9/lyapus/actions/runs/36942161208/job/110636990725) |
| verify | 3m49s | [job110635936621](https://github.com/9AliMay9/lyapus/actions/runs/36942161208/job/110635936621) |

结合workflow定义，本轮提供固定补丁工具在clean runner构建/使用、迁移检查、应用回归、HTTP smoke和容器交付成功证据。Atlas独立漏洞扫描、旧版接续与checksum篡改负例仍是本地专项证据，不混称CI覆盖。

本次仅补文档，不重建数据库或重复本地API验收。push更新同一PR后，须等待最新四项required通过；尚未合并或发布，不提前勾选release，不为记录最终绿灯递归追加提交。

已回读外层 `reference/生涯项目方案书v3.1.docx` 的“M1近期最小版”和阶段复盘要求。七项映射如下：

| 最小交付 | 现有依据 |
| --- | --- |
| 核心模型及归属 | 三类领域模型、schema和阶段契约 |
| CRUD/校验/错误/分页过滤 | 既有业务与HTTP测试、历史smoke、本轮API演示 |
| migration/约束/事务/并发 | PR25直接SQL约束、repository事务/并发测试，本轮迁移与回归 |
| 单元/数据库/race | 本轮make verify及独立integration race |
| 空环境Compose | 本轮新project/新卷/新镜像；当前分支clean runner仍待验证 |
| 查询计划优化前后记录 | PR29已合并，100 Team/100000 Service及原始正反序证据；本轮无SQL变更，复用而不重跑 |
| README模型/API/演示路径 | README与Compose runbook已有，本轮按选定路径复走；不声称全新主机实测五分钟 |

本地证据达到选定覆盖；资源已清理、首轮CI已通过；发布仍须文档补交后的最新四项CI、合并及release。鉴权/RBAC、容量压测、pprof、备份恢复和后续平台组件不临时升级为M1阻塞项。学习入口见 [工程骨架](../../knowledge/go/backend-engineering-baseline.md)、[组件语义](../../knowledge/data/component-contracts.md)、[变更验证](../../knowledge/reliability/change-validation.md)、[生产边界](../../knowledge/reliability/production-readiness.md)。

## 偏差、风险与误操作复盘

- 本轮手工Environment创建走Service初始创建，不是独立POST；独立POST由已有integration/HTTP测试、PR23及后续smoke提供证据，当前分支smoke尚待执行。Team/Service多页与并发不重复手工操作，复用回归；不把演示写成穷尽测试。
- 本次正常退出发生在空闲状态，未测试负载下在途请求排空、连接池耗尽或强制终止。数据库恢复不等于备份恢复、HA或生产SLO。
- Atlas源码扫描有保守调用路径、二进制扫描缺少完整调用链；0符号命中不能证明无可利用漏洞。7项模块级提示保留，后续安全维护仍由项目承担。当前CI构建/使用Atlas，但应用govulncheck不扫描Atlas；本轮独立扫描是本地证据，未擅自新增CI门禁。
- 09-30误复制shell续行提示符 `>` 导致旧工具备份变为零字节；空可执行文件在该Bash调用中退出0且无版本输出，所以只看退出码不够。已从核验过的旧默认工具恢复，哈希及version一致，确认目标库仍0 applied后才重新apply第一份。10-01检查发现同时间根目录零字节 `--dir`、`--url`，疑为同次重定向残留；10-02所有者通过 `rm -i -- ./--dir ./--url` 逐一确认删除，未提交。
- 单引号使变量不展开、schema路径笔误等失败已纠正；只将纠正后带明确输出与退出码的执行计为通过，未用migrate hash“修复”负例。
- 本地依赖补丁增加维护责任：固定来源/哈希/版本、保留扫描证据，上游提供合适修复后重新验收并优先撤补丁；不动态go get、不把工具回退等同于数据库回滚。

## 当前资源与下一次接续

- 2026-10-02暂存补丁后，`git diff --cached --check`识别其上下文标记空格与原文Tab组合，以及表示空上下文行的单个空格为格式问题。此前未暂存的新补丁不在普通 `git diff --check` 覆盖内，故先前通过不能代表它已检查。现仅对该补丁文件在 `.gitattributes` 禁用 `blank-at-eol`、`space-before-tab` 两项检查，其他文件与其他检查保持原状；不修改补丁字节或重算哈希。工具提交需包含该属性文件，重新暂存后再检查。

- 2026-10-02清理补记：所有者先核对两个项目容器与卷标签，再分别执行 `docker compose -p lyapus-integration down --volumes` 和 `docker compose -p lyapus-acceptance down --volumes`。前者容器/网络/卷3项Removed，后者两个容器/网络/卷4项Removed。integration、Atlas接续与acceptance库（含哨兵）随卷删除，未提供备份证据，不承诺恢复；镜像、工具及归档报告未删除。
- 清理后两个项目及目标卷列表均为空；55433、55434、8081无监听。当前shell已取消 `LYAPUS_TEST_DATABASE_URL`、`LYAPUS_ATLAS_UPGRADE_DATABASE_URL`、`LYAPUS_ACCEPTANCE_DATABASE_URL`。不推断其他shell或全主机资源状态；再次运行integration/make verify须先重建并迁移测试库。两个误生成空文件也已删除，Git状态复查不再列出。
- 检查暂存范围，排除根目录误生成空文件、二进制和临时目录；同一PR保留工具升级与文档的清晰提交。等待verify、smoke、atlas-community、compose四项required；记录首轮CI后等待最新文档提交门禁，不递归追加“记录最终绿灯”的提交。
- PR #30首轮CI已有证据，最新文档门禁、合并及release尚待完成。不进入M2；Git操作仍由所有者执行。纯文档修正无需重跑本地数据库/API。
