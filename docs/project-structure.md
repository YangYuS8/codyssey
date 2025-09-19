## 项目结构说明

```
codyssey/
  backend/        # Go 服务（Gin）
  python/         # FastAPI AI 服务
  frontend/       # Next.js 前端
  infra/          # docker compose 与后续 IaC (K8s / Terraform 预留)
  docs/           # 文档集合
  .env.example    # 环境变量模板
```

### backend 关键子目录
```
backend/
  internal/
    config/     # 配置加载
    db/         # 数据库连接 (pgxpool)
    domain/     # 领域模型
    repository/ # 仓储接口与实现 (PG + memory)
    http/
      handler/  # HTTP 处理逻辑（请求绑定+调用领域/仓储+响应）
      router/   # 路由与依赖装配
    server/     # 生命周期管理 (启动/迁移/优雅关闭)
```

### 运行入口
`backend/main.go`：加载配置 -> 构建 server -> 启动 HTTP。

### 配置优先级
环境变量 > `.env` 文件（开发） > 默认值。

### 命名约定
* 仅 `internal/domain` 放纯领域结构。
* 不在 handler 层写 SQL。
* 仓储接口使用 `context.Context`。

---
详细后端分层见 `go-backend.md`。
