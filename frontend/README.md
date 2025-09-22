# Codyssey Frontend

基于 Next.js + Tailwind CSS + React Query + 基础 shadcn 风格组件（手动挑选） 的前端初始骨架。

## 已包含内容
基础功能：
- Next.js 15 App Router
- Tailwind v4 inline `@import "tailwindcss"`
- 基础 UI 组件：Button / Input / Spinner / Pagination / Toast Provider / Skeleton / StatusBadge / 表单抽象 (FormField / FormInput)
- AuthContext：登录（localStorage token）+ 401 事件驱动登出
- React Query Provider（重试策略：仅 5xx + 网络错误，最多 2 次）
- Problems 列表 + 筛选（搜索 / 难度）+ 分页（客户端模拟或服务端，支持 URL 同步）
- Problem 详情 `(/problems/[id])`
- 提交创建页面 `(/problems/[id]/submit)`（长度校验 + 409 冲突 Banner + Monaco 编辑器 + 代码模板/语言记忆）
- Submissions 列表（按 problemId / status / language 过滤 + 分页，URL 同步）
- Submission 详情 `(/submissions/[id])`（判题运行记录 + 轮询 + SSE 优先刷新）
- Zod schema 校验（Problem / Submission / SubmissionDetail / JudgeRun）
- Skeleton 骨架占位（列表 & 详情加载）
- 全局 Toast 通知（成功 / 警告 / 错误）
- Monaco 代码编辑器（动态导入防 SSR 问题）
- Playwright E2E 基础配置与示例用例
- 登录页已迁移 React Hook Form + Zod 校验

Phase 1 新增：
- 服务端分页特性开关：`NEXT_PUBLIC_SERVER_PAGINATION=true` 时直接调用带分页参数接口
- 中间件 Cookie 认证跳转（`middleware.ts`，读取 `auth_token`）
- Web Vitals 采集：在 `app/layout.tsx` 导出 `reportWebVitals`（可通过 `NEXT_PUBLIC_LOG_WEB_VITALS=true` 控制 console 输出）
- SSE Hook：`useSubmissionEvents` 监听提交状态，实时刷新详情；连接失败自动重连，失联时仍由轮询兜底
- 表单体系：`react-hook-form` + `zodResolver` + 统一表单组件（后续可快速扩展其他表单）

## 目录结构（节选）
```
frontend/
  app/
    layout.tsx
    globals.css
    login/page.tsx
    problems/page.tsx
    submissions/page.tsx
  src/
    api/client.ts
    api/schemas.ts
    auth/auth-context.tsx
    hooks/{useProblems,useProblem,useSubmissions,useSubmission,useCreateSubmission}.ts
    components/ui/{button,input,spinner,pagination,toast,toaster,skeleton,status-badge}.tsx
    components/code/CodeEditor.tsx
    lib/{utils,error}.ts
```

## 环境变量
在项目根或 frontend 目录创建 `.env.local`：
```
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_SERVER_PAGINATION=false   # 打开后直接请求服务端分页
NEXT_PUBLIC_LOG_WEB_VITALS=false      # 打开后在控制台输出 Web Vitals
```

## 启动
```bash
pnpm install
pnpm --filter frontend dev
```
或：
```bash
cd frontend
pnpm install
pnpm dev
```

访问：`http://localhost:3000/login` 登录后跳转 `Problems`。

> 注意：当前登录接口依赖后端已有 `/auth/login`，请确保后端已运行并允许跨域（若同域部署则无需额外配置）。

## 下一步建议
1. 后端统一返回分页 envelope（确认字段：`data[]`, `meta: { page,pageSize,total,filtered }`）后移除前端回退路径。
2. Access Token 换成 HttpOnly Cookie（前端移除 localStorage 持久化）。
3. Refresh Token & 静默续期（中间件 + 客户端 401 重试）。
4. SSE 事件类型细化（区分 judge run 增量 vs 最终结果，支持乐观更新局部列表）。
5. E2E：增加提交完整流程、错误态、SSE 实时状态断言。
6. 角色/权限：UI 级别禁用与后端鉴权对齐。
7. 统一日志/埋点：上报前端错误与交互事件。
8. Monaco: 支持自动补全 / LSP 代理（需后端沙箱 API）。

## 分页 & URL 同步说明
开启服务端分页：
```
NEXT_PUBLIC_SERVER_PAGINATION=true
```
Hook `useProblems` / 后续 `useSubmissions` 将直接携带 `page,pageSize,search,difficulty` 请求；关闭时回退到客户端全集过滤 + 切片。

URL 同步：页面内部 state 变更后用 `router.replace` 写回 query，便于刷新/分享保持上下文。

## Toast 使用
通过 `useToast()` 获取 `push` 方法：
```
push({ variant: 'success' | 'error' | 'warning', title: '标题', description?: '描述' })
```

## Monaco 编辑器
组件：`src/components/code/CodeEditor.tsx` 动态导入 `@monaco-editor/react`，自动处理 SSR。语言映射：cpp/python/go/java；未知语言降级为 plaintext。

## 代码模板
文件：`src/lib/codeTemplates.ts`

支持语言：`cpp | python | go | java`

提交页特性：
- 记忆上次语言（`localStorage: submit_language`）
- 空代码时的“填充模板”提示

扩展步骤同前：更新模板映射 + 联合类型 + 选择器与 Monaco 语言映射。

## Web Vitals
在 `app/layout.tsx` 导出 `reportWebVitals`。设置：
```
NEXT_PUBLIC_LOG_WEB_VITALS=true
```
即可在浏览器控制台看到 FCP / LCP / CLS / INP / TTFB / FID 指标。后续可用 `navigator.sendBeacon` 上报到后端。

## SSE 实时更新
Hook：`useSubmissionEvents(submissionId)` -> 建立 `EventSource /submissions/{id}/events`。
- 断线自动重连（默认 5s）
- 与轮询共存：SSE 连接成功后刷新数据，终态后自动停止轮询；若 SSE 不可用继续轮询兜底。
- 可扩展事件：`status_update`, `judge_run_update`, `completed`。



## E2E 测试
基础目录：`tests/e2e`，使用 Playwright。

运行：
```
pnpm --filter frontend test:e2e
```
可通过环境变量 `E2E_BASE_URL` 指向已运行的服务，如：
```
E2E_BASE_URL=http://localhost:3000 pnpm --filter frontend test:e2e
```

## 代码风格约定
- 组件放 `src/components/...`；业务数据 Hook 放 `src/hooks`。
- API 客户端一律使用 `apiFetch`，禁止直接裸用 `fetch`（除非无 envelope）。
- 错误码 → 用户文案映射在 `src/lib/error.ts`。
- 受保护页面调用 `useRequireAuth()`（后续推荐迁移到 middleware + Cookie 方案）。

## 许可证
内部项目（待补充）。
