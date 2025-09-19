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

### 日志 & 追踪 ID

使用 `zap`：
- dev 环境：development encoder（便于阅读）
- prod（规划）：JSON 结构化 + 采样（后续在配置中加入级别与采样率）

已实现：
- Trace ID 中间件（为每个请求注入 trace_id 字段到日志与 context）

待办：
- 与未来分布式追踪 (OpenTelemetry) 对齐（trace_id 复用 /propagation）。

### 迁移机制

当前：使用 goose 版本化 SQL（`backend/migrations`），启动时若配置 `AUTO_MIGRATE=true` 自动执行所有迁移。

策略：
- dev/test：默认允许自动迁移
- prod：建议关闭自动迁移并在部署流水线显式执行

待完善：
- 回滚流程文档化
- 数据变更的向后兼容策略（大表锁风险评估）

### 健康检查

`GET /health`：
- DB Ping (成功 -> up / 失败 -> down)
- 返回版本与环境

### API & OpenAPI
详细接口参考：`go-backend-api.md`。OpenAPI 规范位于 `docs/openapi.yaml`，维护策略见 `openapi.md`。

错误码集中在 `api_errors.md` 与代码 `internal/http/errcode` 包中。新增错误码需同时更新文档。

go test ./...
### 测试策略

当前已包含：
1. Handler 层测试：Gin + httptest，使用内存/或真实仓储（隔离）
2. 权限矩阵测试：验证角色与权限映射是否符合预期（减少安全回退）
3. 指标暴露测试：确认核心 metrics 注册成功
4. JudgeRun / Submission 状态流转测试（内部生命周期）

规划补充：
- 仓储集成测试：使用 PostgreSQL（docker）+ 事务回滚或 schema reset
- Property-based 测试：状态机不变量（不可逆/非法转移拒绝）
- E2E：提交→判题运行（引入 Worker stub 后）

运行：
```
cd backend
go test ./...
```

### 约定与代码风格

- 不在 handler 内写 SQL。
- 公共返回格式后续标准化：`{"data": ..., "error": {"code":..., "message":...}}`。
- 所有外部依赖通过构造函数或依赖注入层 (router/server) 传入。

### 已完成与待办

已完成（持续更新）：
- 统一错误响应 envelope（data/error 格式）
- RBAC 权限体系（含 `judge_run.manage`）
- JWT 鉴权（含 dev/test debug identity）
- 状态机（Submission / JudgeRun）与条件更新防竞态
- Prometheus 最小指标集（HTTP + 状态转移）
- zap 日志 + trace_id 中间件
- goose 迁移体系 + AutoMigrate 受控开关
- golangci-lint 集成

进行中 / 近期：
- 分页与过滤公共库
- OpenAPI 自动生成评估（swag vs oapi-codegen）
- Judge Worker / Sandbox 对接 (队列 + 执行器)
- 统一错误码 → HTTP 映射细化（加入 CONFLICT）
- JudgeRun / Submission 冲突显式 409 区分
- 增加 JudgeRun 执行耗时指标

规划：
- Tracing (OpenTelemetry)
- 分布式调度 / 优先级队列
- 指标：DB 查询耗时 / 沙箱执行耗时 / 重判批次指标
- 安全：速率限制、审计日志、过期策略

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
