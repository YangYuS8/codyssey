## 路线图 (Roadmap)

保持迭代透明与可追踪：已完成事项归档，进行中与后续阶段按价值与依赖排序。

### 已完成
- 基础项目结构（前端 / Go / Python / 基础设施）
- Problem CRUD（统一响应 envelope: data/error）
- 健康检查与版本注入
- 内存与 Postgres 仓储接口抽象
- Submission / JudgeRun 领域建模 + 状态机 + 条件更新防竞态
- RBAC 权限体系（含 `judge_run.manage`）与 JWT 鉴权（dev/test debug identity）
- 统一错误码体系（`api_errors.md`）
- Prometheus 最小指标（HTTP + 状态转移）
- zap 日志 + trace_id 中间件
- 版本化数据库迁移（goose）+ AutoMigrate 开关
- golangci-lint 集成 & 配置清理
- 文档结构初步重组（导航 / domain-model / openapi 维护策略）

### 进行中 / 近期 (Next 4–6 周)
- 分页 / 过滤 / 排序通用参数库
- Judge Worker 初版（队列消费 stub + 状态回写）
- JudgeRun 执行耗时指标 & 冲突 409 显式错误码
- OpenAPI 自动化策略评估（swag 注释 vs oapi-codegen 契约优先）
- AI 题目生成 API 外部供应商接入（第一家）
- Property-based 状态机测试（Submission / JudgeRun 不变量）

### 规划 (MVP-2)
- Judge0 对接封装 (执行 + 资源限制映射)
- 重判（Rejudge）机制与批次实体
- 评测结果细粒度（测试点结果存储与聚合）
- Observability 扩展：DB / Sandbox 耗时指标
- OpenAPI 代码生成 / 动态差异校验脚本

### 中期 (MVP-3)
- AI 质量评估与重复检测
- 竞赛管理 (Ranking / 罚时 / 冻结榜)
- 结果缓存与快速重判
- 社区交互：撤回 / 讨论区 / 标签系统
- 高级权限/审核流（题目上架审批）

### 长期 (SCALE)
- 全链路 Tracing (OpenTelemetry + 采样策略)
- 分布式 & 多集群 Worker 调度 + 优先级 + 限流
- 多租户 / 资源配额 / 隔离策略
- 任务重试、死信队列、指数退避策略
- 多语言编译镜像构建缓存 / 预热调度
- 分库分表与读写分离预案
- 题目内容审核 / 风险检测自动化

### 风险 / 技术债跟踪
| 类别 | 项目 | 风险 | 缓解计划 |
| ---- | ---- | ---- | -------- |
| 安全 | 缺少速率限制 | 滥用/暴力尝试 | 集成令牌桶/全局 limiter (MVP-2) |
| 安全 | JWT 过期/刷新策略未定 | 会话管理不足 | 设计 refresh token + 旋转策略 |
| 数据 | 大规模迁移锁表风险 | 高峰期阻塞 | 预生产影子演练 + 分批迁移 |
| 判题 | 沙箱未接入 | 判题流程缺失 | 优先封装 Judge0 API |
| API | OpenAPI 手工漂移 | 文档不一致 | 差异检测脚本 & CI warning |
| 可观测 | 无 tracing | 瓶颈定位困难 | OTel POC (MVP-2) |
| 性能 | 缺少 DB 指标 | 慢查询不可见 | pgx instrumentation |

### 标签 (Labels) 建议
使用 issue / PR 标签：`area:judging`, `area:api`, `area:infra`, `observability`, `techdebt`, `security`，便于过滤与仪表盘统计。

---
Roadmap 将按迭代回顾稳定更新；若新增方向请在 PR 中附带影响分析与优先级建议。
