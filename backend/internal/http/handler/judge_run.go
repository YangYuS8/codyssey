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

// JudgeRunEnqueueRequest 允许指定判题版本（可选），为空则由后端默认（当前直接透传）。
type JudgeRunEnqueueRequest struct {
    JudgeVersion string `json:"judge_version"`
}

// JudgeRunResponse 输出结构（与 domain 基本一致，仅时间格式化）。
type JudgeRunResponse struct {
    ID           string  `json:"id"`
    SubmissionID string  `json:"submission_id"`
    Status       string  `json:"status"`
    JudgeVersion string  `json:"judge_version"`
    RuntimeMS    int     `json:"runtime_ms"`
    MemoryKB     int     `json:"memory_kb"`
    ExitCode     int     `json:"exit_code"`
    ErrorMessage string  `json:"error_message"`
    CreatedAt    string  `json:"created_at"`
    UpdatedAt    string  `json:"updated_at"`
    StartedAt    *string `json:"started_at"`
    FinishedAt   *string `json:"finished_at"`
}

func toJudgeRunResponse(jr service.JudgeRunDTO) JudgeRunResponse {
    var started, finished *string
    if jr.StartedAt != nil { s := jr.StartedAt.Format(time.RFC3339); started = &s }
    if jr.FinishedAt != nil { f := jr.FinishedAt.Format(time.RFC3339); finished = &f }
    return JudgeRunResponse{
        ID: jr.ID, SubmissionID: jr.SubmissionID, Status: jr.Status, JudgeVersion: jr.JudgeVersion,
        RuntimeMS: jr.RuntimeMS, MemoryKB: jr.MemoryKB, ExitCode: jr.ExitCode, ErrorMessage: jr.ErrorMessage,
        CreatedAt: jr.CreatedAt.Format(time.RFC3339), UpdatedAt: jr.UpdatedAt.Format(time.RFC3339),
        StartedAt: started, FinishedAt: finished,
    }
}

// EnqueueJudgeRun 创建新的判题任务（运行记录）；权限：
// - 学生只能针对自己拥有的 submission
// - 老师 / 管理员可对任意 submission
// 这里暂不校验 submission 是否存在（可在后续服务层扩展校验），当前直接入库（或内存）。
func EnqueueJudgeRun(judgeSvc *service.JudgeRunHTTPAdapter, subSvc *service.SubmissionService) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := auth.GetIdentity(c)
        if id == nil || id.UserID == "guest" { respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "login required"); return }
        submissionID := c.Param("id")
        if strings.TrimSpace(submissionID) == "" { respondError(c, http.StatusBadRequest, "INVALID_ID", "empty submission id"); return }
        // 校验 submission 所属（若非管理员/老师则必须是自己的）
        sub, err := subSvc.Get(c.Request.Context(), submissionID)
        if err != nil { respondError(c, http.StatusNotFound, "SUBMISSION_NOT_FOUND", "submission not found"); return }
        if id.UserID != sub.UserID && !hasAnyRole(id, auth.RoleSystemAdmin, auth.RoleTeacher) {
            respondError(c, http.StatusForbidden, "FORBIDDEN", "not owner")
            return
        }
        var req JudgeRunEnqueueRequest
        _ = c.ShouldBindJSON(&req) // judge_version 可选，解析失败忽略（保持容错）
        jr, err := judgeSvc.Enqueue(c.Request.Context(), submissionID, strings.TrimSpace(req.JudgeVersion))
        if err != nil { respondError(c, http.StatusInternalServerError, "ENQUEUE_FAILED", err.Error()); return }
        respondCreated(c, toJudgeRunResponse(jr))
    }
}

// ListJudgeRuns 列出指定 submission 的所有运行记录。
func ListJudgeRuns(judgeSvc *service.JudgeRunHTTPAdapter, subSvc *service.SubmissionService) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := auth.GetIdentity(c)
        if id == nil || id.UserID == "guest" { respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "login required"); return }
        submissionID := c.Param("id")
        if strings.TrimSpace(submissionID) == "" { respondError(c, http.StatusBadRequest, "INVALID_ID", "empty submission id"); return }
        // 权限：非管理员/老师需验证 owner
        sub, err := subSvc.Get(c.Request.Context(), submissionID)
        if err != nil { respondError(c, http.StatusNotFound, "SUBMISSION_NOT_FOUND", "submission not found"); return }
        if id.UserID != sub.UserID && !hasAnyRole(id, auth.RoleSystemAdmin, auth.RoleTeacher) {
            respondError(c, http.StatusForbidden, "FORBIDDEN", "not owner")
            return
        }
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
        offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
        runs, err := judgeSvc.ListBySubmission(c.Request.Context(), submissionID, limit, offset)
        if err != nil { respondError(c, http.StatusInternalServerError, "LIST_FAILED", err.Error()); return }
        out := make([]JudgeRunResponse, 0, len(runs))
        for _, r := range runs { out = append(out, toJudgeRunResponse(r)) }
        meta := map[string]int{"limit":limit, "offset":offset, "count":len(out)}
        respondOK(c, out, meta)
    }
}

// GetJudgeRun 读取单个运行记录。
func GetJudgeRun(judgeSvc *service.JudgeRunHTTPAdapter, subSvc *service.SubmissionService) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := auth.GetIdentity(c)
        if id == nil || id.UserID == "guest" { respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "login required"); return }
        runID := c.Param("id")
        if strings.TrimSpace(runID) == "" { respondError(c, http.StatusBadRequest, "INVALID_ID", "empty run id"); return }
        jr, err := judgeSvc.Get(c.Request.Context(), runID)
        if err != nil { respondError(c, http.StatusNotFound, "NOT_FOUND", "judge run not found"); return }
        // 拿 submission 校验可见性
        sub, err := subSvc.Get(c.Request.Context(), jr.SubmissionID)
        if err != nil { respondError(c, http.StatusNotFound, "SUBMISSION_NOT_FOUND", "submission not found"); return }
        if id.UserID != sub.UserID && !hasAnyRole(id, auth.RoleSystemAdmin, auth.RoleTeacher) {
            respondError(c, http.StatusForbidden, "FORBIDDEN", "not owner")
            return
        }
        respondOK(c, toJudgeRunResponse(jr), nil)
    }
}
