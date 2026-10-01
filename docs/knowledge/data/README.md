# 数据与中间件

建议按顺序建立：SQL 与索引、事务/MVCC/锁、PostgreSQL、MySQL、Redis、Kafka。每篇必须明确保证的语义和未保证的边界。

- [组件关键语义](component-contracts.md)：pgx、sqlc、Atlas、PostgreSQL及交付边界中值得深入的专项知识。

- [游标分页与查询计划](postgresql-cursor-query-plan.md)：Index Cond 与 Filter、复合索引、行比较、顺序效应和证据边界。

- [PostgreSQL 约束的直接行为证据](postgresql-constraint-evidence.md)：正反边界、SQLSTATE、联合唯一性、外键和引用删除的验证方法。
