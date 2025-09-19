# 开发环境启动指南

## 前置依赖

1. 已安装 Docker 与 Docker Compose
2. 可选：本地安装 Go 1.22+ 与 Python 3.12 以便本地调试（否则使用容器）

## 模式概览

当前支持两种开发模式：

| 模式 | 描述 | 适用场景 |
| ---- | ---- | -------- |
| Full Docker | 前端 + Go + Python + 基础设施全部容器化 | 快速验收、零本地依赖 |
| Hybrid (推荐) | 仅基础设施容器化（PostgreSQL/RabbitMQ/MinIO），应用本地运行 | 日常开发、热重载、调试 |

---

## 模式一：Full Docker 全容器

```bash
cp .env.example .env  # 如需自定义再编辑
cd infra
docker compose up --build -d
```

启动后服务：

- Go 后端: http://localhost:8080/health
- Python AI 服务: http://localhost:8000/health
- 前端: http://localhost:3000
- PostgreSQL: localhost:5432 （用户/密码：codyssey/codyssey）
- RabbitMQ 管理界面: http://localhost:15672 （guest/guest）
- MinIO 控制台: http://localhost:9001 （minio/minio123）

---

## 模式二：Hybrid 本地应用 + Docker 基础设施

1. 启动基础设施：

```bash
cp .env.example .env   # 若尚未创建
cd infra
docker compose -f docker-compose.infra.yml up -d
```

2. 本地启动 Go 后端：

```bash
GO_BACKEND_PORT=${GO_BACKEND_PORT:-8080} go run ./backend
```

或（进入目录）：

```bash
cd backend
go run .
```

3. 本地启动 Python AI 服务（自动热重载）：

```bash
cd python
uvicorn main:app --reload --port ${PY_BACKEND_PORT:-8000}
```

4. 本地启动前端：

```bash
cd frontend
pnpm install  # 或 yarn / npm
pnpm dev
```

5. 验证：

```bash
curl http://localhost:8080/health
curl http://localhost:8000/health
```

6. 停止基础设施：

```bash
cd infra
docker compose -f docker-compose.infra.yml down
```

> 如果需要临时切回全容器模式，只需运行 `docker compose up -d`（默认使用完整文件）。

---

## 常用命令（Full Docker 模式）

```bash
# 查看日志
cd infra
docker compose logs -f go-backend

# 重建某个服务
docker compose up --build -d go-backend

# 运行 Go 测试（容器外本地）
(cd backend && go test ./...)

# 运行 Python 测试（容器外本地）
(cd python && pytest -q)
```

## 目录约定

- backend: Go (Gin) 核心 API / 判题调度入口（后续会扩展 MQ、DB、MinIO 接入）
- python: FastAPI AI 服务（题目生成 + 检测），当前是 stub
- infra: docker-compose 及后续 IaC（可扩展 K8s Manifests / Terraform）

## 下一步规划

- 判题 Worker 与 Judge0 集成
- 数据库迁移工具（如 golang-migrate）
- 统一 API Gateway / Auth 中间件
- AI 生成与检测真实实现

## 故障排查

1. 端口冲突：修改 `.env` 中端口或 compose 映射
2. Python 依赖安装慢：可配置国内镜像源
3. MinIO 启动失败：确认本地 9000/9001 未被占用
