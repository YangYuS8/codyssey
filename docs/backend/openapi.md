<!-- 合并迁移：原位置 docs/openapi.md 已移至此，更新引用请使用 docs/backend/openapi.md -->

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
| Phase 3 | CI 中强制路由与 OpenAPI 差异检测、变更审计 |

## 3. 命名与约定
- Path：小写复数资源名：`/problems`, `/submissions`, `/judge-runs`
- 错误响应结构：
```json
{
  "data": null,
  "error": { "code": "<CODE>", "message": "..." }
}
```
- 添加字段：避免破坏性变更（新增 → 客户端迁移 → 移除旧字段）。

## 4. 更新流程 (Phase 1)
1. 修改/新增 Handler。
2. 更新 `openapi.yaml`：Path/Method/参数/RequestBody/Responses。
3. 本地校验：`spectral lint docs/openapi.yaml`（示例工具）。
4. PR 描述附：`Docs: Updated openapi.yaml for <endpoint>`。
5. Review 检查：响应 schema 一致性 / 错误码列出。
6. （规划）公共接口变更同步 `CHANGELOG`。

## 5. 未来自动化 (Phase 2/3)
| 目标 | 方案 | 工具候选 |
| ---- | ---- | -------- |
| 自动生成骨架 | 注释/契约优先生成 paths + schemas | swag / oapi-codegen |
| 反向校验未文档化路由 | 运行时枚举 Gin 路由 vs OpenAPI | 自定义脚本 |
| JSON Schema 校验 | 启动阶段 Round Trip 测试 | validator + golden tests |
| 变更审计 | CI diff 检测 breaking 删除 | GitHub Action + JSON Patch |

## 6. 差异检测脚本（`routecheck`）
位置：`backend/cmd/routecheck`

运行：
```bash
go run ./backend/cmd/routecheck > route_diff.json
```
参数：
- `--openapi docs/openapi.yaml` 指定路径
- `--include-metrics` 包含默认忽略的 `/metrics`

输出结构：
```json
{
  "implemented_only": [],
  "documented_only": [],
  "method_mismatches": []
}
```

CI 集成策略：
1. 观察期：差异非空 -> 警告构件
2. 稳定期：差异非空 -> 直接失败
3. 白名单：内部路由前缀忽略

改进路线：
- 生成缺失最小模板片段
- 覆盖率统计 (`documented/implemented`)
- 可配置忽略前缀

> 当前仅对 (path, method) 检测，不比较 schema；Schema 校验放入后续阶段。

## 7. 版本与发布策略
- 语义化版本（规划）：`MAJOR.MINOR.PATCH`
- 变更分类：Added / Changed / Deprecated / Removed / Fixed / Security

## 8. FAQ
| 问题 | 说明 |
| ---- | ---- |
| 忘记更新 | PR 模板 checklist + routecheck 提示 |
| 文档太大 | 拆分片段再合并（后续自动化） |
| 响应示例走样 | 使用真实测试响应（脱敏）更新 |

## 9. 参考与后续
- 工具：`swag`, `oapi-codegen`, `spectral`
- 流程：契约优先 → 生成 Server / Client Stubs → 实现

---
最后更新：2025-09-22（迁移版）
