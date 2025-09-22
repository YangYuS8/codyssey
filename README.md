# **Codyssey**

> 一个面向算法竞赛与日常教学的在线判题系统（OJ），集成 AI 出题与智能检测功能。

---

## 快速开始

```bash
# 克隆仓库后
cp .env.example .env
cd infra
docker compose up --build -d
```

服务可用性检查：
- Go 后端: `curl http://localhost:8080/health`
- Python AI: `curl http://localhost:8000/health`
- 前端: 打开浏览器访问 http://localhost:3000

更多细节见 `docs/setup.md`。

---

## 项目简介

**Codyssey** 是一个支持 **算法竞赛、日常练习、考试场景** 的在线判题系统。

与传统 OJ 不同，它引入了 AI 技术：

* **AI 出题** ：通过调用大模型 API 自动生成题目，并保存到题库。
* **AI 检测** ：可选开启，对学生提交的代码进行分析，判断是否可能由 AI 自动生成。

系统设计上强调  **高并发、可扩展、解耦合** ，确保能够应对大规模考试/竞赛场景。

---

## 技术栈

### 前端

* **Next.js** ：用于实现竞赛/练习的交互界面。
* **Tailwind CSS**
* **shadcn/ui** （组件库，用于快速构建统一风格的界面）

### 后端

* **Go（Gin 框架）** ：
  * 用户管理、题目管理、比赛管理（当前初始化阶段仅健康检查与占位）
  * 判题任务调度、高并发处理（规划中）
* **Python（FastAPI 框架）** ：
  * AI 出题（调用大模型 API，当前为 stub）
  * AI 检测（深度学习模型，当前为 stub）
  * 智能化题库管理（规划中）

### 基础设施

* **数据库** ：PostgreSQL（存储用户、题目、提交记录等）
* **消息队列** ：RabbitMQ（分发判题任务、异步解耦）
* **对象存储** ：MinIO（存放题目文件、测试数据、用户上传文件）
* **代码沙箱** ：Judge0（Docker 隔离执行，确保安全性，后续添加）

---

> 更完整的系统架构、开发模式、路线图与 API 文档请参见 `docs/` 及子目录：
> - 通用：`architecture.md`, `project-structure.md`, `development.md`, `roadmap.md`
> - 前端：`frontend/overview.md` 等模块化文件
> - 后端：`backend/overview.md`, `backend/architecture.md`, `backend/api.md`, `backend/domain-model.md`, `backend/metrics.md`, `backend/observability.md`

---

## 项目目录结构（概览）
详见 `docs/project-structure.md`。

---

## 核心功能（当前阶段状态）

| 功能 | 说明 | 当前状态 |
| ---- | ---- | -------- |
| 题目管理 | 题目 CRUD / AI 出题 | 基础 CRUD 已实现 / AI stub |
| 竞赛管理 | 比赛创建、排名 | 规划 |
| 判题系统 | 队列 + Worker + Judge0 | 未实现 |
| AI 出题 | 调用大模型生成题目 | stub |
| AI 检测 | 代码风格检测 | stub |
| 健康检查 | /health | 已实现 |

---

## 开发说明
开发模式、常用命令、故障排查详见 `docs/development.md`。

---

## TODO / 路线图
精细路线与阶段目标参见 `docs/roadmap.md`。

---

## License
待补充
