# API Client

统一封装在 `src/api/client.ts`，目标：可靠（超时/重试）、安全（最小泄露）、一致（Envelope）、可演进（后续 codegen）。

## Envelope 约定
成功：
```json
{ "data": <Payload>, "meta": { ... 可选 ... } }
```
失败：
```json
{ "error": { "code": "ERROR_CODE", "message": "人类可读消息" } }
```

## apiFetch 行为
| 步骤 | 描述 |
| ---- | ---- |
| 1 | 组装 headers（含 Authorization Bearer <token> 可选） |
| 2 | 启动超时计时器（默认 12s，可通过 NEXT_PUBLIC_API_TIMEOUT_MS 调整） |
| 3 | 发起 fetch，credentials: include 以附带 Cookie |
| 4 | 解析 JSON → 如果是成功：`return envelope.data ?? json` |
| 5 | 失败且 401：尝试 refresh（仅一次）成功后重放原请求 |
| 6 | 失败：构造 ApiError（附条件布尔 helper）并抛出；401 会派发 `auth:unauthorized` 事件 |
| 7 | 超时 / Abort：抛出 code = TIMEOUT |

## GET 重试
- 条件：`httpStatus >= 500` 或网络错误。
- 次数：2（指数退避：300ms, 600ms）。

## ApiError 结构
```ts
interface ApiError extends Error {
  code: string;          // 业务或通用错误码
  httpStatus: number;    // HTTP 状态
  conflict?: boolean;    // code === 'CONFLICT'
  unauthorized?: boolean;// code === 'UNAUTHORIZED'
  forbidden?: boolean;   // code === 'FORBIDDEN'
  payloadTooLarge?: boolean; // code === 'PAYLOAD_TOO_LARGE'
  notFound?: boolean;    // *_NOT_FOUND | NOT_FOUND
}
```

## 常见错误码建议
| 场景 | 建议 code | 前端行为 |
| ---- | ---- | ---- |
| 未认证 | UNAUTHORIZED | 尝试 refresh / 失败登出 |
| 无权限 | FORBIDDEN | 提示 403 |
| 资源缺失 | XXX_NOT_FOUND | 显示 404 状态 |
| 并发冲突 | CONFLICT | Banner or inline 重试按钮 |
| 体积过大 | PAYLOAD_TOO_LARGE | 提示压缩/裁剪 |
| 超时 | TIMEOUT | 可重试 |

## 使用示例
```ts
import { apiGet, apiPost } from '@/src/api/client';

const problem = await apiGet<Problem>(`/problems/${id}`);
await apiPost<Submission, CreateSubmissionBody>(`/submissions`, body);
```

## 后续演进
1. OpenAPI 生成基础 TS + runtime 校验 schema -> 减少手写路径。
2. 引入请求级中间件（链）实现日志、追踪 ID 注入。
3. 增加并发取消：合并外部 AbortSignal（当前占位逻辑可升级）。
4. 对非幂等请求加入幂等键（Idempotency-Key）支持。
