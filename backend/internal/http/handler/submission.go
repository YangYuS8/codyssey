package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type SubmissionCreateRequest struct {
    ProblemID string `json:"problem_id" binding:"required"`
    Language  string `json:"language" binding:"required"`
    Code      string `json:"code" binding:"required"`
}

type SubmissionUpdateStatusRequest struct {
    Status string `json:"status" binding:"required"`
}

// SubmissionStatusLogResponse 用于返回日志（结构简单直接复用 domain 字段名）
type SubmissionStatusLogResponse struct {
    ID           string `json:"id"`
    SubmissionID string `json:"submission_id"`
    FromStatus   string `json:"from_status"`
    ToStatus     string `json:"to_status"`
    CreatedAt    string `json:"created_at"`
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

// ListSubmissions 列表：支持按 user_id / problem_id / status 过滤，分页 limit/offset。
// 代码可见性：仅 owner 或 teacher/system_admin 角色保留 code，其余清空。
func ListSubmissions(s *service.SubmissionService) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := auth.GetIdentity(c)
        if id == nil || id.UserID == "guest" { // 经过权限中间件理论上不会出现 guest，这里兜底
            respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "login required")
            return
        }
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
        offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
        filter := service.SubmissionListFilter{
            UserID:    strings.TrimSpace(c.Query("user_id")),
            ProblemID: strings.TrimSpace(c.Query("problem_id")),
            Status:    strings.TrimSpace(c.Query("status")),
            Limit:     limit,
            Offset:    offset,
        }
        subs, total, err := s.ListWithTotal(c.Request.Context(), filter)
        if err != nil { respondError(c, http.StatusInternalServerError, "LIST_FAILED", err.Error()); return }
        // redaction
        if !hasAnyRole(id, auth.RoleSystemAdmin, auth.RoleTeacher) {
            for i := range subs {
                if subs[i].UserID != id.UserID { subs[i].Code = "" }
            }
        }
        meta := map[string]int{"limit": limit, "offset": offset, "count": len(subs), "total": total}
        respondOK(c, subs, meta)
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

// UpdateSubmissionStatus 更新单个提交的判题状态（需 submission.update_status 权限 + 状态机校验）
func UpdateSubmissionStatus(s *service.SubmissionService) gin.HandlerFunc {
    return func(c *gin.Context) {
        subID := c.Param("id")
        if strings.TrimSpace(subID) == "" { respondError(c, http.StatusBadRequest, "INVALID_ID", "empty id"); return }
        var req SubmissionUpdateStatusRequest
        if err := c.ShouldBindJSON(&req); err != nil { respondError(c, http.StatusBadRequest, "INVALID_BODY", err.Error()); return }
        updated, err := s.UpdateStatus(c.Request.Context(), subID, req.Status)
        if err != nil {
            switch err {
            case service.ErrSubmissionNotFound:
                respondError(c, http.StatusNotFound, "NOT_FOUND", "submission not found")
            case service.ErrInvalidStatus:
                respondError(c, http.StatusBadRequest, "INVALID_STATUS", err.Error())
            case service.ErrInvalidStatusTransition:
                respondError(c, http.StatusBadRequest, "INVALID_TRANSITION", err.Error())
            default:
                respondError(c, http.StatusInternalServerError, "UPDATE_STATUS_FAILED", err.Error())
            }
            return
        }
        // 按可见性规则（更新后返回时，非管理员/老师且非 owner 不返回 code）
        id := auth.GetIdentity(c)
        if id == nil || (id.UserID != updated.UserID && !hasAnyRole(id, auth.RoleSystemAdmin, auth.RoleTeacher)) {
            updated.Code = ""
        }
        respondOK(c, updated, nil)
    }
}

// ListSubmissionStatusLogs 获取指定提交的状态流转日志
func ListSubmissionStatusLogs(s *service.SubmissionService) gin.HandlerFunc {
    return func(c *gin.Context) {
        subID := c.Param("id")
        if strings.TrimSpace(subID) == "" { respondError(c, http.StatusBadRequest, "INVALID_ID", "empty id"); return }
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
        offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
        logs, err := s.ListStatusLogs(c.Request.Context(), subID, limit, offset)
        if err != nil { respondError(c, http.StatusInternalServerError, "LIST_LOGS_FAILED", err.Error()); return }
        // 映射输出
        out := make([]SubmissionStatusLogResponse, 0, len(logs))
        for _, l := range logs {
            out = append(out, SubmissionStatusLogResponse{ID: l.ID, SubmissionID: l.SubmissionID, FromStatus: l.FromStatus, ToStatus: l.ToStatus, CreatedAt: l.CreatedAt.Format(time.RFC3339)})
        }
        meta := map[string]int{"limit":limit, "offset":offset, "count": len(out)}
        respondOK(c, out, meta)
    }
}
