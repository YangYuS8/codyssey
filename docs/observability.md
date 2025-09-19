# Observability (可观测性)

> 目标：统一日志 (Logging)、指标 (Metrics)、分布式追踪 (Tracing) 与告警 (Alerting) 的设计与演进策略。当前阶段仅实现基础日志与 Prometheus 指标。

## 1. 总览
| 维度 | 当前状态 | 短期目标 | 中期目标 |
| ---- | -------- | -------- | -------- |
| 日志 Logging | zap + trace_id | prod JSON + 采样 | 结构字段规范 & 审计日志 |
| 指标 Metrics | 最小集 (HTTP / 状态转移) | JudgeRun 耗时 / DB 耗时 | 完整 SLI 集 (延迟/错误/饱和度) |
| 追踪 Tracing | 未实现 | OTel PoC | 采样策略 + 与日志互链 |
| 告警 Alerting | 未实现 | 基础规则 (5xx / P99) | 噪声抑制 & 组合告警 (多信号) |
| Profiling | 未实现 | On-demand pprof | 连续剖析 (continuous profiling) |

## 2. 日志 (Logging)
### 2.1 现状
- 使用 `zap` development logger（dev 环境更易读）
- Trace ID 中间件：每个请求注入 `trace_id`

### 2.2 字段规范（建议执行）
| 字段 | 含义 | 示例 |
| ---- | ---- | ---- |
| ts | 时间戳 | 2025-09-19T10:00:00Z |
| level | 级别 | info / warn / error |
| msg | 描述 | request completed |
| trace_id | 请求追踪 ID | 5f2c1... |
| method | HTTP 方法 | GET |
| route | 匹配路由 | /problems/:id |
| status | HTTP 状态码 | 200 |
| latency_ms | 耗时 ms | 12.4 |

### 2.3 规划
- Prod 模式强制 JSON + 统一字段顺序
- 错误堆栈：使用 `zap.Error(err)` 并保留 cause 链
- 审计日志：后续对权限敏感操作（题目删除 / 评测重判）单独分类

## 3. 指标 (Metrics)
详见 `metrics.md`。

补充即将新增：
| 指标 (计划) | 类型 | 说明 |
| ----------- | ---- | ---- |
| `codyssey_judge_run_duration_seconds` | Histogram | JudgeRun start->terminal 耗时 |
| `codyssey_db_query_duration_seconds` | Histogram | 按语句类型 / 标签统计 DB 时延 |
| `codyssey_submission_conflict_total` | Counter | Submission 状态更新冲突次数 |
| `codyssey_judge_run_conflict_total` | Counter | JudgeRun 状态更新冲突次数 |

SLI 推荐：
- 可用性：成功请求占比 = 1 - 5xx_rate
- 延迟：P95 / P99 HTTP 按关键路由 (提交 / 判题启动)
- 正确性：非法状态转移计数（可通过 error 日志 + 指标补充）

## 4. 分布式追踪 (Tracing)
### 4.1 设计原则
| 原则 | 说明 |
| ---- | ---- |
| 低侵入 | 使用 OTel SDK + Gin instrumentation |
| 可采样 | 默认低采样率，问题排查临时提升 |
| 可关联 | TraceID = 日志 trace_id 字段；可选生成 SpanID |

### 4.2 初期方案 (PoC)
- 引入依赖：`go.opentelemetry.io/otel` + stdout 导出器
- 追踪范围：HTTP inbound → (未来) DB 查询 wrap → JudgeRun 调度
- 后续：导出 Jaeger / Tempo / OTLP Collector

### 4.3 Tracing 与 Metrics 互补
| 问题类型 | 优先工具 | 说明 |
| -------- | -------- | ---- |
| 单请求慢 | Tracing | 查看 Span 树定位卡点 |
| 整体延迟升高 | Metrics | 观察总体分位点变化 |
| 偶发错误 | Logs + Trace 采样 | 通过 trace_id 聚合上下文 |

## 5. 告警 (Alerting)（规划）
### 5.1 基础规则建议
| 规则 | 条件 | 动作 |
| ---- | ---- | ---- |
| 高错误率 | 5xx_rate > 2% 持续 5m | Page / 钉钉 / Slack |
| 高延迟 | P99 > 1.5s 持续 10m | 创建告警事件 |
| 判题停滞 | queued->running 速率接近 0 且队列 backlog 上升 | 人工介入 |
| 冲突激增 | conflict_total 斜率异常 | 观察潜在逻辑竞态 |

### 5.2 噪声抑制
- 使用延迟窗口 (for 5m) 避免瞬时尖峰
- 嵌套告警：先触发“延迟升高”再允许“请求失败率” page

## 6. 数据采集与存储策略
| 数据类型 | 保留策略 | 说明 |
| -------- | -------- | ---- |
| 应用日志 | 7–14 天（热） + 冷存归档 | 依访问量调优 |
| 指标 | 高频 15s 抓取，保留 30–90 天 | 资源占用与容量平衡 |
| Trace | 采样 1–5% / 错误强制保留 | 降低成本 |
| Profiling | 间歇捕获 | 仅在性能调查时启用 |

## 7. 演进路线图
| 阶段 | 交付 | 验收标准 |
| ---- | ---- | -------- |
| O11y-1 | OTel PoC + JudgeRun 耗时指标 | /metrics 增加 duration，Trace 查看 1 条样本 |
| O11y-2 | DB 指标 + 冲突计数 + Alert 规则草稿 | 有告警 YAML 草案 |
| O11y-3 | Tracing 集成生产（Jaeger/Tempo） | Trace 与日志 trace_id 对齐 |
| O11y-4 | 完整 SLI/SLO 仪表盘 | Grafana Dashboard 模板提交 | 

## 8. 风险与缓解
| 风险 | 说明 | 缓解 |
| ---- | ---- | ---- |
| 标签基数膨胀 | route / user-id 等无控制增长 | 限制标签：不加用户级动态标签 |
| 追踪采样过低 | 排查粒度不足 | 动态调采样，错误强制保留 |
| 指标开销 | 高频抓取影响性能 | 控制桶数量，必要时调抓取间隔 |

## 9. 实施清单（近期建议）
- [ ] 增加 JudgeRun started_at / finished_at 字段（迁移）
- [ ] 指标：`judge_run_duration_seconds`
- [ ] 指标：冲突 counter（submission_conflict_total / judge_run_conflict_total）
- [ ] OTel PoC：HTTP + 自定义 Span
- [ ] 采集 DB 查询耗时（wrap pgx）

## 10. 参考
- Prometheus 官方最佳实践
- OpenTelemetry Spec 1.x
- Google SRE Workbook：黄金四指标 (Latency / Traffic / Errors / Saturation)

---
本文件将随阶段推进更新；旧阶段内容不删除而是标注完成状态。