## 角色与权限设计 (Draft)

此文档定义 Codyssey 平台初版角色（Role）与权限（Permission）模型，为后续认证与访问控制实现提供基线。

### 目标原则
1. 最小权限原则 (Least Privilege)
2. 角色可扩展（支持后续自定义或细粒度策略）
3. API 层统一权限检查（中间件 + 策略表 / RBAC 缓存）
4. 权限命名清晰、可读、易于分组

### 核心角色 (初版)
| 角色 | 英文标识 | 描述 | 典型主体 |
| ---- | -------- | ---- | -------- |
| 系统管理员 | `system_admin` | 平台最高权限，运维 / 超级管理 | 平台运维人员 |
| 老师 | `teacher` | 课程/题目/比赛组织、审核提交 | 教师 / 助教 |
| 学生 | `student` | 参与课程学习、提交题解 | 学生用户 |
| 参赛者 | `contestant` | 参与特定竞赛（权限范围限定在竞赛上下文） | 临时/报名选手 |
| 游客 | `guest` | 未登录或未认证用户，只读公共内容 | 未登录访客 |

> 说明：`contestant` 与 `student` 可能在实现阶段合并为 `student` + “参赛报名关联表”；当前分离以便权限语义清晰。

### 资源域 (Domains)
| Domain | 描述 | 示例资源标识 |
| ------ | ---- | ------------ |
| `user` | 用户账户与角色管理 | user_id |
| `problem` | 题目 | problem_id |
| `submission` | 提交记录 | submission_id |
| `contest` | 比赛/活动 | contest_id |
| `course` | 课程（若后续拓展） | course_id |
| `judging` | 判题执行 / 结果访问 | submission_id |
| `ai` | AI 出题/检测能力 | 模型/调用策略 |
| `system` | 系统级运维、配置 | 全局配置项 |
| `storage` | 附件/测试数据 | object_key |

### 动作 (Actions)
常用基础动作集合（组合形成权限 Code）：
| 动作 | 英文 | 说明 |
| ---- | ---- | ---- |
| 创建 | create | 创建新资源 |
| 读取 | read | 读取单项或列表 |
| 更新 | update | 修改资源（幂等部分更新） |
| 删除 | delete | 删除（逻辑或物理） |
| 审核 | review | 审核/批准（题目、比赛、提交特殊状态） |
| 发布 | publish | 公开可见（题目/比赛） |
| 提交 | submit | 发起一次提交（题解、评测） |
| 运行 | run | 触发判题 / AI 处理 |
| 管理 | manage | 汇总类：等价于 create+read+update+delete+… (仅少数高阶角色拥有) |

扩展动作（按需）：`freeze`（比赛封榜）、`unfreeze`、`rejudge`、`download`（下载测试数据）、`approve`、`assign`（分配课程/比赛）。

### 权限命名规范
格式：`<domain>.<action>[.<scope>]`
示例：
* `problem.create`
* `problem.read`
* `problem.update`
* `problem.delete`
* `contest.freeze`
* `submission.rejudge`
* `ai.generate`
* `ai.detect`
* `system.manage`

Scope（可选）用于限定粒度：如 `problem.read.public` / `problem.read.own` / `submission.read.own`。

### 角色权限矩阵 (首版建议)

| Permission | system_admin | teacher | student | contestant | guest |
| ---------- | ------------ | ------- | ------- | ---------- | ----- |
| user.read | ✔ | (自身基本信息) | 自身 | 自身 | - |
| user.update | ✔ | 自身 | 自身 | 自身 | - |
| user.manage | ✔ | - | - | - | - |
| problem.create | ✔ | ✔ | - | - | - |
| problem.read | ✔ | ✔ | ✔ | ✔ | 公共题目 |
| problem.update | ✔ | ✔(限自己创建或被授权) | - | - | - |
| problem.delete | ✔ | ✔(限自己创建且未引用) | - | - | - |
| problem.publish | ✔ | ✔ | - | - | - |
| submission.submit | ✔ | (调试/题解) | ✔ | ✔(竞赛期间) | - |
| submission.read | ✔ | ✔(可查看班级/比赛相关) | ✔(own) | ✔(contest scope) | - |
| submission.read.all | ✔ | 部分（课程/比赛范围） | - | - | - |
| submission.rejudge | ✔ | (授权) | - | - | - |
| judging.run | ✔ | (授权) | - | - | - |
| contest.create | ✔ | ✔ | - | - | - |
| contest.read | ✔ | ✔ | ✔ | ✔(报名范围) | 公共已发布 |
| contest.update | ✔ | ✔(自己创建/协作) | - | - | - |
| contest.delete | ✔ | ✔(限定条件) | - | - | - |
| contest.freeze | ✔ | ✔ | - | - | - |
| contest.unfreeze | ✔ | ✔ | - | - | - |
| contest.register | ✔ | 可代报名 | ✔ | ✔ | - |
| ai.generate | ✔ | ✔ | - | - | - |
| ai.detect | ✔ | ✔ | - | - | - |
| storage.upload | ✔ | ✔ | - | - | - |
| storage.read | ✔ | ✔ | 部分(公开/own) | 比赛关联 | 公共 |
| system.manage | ✔ | - | - | - | - |

> “✔(own)” 表示仅自身资源；“公共题目”= 已发布且公开；表中括号说明需额外业务规则判断。

### 数据范围控制 (Scopes)
推荐在权限判定时二阶段校验：
1. 粗粒度：是否拥有 `<domain>.<action>` 基础权限。
2. 细粒度：资源归属 / 公开状态 / 比赛时间窗口 / 课程成员关系。

可在代码中抽象：
```go
type Permission string

const (
    PermProblemCreate Permission = "problem.create"
    PermProblemRead   Permission = "problem.read"
    // ...
)

// Scope 校验示例
func CanReadSubmission(user User, sub Submission) bool {
    if user.Has(PermSubmissionReadAll) { return true }
    if user.Has(PermSubmissionRead) && sub.UserID == user.ID { return true }
    // contest scope
    if user.HasRole(RoleContestant) && sub.ContestID != nil && user.InContest(*sub.ContestID) { return true }
    return false
}
```

### 推荐数据库结构 (草案)
```
users(id, username, password_hash, status, created_at,...)
roles(id, code, name, description)
user_roles(user_id, role_id)
permissions(id, code, description)
role_permissions(role_id, permission_id)
// 可选：用户自定义附加权限
user_permissions(user_id, permission_id)
```

### JWT Claims 建议
```
{
  "sub": "<user_id>",
  "roles": ["teacher","student"],
  "perms": ["problem.create","submission.submit"], // 可选（或仅 roles）
  "exp": 1735689600,
  "iat": 1735603200,
  "ver": 1  // token schema 版本
}
```
> 若 `perms` 过长，可仅下发 roles，服务端缓存 role->perm 映射。

### 中间件授权流程 (Go 后端草图)
1. 解析/验证 JWT（签名、过期、黑名单）。
2. 将用户角色与缓存映射出的权限集合附加到请求上下文 `context.Context`。
3. Handler 前（或分组路由）声明所需权限：
   ```go
   Require(PermProblemCreate)(handler)
   ```
4. 中间件判断：若无权限 -> 403；若需要资源范围，再调用二次检查函数。

### 缓存策略
* role->permissions 映射：本地 LRU + 失效（10~30s）或基于版本号。
* Admin 修改权限后：发布 “权限变更” 事件，触发缓存刷新。

### 扩展策略
| 需求 | 方案 |
| ---- | ---- |
| 更细粒度（行/列级） | 引入 ABAC：基于属性 (owner, contest_window, is_public) 判定 |
| 临时提升权限 | 写临时票据 (elevation token) 附加限定时长/用途 |
| 审计 | 增加 audit_log：记录 (user, action, resource, result, ts) |
| 租户隔离 | 加列 tenant_id；所有查询加 tenant 过滤；权限表按租户分组 |

### 示例：声明式权限绑定（未来）
可封装：
```go
func Require(perms ...Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        user := auth.FromContext(c)
        if !user.HasAll(perms...) { c.JSON(403, gin.H{"error":"FORBIDDEN"}); c.Abort(); return }
        c.Next()
    }
}
```

### 开发落地优先级建议
1. 建表 & 预置角色 + 权限种子
2. JWT 签发与解析（先内存映射）
3. 基础权限中间件（仅检查是否登录 + 单权限）
4. 替换 Problem 写操作为权限保护（如 `problem.create`）
5. 扩展提交 / 比赛场景（时间窗口校验）
6. 增加权限缓存 + 管理端动态调整接口

### 后续待决事项
| 项目 | 说明 |
| ---- | ---- |
| 权限管理 UI | 是否需要 Web 管理界面（拖拽/搜索） |
| 动态策略 | 是否接入策略引擎（Casbin / OPA） |
| 多租户 | 是否需要在 early stage 设计 tenant_id |

---
本文件为初稿，实施前可在 issue 中补充业务实际差异与约束。
