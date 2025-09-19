package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
)

func TestAuth_Register_Login_Refresh(t *testing.T) {
    gin.SetMode(gin.TestMode)
    mem := repository.NewMemoryUserRepository()
    jwtMgr := auth.NewJWTManager("test-secret", 2*time.Minute, time.Hour)
    svc := auth.NewAuthService(mem, jwtMgr)
    h := NewAuthHandlers(svc)
    r := gin.New()
    r.POST("/auth/register", h.Register)
    r.POST("/auth/login", h.Login)
    r.POST("/auth/refresh", h.Refresh)

    // register
    regBody := map[string]any{"username": "alice", "password": "secret123", "roles": []string{"student"}}
    w := httptest.NewRecorder()
    b, _ := json.Marshal(regBody)
    req, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w, req)
    require.Equal(t, 201, w.Code, w.Body.String())

    // duplicate register
    w2 := httptest.NewRecorder()
    req2, _ := http.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(b))
    req2.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w2, req2)
    require.Equal(t, 409, w2.Code, w2.Body.String())

    // login ok
    loginBody := map[string]any{"username": "alice", "password": "secret123"}
    w3 := httptest.NewRecorder()
    bl, _ := json.Marshal(loginBody)
    req3, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bl))
    req3.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w3, req3)
    require.Equal(t, 200, w3.Code, w3.Body.String())
    var loginResp struct { Data struct { User struct { ID string }; Tokens struct { RefreshToken string `json:"refresh_token"` } } }
    _ = json.Unmarshal(w3.Body.Bytes(), &loginResp)
    require.NotEmpty(t, loginResp.Data.User.ID)
    require.NotEmpty(t, loginResp.Data.Tokens.RefreshToken)

    // login wrong password
    badBody := map[string]any{"username": "alice", "password": "wrong"}
    w4 := httptest.NewRecorder()
    bb, _ := json.Marshal(badBody)
    req4, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bb))
    req4.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w4, req4)
    require.Equal(t, 401, w4.Code)

    // refresh
    refreshBody := map[string]any{"refresh_token": loginResp.Data.Tokens.RefreshToken}
    w5 := httptest.NewRecorder()
    br, _ := json.Marshal(refreshBody)
    req5, _ := http.NewRequest(http.MethodPost, "/auth/refresh", bytes.NewReader(br))
    req5.Header.Set("Content-Type", "application/json")
    r.ServeHTTP(w5, req5)
    require.Equal(t, 200, w5.Code, w5.Body.String())
}
