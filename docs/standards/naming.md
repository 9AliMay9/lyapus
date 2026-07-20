# 命名规范

## 通用

- Git 路径、目录、文件：小写 kebab-case；Go 包目录例外，使用小写单词，如 `catalog`、`transport`。
- 缩写只用业界通用写法：`API`、`HTTP`、`ID`、`URL`、`SLO`、`SQL`；Go 标识符写作 `ServiceID`、`HTTPClient`。
- 名称表达业务语义，不表达实现细节；避免 `util`、`common`、`helper` 等无边界名称。

## Go

- 导出标识符使用 PascalCase；未导出标识符使用 camelCase。
- 接口按行为命名，通常以 `-er` 结尾，如 `ServiceRepository`、`Clock`；单方法接口靠近使用方定义。
- 构造函数为 `NewX`，不使用 `CreateX`；`CreateX` 留给业务操作。
- context 参数始终命名为 `ctx`，并放在函数第一个参数位置。
- 接收者使用简短一致的名称；同一类型不要混用多种接收者名。

## HTTP 与数据

- 对外路径使用复数资源名：`/v1/services`；标识符使用 `service_id`，不使用 `serviceId`。
- JSON 字段使用 snake_case；Go 字段使用 PascalCase。
- 数据库表使用复数 snake_case，如 `services`；主键为 `id`；外键为 `<entity>_id`。
- 时间字段使用 `_at` 后缀并以 UTC 表达，如 `created_at`、`updated_at`。
