# 工作会话：2026-07-27 - M0 终局审计

## 目标

对照文档真源、原始 v3.1 方案书和真实代码确认 M0 是否基本符合预期，补齐终局文档，使下一次对话可以直接准备 M1。

## 实际完成

- 按 `docs/README.md` 的最小阅读顺序复核项目上下文、工程规范、当前进度、M0 施工包、架构、ADR、学习笔记和最近会话。
- 回读原始 v3.1 的 M0 实现、验收、资源口径和四条执行护栏。
- 逐文件核对 `cmd/apiserver`、`config`、`health`、`transport/http`、测试、Makefile、GitHub Actions、示例配置和忽略规则。
- 确认现有实现完整满足仓库内 M0 施工包，并基本符合原始方案书；手工全新 Ubuntu 15 分钟复现是唯一明确延期的原始验收证据。
- 新增 ADR-0002 四条执行护栏；为避免改写已接受的 ADR-0001，新增 ADR-0003 单独记录 PostgreSQL 主库决策。
- 在公开 README 补充固定漏洞扫描器安装方式、资源/磁盘控制口径和 PR 合并流程。
- 系统重写 M0 学习笔记，并把当前交接页切换到 M1 施工包准备入口。

## 修改的文件

- 根入口：`README.md`。
- 长期与当前上下文：`docs/project-context.md`、`docs/progress/current.md`。
- 架构决策：`docs/architecture/decisions/README.md`、ADR-0002、ADR-0003。
- 阶段结果：`docs/stages/README.md`、M0 `checklist.md`、M0 `outcome.md`。
- 学习与交接：`docs/knowledge/go/m0-engineering-baseline.md`、本会话记录。
- 协作流程：`docs/README.md`、`docs/runbooks/github-ssh.md`。

## 验证证据

```text
命令：make verify
结果：gofmt、go vet、普通测试、race 测试和 integration 入口均通过；govulncheck 获取数据库时被协作执行沙箱禁止访问本机代理，该失败不来自项目代码

命令：make vuln（允许主机网络访问）
结果：No vulnerabilities found.

命令：GitHub PR #1 的失败门禁实验
结果：verify 失败、smoke 跳过，GitHub 明确禁止合并；删除故意失败探针后 verify 与 smoke 均通过，PR 以 squash merge 进入 main

命令：git status --short（审计开始前）
结果：无输出，main 与 origin/main 位于 6306d4b
```

## 偏差、风险或待确认事项

- 手工全新 Ubuntu 15 分钟复现由项目所有者决定延期；不得把 clean-runner smoke 描述为同一证据。
- M0 的 secret scan 是忽略规则、信息边界和提交前凭据特征检查；没有引入独立自动化 secret-scanning 产品。
- `integration` 是为 M1 预留的命令入口，不应描述成已有外部依赖集成测试。
- 本批终局文档需由项目所有者阅读，并通过受保护分支的 PR 合并。

## 下一次从这里继续

- 具体文件：`docs/progress/current.md`、`docs/stages/m1-go-backend/README.md`。
- 具体任务：对照 v3.1 建立 M1 施工包，明确近期最小版与深度增强边界。
- 验收命令：M1 施工包先做 Markdown 链接、`git diff --check` 和范围审阅；确认契约后再确定代码验收命令。
- 不要做：不要跳过 M1 施工包直接写 CRUD，也不要把 M1 深度增强项变成首次 release 的阻塞项。
