package auth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
)

// helper 生成测试 token
func makeToken(t *testing.T, secret string, user string, roles []string, perms []string) string {
    t.Helper()
    claims := jwt.MapClaims{
        "sub": user,
        "roles": roles,
        "perms": perms,
        "exp": time.Now().Add(1 * time.Hour).Unix(),
        "iat": time.Now().Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    s, err := token.SignedString([]byte(secret))
    if err != nil { t.Fatalf("sign token: %v", err) }
    return s
}

// 简单受保护处理器
func protectedHandler() gin.HandlerFunc {
    return func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) }
}

func setupTestRouter(perms ...Permission) *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(AttachDebugIdentity())
    r.GET("/protected", Require(perms...), protectedHandler())
    return r
}

func TestRequire_Unauthorized(t *testing.T) {
    r := setupTestRouter(PermProblemCreate)
    w := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
    // guest only has read
    r.ServeHTTP(w, req)
    if w.Code != http.StatusForbidden { t.Fatalf("expected 403 got %d", w.Code) }
}

func TestRequire_WithExplicitPermsHeader(t *testing.T) {
    r := setupTestRouter(PermProblemCreate)
    w := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
    req.Header.Set("X-Debug-Perms", string(PermProblemCreate))
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("expected 200 got %d", w.Code) }
}

func TestRequire_RoleMapping(t *testing.T) {
    r := setupTestRouter(PermProblemDelete)
    w := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
    // teacher role has delete
    req.Header.Set("X-Debug-Roles", RoleTeacher)
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("expected 200 got %d", w.Code) }
}

func TestJWTTokenPermissions(t *testing.T) {
    secret := "dev-secret-change-me" // must match default
    tok := makeToken(t, secret, "u123", []string{"student"}, []string{string(PermProblemUpdate)})
    r := setupTestRouter(PermProblemUpdate)
    w := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+tok)
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("expected 200 got %d body=%s", w.Code, w.Body.String()) }
}

func TestJWTAndDebugMerge(t *testing.T) {
    secret := "dev-secret-change-me"
    tok := makeToken(t, secret, "u999", []string{"student"}, []string{})
    r := setupTestRouter(PermProblemDelete)
    w := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+tok)
    // add delete via debug perms
    req.Header.Set("X-Debug-Perms", string(PermProblemDelete))
    r.ServeHTTP(w, req)
    if w.Code != http.StatusOK { t.Fatalf("expected 200 got %d", w.Code) }
}

func TestInvalidTokenIgnored(t *testing.T) {
    r := setupTestRouter(PermProblemCreate)
    w := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
    req.Header.Set("Authorization", "Bearer "+strings.Repeat("x", 10))
    // no debug perms, should fall back to guest -> 403
    r.ServeHTTP(w, req)
    if w.Code != http.StatusForbidden { t.Fatalf("expected 403 got %d", w.Code) }
}
