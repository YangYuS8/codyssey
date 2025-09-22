# Backend Overview

Codyssey 后端基于 Go 构建，目标：清晰分层、可测试、易观测、便于横向扩展判题与竞赛能力。

## 核心特性
- 统一 Envelope 响应: `{ data, error }`
- Submission / JudgeRun 状态机 + 乐观并发控制
- RBAC 权限模型（角色 -> 权限 -> 资源范围二次校验）
- Prometheus 指标 (HTTP / 状态转移 / 冲突 / JudgeRun 时长)
- 版本化数据库迁移 (goose)
- zap 日志 + trace_id 中间件
- OpenAPI 契约文件 + 路由差异校验脚本 (routecheck)

## 技术栈
| 组件 | 说明 |
| ---- | ---- |
| Go 1.x | 主语言 |
| Gin | HTTP 框架 + 路由 |
| pgx / pgxpool | PostgreSQL 驱动与连接池 |
| goose | SQL 迁移策略 |
| zap | 结构化日志 |
| prometheus client | 指标采集 |
| jwt | 认证（调试阶段 stub / 未来对接发行） |

## 目录分层概述
参见 `architecture.md`。

## 当前边界
- 判题 Worker 尚未接入（队列与执行仍为规划）
- Refresh Token / Session 管理策略初稿（与前端协同）
- Observability Tracing 未启用（规划 OTel）

更多演进方向参考 `roadmap.md`。
