<!--
	本文件由原根目录 `docs/domain-model.md` 与 `docs/backend/domain-model.md` 合并而来。
	目标：统一领域模型（概念 + 实现）单一来源，避免跨文件漂移。
-->

# Backend Domain Model (领域模型与状态机)

本文为后端领域模型权威文档，涵盖：核心实体结构、状态机、错误码、并发控制、指标关联以及实现要点与未来扩展。

## 目录
1. Submission
2. JudgeRun
3. 错误码与状态流转
4. 并发与一致性
5. 未来扩展点
6. 关联图
7. 维护策略
8. 实现要点速览 (原 backend/domain-model.md 实现部分)

---
## 1. Submission

### 1.1 目的
表示用户对某题目的代码提交记录，是用户视角的“最终结果”载体。其评测可由 1..N 个 `JudgeRun` 支撑（未来支持多语言或重判）。

### 1.2 字段（核心）
| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| id | UUID | 主键 |
| user_id | UUID | 提交者 |
| problem_id | UUID | 题目 |
| status | ENUM | 当前聚合状态（由评测结果驱动） |
| created_at | timestamptz | 创建时间 |
| updated_at | timestamptz | 更新时间 |
| version | INT | 乐观锁版本号（每次状态或关键字段更新自增） |

### 1.3 状态机
```
 pending -> judging -> ( accepted | wrong_answer | error )
```
非法流转：
- 任何终止状态（accepted / wrong_answer / error）不可再前进或回退
- 跳过 `judging` 直接终止 → `INVALID_TRANSITION`
- 未知目标状态值 → `INVALID_STATUS`

Mermaid：
```mermaid
graph LR
	P[pending] --> J[judging]
	J --> A[accepted]
	J --> W[wrong_answer]
	J --> E[error]
```

### 1.4 失败与重新评测（规划）
- 重判（rejudge）策略：创建新的 JudgeRun 批次并更新 Submission 聚合规则。
- 聚合策略：首次 AC 即锁定（或允许配置：最新结果 / 最优结果）。

## 2. JudgeRun

### 2.1 目的
表示一次具体判题执行（调度 → 沙箱执行 → 结果采集）。可视为底层工作单元；一个 Submission 可关联多次 JudgeRun（重判、不同执行环境等）。

### 2.2 字段（核心）
| 字段 | 类型 | 说明 |
| ---- | ---- | ---- |
| id | UUID | 主键 |
| submission_id | UUID | 所属 Submission |
| status | ENUM | 判题执行状态 |
| created_at | timestamptz | 创建时间 |
| updated_at | timestamptz | 更新时间 |
| started_at | timestamptz | 进入 running 时间（用于耗时统计） |
| finished_at | timestamptz | 结束时间（与 started_at 差值形成运行总耗时） |

### 2.3 状态机
```
 queued -> running -> ( succeeded | failed | canceled )
```
非法流转：
- 直接从 queued 跳到终止状态（除 running）
- 终止后再次更新
- 未知状态值 → `INVALID_STATUS`
- 不符合链路顺序 → `INVALID_TRANSITION`

并发与冲突语义：
* Submission：使用 `version` 字段进行乐观锁（`UPDATE ... WHERE id=? AND version=?`）。成功后版本号自增；版本不匹配返回 `CONFLICT`。
* JudgeRun：使用状态条件更新 (`WHERE status='queued'` / `WHERE status='running'`) 作为轻量乐观控制。受影响行数为 0 时：
	* 记录不存在 → `JUDGE_RUN_NOT_FOUND` (404)
	* 记录存在但状态已被其它并发操作修改 → `CONFLICT` (409)
* 内存与 PG 实现均暴露对应 `ErrSubmissionConflict` / `ErrJudgeRunConflict`。

运行时长指标：
- 在 `Finish` 成功后，如果 `started_at` 与 `finished_at` 存在且顺序合法，记录 Histogram: `codyssey_judge_run_duration_seconds{status="<terminal>"}`。
- 用途：观察终态执行时间分布；识别超时 / 沙箱性能波动。示例：
	```promql
	histogram_quantile(0.95, sum by (le) (rate(codyssey_judge_run_duration_seconds_bucket[5m])))
	```

Mermaid：
```mermaid
graph LR
	Q[queued] --> R[running]
	R --> S[succeeded]
	R --> F[failed]
	R --> C[canceled]
```

### 2.4 与 Submission 的关系
- 运行完成后（succeeded / failed / canceled）将影响 Submission 聚合逻辑：
	* succeeded + 判分 → 决定 Submission 是否 accepted / wrong_answer / error
	* failed → 可能映射为 Submission.error 或维持 judging（视重试策略）
- 后续：Submission 最终状态可来自最新一次成功的 JudgeRun 结果。

## 3. 错误码与状态流转
| 场景 | 错误码 | HTTP | 触发条件 |
| ---- | ------ | ---- | -------- |
| 目标状态枚举无效 | INVALID_STATUS | 400 | 输入状态不在允许集合 |
| 状态链路非法 | INVALID_TRANSITION | 400 | 不符合状态图定义的有向边 |
| 目标对象缺失 | SUBMISSION_NOT_FOUND / JUDGE_RUN_NOT_FOUND | 404 | ID 不存在 |
| 竞争更新失败 | CONFLICT | 409 | 乐观锁 / 条件更新 0 行 |

## 4. 并发与一致性
Submission：版本号乐观锁防止“最后写入 wins”覆盖：
1. 读取返回当前 `version`。
2. 更新时携带期望版本；若 0 行受影响说明版本已变 → `CONFLICT`。
3. 客户端策略：重新获取最新状态决定是否重试。

JudgeRun：依赖状态机单调（`queued->running->terminal`）的条件更新，避免并行重复启动或结束。

冲突可观测性：`submission_conflicts_total` / `judge_run_conflicts_total` 指标用于监测热点资源竞争，可辅助决定是否需要退避或分片。

## 5. 未来扩展点
| 方向 | 说明 |
| ---- | ---- |
| 分布式调度 | 引入队列优先级、抢占、限流维度（如题目/用户配额）。 |
| 计时指标 | 已实现：`judge_run_duration_seconds`（Histogram，标签：status） |
| 重判批次 | 增加 RejudgeBatch 实体，跟踪一次批量重判影响范围。 |
| 结果细粒度 | 引入 `test_case_results`（每个测试点耗时/内存/错误原因）。 |
| 失败分类 | 扩展 failed 细分类（编译错误/运行超时/内存超限/沙箱异常）。 |

## 6. 关联图（整体）
```mermaid
erDiagram
	SUBMISSION ||--o{ JUDGERUN : has
	SUBMISSION {
		UUID id
		UUID user_id
		UUID problem_id
		STRING status
	}
	JUDGERUN {
		UUID id
		UUID submission_id
		STRING status
	}
```

## 7. 维护策略
- 新增状态：需更新：枚举、服务层校验、指标标签、本文档；错误码文档同步。
- 删除状态：需发布 BREAKING 说明，并迁移旧数据（DB / 代码 / 文档同步）。

---
## 8. 实现要点速览
（以下为原实现侧专注内容，保留便于工程落地）

### Submission 实现要点
- 乐观锁字段 `version INT`
- SQL：`UPDATE submissions SET status=?, version=version+1 WHERE id=? AND version=?`
- 0 行影响 -> `CONFLICT`
- 状态转换记录指标：`submission_status_transitions_total{from,to}`

### JudgeRun 实现要点
- 条件状态更新确保单调：`WHERE status='queued'` → running；`WHERE status='running'` → terminal
- 结束采集耗时：`judge_run_duration_seconds{status}`
- 冲突指标：`judge_run_conflicts_total`

### 聚合策略（规划）
- Submission 终态策略：最新成功 / 首次 AC 锁定 / 最佳分数（可配置）
- 重判：新建 JudgeRun 与旧记录并存；聚合器选择“有效”集合

### 扩展实体占位
| 实体 | 状态(草案) | 说明 |
| ---- | ---- | ---- |
| Contest | draft -> published -> running -> frozen -> finished | 冻结榜逻辑 |
| RejudgeBatch | created -> running -> completed -> failed | 重判批量控制 |
| AIAnalysis | queued -> running -> succeeded -> failed | AI 质量/检测任务 |

新增实体流程：
1. 定义 struct + 状态常量
2. Repository 接口 / 实现
3. 状态转移校验 + 指标埋点
4. 更新本文档

---
最后更新：2025-09-22（合并版）
