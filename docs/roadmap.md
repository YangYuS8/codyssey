## 路线图 (Roadmap)

### 已完成 (MVP-1)
- 基础项目结构（前端 / Go / Python / 基础设施）
- Problem CRUD（统一响应格式）
- 健康检查与版本注入
- 内存与 Postgres 仓储接口抽象
- 测试：Problem 生命周期 + Health

### 进行中 / 近期
- 提交(Submission) 领域建模
- 判题任务消息结构草案
- AI 题目生成 API 真实化（模型/供应商接入）

### 规划 (MVP-2)
- 判题 Worker & 队列消费
- Judge0 对接封装 (执行 + 资源限制映射)
- 统一认证 (JWT + Role)
- 分页/过滤/排序统一库
- OpenAPI 文档自动生成

### 中期 (MVP-3)
- AI 质量评估与重复检测
- 竞赛管理 (Ranking, 罚时, 冻结板) 基础实现
- 评测缓存与重判机制
- 撤回 / 讨论区 / 标签系统

### 长期 (SCALE)
- 观察性 (Prometheus + Grafana + Tracing)
- 水平扩展 Worker & 分布式调度
- 资源隔离 / 优先级队列 / 限流
- 任务重试与死信队列策略
- 多语言编译镜像优化与缓存

### 风险 / 技术债跟踪
- 缺少正式迁移工具 (当前 EnsureSchema)
- 缺少统一错误码枚举文档
- 无 Auth / 无权限控制
- 缺少安全基线（速率限制 / CSRF / JWT 过期策略）

---
在 PR 或 ISSUE 中附上对应阶段标签加速评审。
