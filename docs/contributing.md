## 贡献指南 (Contributing)

欢迎参与 Codyssey 开发！以下规范有助于维持代码质量与一致性。

### 分支与工作流
| 类型 | 约定命名 | 说明 |
| ---- | -------- | ---- |
| 主干 | `main` | 稳定分支（可部署） |
| 功能 | `feat/<scope>-<short>` | 新增功能 |
| 修复 | `fix/<scope>-<issue-id>` | 缺陷修复 |
| 重构 | `refactor/<scope>` | 不改变外部行为的结构调整 |
| 文档 | `docs/<topic>` | 仅文档变更 |

合并策略：优先使用 **Squash**（保持历史整洁），必要时 Rebase 保持线性。

### Commit Message 规范（简化 Conventional Commits）
格式：`<type>(<scope>): <subject>`

常用 type：`feat` / `fix` / `refactor` / `docs` / `test` / `chore` / `perf` / `build` / `ci`。

示例：
```
feat(problem): add update & delete endpoints
fix(repository): return ErrNotFound on update miss
docs(backend): split API doc to dedicated file
```

避免：超过 72 字符的 subject；模糊描述如 `update code`。

### 代码风格要求

#### Go
* `go fmt` / `go vet` 必须通过。
* 推荐引入 `golangci-lint`（后续集成）。
* 不在 handler 写 SQL；领域逻辑留在 domain / service（后续 service 层演进时）。
* 错误处理：向上返回原始错误并在边界统一转为 API 错误码。
* 避免过早抽象：抽象 ≥2 个使用点再提炼。

#### Python
* 使用 `ruff` 或 `flake8`（后续配置）。
* FastAPI 路由函数需类型注解。
* 避免在路由层执行重逻辑，抽离到 service / utils。

#### 前端 (Next.js)
* 使用 TypeScript，避免 `any`。
* 组件划分：`components/` + `app/` 路由层。
* UI 统一使用 shadcn/ui + Tailwind，避免内联 style。
* 数据获取封装到专门的 `lib/api.ts`（后续再建）。

### 测试要求
| 层级 | 要求 |
| ---- | ---- |
| Handler | 主要路径 + 失败场景（至少 1 个校验失败 / NotFound） |
| 仓储 | Memory 版单测，PG 集成测试后续补充 |
| AI 服务 | Stub 接口返回结构测试 |

新增功能需至少附带 1 个测试；Bug 修复需添加回归测试。

### 提交前自查清单 (Checklist)
```
[ ] go test ./... 通过
[ ] 无未提交的 go.mod / go.sum 变更
[ ] 文档（若影响使用方式）已更新
[ ] 错误码（若新增）符合命名规则
[ ] API 变更已同步 go-backend-api.md
```

### 错误码建议命名
`<DOMAIN>_<ACTION>_<RESULT>` 或通用：`INVALID_REQUEST` / `NOT_FOUND` / `CREATE_FAILED`。

### 性能与安全
* 优先保证正确性，其次再做性能优化。
* 涉及用户输入的功能需考虑：注入、越权、资源滥用（后续提供安全基线文档）。

### 开发工具建议
| 目的 | 工具 | 备注 |
| ---- | ---- | ---- |
| 接口调试 | HTTPie / curl / Bruno | 轻量快速 |
| 负载测试 | k6 / vegeta | 规划阶段 |
| 日志观察 | jq / lnav | 分析结构化日志 |

### 讨论与决策
* 架构变更：开 issue 标记 `discussion`。
* 快速决策：短期 Slack/IM 讨论后需回填 issue 结论。

---
欢迎 PR！发现缺失规范可先提 issue 讨论再补文档。
