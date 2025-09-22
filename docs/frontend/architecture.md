# Frontend Architecture

本文件描述前端的结构分层、核心模块与数据流。

## 分层概览
```
app/ (Next.js App Router 页面 + 布局)
  ├─ layout.tsx (Providers 注入 / Web Vitals / 全局样式)
  ├─ login/            (认证入口)
  ├─ problems/         (题目列表)
  ├─ submissions/      (提交列表 + 详情)
  └─ problems/[id]/submit (提交创建)

src/
  ├─ api/        (client.ts + schemas.ts 统一访问层)
  ├─ auth/       (auth-context.tsx 处理 token / 用户 / 401 登出)
  ├─ hooks/      (业务数据 hooks: useProblems/useSubmission 等)
  ├─ components/ (UI 与代码编辑器等)
  ├─ lib/        (工具与错误映射 / 模板)
  └─ types/      (核心类型统一 - 位于 api/types 或简化合并)
```

## 关键设计点
### 1. App Router + Client Hooks
- 页面对外展示，数据逻辑内聚在自定义 Hooks 中，便于测试与复用。
- 通过 `useQuery` / `useMutation` 保持 SSR 友好（当前多为 Client 组件，可逐步引入 RSC 优化）。

### 2. API 客户端 (api/client.ts)
- 统一 `apiFetch`：附加认证、Envelope 解析、401 -> refresh、GET 指数退避、超时中断。
- 提供 `apiGet/apiPost/apiPatch` 语义化方法；调用方只关心类型与路径。

### 3. 状态与缓存
- React Query 负责数据缓存与失效；SSE 事件 -> 精确 setQueryData。
- 轮询与 SSE 互斥：SSE 成功后关闭轮询，断线再恢复。

### 4. 实时层 (SSE)
- `useSubmissionEvents` 监听单条提交：事件枚举 + 增量合并。
- 未来扩展：列表页接收 summary 事件、push-based invalidation。

### 5. 认证与权限
- `auth-context` 保存当前用户与 roles；监听 `auth:unauthorized` 事件执行 logout。
- `middleware.ts` 在 Edge 层检查 Cookie：`auth_token`、`roles` → 未匹配重定向 `/login`。
- `useRequireRole` 客户端守卫（二次防御 + 渐进体验）。

### 6. 表单与校验
- 采用 `react-hook-form` + Zod：schema 既作为校验又指导类型推断，减少重复定义。

### 7. 错误处理
- ApiError 提供语义标记：`unauthorized/conflict/forbidden/payloadTooLarge/notFound`。
- UI 根据标记决定展示模式（如 409 Banner、404 文案）。

### 8. 可观测性
- Web Vitals：`reportWebVitals` 可收集 FCP/LCP/CLS/INP/TTFB/FID。
- 后续：统一 error boundary + 上报、性能埋点。

### 9. 可扩展性策略
| 方向 | 现状 | 扩展策略 |
| ---- | ---- | ---- |
| 认证 | localStorage + refresh | 迁移双 Cookie + 静默刷新调度器 |
| 实时 | 单资源 SSE | 列表增量 + 合并 patch 协议 |
| 权限 | 路径前缀 + Hook | 语义组件与服务端布局条件渲染 |
| 数据请求 | 手写 path | 生成型 SDK（OpenAPI codegen + 包装保留超时/刷新） |
| 监控 | Web Vitals | Error/Interaction telemetry pipeline |

## 数据流 (Submission 详情)
```
[Page] -> useSubmission (queryKey: submission/:id)
        -> apiGet /submissions/:id

SSE: /submissions/:id/events
  status_update -> setQueryData(status)
  judge_run_update -> merge judgeRuns[]
  completed -> invalidate(queryKey) -> fetch final snapshot
```

## 依赖治理
- 对第三方依赖保持最小集合；Monaco 动态 import 避免 SSR break。
- 后续引入代码生成时需引入前置脚本或 prebuild 步骤，避免运行期增加体积。

## 目录演进建议
- `/src/types` 抽离所有共享 TS 类型，避免 hooks 间循环导入。
- `/src/services`（可选）：聚合更高阶业务操作（组合多个 API 调用与状态变化）。

