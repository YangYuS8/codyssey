## Go 后端架构说明

本文档说明 `backend/` Go 服务当前的分层设计、关键组件、运行方式、版本注入策略以及测试与后续演进计划。

### 目录结构概览

```
backend/
  main.go                # 入口：加载配置 -> 创建 server -> 启动 HTTP 与优雅退出
  go.mod / go.sum
  internal/
    config/              # 配置加载与结构体 (Env, Port, Version, DB...)
    db/                  # PostgreSQL 连接与资源管理 (pgxpool)
    domain/              # 领域模型 (e.g. Problem)
    repository/          # 仓储实现：PostgreSQL + 内存实现 (测试友好)
    http/
      handler/           # 适配 HTTP 的 handler（纯粹协调 + 绑定/校验）
      router/            # 路由集中定义 (依赖注入)
    server/              # Server 组合根：启动顺序、迁移、优雅关闭
```

### 分层职责

- domain: 只包含业务核心结构与构造逻辑（不依赖 Gin、DB）。
- repository: 领域数据访问接口与实现。接口依赖 `context.Context` 而非 Gin，利于测试与解耦。
- http/handler: 处理 HTTP 入口，负责：
  * 解析/绑定/校验请求
  * 调用 repository/领域逻辑
  * 组合响应（不做复杂业务）
- http/router: 统一构建 `*gin.Engine`，在此注入依赖（如 ProblemRepo, healthProbe, Version, Env）。
- server: 统筹生命周期：
  1. 连接数据库
  2. 执行最小化 schema 迁移 (EnsureSchema)
  3. 启动 HTTP 监听
  4. 监听信号，触发优雅关闭 (shutdown -> close DB)

### 运行流程（main.go）

1. 加载 `.env`（若存在）与环境变量。
2. 读取配置 -> 初始化 `server.Server`。
3. `Server.Start()`：DB 连接 + 迁移 + HTTP 启动。
4. 等待信号触发 `Server.WaitForShutdown()`：优雅关闭。

### 配置与环境

`internal/config.Config` 提供：
- `Env`：`dev` / `prod` 等，影响 Gin Mode、日志格式等。
- `Version`：通过构建注入（详见下节）。
- `DB`：PostgreSQL 连接参数。

`.env.example` 中列出了所需变量。运行时优先环境变量，其次可通过 `.env` 便捷加载（开发场景）。

### 版本注入策略

在 `main.go` 顶部：
```
var buildVersion = "dev"
```
构建时可以通过：
```
go build -ldflags "-X main.buildVersion=$(git rev-parse --short HEAD)" -o bin/backend ./backend
```
运行时 `/health` 返回：
```
{
  "status": "ok",
  "db": "up",
  "version": "<buildVersion>",
  "env": "dev"
}
```

### 日志

当前使用 `zap` development logger：
- Dev 环境：人类可读
- TODO：为 prod 环境增加结构化 JSON Logger + 请求追踪 (trace_id)

### 最小迁移机制

`repository.EnsureSchema(ctx, pool)` 在启动时创建 `problems` 表（若不存在）。后续需要：
- 引入真正迁移工具（如 `golang-migrate`）
- 版本化迁移与回滚策略

### 健康检查

`GET /health`：
- DB Ping (成功 -> up / 失败 -> down)
- 返回版本与环境

### API 说明
详细接口参考：`go-backend-api.md`。

### 测试策略

类型：
1. Handler 层 API 测试：使用 Gin + httptest + 内存仓储 (`memory_problem_repo.go`) → 避免 IO 依赖。
2. 未来可加：
   - 仓储集成测试（真实 PostgreSQL，使用事务回滚）
   - 端到端 (E2E)：容器化环境 + 判题队列交互。

运行：
```
cd backend
go test ./...
```

### 约定与代码风格

- 不在 handler 内写 SQL。
- 公共返回格式后续标准化：`{"data": ..., "error": {"code":..., "message":...}}`。
- 所有外部依赖通过构造函数或依赖注入层 (router/server) 传入。

### 待办 / 演进路线

- [ ] 标准化错误响应格式
- [ ] 分页与过滤 (`GET /problems?page=1&page_size=20`)
- [ ] OpenAPI / Swagger 生成
- [ ] Auth/JWT 中间件 & RBAC（题目创建权限）
- [ ] Judge 执行管线：提交、任务调度、结果回传
- [ ] 统一日志与请求 ID (中间件)
- [ ] golangci-lint 集成 + CI
- [ ] 正式迁移工具引入 (migrate)
- [ ] Metrics (Prometheus) & Tracing (OTel)

### 快速 FAQ

Q: 为什么仓储接口用 `context.Context` 而不是 `*gin.Context`？
A: 降低耦合，支持命令行 / gRPC / 测试环境直接调用。

Q: 如何扩展一个新资源？
1. 在 domain 定义结构/构造器。
2. 在 repository 添加接口与实现（PG + memory）。
3. 在 handler 添加 HTTP 绑定/校验与响应。
4. 在 router 注入并注册路由。
5. 添加相应测试（优先使用内存仓储）。

Q: 失败的迁移如何处理？
A: 当前版本：启动即失败并退出。未来引入版本化迁移后可做幂等与回滚。

---
如需补充或调整，请在 issue 中记录或直接提交 PR。
