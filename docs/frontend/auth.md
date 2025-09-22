# Authentication & Authorization

## 当前实现
| 维度 | 实现 | 说明 |
| ---- | ---- | ---- |
| 登录 | /auth/login -> access_token | 存储于 localStorage (key: access_token) + Cookie(auth_token) 供 middleware 校验 |
| 刷新 | /auth/refresh (POST) | 401 时触发一次；成功后重放原请求；失败派发事件 -> 登出 |
| 登出 | auth-context logout() | 清理 token + 用户信息；跳转 login |
| 角色 | roles Cookie + context.roles | middleware 与 useRequireRole 双层守卫 |
| 401 监听 | window event: auth:unauthorized | apiFetch 构造并派发；context 捕获执行 logout |

## Token 刷新流程
```
apiFetch -> fetch() -> 401?
  └─ yes -> refreshAccessToken()
         -> 成功? -> 重放原请求
         -> 失败? -> dispatch(auth:unauthorized) -> context logout
```

## 角色守卫
- Edge (middleware.ts): 判断路径是否命中受限前缀（如 /admin），若用户 roles 不包含要求角色 -> 重定向 /login。
- Client (useRequireRole): 渲染期再次校验，提升 UX（可改为提示 403 页面）。

示例：
```ts
useRequireRole('ADMIN');
```

## 待改进
1. 改为 **HttpOnly Cookie + SameSite**，前端不再显式存储 access token。
2. 引入刷新提前调度：token 剩余 < 2 分钟时主动刷新，降低突发 401。
3. 增加后端下发 `expiresAt` 字段，避免前端解析 JWT。
4. 为 roles 设计枚举与映射（防止魔法字符串）。
5. 增加 `<Authorized roles={['ADMIN']}>...</Authorized>` 组件封装条件渲染逻辑。

## 错误与权限场景行为
| 场景 | 行为 |
| ---- | ---- |
| 401 且刷新失败 | 派发事件 -> logout -> 重定向登录页 |
| 403 | 展示“无权限”提示 / 可重定向 dashboard |
| 404 | 资源不存在，返回轻量文案 |
| 409 | 前端显示冲突 Banner（提交重复等并发冲突） |

