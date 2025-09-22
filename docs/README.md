## 文档导航 (Documentation Index)

> 本目录为文档“总导航”。按角色（使用者 / 开发 / 运维 / 架构 / 管理）与主题分层，便于快速定位。

### 0. Overview / 总览
| 文件 | 说明 |
| ---- | ---- |
| [architecture.md](architecture.md) | 系统整体架构、核心组件、演进阶段 (含 Mermaid 图) |
| [project-structure.md](project-structure.md) | 仓库目录结构与分层职责 |
| [roadmap.md](roadmap.md) | 路线图与阶段目标、技术债跟踪 |

### 1. Development / 开发
| 文件 | 说明 |
| ---- | ---- |
| [development.md](development.md) | 本地 (Hybrid) / 全容器 (Full) 启动、端口、常用命令、故障排查 |
| [contributing.md](contributing.md) | 贡献流程、提交规范、分支策略、代码与测试要求 |
| [api_errors.md](api_errors.md) | 统一错误码与使用指南（后端实现详见 `backend/errors.md`） |
| [auth-roles-permissions.md](auth-roles-permissions.md) | 角色与权限矩阵（顶层概览，实现详见 `backend/permissions.md`） |
| Frontend: [frontend/overview.md](frontend/overview.md) | 前端概览、技术栈、能力矩阵 |
| Frontend: [frontend/architecture.md](frontend/architecture.md) | 前端架构分层、数据流、扩展策略 |
| Frontend: [frontend/auth.md](frontend/auth.md) | 前端认证 / 刷新 / 角色守卫机制 |
| Frontend: [frontend/api-client.md](frontend/api-client.md) | API 客户端封装、错误与重试策略 |
| Frontend: [frontend/realtime.md](frontend/realtime.md) | SSE + 轮询协调、事件模型与演进 |
| Frontend: [frontend/environment.md](frontend/environment.md) | 前端环境变量说明 |
| Frontend: [frontend/development.md](frontend/development.md) | 前端开发指南 / FAQ / 提交规范 |
| Backend: [backend/overview.md](backend/overview.md) | 后端特性概览与能力矩阵 |
| Backend: [backend/architecture.md](backend/architecture.md) | 后端分层架构、请求生命周期、并发策略 |
| Backend: [backend/development.md](backend/development.md) | 后端开发环境、命令、调试与 FAQ |
| Backend: [backend/api.md](backend/api.md) | API 端点列表、统一 Envelope、示例与演进计划 |
| Backend: [backend/metrics.md](backend/metrics.md) | 后端指标（统一权威，替代已删除旧 `metrics.md`） |
| Backend: [backend/errors.md](backend/errors.md) | 错误码实现细节与映射策略 |
| Backend: [backend/permissions.md](backend/permissions.md) | RBAC 实现与缓存策略 |
| Backend: [backend/domain-model.md](backend/domain-model.md) | 领域模型 / 状态机 / 并发控制 |
| Backend: [backend/observability.md](backend/observability.md) | 日志 / 指标 / Tracing / Alert 路线图 |
| Backend: [backend/testing.md](backend/testing.md) | 测试分层策略与示例 |

### 2. Domain / 领域模型
| 文件 | 说明 |
| ---- | ---- |
| [backend/domain-model.md](backend/domain-model.md) | Submission / JudgeRun 状态机、并发控制与扩展点（取代旧 `domain-model.md`） |

### 3. Operations & Observability / 运行与可观测
| 文件 | 说明 |
| ---- | ---- |
| *(规划) observability.md* | 日志 / 指标 / Tracing / 告警策略统一指引 |
| *(规划) runbook.md* | 常见告警、排障步骤、人工干预操作清单 |
| *(规划) deployment.md* | 部署模式、环境矩阵、配置要求、升级策略 |

### 4. Architecture Decisions / 架构决策
| 文件 | 说明 |
| ---- | ---- |
| [adr/ADR-0001-judging-state-machine-and-concurrency.md](adr/ADR-0001-judging-state-machine-and-concurrency.md) | 判题状态机与并发控制决策 |
| *(规划) 其它 ADR* | 后续新增架构决策记录 |

### 5. Governance & Compliance / 治理与合规（规划）
| 文件 | 说明 |
| ---- | ---- |
| *(规划) security.md* | 安全策略：JWT、权限、密钥管理、最小权限原则 |
| *(规划) docs-maintenance.md* | 文档更新流程、陈旧检测策略、质量守护 |
| [changelog.md](changelog.md) | 版本与变更记录（语义化发布参考） |

### 6. API Schema
| 文件 | 说明 |
| ---- | ---- |
| [openapi.yaml](openapi.yaml) | OpenAPI 规范（维护策略见 `backend/openapi.md`） |
| [backend/openapi.md](backend/openapi.md) | OpenAPI 维护策略与工具链规划 |

### 快速使用场景
- 第一次了解项目：`architecture.md` -> `project-structure.md`
- 新开发者上手：`development.md` -> `backend/overview.md` -> `auth-roles-permissions.md`
- 调试 / 指标：`backend/metrics.md` (可结合 `backend/observability.md`)
- API 对接：`backend/api.md` + `openapi.yaml`
- 错误处理：`api_errors.md` + `backend/errors.md`
- 判题/状态机：`backend/domain-model.md`
- 前端实时 / 刷新：`frontend/realtime.md` + `frontend/api-client.md` + `frontend/auth.md`

### 待办与占位说明
本导航中标记 (规划) / (待添加) 的文件为即将创建的文档占位，PR 提交时可逐步补充；避免一次性重写导致评审阻塞。

> 已删除的旧版聚合文档（`go-backend.md`, `go-backend-api.md`, `metrics.md`, `observability.md`, `domain-model.md`, `openapi.md`）若仍在外部引用，请更新链接指向对应 `backend/` 下的新模块化文件。

### 贡献文档的建议流程
1. 若是新增主题：先在此导航添加占位行。
2. 编写文档并遵循：标题层级从 `#` 开始、使用表格描述结构化信息、必要时附 Mermaid 图。
3. 若涉及架构决策：使用 ADR 模板放入 `adr/`。
4. 若更改 API：同步 `openapi.yaml` & 更新 `openapi.md` 所描述流程。
5. 提交 PR：在描述中列出“受影响文档”列表。

---
欢迎在 Issue 中提出新增或改进建议。文档质量是演进速度的保障。
