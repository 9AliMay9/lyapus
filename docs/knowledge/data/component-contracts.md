# 组件专项能力：不必背全部 API，但必须知道关键语义

## 一句话模型

组件不是关键词；学习它要能回答“解决什么问题、何时失败、谁持有资源、什么不保证、怎么验证”。不泛用的细节，只要处在正确性边界上，就值得深入。

## 必须掌握

| 组件 | 本项目必须理解的语义 | 易忽略的边界 |
| --- | --- | --- |
| PostgreSQL | 联合唯一/外键/非空、事务、锁、排序索引与执行计划 | 应用校验不防并发竞争；有索引不等于游标成为Index Cond |
| pgx/pgxpool | 创建pool与Ping的区别；ctx、连接归还、tx结束、SQLSTATE | 连接池对象创建不证明连通；池大小与数据库总连接预算相关 |
| sqlc | SQL为真源、参数类型、Scan顺序、WithTx、生成一致性 | 编译通过不证明业务正确或SQL高效；WithTx绑定事务，不替你提交/回滚 |
| Atlas Community | schema与版本历史、revision、checksum、diff/apply/status | 不自动证明零停机；工具升级不是数据库回滚；checksum不等于数字签名 |
| Compose | project隔离、服务名DNS、host/container端口、命名卷 | 改POSTGRES_DB不会改旧卷中的库；健康不证明schema已迁移 |
| net/http与chi | Handler边界、middleware顺序、状态码、body限制、shutdown | ReadHeaderTimeout不等于整条请求/SQL超时；包装器可能遮蔽可选接口 |
| CI/shell | 退出码、参数引用、管道失败、required check与提交关联 | 无检查记录不等于成功；旧提交成功不替代新提交；后台符号不是URL字符 |

具体工具API随版本查官方资料；上表中的资源所有权、失败语义与安全边界应当能够不依赖复制解释清楚。

## 容易混淆或踩坑

- PostgreSQL池用连接URL解析后的实际配置，不能仅看URL字符串尾巴；库名防护也不是权限隔离。
- 创建事务后要显式Commit/Rollback；sqlc生成函数不会凭函数名保证原子性。取消请求也不能当成已经完成事务清理的证据。
- `atlas.sum`可由有权限者重算，因此它能发现意外修改，但不认证作者或证明迁移安全；仍需代码审阅与权限控制。
- Docker daemon、构建客户端、构建RUN、运行容器是不同网络/代理作用域；构建临时host网络不应直接变成通用生产网络配置。
- Python、shell、SQL分别有自己的语法和断言工具；会读、能测试失败分支比背库方法更重要。尤其 Python assert 可被优化选项禁用，不能当生产输入安全校验。

## 在本项目中的落点

参见[迁移ADR](../../architecture/decisions/ADR-0004-atlas-community-migrations.md)、[数据库约束证据](postgresql-constraint-evidence.md)、[游标计划](postgresql-cursor-query-plan.md)、[网络边界](../linux/daemon-network-boundaries.md)。这些实际踩坑比记一张组件名清单更有学习价值。

## 最小验证实验

已有：直接SQL制造约束冲突并核对SQLSTATE；一个事务中途失败检查零残留；停止数据库区分livez/readyz；删除测试项目后确认另一项目不受影响。尚未演练的连接池耗尽、取消与清理、备份恢复等应明确标为后续，不为读笔记擅自执行破坏操作。

## 面试表达

“我不承诺记住所有命令参数，但能解释池、事务、生成器、迁移历史和容器卷的边界，并用负例证明失败行为；选定版本的API再查官方文档。”

## 延伸阅读

- [pgxpool](https://pkg.go.dev/github.com/jackc/pgx/v5/pgxpool)：使用时切换到go.mod锁定版本，避免把新版本文档当旧版能力。
- [sqlc事务](https://docs.sqlc.dev/en/stable/howto/transactions.html)、[项目Compose手册](../../runbooks/compose-delivery.md)。
