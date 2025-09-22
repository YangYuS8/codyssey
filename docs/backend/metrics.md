<!-- 合并说明：本文件已整合原根目录 metrics.md 全量内容，成为唯一指标文档。 -->

# Metrics 监控说明

本文件描述当前 Go 后端已暴露的 Prometheus 指标、抓取方式与后续可扩展点。

## 1. 暴露端点

- 路径：`/metrics`
- 协议：HTTP GET
- 格式：Prometheus / OpenMetrics (`Content-Type: text/plain; version=0.0.4` 或 `application/openmetrics-text`)
- 典型抓取间隔：`15s`（可按部署规模与开销调优）

## 2. 指标列表（最小可用集）

| 指标名 | 类型 | 标签 | 说明 | 示例用途 |
| ------ | ---- | ---- | ---- | -------- |
| `codyssey_http_requests_total` | Counter | `method`, `route`, `status` | 累积 HTTP 请求总数 | QPS、错误率 (`status` 聚合) |
| `codyssey_http_request_duration_seconds` | Histogram | `method`, `route` | HTTP 请求耗时分布 | P95/P99 延迟、慢路由识别 |
| `codyssey_http_in_flight_requests` | Gauge | (无) | 当前正在处理的请求数 | 过载/阻塞识别 |
| `codyssey_submission_status_transitions_total` | Counter | `from`, `to` | Submission 状态跳转计数 | 观察状态机健康、失败激增 |
| `codyssey_judge_run_status_transitions_total` | Counter | `from`, `to` | JudgeRun 状态跳转计数 | 调度/执行阶段异常监测 |
| `codyssey_judge_run_duration_seconds` | Histogram | `status` | JudgeRun 从 start->finish 总耗时 | 评测性能、长尾分析 |
| `submission_conflicts_total` | Counter | (无) | Submission 状态/版本更新时发生乐观锁冲突次数 | 并发写入热点、重试放大识别 |
| `judge_run_conflicts_total` | Counter | (无) | JudgeRun 状态更新（queued→running / running→终态）冲突次数 | 竞争队列/执行阶段冲突诊断 |

### 2.1 直方图桶
`codyssey_http_request_duration_seconds` 直方图桶：
```
0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10 (秒)
```
适配常规 API（毫秒级到数秒级）。

## 3. Prometheus 抓取配置示例
在 Prometheus `prometheus.yml`:
```yaml
scrape_configs:
	- job_name: codyssey-backend
		metrics_path: /metrics
		scrape_interval: 15s
		static_configs:
			- targets: ["backend-service:8080"]
```
Kubernetes：
```yaml
	- job_name: codyssey-backend
		kubernetes_sd_configs:
			- role: pod
		relabel_configs:
			- source_labels: [__meta_kubernetes_pod_label_app]
				action: keep
				regex: codyssey-backend
			- source_labels: [__meta_kubernetes_pod_container_port_number]
				action: keep
				regex: "8080"
```

## 4. 常见查询 (PromQL)
| 目标 | 示例查询 | 说明 |
| ---- | -------- | ---- |
| 总 QPS | `sum(rate(codyssey_http_requests_total[5m]))` | 整体吞吐 |
| 错误率 | `sum(rate(codyssey_http_requests_total{status=~"5.."}[5m])) / sum(rate(codyssey_http_requests_total[5m]))` | 5xx 占比 |
| 路由 P95 | `histogram_quantile(0.95, sum by (le, route) (rate(codyssey_http_request_duration_seconds_bucket[5m])))` | 按路由分位延迟 |
| In-flight | `codyssey_http_in_flight_requests` | 当前并发 |
| Submission 状态跳转速率 | `sum(rate(codyssey_submission_status_transitions_total[5m]))` | 状态机活跃度 |
| JudgeRun queued -> running | `rate(codyssey_judge_run_status_transitions_total{from="queued",to="running"}[5m])` | 调度吞吐 |

## 5. Grafana 面板建议
| 面板 | 建议类型 | 关键指标 |
| ---- | ------- | -------- |
| 请求延迟 | Heatmap / Time series | `codyssey_http_request_duration_seconds` 分位 |
| 请求量 & 错误 | Stacked / Time series | 成功 vs 4xx vs 5xx |
| In-flight | SingleStat / Gauge | `codyssey_http_in_flight_requests` |
| 状态机活跃度 | Time series | Submission / JudgeRun transitions 速率 |
| 异常报警 | Alert rules | 高 5xx 比例、P99 延迟飙升、transition 异常变化 |

## 6. 状态机监控要点
- `queued -> running` 速率下降：调度阻塞 / Worker 饱和
- `running -> failed` 占比升高：沙箱或题目数据异常
- `submission_conflicts_total` 升高：热点或重试策略不当

## 7. 扩展点（未来可添加）
| 类别 | 指标建议 | 备注 |
| ---- | -------- | ---- |
| Submission 首字节延迟 | `submission_enqueue_latency_seconds` | 提交 -> 第一次调度 |
| DB 交互 | `db_query_duration_seconds`、`db_connections_in_use` | 包装 `pgx` 统计 |
| 外部依赖 | `sandbox_exec_duration_seconds` | Judge0 / sandbox 耗时 |
| 请求体大小分布 | `request_body_bytes` Histogram | 分析拒绝前的典型体量 |

## 8. 性能与开销
标签基数控制：
- `route` 基数 ≈ 主要 API 数量（`c.FullPath()`）
- 状态标签来自有限枚举
- 当前无高基数动态标签（user_id 等）

## 9. FAQ
| 问题 | 说明 |
| ---- | ---- |
| 指标没有输出？ | 确认已注册中间件与 Handler，访问 `/metrics`。 |
| 延迟不准 | 检查时钟 / NTP，同步时间。 |
| 桶不匹配 | 调整代码中 `Buckets`，注意历史不可比。 |
| 标签爆炸风险？ | 当前安全，新增标签需评审。 |

## 10. 版本策略
新增指标：向后兼容（只增不删）。
重命名/删除：标注 BREAKING，发布迁移指引。

---
附加说明：
1. 请求体 / 代码大小限制目前仅通过 HTTP 错误可观测（413/400/409），若需分布增加 Histogram。
2. 冲突计数器已实现，用于识别重试或分片需求。

最后更新：2025-09-22（合并版）
