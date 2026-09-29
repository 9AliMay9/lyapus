# 2026-09-26：旧开发环境清理

## 原因与授权

PR #28（squash `1405708`）将新开发库默认名改为 `lyapus_dev`，并保护旧开发库不被集成测试误清。配置改变不会重命名已有数据库。所有者实际查询确认旧实例仍只有 `lyapus_dev_test` 和 `postgres` 两个非模板库，随后明确选择删除旧开发环境及演示数据，而非保留或迁移旧卷。

## 删除前确认

- `docker compose -p lyapus-dev ps -a`：仅开发 API 和 PostgreSQL 两个容器，分别发布 8080 和 55432 到宿主机回环地址。
- 卷 `lyapus-dev_postgres_data` 的 `com.docker.compose.project` 标签为 `lyapus-dev`。
- PostgreSQL 容器挂载为 `volume lyapus-dev_postgres_data -> /var/lib/postgresql/data`。
- 已向所有者说明该操作删除数据库和演示数据；没有备份就不能恢复。没有提供本次备份证据。

## 已执行与结果

所有者执行：

```bash
docker compose -p lyapus-dev down --volumes
unset LYAPUS_DATABASE_URL
```

输出确认以下四项均 Removed：

- `lyapus-dev-apiserver-1`
- `lyapus-dev-postgres-1`
- `lyapus-dev_postgres_data`
- `lyapus-dev_default`

随后 `docker compose -p lyapus-dev ps -a` 和目标卷列表只显示表头；`ss -ltn 'sport = :55432'` 与 `ss -ltn 'sport = :8080'` 均无监听。这是该项目与端口在检查时的状态，不是全机资源清空声明。

## 数据与后续操作边界

- 旧库 `lyapus_dev_test`、Team `compose-demo`、Service `compose-api`、staging/production 环境随卷删除，不再可回读；历史成功验证证据仍保留。
- 本次没有使用删除镜像选项；不能把保留的 `lyapus-apiserver:m1-dev` 标签当作最新源码已构建的证据。
- 没有执行数据库重命名、备份恢复或新开发库初始化。当前没有运行中的开发 API，curl 8080 连接失败不应直接诊断为程序缺陷。
- `unset` 只作用于当前 shell；后续检查 `.env` 和 `LYAPUS_POSTGRES_DB` 等覆盖项，避免重新初始化为旧库名。
- 需要恢复开发时按 Compose runbook 从新卷建立 `lyapus_dev`：重新构建当前源码镜像，启动 PostgreSQL，显式应用 versioned migrations，启动 API，再新建演示数据并读取实际 ID。
- 下一步查询计划实验使用独立实验 project/库/卷，不要求先恢复开发环境。旧库名防护仍保留，以保护其他环境或后来恢复的历史数据。

该记录是合并后真实资源操作的新增证据，可随下一施工包提交；不修改历史迁移，不为只记录 PR 合并递归新增 PR。
