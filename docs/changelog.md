# Changelog

采用语义化版本（规划中）。格式参考 Keep a Changelog。

## [Unreleased]
### Added
 - 错误码 `CONFLICT`：用于并发/条件更新 0 行场景（返回 HTTP 409）
 - JudgeRun 冲突区分：`UpdateRunning` / `UpdateFinished` 区分不存在与状态冲突，冲突返回 409
 - Histogram 指标：`codyssey_judge_run_duration_seconds`（按终态标签记录运行耗时）
### Changed
- 
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

