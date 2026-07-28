# P-0001：验证 Atlas Community migration 工作流

- 状态：experiment required
- 日期：2026-07-28
- 影响范围：M1 PostgreSQL schema、migration、CI 与开发者工作流

## 问题

M1 需要一个免费、可离线复现、能保留 versioned SQL 审阅历史的 PostgreSQL migration 工具。项目所有者希望学习声明式 schema 工作流，但不能以隐藏 SQL 基础、依赖付费服务或追逐“未来主流”作为代价。

## 候选方案

推荐先验证 Atlas Community：

```text
db/schema.sql（期望状态）
        ↓ migrate diff
db/migrations/*.sql（可审阅版本）
        ↓ migrate apply
PostgreSQL 16.14
```

选择理由：

- 同时保留声明式期望状态与 versioned SQL，适合练习 schema-as-code、diff、评审和可重复部署。
- Community Edition 公开支持 PostgreSQL 以及 `migrate diff`、`apply`、`status` 等核心 versioned migration 能力。
- `atlas.sum` 可以暴露 migration 目录被历史修改的情况。
- 这类工作流与平台工程的变更治理有关，但价值来自可验证流程，不来自品牌本身。

主要替代方案是 `golang-migrate`。它成熟、简单、显式执行手写 up/down SQL，且项目所有者已有使用经验；若 Atlas 的免费路径不稳定、有关键能力门槛或让学习重心偏向工具排障，应直接回退。M1 不比较 ORM 自动迁移，因为它不满足独立、可审阅部署迁移的目标。

## 免费版边界

M1 只接受不登录 Atlas Cloud、不使用 token、不要求试用或付费的能力。migration lint、drift detection、云 Registry、预迁移高级检查等不作为 M1 v0.1 能力，也不得在作品说明中暗示已经具备。

Atlas 是否会成为必然主流、以及特定大厂是否采用它，都没有被本项目验证，不作为选型理由。

## 小实验

在写第一行业务源码前，用固定版本的 Atlas Community 和 PostgreSQL 16.14 完成：

1. 从一份含 Team、Service、Environment 的最小 `schema.sql` 生成首个 versioned SQL，并确认同一文件可被固定版本 sqlc 解析。
2. 人工逐句解释生成 SQL，并确认没有意外删除、非预期 extension 或不可解释语句。
3. 对空库 apply，检查 revision 状态和实际表/约束。
4. 再次执行 apply，确认无重复变更；未修改 schema 时再次 diff 应为空。
5. 修改期望 schema 新增一个安全字段或索引，生成第二个 migration，验证旧库前滚。
6. 人工修改已生成 migration，确认完整性校验能阻止或显著暴露历史漂移。
7. 全程断网登录态、Cloud token 和 Pro 功能均不应成为核心路径前提。

实验只使用可丢弃的 `_test` 数据库；删除或重建前必须验证数据库名。

## 通过条件

- 核心命令全部来自免费 Community Edition，并可固定版本、在本地与 CI 非交互执行。
- 期望 schema、生成 SQL、revision 历史和完整性文件都可进入 Git 并由人审阅。
- 同一 `db/schema.sql` 可同时作为 Atlas 期望状态和 sqlc schema 输入，不需要维护两份结构定义。
- 能从空库可靠建立 schema，也能对已有库生成并应用下一版 migration。
- 项目所有者能够解释实际 SQL、锁和约束，不把生成器当作正确性证明。

## 退出条件

任一情况成立就停止投入并使用 `golang-migrate`：

- 核心 diff/apply/status 路径需要登录、token、试用或付费功能。
- 固定 Community 版本无法在本地与 CI 稳定安装或非交互运行。
- PostgreSQL diff 产生无法可靠解释或控制的核心变更。
- 为适配工具增加的配置和排障成本明显超过 M1 schema 规模的学习收益。
- 关键安全检查只能靠不可用的付费能力，而本项目无法用人工评审和数据库测试补足。

回退不是失败；它证明项目按可复现性、可解释性和阶段目标选择工具。

## 决策产物

实验通过后新建 ADR，固定：

- Atlas Community 精确版本与安装校验方式。
- `schema.sql`、migration 目录和完整性文件路径。
- diff、人工评审、apply、status 和 CI 命令。
- 已应用 migration 的不可修改规则和 dev/test 恢复边界。

实验不通过则在同一 ADR 记录退出证据，并固定 `golang-migrate` 版本与手写 SQL 工作流。提案在 ADR 接受后改为 resolved 并互相链接。

## 上游依据

- Atlas Community Edition：https://atlasgo.io/community-edition
- 声明式与 versioned workflow：https://atlasgo.io/concepts/declarative-vs-versioned
- Migration directory integrity：https://atlasgo.io/concepts/migration-directory-integrity
- Atlas CLI reference：https://atlasgo.io/cli-reference
- golang-migrate：https://github.com/golang-migrate/migrate
