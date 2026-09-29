# 数据库 migration

## 用途与边界

本 runbook 管理 M1 的 PostgreSQL versioned migration。它只适用于可确认的 dev/test 或经变更审批的目标库；应用启动不执行 migration。真实连接串、密码和私有网络地址不能进入 Git 或终端记录。

## 前置条件

- 从仓库根目录执行，目标数据库与当前工作负载已确认。
- 本机已有 `.tools/bin/atlas`；缺失时按 `scripts/install-atlas-community.sh` 构建。该脚本需要下载时使用 [临时代理下载](temporary-proxy-downloads.md) 的单命令环境前缀。
- `LYAPUS_DATABASE_URL` 已在当前 shell 私有设置；开发库默认 `lyapus_dev`，一次性测试库以 `_test` 结尾。开发库不应为了通过测试库检查而改成 `_test` 名称。
- 已阅读本次 schema 变更与生成的 SQL；绝不让应用自动运行 migration。

## 生成与审阅新 migration

只改 `db/schema.sql`，不编辑已应用的 `db/migrations/*.sql`。然后生成一个语义化名称的新 migration：

```bash
./.tools/bin/atlas migrate diff <change_name> \
  --dir "file://db/migrations" \
  --to "file://db/schema.sql" \
  --dev-url "docker://postgres/16.14/dev?search_path=public"
```

人工审阅新增 SQL 与 `atlas.sum`；确认没有意外删除、表重写、extension 或不可解释语句。未改 schema 时同一命令应报告 migration 目录已与期望状态同步，且不生成文件。

## Apply 与状态确认

先确认当前目标的 Compose project、容器、端口、卷和实际数据库名，并通过该目标的 psql 连接执行 `SELECT current_database();`。具体开发与测试命令见 [Compose runbook](compose-delivery.md)。本节命令使用 `LYAPUS_DATABASE_URL`；独立集成测试使用 `LYAPUS_TEST_DATABASE_URL`，实验使用专用变量，不相互覆盖，也不打印连接串。

不能仅用 URL 路径的 `_test` 后缀推断实际连接目标：查询参数可能覆盖库名，编码也可能改变解析结果。Go 集成测试的有效库名防护不包裹 Atlas 命令；Atlas dry-run/apply 前仍须人工确认。清表测试不允许连接开发库或需要保留数据的查询计划实验库。

先 dry-run；`migrate apply` 会校验既有 `atlas.sum`：

```bash
./.tools/bin/atlas migrate apply --dry-run \
  --dir "file://db/migrations" \
  --url "$LYAPUS_DATABASE_URL"
```

经人工审阅后才执行真正 apply，并检查 revision 状态：

```bash
./.tools/bin/atlas migrate apply \
  --dir "file://db/migrations" \
  --url "$LYAPUS_DATABASE_URL"

./.tools/bin/atlas migrate status \
  --dir "file://db/migrations" \
  --url "$LYAPUS_DATABASE_URL"
```

预期 status 显示 `Migration Status: OK`、无 pending files。重复 apply 应输出没有 migration 需要执行。

## 历史完整性与恢复边界

- `atlas.sum` 与 migration 文件一同提交。`migrate apply`/`status` 会校验它；若 Atlas 报 checksum mismatch，停止操作并恢复被编辑的历史文件。
- `atlas migrate hash` 会重算并写入 `atlas.sum`，只可用于已人工审阅的**新、未应用** migration 的受控维护；绝不能用它接受已应用历史或未知改动。
- M1 只承诺可丢弃 dev/test 数据库的重建；尚未演练备份恢复、生产 rollback 或零停机迁移。
- 需要撤销本机测试时，先验证容器/数据库名称，再停止并删除一次性测试容器或重建 `_test` 数据库；不要将清理命令泛化到未知目标。

## 已验证证据

Atlas Community v1.2.0、PostgreSQL 16.14 与本 runbook 的 diff/apply/status 核心路径已在 P-0001 和 GitHub-hosted CI 实测。细节见 [ADR-0004](../architecture/decisions/ADR-0004-atlas-community-migrations.md)。

### CI 中的 PostgreSQL 启动稳定性

在 GitHub-hosted runner 上，不能把一次 `pg_isready` 成功直接视为数据库已可供 migration 使用：官方镜像的初始化路径可能先运行临时 PostgreSQL，再切换到正式服务。CI 先轮询 `pg_isready`，成功后短暂等待，再通过容器内 `psql -v ON_ERROR_STOP=1 -c 'SELECT 1'` 确认真实 SQL 往返；任一步失败都输出容器日志并失败。不要以盲目重试替代这个就绪确认。
