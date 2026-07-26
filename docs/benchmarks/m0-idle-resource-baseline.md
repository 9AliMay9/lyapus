# M0 空闲资源基线

## 目的

记录 M0 尚未启动应用或业务容器时的主机资源状态，作为 M1、M2 和 M3 同类测量的对照基线。它不用于声明应用容量、生产可用性或容器资源上限。

## 证据范围

- Git commit：`fe1ab2e`（`ci: add vulnerability scanning`）
- 采集时间：2026-07-26T12:28:31Z
- 操作系统内核：Linux 6.8.0-124-generic
- 采集状态：Go M0 服务未启动；Docker daemon 运行，但 `docker stats --no-stream` 没有返回运行中的容器。

未记录主机名、IP、账单、代理端口或其他私有运维信息。

## 采集命令

```bash
date -u +%Y-%m-%dT%H:%M:%SZ
uname -sr
free -h
df -h /
vmstat 1 5
cat /proc/pressure/memory
ps -eo pid,comm,%cpu,%mem,rss --sort=-rss | head -n 15
docker stats --no-stream
```

## 结果摘要

| 维度 | 结果 |
| --- | --- |
| 内存 | 总计 7.5 GiB；已用 858 MiB；available 6.7 GiB；Swap 1.9 GiB 且已用 0 B。 |
| 磁盘 | 根分区 178 GiB；已用 7.8 GiB（5%）；可用 162 GiB。 |
| CPU | 排除 `vmstat` 首个开机以来平均样本后，连续四个 1 秒样本均为约 99% idle。 |
| 内存压力 | memory PSI 的 `some` 与 `full` 在 10/60/300 秒窗口均为 0.00。 |
| 进程 | 最大 RSS 约 230 MiB；Docker daemon RSS 约 85 MiB；未观察到高 CPU 或高内存占用进程。 |
| 容器 | Docker daemon 存在，但没有运行中的容器，因此没有可报告的容器级资源数据。 |

## 结论与限制

当前主机在此采样窗口内没有持续 CPU、内存、Swap 或磁盘压力。后续每次引入 M1 数据库、M2 工作负载或 M3 遥测组件后，应使用相同命令重复采样，并同时记录容器上限、实际峰值、数据量和保留期。

单次空闲快照不能推导 API 吞吐、数据库容量、容器并发上限、网络体验或生产可靠性。
