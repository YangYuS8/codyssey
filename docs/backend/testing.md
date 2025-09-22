# Backend Testing Strategy

## 层次划分
| 层 | 目标 | 工具 |
| ---- | ---- | ---- |
| 单元测试 | 纯函数 / 状态机校验 | go test + table driven |
| 仓储测试 | SQL 正确性 / 事务语义 | docker postgres / tx rollback |
| Handler 测试 | 入参绑定 / 错误映射 / 权限 | httptest + gin.Engine |
| 属性测试 | 状态机不变量 | fastcheck-like (自实现/库) (规划) |
| 集成 / E2E | 提交→判题完整链路 | 后续接入 Worker stub |

## 状态机测试示例
```go
func TestSubmissionTransitions(t *testing.T) {
  cases := []struct{ from, to string; ok bool }{
    {"pending","judging",true},
    {"judging","accepted",true},
    {"judging","wrong_answer",true},
    {"judging","error",true},
    {"pending","accepted",false},
  }
  for _, c := range cases { /* 调用校验函数断言 */ }
}
```

## 并发冲突测试
- 构造两个 goroutine 同时 `UpdateSubmissionStatus(id, version=X)`
- 断言：只有一个成功，另一个返回冲突错误

## 工具建议
| 目标 | 工具 |
| ---- | ---- |
| 覆盖率报告 | `go test -coverprofile=cover.out` + CI 展示 |
| 数据库隔离 | 每测试用例开启事务并回滚 / schema reset |
| 伪随机生成 | `math/rand` + 种子固定复现 |
| Mock | interface + 轻量手写 stub（避免大型 mocking 框架） |

## CI 建议
1. `golangci-lint run`
2. `go vet ./...`
3. `go test -race -cover ./...`
4. routecheck 差异检测
5. 若差异或覆盖率 < 阈值 -> 失败

## 未来
- 引入 fuzz test（Go 1.18+）针对解析/序列化
- 状态机 property-based：随机序列生成 + 最终不变量断言
