# Backend Permissions & RBAC

主文档：`../auth-roles-permissions.md`。这里聚焦后端实现细节与调用约定。

## 数据结构建议
```
users
roles
user_roles
permissions
role_permissions
```
可选：`user_permissions` 追加粒度补丁。

## 校验流程
1. 认证中间件解析 JWT (sub, roles, perms?)
2. 构造用户权限集合（roles -> permissions 缓存映射）
3. Handler 前包装：`Require("problem.create")`
4. 范围校验：若需资源归属检查（own/public/contest），调用辅助函数。

## 代码示例
```go
func Require(perms ...Permission) gin.HandlerFunc {
  return func(c *gin.Context) {
    u := MustUser(c)
    if !u.HasAll(perms...) { c.JSON(403, ErrForbidden()); c.Abort(); return }
    c.Next()
  }
}
```

## 缓存策略
| 层 | 说明 |
| ---- | ---- |
| 内存 LRU | role -> permission list 映射 |
| 失效策略 | TTL 10~30s 或版本号刷新 |

## 演进
- ABAC：加入资源属性 (owner_id, contest_window)
- 多租户：所有权限附加 tenant_id 维度
- 动态策略：可评估 OPA / Casbin 适配
