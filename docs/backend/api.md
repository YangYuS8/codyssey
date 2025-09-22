# Backend API Guide

## 响应 Envelope
成功：
```json
{ "data": {"...": "..."}, "error": null }
```
失败：
```json
{ "data": null, "error": {"code": "NOT_FOUND", "message": "..."} }
```
错误码详见 `../api_errors.md`。

## 典型资源：Problem
| 方法 | 路径 | 描述 |
| ---- | ---- | ---- |
| POST | /problems | 创建题目 |
| GET | /problems | 列表（分页/过滤 规划中） |
| GET | /problems/{id} | 获取单题 |
| PUT | /problems/{id} | 更新 |
| DELETE | /problems/{id} | 删除 |

健康检查：`GET /health`；版本：`GET /version`。

## OpenAPI 契约
- 文件：`docs/openapi.yaml`
- 维护策略：参见 `./openapi.md`
- 差异检测：`go run ./backend/cmd/routecheck --openapi ../docs/openapi.yaml`

## 扩展计划
| 方向 | 内容 |
| ---- | ---- |
| 分页统一 | limit/offset -> page/pageSize 标准化元信息 |
| 过滤参数 | 通用结构：`?filter[field]=value` 或 `?status=xxx` 简化方案 |
| 排序 | `sort=created_at,-title` 多字段支持 |
| 部分更新 | PATCH + JSON Merge / JSON Patch（评估） |

## 版本策略
- 少量新增字段：后端向前兼容，客户端逐步采用
- 删除字段：标记 deprecated → N 版本后移除

## 示例（创建 Problem）
Request:
```bash
curl -X POST http://localhost:8080/problems \
  -H 'Content-Type: application/json' \
  -d '{"title":"Two Sum","description":"..."}'
```
Response:
```json
{ "data": { "id": "uuid", "title": "Two Sum" }, "error": null }
```

更多状态机相关接口：参考 `domain-model.md`。
