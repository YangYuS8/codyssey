package handler

import (
	"net/http"
	"strings"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/gin-gonic/gin"
)

type AuthHandlers struct { Service *auth.AuthService }

func NewAuthHandlers(s *auth.AuthService) *AuthHandlers { return &AuthHandlers{Service: s} }

type registerRequest struct {
    Username string   `json:"username"`
    Password string   `json:"password"`
    Roles    []string `json:"roles"`
}

type loginRequest struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

type refreshRequest struct {
    RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandlers) Register(c *gin.Context) {
    var req registerRequest
    if err := c.ShouldBindJSON(&req); err != nil { respondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error()); return }
    user, tokens, err := h.Service.Register(c, req.Username, req.Password, req.Roles)
    if err != nil {
        switch err {
        case auth.ErrUsernameTaken:
            respondError(c, http.StatusConflict, "USERNAME_TAKEN", err.Error())
        case auth.ErrWeakPassword:
            respondError(c, http.StatusBadRequest, "WEAK_PASSWORD", err.Error())
        default:
            respondError(c, http.StatusBadRequest, "REGISTER_FAILED", err.Error())
        }
        return
    }
    respondCreated(c, gin.H{"user": user, "tokens": tokens})
}

func (h *AuthHandlers) Login(c *gin.Context) {
    var req loginRequest
    if err := c.ShouldBindJSON(&req); err != nil { respondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error()); return }
    user, tokens, err := h.Service.Authenticate(c, req.Username, req.Password)
    if err != nil {
        if err == auth.ErrInvalidLogin { respondError(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error()); return }
        respondError(c, http.StatusBadRequest, "LOGIN_FAILED", err.Error())
        return
    }
    respondOK(c, gin.H{"user": user, "tokens": tokens}, nil)
}

func (h *AuthHandlers) Refresh(c *gin.Context) {
    var req refreshRequest
    if err := c.ShouldBindJSON(&req); err != nil { respondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error()); return }
    tok := strings.TrimSpace(req.RefreshToken)
    if tok == "" { respondError(c, http.StatusBadRequest, "MISSING_REFRESH_TOKEN", "refresh token required"); return }
    pair, err := h.Service.Refresh(c, tok)
    if err != nil { respondError(c, http.StatusUnauthorized, "INVALID_REFRESH", err.Error()); return }
    respondOK(c, gin.H{"tokens": pair}, nil)
}
