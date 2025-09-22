# Frontend Development Guide

## 安装与启动
```bash
pnpm install
pnpm --filter frontend dev
# 或
cd frontend && pnpm install && pnpm dev
```
访问: http://localhost:3000/login

## 目录约定
| 路径 | 用途 |
| ---- | ---- |
| app/ | 页面与布局（App Router） |
| src/api | API 客户端与 schema 定义 |
| src/auth | AuthContext / 事件监听登出 |
| src/hooks | 业务数据 Hooks |
| src/components | UI + 代码编辑器组件 |
| src/lib | 工具、错误映射、模板等 |

## 代码风格
- TypeScript 严格模式；避免 any（必要时通过 // eslint-disable-line 标注原因）
- 命名：数据获取 Hook 以 use<Resource> 或 use<Resource>List 结尾
- 组件 UI 放置 `components/ui/`；复合业务组件可放 `components/feature/`
- 不直接裸用 fetch，统一走 apiFetch 以保持刷新/错误处理一致性

## 表单
- 使用 `react-hook-form` + `@hookform/resolvers/zod`
- schema 放置在 `api/schemas.ts` 或按领域拆分未来 `schemas/` 目录

## 实时调试
- 开启网络面板观察 SSE：/submissions/:id/events
- 模拟断线：Chrome DevTools -> Offline -> 观察重连与轮询切换

## 错误调试
- 在 `src/lib/error.ts` 定义 code -> message 映射
- ApiError 捕获：组件内区分 `unauthorized/notFound/conflict`

## 测试
- Playwright：`pnpm --filter frontend test:e2e`
- 建议新增：提交创建 -> 状态流转 -> SSE 事件出现顺序断言

## 提交规范 (建议)
- feat(frontend): 新增 xxx
- fix(frontend): 修复 xxx
- docs(frontend): 更新文档
- refactor(frontend): 重构不改行为

## 性能优化点
| 场景 | 策略 |
| ---- | ---- |
| 大型列表 | 分页 + Skeleton + 按需加载 |
| SSE 高频 | 批量聚合 / 限流处理 | 
| 首屏 | RSC + 延迟加载 Monaco | 

## 常见问题 FAQ
| 问题 | 排查 |
| ---- | ---- |
| 401 循环 | 检查 refresh 是否成功返回 data.access_token；Cookie 是否跨域受限 |
| SSE 不触发 | 确认后端 CORS + Content-Type: text/event-stream 且不缓存 |
| 超时报错 TIMEOUT | 调整 NEXT_PUBLIC_API_TIMEOUT_MS 或排查后端性能 |

## 后续
详见 `roadmap.md` 与本目录其他文件（architecture / realtime / auth）。
