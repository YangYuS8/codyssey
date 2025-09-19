package auth

import (
	"net/http"
	"os"
	"strings"

	h "github.com/YangYuS8/codyssey/backend/internal/http/handler"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

const ctxKeyIdentity = "__identity"

// AttachDebugIdentity 中间件：开发阶段用，读取 X-Debug-Perms:
// 示例：X-Debug-Perms: problem.create,problem.update
// jwtCustomClaims 简化 claims 结构
type jwtCustomClaims struct {
    UserID string   `json:"sub"`
    Roles  []string `json:"roles"`
    Perms  []string `json:"perms"`
    jwt.RegisteredClaims
}

// AttachDebugIdentity 同时支持：
// 1) Authorization: Bearer <token>
// 2) X-Debug-Roles / X-Debug-Perms (用于本地调试叠加)
// 优先 JWT，再叠加 debug 头。
func AttachDebugIdentity() gin.HandlerFunc {
    return func(c *gin.Context) {
        var id = &Identity{UserID: "guest", Roles: []string{RoleGuest}, Permissions: map[Permission]struct{}{}}

        // 解析 JWT（可选）
        authz := c.GetHeader("Authorization")
        if strings.HasPrefix(strings.ToLower(authz), "bearer ") {
            tokenStr := strings.TrimSpace(authz[7:])
            if tokenStr != "" {
                claims := &jwtCustomClaims{}
                // 暂时从环境变量里拿密钥（之后可注入）
                secret := []byte(getJWTSecret())
                t, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) { return secret, nil })
                if err == nil && t.Valid {
                    if claims.UserID != "" { id.UserID = claims.UserID }
                    if len(claims.Roles) > 0 { id.Roles = append(id.Roles, claims.Roles...) }
                    for _, p := range claims.Perms { if p != "" { id.Permissions[Permission(p)] = struct{}{} } }
                }
            }
        }

        // Debug headers 叠加
        rawPerms := c.GetHeader("X-Debug-Perms")
        if rawPerms != "" {
            for _, part := range strings.Split(rawPerms, ",") {
                p := strings.TrimSpace(part)
                if p != "" { id.Permissions[Permission(p)] = struct{}{} }
            }
        }
        rawRoles := c.GetHeader("X-Debug-Roles")
        if rawRoles != "" {
            for _, part := range strings.Split(rawRoles, ",") {
                r := strings.TrimSpace(part)
                if r != "" { id.Roles = append(id.Roles, r) }
            }
        }
        mergeRolePermissions(id)
        // read 系列兜底
        id.Permissions[PermProblemRead] = struct{}{}
        c.Set(ctxKeyIdentity, id)
        c.Next()
    }
}

// getJWTSecret 暂时简单实现：从环境变量读取（避免循环依赖 config 包）
func getJWTSecret() string {
    if v := strings.TrimSpace(os.Getenv("JWT_SECRET")); v != "" { return v }
    return "dev-secret-change-me"
}

// 从 context 取出身份
func GetIdentity(c *gin.Context) *Identity {
    if v, ok := c.Get(ctxKeyIdentity); ok {
        if id, ok2 := v.(*Identity); ok2 { return id }
    }
    return nil
}

// Require 权限校验（简单版）
func Require(perms ...Permission) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := GetIdentity(c)
        if id == nil {
            c.JSON(http.StatusUnauthorized, h.ErrorResponse{Data: nil, Err: &h.APIError{Code: "UNAUTHORIZED", Message: "missing identity"}})
            c.Abort(); return
        }
        for _, p := range perms {
            if !id.Has(p) {
                c.JSON(http.StatusForbidden, h.ErrorResponse{Data: nil, Err: &h.APIError{Code: "FORBIDDEN", Message: string(p)}})
                c.Abort(); return
            }
        }
        c.Next()
    }
}
// end
