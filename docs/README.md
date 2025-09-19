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
| [go-backend.md](go-backend.md) | Go 后端分层、运行流程、配置、日志、迁移、测试策略 |
| [go-backend-api.md](go-backend-api.md) | 当前已实现的公开 API 参考（响应格式、示例） |
| [api_errors.md](api_errors.md) | 统一错误码与使用指南 |
| [auth-roles-permissions.md](auth-roles-permissions.md) | 角色与权限矩阵（RBAC） |
| [metrics.md](metrics.md) | 指标说明（后续将合并到 observability 文档） |

### 2. Domain / 领域模型
| 文件 | 说明 |
| ---- | ---- |
| [domain-model.md](domain-model.md) | Submission / JudgeRun 等状态机与领域对象关系 |

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
| [openapi.yaml](openapi.yaml) | OpenAPI 规范（当前手工维护，后续自动化计划见 `openapi.md`） |
| [openapi.md](openapi.md) | OpenAPI 维护策略与工具链规划 |

### 快速使用场景
- 第一次了解项目：`architecture.md` -> `project-structure.md`
- 新开发者上手：`development.md` -> `go-backend.md` -> `auth-roles-permissions.md`
- 调试 / 指标：`metrics.md` (未来: `observability.md`)
- API 对接：`go-backend-api.md` + `openapi.yaml`
- 错误处理：`api_errors.md`
- 未来判题/状态机：`domain-model.md`（补齐后）

### 待办与占位说明
本导航中标记 (规划) / (待添加) 的文件为即将创建的文档占位，PR 提交时可逐步补充；避免一次性重写导致评审阻塞。

### 贡献文档的建议流程
1. 若是新增主题：先在此导航添加占位行。
2. 编写文档并遵循：标题层级从 `#` 开始、使用表格描述结构化信息、必要时附 Mermaid 图。
3. 若涉及架构决策：使用 ADR 模板放入 `adr/`。
4. 若更改 API：同步 `openapi.yaml` & 更新 `openapi.md` 所描述流程。
5. 提交 PR：在描述中列出“受影响文档”列表。

---
欢迎在 Issue 中提出新增或改进建议。文档质量是演进速度的保障。
