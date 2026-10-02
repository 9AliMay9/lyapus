# Atlas 升级扫描证据

2026-09-30 经所有者授权执行，2026-10-01从临时目录原样归档；文件SHA-256与原件逐一核对一致。扫描器govulncheck v1.6.0，Go1.26.6，漏洞库更新时间2026-09-28 16:43:40 UTC。退出码来自执行记录，不是报告文本自行证明。

| 文件 | 扫描目标 / 命令 | 退出码 | 结果 |
| --- | --- | --- | --- |
| [upstream-source-scan.txt](upstream-source-scan.txt) | 上游9a6bc601212130aaaefcbc8dd36c710baf9716ff，cmd/atlas内 `govulncheck -show=version,traces .` | 3 | 符号3、包2、模块10 |
| [patched-source-scan.txt](patched-source-scan.txt) | 同一基础加固定补丁，cmd/atlas内 `govulncheck -show=version,traces .` | 0 | 符号0、包0、模块7 |
| [patched-binary-scan.txt](patched-binary-scan.txt) | 补丁实验二进制，`govulncheck -mode=binary -show=version <candidate>` | 0 | 符号0、包0、模块7 |

归档来源分别为 `/tmp/lyapus-atlas-audit.VmTq68/source-scan-network.txt`、`/tmp/lyapus-atlas-patch.aJpoVX/source-scan.txt` 和同目录 `binary-scan.txt`。临时目录不参与后续构建。

正式脚本重建产物与实验二进制SHA-256均为 `aa7ce349a974cddf071c8d8f5b164a9e535798af1cb65265ca7388c428276295`；所有者另行运行的正式产物binary复扫同样退出0，其结果来自终端回传，不伪造为上述原始报告。

模块级7项提示没有消失，0退出不是“所有依赖无漏洞”或攻击不可达的证明。扫描只代表上述数据库时点；应用 `make verify` 不替代工具扫描，当前CI尚未加入Atlas独立漏洞扫描。后续版本/补丁/工具链变更需复评和重扫。

构建输入及补丁维护见 [补丁说明](../../../../scripts/patches/README.md)，验收与风险见 [复盘](../../sessions/2026-10-01-m1-final-acceptance.md)。
