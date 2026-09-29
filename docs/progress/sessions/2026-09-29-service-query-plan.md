# 2026-09-29：Service 查询计划施工包复盘

## 范围与审阅结论

所有者明确确认模型参数足够后，助手核对当前分支 `perf/m1-service-query-plan`、基线1405708、SQL/生成代码/adapter/测试、四份 M1 施工包文档、ADR-0004、CI 入口及外层 v3.1 原始方案的 M1 七项最小交付。本轮属于“查询计划及优化前后记录”，不扩大到 k6、pprof、其他资源查询或 M2。

- 已实现且有证据：100 Team / 100000 Service 的确定性造数、历史与现有索引对照、深页 OR/行比较正反序六轮测量、同时间桶/深页/末页/空页手工等价检查；按 Team 后续页业务 SQL 改为行比较，生成参数保持兼容。完整证据见[报告](../../benchmarks/m1-service-list-query-plan.md)。
- 本地回归：生成及生成一致性、vet、普通与非缓存 race、真实分页 integration、全 adapter integration race（4.606s）、make verify 均通过，无漏洞发现。执行者为所有者，助手未代跑数据库操作。
- 未发现本轮 SQL、生成参数、分层或公共契约的阻塞缺陷。没有新增依赖、schema、migration 或 API 参数；历史迁移完整保留。
- 有意边界：旧索引是事务内重建的历史形态，不是恢复整个历史应用版本；只有单次索引对照。重复统计只覆盖现有索引下的深页谓词，不证明 HTTP 容量或 prepared generic plan 表现。
- 测试覆盖限制：现有 Go 分页 integration 覆盖过滤及真实参数调用，但没有固定同时间戳和最后游标空页场景；本轮以手工 SQL 样本补证据，未将其伪称为自动回归。
- 文档缺陷已修复：阶段入口仍称查询实验待开始、PR #28 未合并、Compose 仅有本地证据；migration runbook 仍把开发库限定为 `_test` 且示例只检查 URL 后缀。现在区分当前状态与历史记录，强调实际目标确认。

## 原始方案对齐

v3.1 M1 最小项6要求一份查询计划及优化前后记录；本轮提供对应本地证据。没有引入原方案深度增强项。M1 最终 clean-runner/空环境验收、学习总结与 release 仍未完成，不进入 M2。

## Atlas 升级决定

2026-09-29 查阅 [v1.3.0 官方发布页](https://github.com/ariga/atlas/releases/tag/v1.3.0)。发布介绍涉及 Cloud Security Graph、Scripts、Registry CLI 和数据库驱动改进；它同时区分默认发行二进制与 Community 许可，不能把全部发布能力视为本项目源码构建 Community 的能力。

本轮继续锁定 ADR-0004 的 Community v1.2.0 / commit47daa88aea519f7f4c4aab5adfde2beab9b10b13。当前 migration/CI 基线工作正常，没有本轮证据证明必须升级；这不等于完整漏洞审计，也不能用应用的 govulncheck 结果证明 Atlas 工具链安全。

推荐时间：本查询优化 PR 合并后做独立工具维护评估；无相关安全/正确性修复或兼容性阻碍时，不阻塞 M1 release，可安排在 release 后、M2 开工前的维护窗口。若确认现版本存在影响本项目的漏洞、迁移错误或构建失效，则提前处理，不机械等待阶段结束。

升级验证清单：审阅 Community 源码差异和构建要求；更新固定 tag 与核实后的完整 commit；用独立空库验证 diff/apply/status、重复 apply、无变更 diff及历史完整性拒绝；验证已有迁移链与 sqlc 生成一致性；运行 make verify 和最新四项 required CI；同步 ADR 的版本补充记录、安装脚本及 runbook。保留旧版本可复现安装入口，不改已应用 migration 或用 migrate hash 掩盖差异。仅同一工具版本更新不必另造选型 ADR，若授权/功能边界改变则重新决策。

## PR #29 首轮 CI 证据

所有者推送 `bdc9517` 并创建 [PR #29](https://github.com/9AliMay9/lyapus/pull/29)，随后提供 `gh pr checks --required --watch` 的最终输出：4 successful，0 failing/pending/skipped/cancelled。

| Required check | 耗时 | 运行证据 |
| --- | --- | --- |
| atlas-community | 36s | [job](https://github.com/9AliMay9/lyapus/actions/runs/36524867702/job/109265564342) |
| compose | 1m24s | [job](https://github.com/9AliMay9/lyapus/actions/runs/36524867702/job/109265564409) |
| smoke | 59s | [job](https://github.com/9AliMay9/lyapus/actions/runs/36524867702/job/109266202642) |
| verify | 2m35s | [job](https://github.com/9AliMay9/lyapus/actions/runs/36524867702/job/109265564583) |

证据来源为所有者终端回传，助手未另行读取远端 job 日志。结合仓库 workflow，该轮覆盖既有迁移、生成一致性、测试、HTTP smoke 与容器交付门禁；CI 没有重跑10万行查询计划实验，不将本地性能数字描述为 clean-runner 测量，也不宣称长期稳定性。

本次只补充文档，不改 SQL 或运行配置，不递归开启同级收口审阅。补交后必须等待最新提交的四项 required checks；本记录不提前宣称 PR 已合并，M1 总验收仍待完成。

## 当前资源与下一步

- 旧开发 project/库/卷已于09-26删除，详见[记录](2026-09-26-dev-cleanup.md)，没有恢复开发 API。
- 查询实验：lyapus-query-plan / lyapus_query_plan_test / 55434 / lyapus-query-plan_postgres_data。
- 集成测试：lyapus-integration / lyapus_integration_test / 55433 / lyapus-integration_postgres_data。
- 上述两项目已由所有者于2026-09-29清理，详情如下；实验原始十二份输出已归档，不再只依赖 /tmp。助手未代执行删除。
- 本地审阅完成；PR #29 提交 `bdc9517` 首轮四项 required CI 已通过，见下方运行记录。本次补充文档仍在同一分支，需等待补交后最新四项检查成功再合并。M1 最终验收另行安排。

### 临时资源清理实证

删除前项目列表各有一个 healthy PostgreSQL，分别监听回环55433和55434；两卷的 `com.docker.compose.project` 标签与目标项目一致。所有者执行：

```bash
docker compose -p lyapus-integration down --volumes
docker compose -p lyapus-query-plan down --volumes
unset LYAPUS_TEST_DATABASE_URL LYAPUS_QUERY_PLAN_DATABASE_URL
```

输出确认以下六项 Removed：

- `lyapus-integration-postgres-1`、`lyapus-integration_postgres_data`、`lyapus-integration_default`。
- `lyapus-query-plan-postgres-1`、`lyapus-query-plan_postgres_data`、`lyapus-query-plan_default`。

随后两个 `compose ps -a`、两个目标卷列表都只有表头，55433/55434 的 `ss -ltn` 查询均无监听。这仅证明被检查项目和端口在该时刻的状态，不代表全机 Docker 已清空。unset 仅影响当前 shell。

两库数据随卷删除，包括实验的100个 Team 和100000个 Service；未提供备份，不承诺恢复原库。可在新卷显式迁移后按造数脚本重建实验数据，但测量耗时不保证相同。仓库中的脚本、报告、十二份原始输出不受影响；没有执行镜像删除、全局 prune 或 /tmp 清理。之后如需重跑本地 integration 或 make verify，必须重建独立测试库并迁移，不能使用已失效连接或把实验库替作测试库。
