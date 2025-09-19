## 开发环境与运行模式

### 前置依赖

1. Docker + Docker Compose
2. 可选：Go 1.22+、Python 3.12、Node.js 18+（Hybrid 模式下本地运行）

### 模式概览

| 模式 | 描述 | 适用场景 |
| ---- | ---- | -------- |
| Full Docker | 前端 + Go + Python + 基础设施全部容器化 | 快速验收、零本地语言依赖、CI 对齐 |
| Hybrid (推荐) | 仅基础设施容器化，应用本地运行 | 日常开发、热重载、调试效率 |

### Hybrid 启动流程（推荐）
```bash
cp .env.example .env
docker compose -f infra/docker-compose.infra.yml up -d

# Go 后端
go run ./backend

# Python AI
uvicorn python.main:app --reload --port ${PY_BACKEND_PORT:-8000}

# 前端
cd frontend && pnpm install && pnpm dev
```

### Full Docker 启动流程
```bash
cp .env.example .env
cd infra
docker compose up --build -d
```

### 服务端口（默认）

| 服务 | 端口 | 说明 |
| ---- | ---- | ---- |
| Go Backend | 8080 | 可通过 `GO_BACKEND_PORT` 覆盖 |
| Python AI | 8000 | `PY_BACKEND_PORT` |
| Frontend | 3000 | Next.js Dev Server |
| PostgreSQL | 5432 | 用户/密码：`codyssey/codyssey` |
| RabbitMQ 管理 | 15672 | 账号：`guest/guest` |
| MinIO API | 9000 | 账号：`minio/minio123` |
| MinIO Console | 9001 | Web 管理 |

### 常用命令
```bash
# Go 测试
go test ./backend/...

# Python 测试 (若添加 pytest)
cd python && pytest -q

# 关闭基础设施
docker compose -f infra/docker-compose.infra.yml down
```

### 故障排查速览
| 问题 | 可能原因 | 解决 |
| ---- | -------- | ---- |
| 端口占用 | 本地已有服务 | 修改 `.env` 端口或停掉冲突进程 |
| MinIO 起不来 | 9000/9001 占用 | 调整 compose 端口映射 |
| Go 依赖校验失败 | go.sum 错误/缓存 | `go mod tidy` / 清理缓存 |

---
### 目录与约定速览

| 目录 | 说明 |
| ---- | ---- |
| backend | Go 核心 API / 未来判题调度 |
| python | FastAPI AI 服务（生成 / 检测 stub） |
| frontend | Next.js 前端界面 |
| infra | docker-compose 与未来 IaC |

> 更详细结构：`project-structure.md`

### 后续计划 (节选)
详见 `roadmap.md`：判题 Worker、迁移工具、认证、OpenAPI、可观测性。

---
更多章节（测试数据、Mock、自动化脚本）将在后续补充。
