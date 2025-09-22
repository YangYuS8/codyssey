# Backend Development Guide

## 本地运行
```bash
cd backend
# 运行（带版本注入示例）
go build -ldflags "-X main.buildVersion=$(git rev-parse --short HEAD)" -o bin/backend ./backend
./bin/backend
```
或直接：
```bash
go run ./backend
```

## 环境变量 (示例)
| 变量 | 说明 |
| ---- | ---- |
| APP_ENV | dev / prod 决定日志模式等 |
| HTTP_PORT | 监听端口 (默认 8080) |
| DATABASE_URL | PostgreSQL 连接串 |
| AUTO_MIGRATE | 是否自动执行 goose 迁移 (true/false) |
| JWT_SECRET | 开发调试用 secret (正式需轮换策略) |

`.env.example` 中应列出全部必要变量（若不存在可补充）。

## 常用命令
| 操作 | 命令 |
| ---- | ---- |
| 运行测试 | `go test ./...` |
| Lint | `golangci-lint run` |
| 迁移创建 | `goose create add_table_submission sql` |
| 迁移执行 | `goose up` |
| 路由检查 | `go run ./backend/cmd/routecheck` |

## 代码风格约定
- handler 不写 SQL / 不包含业务核心逻辑
- repository 不依赖 gin / 仅使用 context + 领域模型
- 错误统一返回 Go `error`，上层映射为错误码
- 响应必须使用 Envelope：`{"data":...,"error":null}` 或 `{"data":null,"error":{...}}`

## 调试技巧
| 场景 | 手段 |
| ---- | ---- |
| 查看路由 | `go run ./backend/cmd/routecheck | jq` |
| DB 连接失败 | 确认 `DATABASE_URL`，使用 `psql` 手动连接 |
| 状态流转异常 | 打开 DEBUG 日志，观察 from->to 记录 |
| 性能慢 | 先看 P95 延迟 (`/metrics`) 再定位特定 route |

## FAQ
| 问题 | 说明 |
| ---- | ---- |
| 迁移失败退出 | 检查 goose 版本 / SQL 语法；禁用 AUTO_MIGRATE 手动验证 |
| Permission 拒绝 | 确认 JWT roles/permissions claim 是否完整 |
| 指标不输出 | 确认 metrics 中间件注册并访问 /metrics |

更多高级主题：见 `observability.md`, `permissions.md`, `api.md`。
