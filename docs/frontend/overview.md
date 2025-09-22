# Frontend Overview

本目录描述 Codyssey 前端整体能力、技术选型与现状。

## 技术栈
- Next.js 15 (App Router + React 19)
- Tailwind CSS v4 inline 方式
- React Query 进行数据缓存与请求状态管理
- Zod 运行时数据验证 + TypeScript 静态类型
- 基础 UI：轻量自定义 + shadcn 风格抽取组件 (Button / Input / Pagination / Toast / Skeleton / StatusBadge / Form 抽象)
- Monaco Editor 动态加载（判题代码编辑）
- Playwright E2E 基础测试
- SSE (Server-Sent Events) + 轮询协同的实时更新

## 已实现能力分组
| 领域 | 能力 | 说明 |
| ---- | ---- | ---- |
| 认证与权限 | 登录 / 401 事件登出 / 角色守卫 | localStorage token（待迁移 Cookie）+ middleware 重定向 + useRequireRole Hook |
| 数据获取 | 统一 API 客户端 / Envelope 解析 | apiFetch 封装：超时、401 刷新、GET 重试、错误映射 |
| 列表与分页 | Problems / Submissions | 支持客户端 fallback 与服务端分页开关 NEXT_PUBLIC_SERVER_PAGINATION |
| 实时 | Submission 详情 | SSE 事件增量 + 轮询兜底 + 枚举化事件类型 |
| 表单 | 登录、提交代码 | react-hook-form + zodResolver 组合，统一 FormField/FormInput |
| 反馈 | Toast / Skeleton / StatusBadge | 统一状态标签与异步加载光栅 |
| 代码体验 | Monaco + 模板 | 语言记忆、空内容提示填充模板 |
| 可靠性 | 超时 / 重试 / 刷新 | 超时 12s 默认，可配置；GET 指数退避 2 次；刷新一次后不递归 |
| 观测 | Web Vitals | reportWebVitals + 环境变量开关日志输出 |

## 快速特性总结
- Envelope 解包：所有成功请求 prefer `data` 字段；失败统一构造 ApiError。
- 增量实时：仅对变更字段进行 React Query setQueryData（减少不必要网络）。
- 角色保护：服务端（edge middleware）+ 客户端（Hook）双层。
- 内聚错误层：错误码 → 用户文案集中映射（lib/error.ts），便于国际化扩展。

## 仍在计划（详见 roadmap）
- Cookie + Refresh Token 静默续期
- SSE 列表级增量广播
- 前端监控 / 错误上报管线
- 更细粒度的权限组件包装（<Authorized />）

> 详细架构与交互流程参见 architecture.md 与 realtime.md。
