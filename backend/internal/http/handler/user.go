package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/YangYuS8/codyssey/backend/internal/repository"
	"github.com/YangYuS8/codyssey/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type createUserReq struct {
    Username string   `json:"username" binding:"required,min=3"`
    Roles    []string `json:"roles"`
}

type updateUserRolesReq struct { Roles []string `json:"roles" binding:"required"` }

func ListUsers(us *service.UserService) gin.HandlerFunc {
    return func(c *gin.Context) {
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
        offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
        users, err := us.List(c, limit, offset)
        if err != nil { respondError(c, http.StatusInternalServerError, "LIST_FAILED", err.Error()); return }
        meta := map[string]int{"limit": limit, "offset": offset, "count": len(users)}
        respondOK(c, users, meta)
    }
}

func CreateUser(us *service.UserService) gin.HandlerFunc {
    return func(c *gin.Context) {
        var req createUserReq
        if err := c.ShouldBindJSON(&req); err != nil { respondError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error()); return }
        filtered := make([]string, 0, len(req.Roles))
        for _, r := range req.Roles { if s := strings.TrimSpace(r); s != "" { filtered = append(filtered, s) } }
        u, err := us.Create(c, req.Username, filtered)
        if err != nil {
            if err == service.ErrUserDuplicate { respondError(c, http.StatusConflict, "USER_EXISTS", err.Error()); return }
            respondError(c, http.StatusInternalServerError, "CREATE_FAILED", err.Error()); return }
        respondCreated(c, u)
    }
}

func GetUser(us *service.UserService) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        u, err := us.Get(c, id)
        if err != nil {
            if err == service.ErrUserNotFound { respondError(c, http.StatusNotFound, "NOT_FOUND", "user not found"); return }
            respondError(c, http.StatusInternalServerError, "GET_FAILED", err.Error()); return }
        respondOK(c, u, nil)
    }
}

func UpdateUserRoles(us *service.UserService) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        var req updateUserRolesReq
        if err := c.ShouldBindJSON(&req); err != nil { respondError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error()); return }
        if err := us.UpdateRoles(c, id, req.Roles); err != nil {
            if err == service.ErrUserNotFound { respondError(c, http.StatusNotFound, "NOT_FOUND", "user not found"); return }
            respondError(c, http.StatusInternalServerError, "UPDATE_FAILED", err.Error()); return }
        respondOK(c, gin.H{"id": id, "roles": req.Roles}, nil)
    }
}

func DeleteUser(us *service.UserService) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := c.Param("id")
        if err := us.Delete(c, id); err != nil {
            if err == service.ErrUserNotFound { respondError(c, http.StatusNotFound, "NOT_FOUND", "user not found"); return }
            respondError(c, http.StatusInternalServerError, "DELETE_FAILED", err.Error()); return }
        respondOK(c, gin.H{"deleted": id}, nil)
    }
}

// 确保引用 repository 错误以保持编译期校验
var _ = repository.ErrUserNotFound
