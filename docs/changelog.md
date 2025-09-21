# Changelog

采用语义化版本（规划中）。格式参考 Keep a Changelog。

## [Unreleased]
### Added
 - 错误码 `CONFLICT`：用于并发/条件更新 0 行场景（返回 HTTP 409）
 - JudgeRun 冲突区分：`UpdateRunning` / `UpdateFinished` 区分不存在与状态冲突，冲突返回 409
 - Histogram 指标：`codyssey_judge_run_duration_seconds`（按终态标签记录运行耗时）
 - Submission 使用 `version` 字段（乐观锁）替换基于 status 的条件更新，防止误报冲突
 - Prometheus 冲突计数器：`submission_conflicts_total`、`judge_run_conflicts_total`
 - 全局请求体大小限制中间件（`MAX_REQUEST_BODY_BYTES`，超限返回 413 `PAYLOAD_TOO_LARGE`）
 - Submission 代码长度限制（`MAX_SUBMISSION_CODE_BYTES`，超限返回 400 `CODE_TOO_LONG`）
 - 启动 Bootstrap 日志：输出 env、版本、最大请求体与代码限制参数
 - OpenAPI：为 Submission 状态更新与内部 JudgeRun start/finish 增补 409 响应；为创建 Submission 增补 413 响应
 - 文档：更新 `metrics.md`（冲突计数器）、`api_errors.md`（409/413/代码超限）、`domain-model.md`（version 乐观锁）
### Changed
 - 迁移 Submission 并发控制：由 `WHERE status=?` 条件更新切换为 `WHERE id=? AND version=?` 乐观锁语义
 - metrics 扩展章节移除已上线的冲突计数器占位
### Deprecated
- 
### Removed
- 
### Fixed
- 
### Security
- 

## [0.1.0] - 2025-09-19
### Added
- 初始项目结构（frontend / backend / python / infra）
- Problem CRUD 与统一响应 envelope
- 健康检查与版本注入 (buildVersion)
- Submission / JudgeRun 领域模型与状态机 + 条件更新防竞态
- RBAC 权限体系与 `judge_run.manage` 权限
- JWT 鉴权（prod 模式强制 / dev/test debug identity）
- 统一错误码体系（api_errors.md）
- Prometheus 最小指标集（HTTP 指标 + 状态转移 counters）
- zap 日志与 trace_id 中间件
- 版本化数据库迁移（goose）与 `AUTO_MIGRATE` 控制
- golangci-lint 集成 + 配置清理
- 文档导航重构、OpenAPI 维护策略、领域模型文档、Roadmap 更新

### Changed
- 调整旧文档结构，集中到分层导航

### Fixed
- 权限矩阵测试添加缺失中间件导致的错误授权
- ineffassign 与 unused type 相关 lint 问题

