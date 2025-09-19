package handler

import (
	"net/http"
	"strings"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type SubmissionCreateRequest struct {
    ProblemID string `json:"problem_id" binding:"required"`
    Language  string `json:"language" binding:"required"`
    Code      string `json:"code" binding:"required"`
}

// CreateSubmission 处理创建提交
func CreateSubmission(s *service.SubmissionService) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := auth.GetIdentity(c)
        if id == nil || id.UserID == "guest" {
            respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "login required")
            return
        }
        var req SubmissionCreateRequest
        if err := c.ShouldBindJSON(&req); err != nil {
            respondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error())
            return
        }
        // 简单清洗
        lang := strings.TrimSpace(req.Language)
        code := req.Code
        sub, err := s.Create(c.Request.Context(), id.UserID, req.ProblemID, lang, code)
        if err != nil {
            switch err {
            case service.ErrEmptyCode:
                respondError(c, http.StatusBadRequest, "EMPTY_CODE", err.Error())
            case service.ErrLanguageRequired:
                respondError(c, http.StatusBadRequest, "LANGUAGE_REQUIRED", err.Error())
            default:
                respondError(c, http.StatusBadRequest, "CREATE_SUBMISSION_FAILED", err.Error())
            }
            return
        }
        respondCreated(c, sub)
    }
}

// GetSubmission 获取提交详情：仅 owner 或 带 teacher/system_admin 角色可见代码；其他拥有读取权限的用户可看到除 code 以外字段（当前暂不对外给未授权角色列出权限常量，后续补充 submission.get 权限）
func GetSubmission(s *service.SubmissionService) gin.HandlerFunc {
    return func(c *gin.Context) {
        subID := c.Param("id")
        if strings.TrimSpace(subID) == "" {
            respondError(c, http.StatusBadRequest, "INVALID_ID", "empty id")
            return
        }
        sub, err := s.Get(c.Request.Context(), subID)
        if err != nil {
            if err == service.ErrSubmissionNotFound {
                respondError(c, http.StatusNotFound, "NOT_FOUND", "submission not found")
                return
            }
            respondError(c, http.StatusInternalServerError, "GET_FAILED", err.Error())
            return
        }
        id := auth.GetIdentity(c)
        if id == nil { // 理论上中间件已经放 guest 身份
            respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "missing identity")
            return
        }
        // 不是 owner 且 不是 teacher/system_admin -> 隐去代码
        if id.UserID != sub.UserID && !hasAnyRole(id, auth.RoleSystemAdmin, auth.RoleTeacher) {
            sub.Code = ""
        }
        respondOK(c, sub, nil)
    }
}

func hasAnyRole(id *auth.Identity, roles ...string) bool {
    if id == nil { return false }
    for _, r := range id.Roles {
        for _, want := range roles {
            if r == want { return true }
        }
    }
    return false
}
