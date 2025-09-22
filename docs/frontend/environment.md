# Environment Variables

| 变量 | 示例 | 说明 |
| ---- | ---- | ---- |
| NEXT_PUBLIC_API_BASE_URL | http://localhost:8080 | 后端 API 基地址 |
| NEXT_PUBLIC_SERVER_PAGINATION | false | 是否启用服务端分页（Problems/Submissions） |
| NEXT_PUBLIC_LOG_WEB_VITALS | false | 控制是否在控制台输出 Web Vitals 指标 |
| NEXT_PUBLIC_API_TIMEOUT_MS | 12000 | apiFetch 请求超时（毫秒） |

## 本地配置
在 `frontend/.env.local` 或根 `.env.local`：
```
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
NEXT_PUBLIC_SERVER_PAGINATION=false
NEXT_PUBLIC_LOG_WEB_VITALS=false
NEXT_PUBLIC_API_TIMEOUT_MS=12000
```

> 所有以 NEXT_PUBLIC_ 开头的变量会被暴露到客户端，请避免放置敏感密钥。

## 后续规划
- 引入构建时注入 GIT_SHA / BUILD_TIME 便于调试
- FEATURE FLAG：如 `NEXT_PUBLIC_ENABLE_LIVE_LIST_SSE` 控制列表实时能力
