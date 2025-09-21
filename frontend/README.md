# Codyssey Frontend

基于 Next.js + Tailwind CSS + React Query + 基础 shadcn 风格组件（手动挑选） 的前端初始骨架。

## 已包含内容
- Next.js 15 App Router
- Tailwind v4 inline `@import "tailwindcss"`
- 基础 UI 组件：Button / Input / Spinner / Pagination / Toast Provider
- AuthContext：登录（保存 access_token 到 localStorage）
- React Query Provider（重试策略：仅 5xx + 网络错误，最多 2 次）
- Problems 列表 + 筛选（搜索 / 难度）+ 分页（客户端模拟）
- Problem 详情 `(/problems/[id])`
- 提交创建页面 `(/problems/[id]/submit)`（长度校验 + 409 冲突 Banner + Monaco 编辑器）
- Submissions 列表（按 problemId / status / language 过滤 + 分页）
- Submission 详情 `(/submissions/[id])`（判题运行记录 + 轮询）
- Zod schema 校验（Problem / Submission / SubmissionDetail / JudgeRun）
- 401 自动登出与跳转（apiFetch 分发事件，AuthProvider 监听）
- Submission 非终态轮询刷新（4s 间隔）
- 全局 Toast 通知（成功 / 警告 / 错误）
- Monaco 代码编辑器（动态导入防 SSR 问题）
- Playwright E2E 基础配置与示例用例（problems / submissions 占位）
- 登录页面

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
    components/ui/{button,input,spinner,pagination,toast,toaster}.tsx
    components/code/CodeEditor.tsx
    lib/{utils,error}.ts
```

## 环境变量
在项目根或 frontend 目录创建 `.env.local`：
```
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
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
1. 后端提供真实分页 meta 后：去除前端模拟分页逻辑，直接使用服务端 total/page。
2. 代码模板（按语言为新提交预填样板）。
3. Token 刷新（若后端提供 refresh token）。
4. SSE / WebSocket 实时推送判题状态，移除轮询。
5. 细化 E2E：补充登录（若需要表单）、提交成功流程、冲突处理分支。
6. 访问控制（基于角色隐藏/禁用操作）。
7. Editor 增强：快捷运行本地测试（需沙箱后端接口支持）。

## 分页说明
目前 Problems / Submissions 均在获取全集后于前端执行过滤 + 切片，以 `meta.filtered` 提供过滤后数量。待后端分页实现后：
```
useProblems({ page, pageSize, ... }) => 返回服务端 total/hasNext 等字段
```
替换掉前端切片与 total 计算逻辑即可。

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

在提交页面：
- 自动记忆上一次选择的语言（localStorage: `submit_language`）。
- 语言切换后若当前代码为空，会显示“填充模板”提示；点击按钮插入默认模板。

扩展步骤：
1. 在 `codeTemplates.ts` 添加新语言常量与映射。
2. 更新联合类型 `SupportedLanguage` 与数组校验。
3. 在提交页 `<select>` 补充选项，并在 `CodeEditor` 的 `mapLanguage` 增加映射。
4. 若需要高亮需确认 Monaco 内置支持或引入额外语言定义。


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

## 许可证
内部项目（待补充）。
