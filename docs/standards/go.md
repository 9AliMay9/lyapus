# Go 实现规范

- `cmd/` 只做依赖装配、配置加载、启动和信号处理；业务逻辑进入 `internal/`。
- 不把 `context.Context` 存进结构体；请求链路必须向下传递 `ctx`。
- 错误必须保留上下文：`fmt.Errorf("load service %s: %w", id, err)`；调用方通过 `errors.Is` 或 `errors.As` 判断语义。
- 不记录密钥、认证头、完整个人数据或未脱敏遥测；日志字段使用稳定的 snake_case 键。
- 每个公开行为至少有单元测试；跨数据库、网络或容器边界的测试放在 integration/e2e 层。
- 格式化以 `gofmt` 为准；提交前运行阶段指定的 `make` 或 `go` 校验命令。
