# M0 契约与命名

本文件是 M0 的实现约束。它先于代码存在，防止代码、讲解和后续阶段使用不同名称。

## 服务

- 二进制命令：`apiserver`
- 默认 HTTP 地址：`127.0.0.1:8080`
- 环境变量：`LYAPUS_HTTP_ADDR`
- 配置类型：`config.Config`
- 构造函数：`config.Load()`
- HTTP server 构造函数：`http.NewServer(cfg config.Config, logger *slog.Logger)`

## 健康检查

| 路径 | 方法 | 成功状态 | 语义 |
| --- | --- | --- | --- |
| `/livez` | GET | 200 | 进程存活；不检查外部依赖。 |
| `/readyz` | GET | 200 | 服务可以接收流量；M0 暂无外部依赖。 |

两条端点都返回 JSON：

```json
{"status":"ok"}
```

## 日志与错误

- 使用标准库 `log/slog`。
- 日志键使用 snake_case，例如 `http_addr`、`request_id`、`error`。
- 启动失败返回错误；`main` 负责记录错误并以非零状态退出。
- 不在 M0 引入全局 logger 或自定义日志框架。

## 后续变更规则

一旦 M0 commit 完成，以上名称视为内部基线。未来如需对外暴露或重命名，必须在对应阶段新增 ADR 或迁移说明。
