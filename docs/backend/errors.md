# Backend Error Model

统一错误码定义：参见上级 `../api_errors.md`。本文件对后端实现关注的落地细节做补充。

## 映射策略
| 层 | 行为 |
| ---- | ---- |
| repository | 返回语义化 Go error（如 ErrSubmissionConflict） |
| handler | 转换为标准 {data:null,error:{code,message}} JSON |
| 中间件 recover | panic -> 500 / UNKNOWN |

## 冲突 (CONFLICT)
产生条件：
- Submission: 乐观锁版本不匹配 (UPDATE ... WHERE id=? AND version=?)
- JudgeRun: 条件状态更新 0 行 (WHERE status=...)
客户端建议：刷新最新状态后重试（幂等安全的流转操作）。

## 状态错误
| Code | 条件 |
| ---- | ---- |
| INVALID_STATUS | 输入的目标状态不在枚举集合 |
| INVALID_TRANSITION | 违反状态有向图 |

## 过大请求体
- HTTP 层中间件限制触发 413 -> PAYLOAD_TOO_LARGE
- 业务特定字段（代码大小）触发 CODE_TOO_LONG (400)

## 超时 (TIMEOUT)
- 由前端 apiFetch 产生；后端无需实现，保持幂等即可。

## 增量维护指南
新增错误码时：
1. `internal/http/errcode` 添加常量与默认消息
2. 更新 `../api_errors.md`
3. 若带并发语义，考虑新增指标（计数冲突/失败）
