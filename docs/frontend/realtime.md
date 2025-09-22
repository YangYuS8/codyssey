# Realtime Updates (SSE + Polling)

## 目标
- 降低延迟：判题状态接近实时反馈
- 控制成本：最小化全量 refetch 次数
- 可恢复：断线后自动重连；不可用时回退轮询

## 组成
| 模块 | 作用 |
| ---- | ---- |
| useSubmissionEvents | 建立 EventSource、解析事件、调度增量更新 |
| useSubmission | 基础数据获取 + 可选轮询 enablePolling 控制 |
| React Query | 缓存结构与 setQueryData 局部合并 |

## 事件模型
枚举：
```
status_update
judge_run_update
completed
queued
running
```

### 语义
| 事件 | 说明 | 缓存处理 |
| ---- | ---- | ---- |
| queued/running | 过渡阶段（可选） | 更新 status |
| status_update | 提交整体状态变更 | 更新 status 字段 |
| judge_run_update | 单个判题执行增量 | merge/append judgeRuns[] |
| completed | 终态 | invalidate -> 重新获取快照 |

## 数据流示意
```
EventSource -> onmessage -> parse -> switch(evt.type)
  status_update / running / queued -> setQueryData(status)
  judge_run_update -> setQueryData(merge judgeRuns)
  completed -> invalidateQueries(submission/:id)
```

## 重连策略
- onerror: 标记 connected=false -> setTimeout(5s) -> reconnect
- 避免重复计时：单一 reconnectTimer 引用

## 轮询协调
- 页面初始：useSubmission(enablePolling: true)
- SSE 打开：onEvent 首次成功 -> setSseConnected(true) -> polling 自动停止
- SSE 断开：connected=false -> 轮询仍保留（enablePolling 计算可扩展）

## 性能优化点
| 风险 | 处理 |
| ---- | ---- |
| 高频 judge_run_update | 合并节流（未来可批量事件） |
| 大量同时打开详情页 | 考虑共享单链接多订阅或 WebSocket 汇聚 |
| 列表页实时刷新 | 仅推送 summary（id + status + score）减少数据体积 |

## 演进
1. WebSocket 替代：需要双向（取消、交互指令）时升级。
2. 事件签名：加入版本/签名防重放。
3. Patch 协议：支持 JSON-Patch / RCU 减少复杂合并。
