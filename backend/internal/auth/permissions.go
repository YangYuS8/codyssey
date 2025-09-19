package auth

// Permission 类型定义
type Permission string

// 角色常量（最小示例）
const (
    RoleSystemAdmin = "system_admin"
    RoleTeacher     = "teacher"
    RoleStudent     = "student"
    RoleContestant  = "contestant"
    RoleGuest       = "guest"
)

// 基础权限常量（首版，仅与 Problem 相关 + 示例）
const (
    PermProblemCreate Permission = "problem.create"
    PermProblemRead   Permission = "problem.read"
    PermProblemUpdate Permission = "problem.update"
    PermProblemDelete Permission = "problem.delete"
    // 更细粒度（后续可替换掉 problem.read）
    PermProblemList Permission = "problem.list"
    PermProblemGet  Permission = "problem.get"
    // 用户相关权限
    PermUserCreate Permission = "user.create"
    PermUserRead   Permission = "user.read" // 汇总性读（后续细化）
    PermUserList   Permission = "user.list"
    PermUserGet    Permission = "user.get"
    PermUserUpdateRoles Permission = "user.update_roles"
    PermUserDelete Permission = "user.delete"
    // 提交相关权限（初版占位）
    PermSubmissionCreate Permission = "submission.create"
    PermSubmissionGet    Permission = "submission.get"
)

// 简单用户身份模型（后续替换为 JWT 解析结果）
type Identity struct {
    UserID      string
    Roles       []string
    Permissions map[Permission]struct{}
}

func (i *Identity) Has(p Permission) bool {
    if i == nil { return false }
    _, ok := i.Permissions[p]
    return ok
}

// Mock：根据请求头 X-Debug-Perms 逗号列表生成权限集
func NewDebugIdentity(perms []Permission) *Identity {
    m := make(map[Permission]struct{}, len(perms))
    for _, p := range perms { m[p] = struct{}{} }
    return &Identity{UserID: "debug", Roles: []string{"debug"}, Permissions: m}
}
// 角色到权限的静态初版映射（后续可迁移 DB / 缓存）
var rolePermissionMap = map[string][]Permission{
    RoleSystemAdmin: {PermProblemCreate, PermProblemUpdate, PermProblemDelete, PermProblemRead, PermProblemList, PermProblemGet,
        PermUserCreate, PermUserRead, PermUserList, PermUserGet, PermUserUpdateRoles, PermUserDelete,
        PermSubmissionCreate, PermSubmissionGet},
    RoleTeacher:     {PermProblemCreate, PermProblemUpdate, PermProblemDelete, PermProblemRead, PermProblemList, PermProblemGet,
        PermUserRead, PermUserList, PermUserGet,
        PermSubmissionCreate, PermSubmissionGet},
    RoleStudent:     {PermProblemRead, PermProblemList, PermProblemGet, PermUserGet, PermSubmissionCreate, PermSubmissionGet},
    RoleContestant:  {PermProblemRead, PermProblemList, PermProblemGet, PermUserGet, PermSubmissionCreate, PermSubmissionGet},
    RoleGuest:       {PermProblemRead, PermProblemList, PermProblemGet},
}

func mergeRolePermissions(id *Identity) {
    if id == nil { return }
    if id.Permissions == nil { id.Permissions = make(map[Permission]struct{}) }
    for _, r := range id.Roles {
        if ps, ok := rolePermissionMap[r]; ok {
            for _, p := range ps { id.Permissions[p] = struct{}{} }
        }
    }
}
// end
