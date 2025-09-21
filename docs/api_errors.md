# 统一错误码说明

所有接口返回统一 Envelope：

```
成功: { "data": <payload>, "meta": { ... 可选 }, "error": null }
失败: { "data": null, "error": { "code": "<CODE>", "message": "..." } }
```

## 通用分类

| Code | HTTP 典型状态 | 含义 | 备注 |
|------|----------------|------|------|
| UNAUTHORIZED | 401 | 未登录或身份无效 | 身份缺失/guest |
| FORBIDDEN | 403 | 已登录但无权限 | RBAC 拒绝或资源非 owner |
| INVALID_ID | 400 | Path/参数中 ID 非法 | 空/格式错误 |
| NOT_FOUND | 404 | 泛型资源未找到 | 早期遗留，后续用更具体的 *_NOT_FOUND 覆盖 |
| SUBMISSION_NOT_FOUND | 404 | Submission 不存在 | |
| JUDGE_RUN_NOT_FOUND | 404 | JudgeRun 不存在 | |
| ENQUEUE_FAILED | 500 | JudgeRun 入队失败 | 底层存储错误 |
| LIST_FAILED | 500 | 列表查询失败 | 底层存储错误 |
| INVALID_STATUS | 400 | 提交或运行的目标状态非法 | 值不在允许集合内 |
| INVALID_TRANSITION | 400 | 状态流转不被允许 | 违反状态机规则 |
| CONFLICT | 409 | 并发写入冲突（乐观锁失败） | Submission 版本号不匹配；JudgeRun 条件更新被抢占 |
| PAYLOAD_TOO_LARGE | 413 | 请求体超过全局限制 | 由全局 BodyLimit 中间件返回 |
| CODE_TOO_LONG | 400 | 代码字段超过配置上限 | Create Submission 时校验 `MAX_SUBMISSION_CODE_BYTES` |

## 使用指引

- 4xx 均表示客户端可修复（身份/权限/参数/状态流转）。
- 5xx 表示服务器内部错误（DB/IO/未捕获异常）。
- 新资源专用 *NOT_FOUND 应优先于通用 NOT_FOUND。

## 状态机相关

Submission 状态流转：`pending -> judging -> (accepted|wrong_answer|error)`；其余流转返回 `INVALID_TRANSITION`。状态更新使用“版本号乐观锁”防止并发覆盖：请求必须携带当前版本（服务内部读取后 Compare & Update），冲突返回 `CONFLICT`。

JudgeRun 状态流转：
- `queued -> running -> (succeeded|failed|canceled)`
- 非法目标值返回 `INVALID_STATUS`
- 不符合上述链路返回 `INVALID_TRANSITION`

## 扩展规范

新增错误码时：
1. 在 `internal/http/errcode/errcode.go` 中添加常量与默认消息。
2. 若为领域专用（如 JUDGE_RUN_*），优先使用前缀分类。
3. 在此文档补充表格，说明典型 HTTP 状态与语义。
4. 更新 OpenAPI 文档的响应描述（若影响外部接口）。
5. 涉及并发控制（版本字段、条件更新）需在 `domain-model.md` 与本表中同步说明。

## 示例
并发冲突示例：
```
HTTP/1.1 409 Conflict
{
  "data": null,
  "error": { "code": "CONFLICT", "message": "conflict" }
}
```

请求体超限：
```
HTTP/1.1 413 Payload Too Large
{
  "data": null,
  "error": { "code": "PAYLOAD_TOO_LARGE", "message": "request body too large" }
}
```


```
HTTP/1.1 400 Bad Request
{
  "data": null,
  "error": { "code": "INVALID_TRANSITION", "message": "cannot start: not in queued status" }
}
```
