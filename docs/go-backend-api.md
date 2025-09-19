## Go 后端 API 参考

基础路径：当前直接由 Gin 提供（后续可挂载在 `/api` 前缀）。

### 响应格式
成功：
```json
{ "data": <任意>, "meta": { ...可选... }, "error": null }
```
失败：
```json
{ "data": null, "error": { "code": "NOT_FOUND", "message": "problem not found" } }
```

### Problem 资源

#### 创建题目
POST /problems
```json
Request: { "title": "Two Sum", "description": "Given an array..." }
Response: { "data": { "id": "<uuid>", "title": "Two Sum", ... }, "error": null }
```
错误：`400 INVALID_REQUEST` / `500 CREATE_FAILED`

#### 获取题目列表
GET /problems?limit=20&offset=0
```json
{ "data": [ {"id": "..."} ], "meta": {"limit":20,"offset":0,"count":1}, "error": null }
```

#### 获取单题
GET /problems/:id
```json
{ "data": { "id": "...", "title": "..." }, "error": null }
```
错误：`404 NOT_FOUND`

#### 更新题目
PUT /problems/:id
```json
Request: { "title": "New Title" }
Response: { "data": { "id": "...", "title": "New Title" }, "error": null }
```

#### 删除题目
DELETE /problems/:id
```json
{ "data": { "deleted": "<uuid>" }, "error": null }
```

### 健康检查
GET /health
```json
{ "status": "ok", "db": "up", "version": "<hash>", "env": "dev" }
```

### 版本
GET /version
```json
{ "version": "<hash>" }
```

---
后续计划：统一错误码文档、OpenAPI 自动生成、分页/过滤参数说明统一化。
