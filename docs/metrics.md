# Metrics 监控说明

> 提示：后续计划将本文件与日志 / Tracing 统一归并为 `observability.md`，本文件暂作为指标详细说明的独立入口。

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
      - targets: ["backend-service:8080"]  # 或部署中的实际主机 / K8s Service
```
若使用 Kubernetes，可改为：
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
| JudgeRun 从 queued -> running | `rate(codyssey_judge_run_status_transitions_total{from="queued",to="running"}[5m])` | 调度吞吐 |

## 5. Grafana 面板建议
| 面板 | 建议类型 | 关键指标 |
| ---- | ------- | -------- |
| 请求延迟 | Heatmap / Time series | `codyssey_http_request_duration_seconds` 分位 |
| 请求量 & 错误 | Stacked Bar / Time series | 成功 vs 4xx vs 5xx |
| In-flight | SingleStat / Gauge | `codyssey_http_in_flight_requests` |
| 状态机活跃度 | Time series | Submission / JudgeRun transitions 速率 |
| 异常报警 | Alert rules | 高 5xx 比例、P99 延迟飙升、transition 异常减少或暴增 |

## 6. 状态机监控要点
- 若 `queued -> running` transition 速率明显下降：可能是调度阻塞或 Worker 饱和。
- 若 `running -> finished` 少而 `running -> failed` 或中间停滞：内部执行组件异常。
- 可加派生报警：特定窗口内某 `to="failed"` 占比 > X%。

## 7. 扩展点（未来可添加）
| 类别 | 指标建议 | 备注 |
| ---- | -------- | ---- |
| JudgeRun 冲突计数 | `judge_run_conflict_total` | 统计并发冲突（当前可用 transitions+HTTP 409 推导，后续可单独暴露） |
| Submission 首字节延迟 | `submission_enqueue_latency_seconds` | 提交 -> 第一次调度 |
| DB 交互 | `db_query_duration_seconds`、`db_connections_in_use` | 包装 `pgx` 统计 |
| 外部依赖 | `sandbox_exec_duration_seconds` | Judge0 / sandbox 耗时 |
| 竞争失败 | `submission_conflict_total` / `judge_run_conflict_total` | 乐观锁冲突计数（支持 409） |

## 8. 性能与开销
当前指标全部为 Counter/Gauge/单一直方图，标签基数有限：
- `route` 基数 ≈ 主要 API 数量（受限于使用 `c.FullPath()`，未包含查询参数）
- 状态标签仅状态机有限集合
- 风险较低，可安全在早期环境启用

## 9. 常见问题 (FAQ)
| 问题 | 说明 |
| ---- | ---- |
| 指标没有输出？ | 确认已调用 `metrics.Middleware()` 与 `metrics.Handler()` 注册；访问 `/metrics`。 |
| 指标延迟不准 | 确认部署未发生时钟漂移；必要时统一 NTP。 |
| 直方图桶不匹配 | 可在代码中调整 `Buckets`；变更会导致历史对比不兼容。 |
| 标签爆炸风险？ | 当前无动态高基数标签（如用户 ID），安全。 |

## 10. 版本策略
未来新增指标：向后兼容（新增不删）。重命名/删除需在发布说明中显式标记 BREAKING，并提供迁移指引。

---
最后更新：自动生成日期占位（请在需要时手动维护）
