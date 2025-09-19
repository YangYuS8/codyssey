# ADR-0001: 判题状态机与乐观并发控制策略

Status: Accepted
Date: 2025-09-19
Author: System

## 1. 背景 (Context)
系统需要对用户提交 (Submission) 与具体判题执行 (JudgeRun) 管理状态流转，并保证在并发环境下状态单调且不被回退或非法跳转。

挑战：
- 多个并发 Worker/内部 API 可能尝试更新同一条记录。
- 需要收集状态转移指标 (metrics) 以支持监控。
- 后续将扩展为分布式调度（更多竞态场景）。

## 2. 目标 (Goals)
- 明确定义状态集合与合法有向边。
- 防止非法状态跳转 (Invalid Transition)。
- 防止并发覆盖（后写覆盖前写）导致的回退。
- 提供可观测指标以洞察流转速率与异常。

## 3. 选项 (Considered Options)
| 方案 | 描述 | 优点 | 缺点 |
| ---- | ---- | ---- | ---- |
| A. 乐观锁（条件更新） | `UPDATE ... WHERE id=? AND status=?` | 实现简单；无额外列；冲突检测显式 | 需要调用方处理 0 行更新语义 |
| B. 版本号字段 (version) | 每次更新 +1 | 可扩展到多字段变更 | 多一次字段维护；对本用例略显冗余 |
| C. 分布式锁 | 外部锁 (Redis) | 强制串行 | 引入外部依赖；性能瓶颈；单点风险 |
| D. 事件溯源 | 仅追加事件重建状态 | 审计强；历史保留 | 实现复杂；当前阶段超出需求 |

## 4. 决策 (Decision)
采用方案 A：条件更新 + 状态机校验。

## 5. 细节 (Details)
### 5.1 状态机
Submission:
```
 pending -> judging -> ( accepted | wrong_answer | error )
```
JudgeRun:
```
 queued -> running -> ( succeeded | failed | canceled )
```
非法跳转 → 返回 `INVALID_TRANSITION`；无效状态值 → `INVALID_STATUS`。

### 5.2 更新逻辑流程
1. 读取当前记录状态（或直接用预期 from 状态）。
2. 校验目标状态是否属于允许集合。
3. 生成 `UPDATE ... SET status=$1 WHERE id=$2 AND status=$3`。
4. 受影响行数 == 0 → 表示并发冲突或状态已改变；未来映射为 `CONFLICT`（当前阶段可能返回 NOT_FOUND/INVALID_TRANSITION 之一 —— TODO）。
5. 成功后：
   - 写 `updated_at`
   - 记录状态转移指标 `*_status_transitions_total{from, to}`

### 5.3 指标 (Metrics)
- `codyssey_submission_status_transitions_total{from,to}`
- `codyssey_judge_run_status_transitions_total{from,to}`
未来扩展：失败占比、持续时间直方图 (running -> terminal)。

## 6. 后果 (Consequences)
正面：
- 低实现复杂度；无需新字段；对读写比高的场景友好。
- 冲突即 0 行更新，天然适配幂等重试策略。

负面：
- 调用方需要区分“对象不存在”与“状态已变”——需引入显式 `CONFLICT` 错误码。
- 无法表达更精细的并发意图（例如针对多字段的同时乐观控制）。

## 7. 不做的事情 (Out of Scope)
- 事件溯源 / 完整状态历史；可由独立表 (status_log) 后续实现。
- 分布式锁或悲观锁；避免影响吞吐与引入集中瓶颈。

## 8. 后续计划 (Follow-ups)
| 编号 | 项目 | 说明 |
| ---- | ---- | ---- |
| F1 | 引入 `CONFLICT` 错误码 | 消除歧义（0 行更新 vs 真 NOT_FOUND） |
| F2 | JudgeRun 时长指标 | 需要 started_at / finished_at 字段支持 |
| F3 | 状态日志表 | 审计与历史分析支持 (submission_status_log, judge_run_status_log) |
| F4 | Property-based 测试 | 校验状态机不变量与随机序列合法性 |

## 9. 状态 (Status)
Accepted — 当前实现已按此策略落地。

---
若未来改为版本号策略需新增列并统一迁移，届时本 ADR 会被 superseded。
