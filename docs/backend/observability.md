<!-- 合并：已整合原根 observability.md + backend/observability.md。 -->

# Observability (可观测性)

目标：统一日志 (Logging)、指标 (Metrics)、分布式追踪 (Tracing) 与告警 (Alerting) 的设计与演进策略。当前阶段已实现基础日志与 Prometheus 指标；Tracing / Alerting 正在规划。

## 1. 总览矩阵
| 维度 | 当前状态 | 短期目标 | 中期目标 |
| ---- | -------- | -------- | -------- |
| 日志 Logging | zap + trace_id | prod JSON + 采样 | 结构字段规范 & 审计日志 |
| 指标 Metrics | HTTP / 状态转移 / 冲突 / 耗时 | DB / 队列 / 沙箱 指标 | 完整 SLI 集 (延迟/错误/饱和度) |
| 追踪 Tracing | 未实现 | OTel PoC | 采样策略 + 与日志互链 |
| 告警 Alerting | 未实现 | 基础规则 (5xx / P99) | 噪声抑制 & 组合告警 |
| Profiling | 未实现 | On-demand pprof | 连续剖析 |

## 2. 日志 (Logging)
### 2.1 现状
- `zap` development logger（dev 环境更易读）
- Trace ID 中间件：每请求注入 `trace_id`

### 2.2 字段规范（建议）
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
| user_id | 认证用户 | UUID |
| resource_id | 业务主资源 | submission_id |

### 2.3 规划
- Prod 强制 JSON + 字段顺序固定
- 错误堆栈：`zap.Error(err)` + cause 链
- 审计日志：敏感操作单独 logger（后续）

## 3. 指标 (Metrics)
详见 `metrics.md`。补充计划：
| 指标 (计划) | 类型 | 说明 |
| ----------- | ---- | ---- |
| `codyssey_db_query_duration_seconds` | Histogram | 按语句类型 / 标签统计 DB 延迟 |
| `sandbox_exec_duration_seconds` | Histogram | 沙箱执行耗时 |
| `queue_enqueue_total` | Counter | 判题任务入队次数 |
| `queue_latency_seconds` | Histogram | 入队到开始执行延迟 |

SLI 推荐：
- 可用性：1 - 5xx_rate
- 延迟：P95/P99 关键路由
- 正确性：非法状态转移计数（future 指标）
- 饱和度：in-flight 请求 / 队列 backlog（future）

## 4. 分布式追踪 (Tracing)
### 4.1 原则
| 原则 | 说明 |
| ---- | ---- |
| 低侵入 | OTel SDK + Gin instrumentation |
| 可采样 | 默认低采样率，调试时提升 |
| 可关联 | TraceID 与日志 trace_id 对齐 |

### 4.2 初期 PoC
- 依赖：`go.opentelemetry.io/otel` + stdout/OTLP 导出器
- 覆盖：HTTP inbound → (预留) DB / sandbox
- CLI 触发：可临时开启 100% 采样调试

### 4.3 Metrics 互补
| 问题 | 首选 | 说明 |
| ---- | ---- | ---- |
| 单请求慢 | Trace | 查看 Span 树 |
| 整体延迟升高 | Metrics | 分位点趋势 |
| 偶发错误 | Logs + Trace | 通过 trace_id 聚合上下文 |

## 5. 告警 (Alerting)
### 5.1 基础规则建议
| 规则 | 条件 | 动作 |
| ---- | ---- | ---- |
| 高错误率 | 5xx_rate > 2% 持续 5m | Page/IM |
| 高延迟 | P99 > 1.5s 持续 10m | 创建事件 |
| 判题停滞 | queued->running 速率≈0 且队列 backlog 上升 | 人工介入 |
| 冲突激增 | conflicts_total 斜率异常 | 代码/负载排查 |

### 5.2 噪声抑制
- 延迟窗口 (for 5m) 避免瞬时尖峰
- 分层告警：先性能后错误

## 6. 数据采集与存储策略
| 数据类型 | 保留 | 说明 |
| -------- | ---- | ---- |
| 应用日志 | 7–14 天热 + 归档 | 按成本调节 |
| 指标 | 15s 抓取 30–90 天 | 关注存储膨胀 |
| Trace | 1–5% 采样 + 错误全量 | 成本平衡 |
| Profiling | On-demand | 调试期启用 |

## 7. 演进路线图
| 阶段 | 交付 | 验收 |
| ---- | ---- | ---- |
| O11y-1 | OTel PoC + JudgeRun 耗时 | 样本 trace 可见 |
| O11y-2 | DB 指标 + 冲突计数 | /metrics 更新 |
| O11y-3 | Tracing 上线 | Trace 与日志关联 |
| O11y-4 | 全量 SLI/SLO Dashboard | Grafana 模板提交 |

## 8. 风险与缓解
| 风险 | 说明 | 缓解 |
| ---- | ---- | ---- |
| 标签基数膨胀 | 动态值过多 | 审核新增标签 |
| 采样过低 | 排查不足 | 动态调采样 + 错误保留 |
| 指标开销 | Histogram 过多 | 控制桶 + 采样抓取 |

## 9. 实施清单（近期）
- [ ] OTel PoC (HTTP + 基础 span)
- [ ] 增加 DB 耗时指标
- [ ] Sandbox 执行耗时埋点
- [ ] 冲突指标接入仪表盘报警
- [ ] Grafana Dashboard 初稿

## 10. Runbook (占位)
1. 发现：告警触发 → 查看近期部署 / 主要指标面板
2. 限定：通过 route + status 聚合确定受影响范围
3. 深挖：抽样 trace & 相关错误日志（trace_id 关联）
4. 缓解：扩容 / 降级 / 回滚 / 热修复
5. 复盘：记录根因 + 指标改进项

---
最后更新：2025-09-22（合并版）
