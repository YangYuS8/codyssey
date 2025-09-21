# Codyssey Frontend

基于 Next.js + Tailwind CSS + React Query + 基础 shadcn 风格组件（手动挑选） 的前端初始骨架。

## 已包含内容
- Next.js 15 App Router
- Tailwind v4 inline `@import "tailwindcss"`
- 基础 UI 组件：Button / Input / Spinner
- AuthContext：登录（保存 access_token 到 localStorage）
- React Query Provider（重试策略：仅 5xx + 网络错误，最多 2 次）
- Problems 列表页面（调用 `/problems`）
- Submissions 占位页面
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
    auth/auth-context.tsx
    hooks/useProblems.ts
    components/ui/{button,input,spinner}.tsx
    lib/utils.ts
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
1. 添加 `/problems/[id]` 详情页 + Hook。
2. 添加创建 Submission 表单（代码长度前端校验）。
3. 增加 401 自动刷新/跳转逻辑（集中在 apiFetch 拦截）。
4. 引入更多 shadcn 组件（Dialog / DropdownMenu / Table）。
5. 处理 409 冲突 Banner（Submission 状态更新时）。
6. 使用 zod 对请求/响应做 schema 校验与类型推导。
7. 将 meta（分页）从 OpenAPI 同步并在 useProblems 中返回。

## 代码风格约定
- 组件放 `src/components/...`；业务数据 Hook 放 `src/hooks`。
- API 客户端一律使用 `apiFetch`，禁止直接裸用 `fetch`（除非无 envelope）。
- 错误码处理集中在后续可添加的 `src/lib/error.ts`。

## 许可证
内部项目（待补充）。
