# M1：Service 按 Team 游标查询计划实验

## 问题、版本与结论

在相同数据上分别回答两个问题：历史 `(team_id, id)` 索引与现有排序复合索引有什么差别？已有复合索引时，深页 OR 游标条件能否变为索引范围条件？这不是 HTTP 吞吐或容量压测。

- 基线：`1405708`（PR #28）；实验及实现位于 `perf/m1-service-query-plan` 工作区，尚未提交的变更不能冒充基线已有能力。最终源码和证据由包含本文的提交固定。
- PostgreSQL 实验执行输出由项目所有者提供；正反序十二份原始文件由助手直接读取归档。本地回归于 2026-09-29 通过；本分支 required CI、合并和 M1 最终验收尚未完成。
- 采用现有 `(team_id, created_at DESC, id DESC)` 索引，将且仅将 `ListServicesAfterCursorByTeamID` 改为 `(created_at, id) < (cursor_time, cursor_id)`。不修改 schema、历史 migration、API cursor 或其他列表查询。
- 在本数据集深页中，OR 过滤掉800行；行比较把游标边界纳入 `Index Cond`。正反执行顺序均观察到行比较更快，但耗时明显受顺序影响，不宣称固定倍数提速。

## 环境与负载

- 单台开发云主机上的 PostgreSQL 容器，容器内 psql 发起串行 SQL；无 HTTP、连接池负载、公网请求或并发发压。
- 项目既有主机规格为 Ubuntu 24.04、4 vCPU / 8 GB；本轮未重新采集 CPU 型号或整机背景负载，不能据此形成硬件横向比较。
- PostgreSQL `16.14 (Debian 16.14-1.pgdg13+1)`，x86_64，gcc Debian 14.2.0；Compose 镜像标签 `postgres:16.14`，本轮未保存镜像 digest。
- 独立 project `lyapus-query-plan`，数据库 `lyapus_query_plan_test`，回环端口55434，卷 `lyapus-query-plan_postgres_data`；仅启动 postgres。
- `docker inspect` 的 Memory、NanoCpus、CpuQuota、CpuPeriod 均0，CpusetCpus 为空：未设置这些容器限制，不等于无限资源或专用主机。
- `block_size=8192`；`shared_buffers=16384 × 8kB`（128 MiB）；`work_mem=4096kB`；`effective_cache_size=524288 × 8kB`（4 GiB，规划估计而非内存预留）；`random_page_cost=4`；`seq_page_cost=1`；`max_parallel_workers_per_gather=2`；`jit=on`（不表示本查询实际使用 JIT）。
- services 表19 MB，全部索引15 MB，总计34 MB，为 `pg_size_pretty` 输出；不是内存占用。初始磁盘可用157G；没有记录实验期间峰值资源。

## 数据与安全边界

[造数脚本](../../scripts/seed-service-query-plan.sql)在精确库名校验、空表检查和表锁后写入100个 Team、100000个 Service、0个 Environment，并执行 ANALYZE。每个 Team 恰好1000个 Service，description 为100个 ASCII 字符。

时间公式为固定起点加 `(((n * 37) % 1000) / 10)` 秒。37与1000互质，使0–999余数重排；整数除10形成100个时间桶，每桶10条。它是确定性数据，不是随机生产流量。插入顺序 `n, team_id` 使不同团队记录交错。Team 1 的时间范围为 `2026-01-01 00:00:00+00` 至 `00:01:39+00`。

脚本只检查表为空，不检查 identity 序列是否曾被消耗。复现硬编码 ID 必须使用全新卷并按现有两份 migration 建库；清空旧表不等价于全新序列。执行比较前重新查询边界，若 ID 不同应停止并核查，不能机械套用本文数值。

查询全部返回七个 Service 字段，`team_id=1`，排序 `created_at DESC, id DESC`，SQL `LIMIT 21` 对应应用20条加1条探测下一页。时间与 id 都非空、两列同向排序，行比较与原 OR 具有相同边界语义。

| 游标位置（1起算） | created_at（UTC，2026-01-01） | id |
| --- | --- | --- |
| 15，同时间桶内部 | 00:01:38 | 43101 |
| 20，第二页边界 | 00:01:38 | 29601 |
| 800，深页 | 00:00:20 | 35601 |
| 990，末页 | 00:00:01 | 48601 |
| 1000，最后一条 | 00:00:00 | 75601 |

游标通过按上述排序的 `OFFSET 位置减1 LIMIT 1` 取得；OFFSET 仅用于实验定位，不进入业务分页。

## 索引对照与单次观察

[对照脚本](../../scripts/compare-service-index.sql)在事务内临时移除现有排序索引，建立历史 `(team_id, id)` 索引，再 EXPLAIN 首页和深页 OR；最终 ROLLBACK。它含 DDL、可能持锁，只允许独占可丢弃实验库。不是上线 migration，不可用于生产。所有者随后查询 `pg_indexes`，确认现有复合索引恢复、临时历史索引不存在，主键和 `(team_id, slug)` 唯一索引仍在。

| 索引 / 查询 | 实际计划关键节点 | 执行 ms | 执行 Buffers | 过滤丢弃 |
| --- | --- | ---: | --- | ---: |
| 历史索引，首页 | Bitmap Heap Scan 1000行 → Sort → Limit | 1.250 | hit1006/read6 | 无 |
| 现有索引，首页 | Index Scan → Limit，无 Sort | 0.140 | hit27 | 无 |
| 历史索引，深页 OR | Bitmap Heap Scan → Sort → Limit | 0.756 | hit1006 | 800 |
| 现有索引，第二页 OR | Index Scan，游标在 Filter | 0.208 | hit47 | 20 |
| 现有索引，深页 OR | Index Scan，游标在 Filter | 1.784 | hit833 | 800 |
| 现有索引，深页行比较 | Index Scan，游标在 Index Cond | 0.168 | hit27 | 无 |

历史索引两个 Sort 都是 top-N heapsort，Memory 34kB；输入仍须读取该 Team 的全部1000条。深页保留200条再取21条。不能只引用首页收益而隐去历史索引深页单次0.756ms快于现有索引 OR 的1.784ms：缓存、访问方式和先后顺序并未受控，DDL 本身也影响缓存。该索引对照仅证明本次计划差异，未做多轮性能统计。

以上单次数字来自所有者终端输出，未保存完整原始终端文件；不可将下方后续多轮文件冒充这些单次输出。深页原始计划的可核查完整样本见归档。

## 深页 OR 与行比较的重复测量

使用 [正序脚本](../../scripts/measure-service-cursor.sql)及[反序脚本](../../scripts/measure-service-cursor-reverse.sql)。二者仅交换两条查询的顺序，均校验库名、设置超时、开启只读事务并 ROLLBACK。每轮新 psql 会话，在同一会话内依次运行两条 EXPLAIN；不是两组同时执行。每组第0轮作预热，第1–5轮用于统计。没有清 PostgreSQL 或 OS 缓存；后执行查询也可能受同一会话已初始化状态影响，不能把差异全部归因于磁盘缓存。

从项目根目录执行的调用形式（示例是复现入口，不是本轮重新执行记录）：

```bash
docker compose -p lyapus-query-plan exec -T \
  -e PGPASSWORD=lyapus postgres \
  psql -X -P pager=off -h 127.0.0.1 -U lyapus \
  -d lyapus_query_plan_test -v ON_ERROR_STOP=1 \
  < scripts/measure-service-cursor.sql
```

凭据仅为 Compose 本地占位值。正反两组分别执行0–5轮，保存 stdout/stderr；若任一轮退出失败则停止，不计入成功样本。反序将输入文件改为 `scripts/measure-service-cursor-reverse.sql`。历史索引对照输入改为 `scripts/compare-service-index.sql`，需额外确认其 DDL 风险。

| 轮次 | 正序 OR ms | 正序行比较 ms | 反序行比较 ms | 反序 OR ms |
| --- | ---: | ---: | ---: | ---: |
| 0（排除） | 1.755 | 0.033 | 0.150 | 1.628 |
| 1 | 1.752 | 0.034 | 0.139 | 1.633 |
| 2 | 1.805 | 0.035 | 0.154 | 1.644 |
| 3 | 1.804 | 0.033 | 0.142 | 1.621 |
| 4 | 1.735 | 0.033 | 0.137 | 1.628 |
| 5 | 1.805 | 0.034 | 0.146 | 1.631 |

最小 / 中位 / 最大（ms）：正序 OR 1.735 / 1.804 / 1.805，正序行比较0.033 / 0.034 / 0.035；反序行比较0.137 / 0.142 / 0.154，反序 OR 1.621 / 1.631 / 1.644。不将两组混合成一个掩盖顺序影响的数字。

十二份输出在 [evidence/m1-service-query-plan](evidence/m1-service-query-plan/README.md)，仅去掉行尾空格。第1轮正序执行 hit833 / hit24，反序 hit27 / hit830；OR 都过滤800行，行比较无该 Filter。Buffers 为访问计数而非不同页数，父子节点数值不相加；Planning 与 Execution 分开解释。

## 正确性与实现回归

对原 OR 和候选行比较分别按相同排序取21条，再比较 count 与 `array_agg(id ORDER BY created_at DESC, id DESC) IS NOT DISTINCT FROM ...`，不是仅比较集合或数量。

| 场景 | 原 / 候选行数 | 有序 ID 相同 | 附加断言 |
| --- | --- | --- | --- |
| 第800条之后 | 21 / 21 | t | 深页样本 |
| 第15条之后 | 21 / 21 | t | 同时间桶保留5条，不含游标自身 |
| 第990条之后 | 10 / 10 | t | 不含游标自身 |
| 第1000条之后 | 0 / 0 | t | 不含游标自身；两个 NULL 聚合用 NULL-safe 比较 |

这些是手工 SQL 样本，不是所有输入的自动证明。实际修改仅为 `db/queries/services.sql` 的命名参数行比较及对应 sqlc 生成 SQL；Go 字段 TeamID / CreatedAt / ID / Limit 的类型和顺序未变，adapter、domain、HTTP 层不变。

2026-09-29 所有者在独立 `lyapus-integration` / `lyapus_integration_test` / 55433 上应用两份 migration（版本20260729030502，pending0）后验证：

- `make generate`、`make generate-check` 成功。
- `go vet ./...`、普通测试和 `go test -race -count=1 ./...` 通过。
- `TestServiceRepositoryIntegrationListPaginationAndTeamFilter` 实际运行通过（0.054s 包耗时），覆盖真实 pgx/sqlc 调用及 Team 过滤分页；现有用例不固定同时间戳数据，也未专设最后游标之后空页断言。
- `go test -tags=integration -race -count=1 ./internal/catalog/postgres` 通过（4.606s）。
- `make verify` 通过，全包 integration 非缓存执行，漏洞扫描无发现；普通及 race 部分有缓存，另有上方 `-count=1` 证据。

集成测试会清表，绝不连接55434实验库。实验库虽以 `_test` 结尾，仍会通过通用测试库名防护；名称防护不能替代操作者选择正确目标。

## 限制与后续

- EXPLAIN 使用字面量，不是 pgx prepared statement 的 generic/custom plan 对照；真实 repository 测试证明正确执行，不证明所有参数分布下沿用实验计划。
- 不证明冷缓存、并发写入、偏斜团队、大量不同参数、HTTP 延迟、QPS 或生产容量；未新增索引或量化写放大收益。
- 本轮保留已有完整测试及选定手工边界证据；将同时间戳与空页固化为专门 Go 回归用例是后续增强，不称为已实现。
- 2026-09-29 两个实验/集成测试项目的容器、网络及卷已由所有者逐一确认后删除；目标项目与卷列表为空，55433/55434 无监听，当前 shell 连接变量已取消。数据库数据不再保留，原始测量已归档；重新执行须新建实验环境，不能假定原库仍在。见[本轮清理实证](../progress/sessions/2026-09-29-service-query-plan.md)。旧开发库已于09-26删除，详见[清理记录](../progress/sessions/2026-09-26-dev-cleanup.md)。
- 本轮本地审阅不替代最新提交四项 required CI，也不代表 M1 v0.1 总验收或 release 已完成。

参考：[PostgreSQL 16 行比较](https://www.postgresql.org/docs/16/functions-comparisons.html#ROW-WISE-COMPARISON)、[EXPLAIN 解读](https://www.postgresql.org/docs/16/using-explain.html)。
