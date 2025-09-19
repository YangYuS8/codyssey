# OpenAPI 维护策略 (openapi.yaml)

## 1. 现状
- 文件位置：`docs/openapi.yaml`
- 维护方式：手工更新（Manual Source of Truth）
- 同步风险：代码 Handler 与文档可能漂移 → 需要最小流程保障。

## 2. 目标
| 阶段 | 目标 |
| ---- | ---- |
| Phase 1 (当前) | 手动维护 + 校验规范一致性 |
| Phase 2 | 引入结构化注释或代码生成工具 (swag / chi-openapi / oapi-codegen) |
| Phase 3 | 在 CI 中强制路由与 OpenAPI 差异检测、变更审计 |

## 3. 命名与约定
- Path 统一使用小写复数资源名：`/problems`, `/submissions`, `/judge-runs`
- 错误响应统一结构：
```json
{
  "data": null,
  "error": { "code": "<CODE>", "message": "..." }
}
```
- 添加新字段时：避免破坏性变更（先新增、再迁移客户端、最后删旧字段）。

## 4. 更新流程 (Phase 1)
1. 修改/新增 Handler。
2. 更新 `openapi.yaml`：
   - Path + 方法 + 参数 + RequestBody + Responses
   - 示例：成功与失败（至少一条 4xx）
3. 本地用在线或 CLI 工具校验：
   - 推荐工具：`spectral lint docs/openapi.yaml`
4. PR 描述中加入：`Docs: Updated openapi.yaml for <endpoint>`。
5. Code Review 检查点：
   - Response schema 是否与代码返回结构匹配
   - 错误码是否已在 `api_errors.md` 列入
6. 合并后：若是公开接口，考虑更新 `CHANGELOG` (规划中)。

## 5. 未来自动化 (Phase 2/3)
| 目标 | 方案 | 工具候选 |
| ---- | ---- | -------- |
| 自动生成骨架 | 基于注释/tag 生成 paths + schemas | swag (注释式) |
| 反向校验未文档化路由 | 运行时枚举 Gin 路由对比 OpenAPI | 自定义脚本 (`go list + reflect`) |
| JSON Schema 校验 | 启动时针对关键结构执行 Round Trip 测试 | go-playground/validator + golden tests |
| 变更审计 | CI Diff 检测有 breaking 字段删除 | GitHub Action + JSON Patch 分析 |

## 6. 差异检测脚本（`routecheck` 已实现）
位置：`backend/cmd/routecheck`

用途：比较运行时注册的 Gin 路由与 `docs/openapi.yaml` 中声明的 paths，发现：
1. `implemented_only`: 代码里有但文档缺失的 (path+method)
2. `documented_only`: 文档写了但代码不存在的 (path+method)
3. `method_mismatches`: 同一路径方法集合不一致（列出实现与文档各自的完整集合）

运行示例（在 `backend` 目录下）：
```
go run ./cmd/routecheck > route_diff.json
```
参数：
- `--openapi ../docs/openapi.yaml` 可指定自定义路径
- `--include-metrics` 包含默认忽略的 `/metrics` 端点

输出 JSON 结构：
```json
{
  "implemented_only": [ { "path": "/submissions/{id}/status", "method": "PATCH" } ],
  "documented_only": [ { "path": "/submissions/{id}", "method": "PATCH" } ],
  "method_mismatches": []
}
```

集成建议：
1. 短期（观察阶段）：在 CI 中执行并把结果文件作为构件；若 `implemented_only` 或 `documented_only` 非空则提示 Warning。
2. 稳定后：若差异非空直接 `exit 1` 使 CI 失败，迫使同步文档。
3. 可后续加白名单（例如内部调度接口前缀 `/internal/`）或生成 markdown 注释到 PR。

技术要点：
- 通过 `router.Setup()` 在 development 模式下注册全部路由（无需依赖真实 Repo）。
- 将 Gin 的 `:id` 形式规范化为 OpenAPI `{id}`。
- 解析 OpenAPI 仅关心 `paths` → method 键集合。
- 忽略 `/metrics`（可通过 flag 包含）。

未来改进：
- 支持输出“建议添加的最小 OpenAPI 片段模板”。
- 统计覆盖率：`documented_paths / implemented_paths` 百分比。
- 可选忽略前缀列表（通过 flag）。

> 注意：当前脚本不会解析 schema 级别差异，只针对 (path, method) 覆盖度。Schema 结构对齐将放到 Phase 2/3（参见下节）。

## 7. 版本与发布策略
- 语义化版本（规划）：`MAJOR.MINOR.PATCH`
- 变更分类：
  * Added: 新增非破坏接口
  * Changed: 兼容性行为改变（需挪步）
  * Deprecated: 标记将废弃
  * Removed: 破坏性删除（仅大版本）
  * Fixed / Security: 修复 / 安全补丁

## 8. 常见问题 (FAQ)
| 问题 | 说明 |
| ---- | ---- |
| 忘记更新 openapi.yaml | PR 模板中加入检查项；代码审核中列出受影响端点 |
| 文档结构巨大不易维护 | 拆分为多个片段再合并（如后续引入生成脚本） |
| 响应示例不一致 | 建议用真实测试响应捕获后贴入（脱敏） |

## 9. 参考与后续
- 工具评估：
  * `swag` 快速但侵入注释多
  * `oapi-codegen` 更适合“契约优先”
- 若未来采用契约优先：开发流转 → 修改 openapi.yaml → 生成 stubs → 实现 Handler。

---
更新本文件请遵循“阶段”语义，不直接删除历史阶段，而是在其下补充完成标记。
