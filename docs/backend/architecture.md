# Backend Architecture

## 分层结构
```
backend/
  main.go
  internal/
    config/       配置加载 (env -> struct)
    db/           数据库连接与生命周期
    domain/       领域模型 (纯业务结构, 无外部依赖)
    repository/   仓储接口与实现 (postgres + memory)
    http/
      handler/    HTTP 适配层 (绑定/校验/调用/组装响应)
      router/     组装 gin.Engine (依赖注入)
    server/       启动、迁移、优雅关闭 orchestrator
    metrics/      Prometheus 指标帮助 (若已存在)
    auth/         JWT/RBAC 校验 (规划细化)
```

## 设计原则
| 原则 | 说明 |
| ---- | ---- |
| 无环依赖 | 领域层不依赖 handler 或具体存储 |
| 接口隔离 | repository 暴露最小必要接口（上下文 + 参数） |
| 状态明确 | 状态机枚举/流转集中定义便于审计与测试 |
| 可观测性内建 | HTTP / 状态跳转 / 冲突 1st-class 指标 |
| 失败即显式 | 错误码一致、冲突/非法流转单独分类 |

## 请求生命周期
1. `main.go` 初始化配置与 server
2. `server.Start()`：连接 DB → 可选自动迁移 → 构建 gin 引擎 → 注册中间件/路由 → 监听端口
3. 中间件注入 trace_id / 认证信息
4. handler 绑定/校验输入 → 调用 repository / 领域操作
5. 响应以 Envelope 返回 (JSON)
6. 链路完成记录指标 (时延 / 计数)

## 并发与一致性
| 资源 | 策略 | 冲突表现 | 指标 |
| ---- | ---- | ---- | ---- |
| Submission | 版本号乐观锁(version) | UPDATE 0 行 -> CONFLICT | submission_conflicts_total |
| JudgeRun | 条件状态更新 | UPDATE 0 行 -> CONFLICT | judge_run_conflicts_total |

## 关键中间件
| 名称 | 作用 |
| ---- | ---- |
| trace_id | 为每个请求注入唯一 ID，日志与响应可关联 |
| recover | 捕获 panic 返回 500 并记录日志 |
| metrics | 记录 HTTP 时延 / 计数 / 并发 |
| auth (规划) | 解析 JWT / 角色 / 权限集合注入 context |

## 迁移策略
- 开发：`AUTO_MIGRATE=true` 自动执行 goose migration
- 生产：CI/CD 步骤显式执行 `goose up`，应用仅运行时校验证版本

## OpenAPI 与路由对齐
- `docs/openapi.yaml` 为手工契约
- `cmd/routecheck` 运行对比：代码 vs OpenAPI
- 差异输出 JSON，后续纳入 CI，最终执行“差异阻断”策略

## 日志策略
| 级别 | 使用场景 |
| ---- | ---- |
| DEBUG | 调试模式下的详细内部数据 (dev) |
| INFO | 正常状态转换、启动/关闭事件 |
| WARN | 非预期但可恢复（重试、降级） |
| ERROR | 业务失败、外部依赖不可用 |

结构：JSON（生产）+ trace_id + path + latency（可扩展 user_id、resource_id）

## 后续演进
| 方向 | 内容 | 状态 |
| ---- | ---- | ---- |
| 判题执行 | Worker + 队列 (Redis/NATS) | 规划 |
| Tracing | OpenTelemetry + 采样策略 | 规划 |
| 缓存层 | 题目/权限热数据 Cache | 规划 |
| 限流/熔断 | 中央治理 (token bucket) | 规划 |

详见 `overview.md` 与根目录 `roadmap.md`。
