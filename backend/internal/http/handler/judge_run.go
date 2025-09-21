package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/http/errcode"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
	"github.com/YangYuS8/codyssey/backend/internal/service"
	"github.com/gin-gonic/gin"
)

// JudgeRunEnqueueRequest 允许指定判题版本（可选），为空则由后端默认（当前直接透传）。
type JudgeRunEnqueueRequest struct {
    JudgeVersion string `json:"judge_version"`
}

// InternalFinishRequest 内部 Finish 请求体
type InternalFinishRequest struct {
    Status     string `json:"status"`
    RuntimeMS  int    `json:"runtime_ms"`
    MemoryKB   int    `json:"memory_kb"`
    ExitCode   int    `json:"exit_code"`
    ErrorMessage string `json:"error_message"`
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
        if id == nil || id.UserID == "guest" { respondError(c, http.StatusUnauthorized, errcode.CodeUnauthorized, errcode.Text(errcode.CodeUnauthorized)); return }
        submissionID := c.Param("id")
        if strings.TrimSpace(submissionID) == "" { respondError(c, http.StatusBadRequest, errcode.CodeInvalidID, "empty submission id"); return }
        // 校验 submission 所属（若非管理员/老师则必须是自己的）
        sub, err := subSvc.Get(c.Request.Context(), submissionID)
        if err != nil { respondError(c, http.StatusNotFound, errcode.CodeSubmissionNotFound, errcode.Text(errcode.CodeSubmissionNotFound)); return }
        if id.UserID != sub.UserID && !hasAnyRole(id, auth.RoleSystemAdmin, auth.RoleTeacher) {
            respondError(c, http.StatusForbidden, errcode.CodeForbidden, "not owner")
            return
        }
        var req JudgeRunEnqueueRequest
        _ = c.ShouldBindJSON(&req) // judge_version 可选，解析失败忽略（保持容错）
        jr, err := judgeSvc.Enqueue(c.Request.Context(), submissionID, strings.TrimSpace(req.JudgeVersion))
        if err != nil { respondError(c, http.StatusInternalServerError, errcode.CodeEnqueueFailed, err.Error()); return }
        respondCreated(c, toJudgeRunResponse(jr))
    }
}

// ListJudgeRuns 列出指定 submission 的所有运行记录。
func ListJudgeRuns(judgeSvc *service.JudgeRunHTTPAdapter, subSvc *service.SubmissionService) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := auth.GetIdentity(c)
        if id == nil || id.UserID == "guest" { respondError(c, http.StatusUnauthorized, errcode.CodeUnauthorized, errcode.Text(errcode.CodeUnauthorized)); return }
        submissionID := c.Param("id")
        if strings.TrimSpace(submissionID) == "" { respondError(c, http.StatusBadRequest, errcode.CodeInvalidID, "empty submission id"); return }
        // 权限：非管理员/老师需验证 owner
        sub, err := subSvc.Get(c.Request.Context(), submissionID)
        if err != nil { respondError(c, http.StatusNotFound, errcode.CodeSubmissionNotFound, errcode.Text(errcode.CodeSubmissionNotFound)); return }
        if id.UserID != sub.UserID && !hasAnyRole(id, auth.RoleSystemAdmin, auth.RoleTeacher) {
            respondError(c, http.StatusForbidden, "FORBIDDEN", "not owner")
            return
        }
        limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
        offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
        runs, err := judgeSvc.ListBySubmission(c.Request.Context(), submissionID, limit, offset)
        if err != nil { respondError(c, http.StatusInternalServerError, errcode.CodeListFailed, err.Error()); return }
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
        if id == nil || id.UserID == "guest" { respondError(c, http.StatusUnauthorized, errcode.CodeUnauthorized, errcode.Text(errcode.CodeUnauthorized)); return }
        runID := c.Param("id")
        if strings.TrimSpace(runID) == "" { respondError(c, http.StatusBadRequest, errcode.CodeInvalidID, "empty run id"); return }
        jr, err := judgeSvc.Get(c.Request.Context(), runID)
        if err != nil { respondError(c, http.StatusNotFound, errcode.CodeNotFound, "judge run not found"); return }
        // 拿 submission 校验可见性
        sub, err := subSvc.Get(c.Request.Context(), jr.SubmissionID)
        if err != nil { respondError(c, http.StatusNotFound, errcode.CodeSubmissionNotFound, errcode.Text(errcode.CodeSubmissionNotFound)); return }
        if id.UserID != sub.UserID && !hasAnyRole(id, auth.RoleSystemAdmin, auth.RoleTeacher) {
            respondError(c, http.StatusForbidden, errcode.CodeForbidden, "not owner")
            return
        }
        respondOK(c, toJudgeRunResponse(jr), nil)
    }
}

// InternalStartJudgeRun 仅内部/管理员调用：将 queued -> running
func InternalStartJudgeRun(judgeSvc *service.JudgeRunHTTPAdapter) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := auth.GetIdentity(c)
        if id == nil || id.UserID == "guest" { respondError(c, http.StatusUnauthorized, errcode.CodeUnauthorized, errcode.Text(errcode.CodeUnauthorized)); return }
        runID := strings.TrimSpace(c.Param("id"))
        if runID == "" { respondError(c, http.StatusBadRequest, errcode.CodeInvalidID, "empty run id"); return }
        // 早期校验：必须当前为 queued（减少一次无意义的 DB 写尝试, 可选优化）
        jrExisting, err := judgeSvc.Get(c.Request.Context(), runID)
        if err != nil {
            respondError(c, http.StatusNotFound, errcode.CodeJudgeRunNotFound, errcode.Text(errcode.CodeJudgeRunNotFound))
            return
        }
        if jrExisting.Status != domain.JudgeRunStatusQueued {
            respondError(c, http.StatusBadRequest, errcode.CodeInvalidTransition, "cannot start: not in queued status")
            return
        }
        jrDomain, err := judgeSvc.Service().Start(c.Request.Context(), runID)
        if err != nil {
            if err == repository.ErrJudgeRunNotFound {
                respondError(c, http.StatusNotFound, errcode.CodeJudgeRunNotFound, errcode.Text(errcode.CodeJudgeRunNotFound))
                return
            }
            if err == repository.ErrJudgeRunConflict {
                respondError(c, http.StatusConflict, errcode.CodeConflict, errcode.Text(errcode.CodeConflict))
                return
            }
            respondError(c, http.StatusBadRequest, errcode.CodeInvalidTransition, err.Error())
            return
        }
        respondOK(c, toJudgeRunResponse(service.JudgeRunDTO{ID: jrDomain.ID, SubmissionID: jrDomain.SubmissionID, Status: jrDomain.Status, JudgeVersion: jrDomain.JudgeVersion, RuntimeMS: jrDomain.RuntimeMS, MemoryKB: jrDomain.MemoryKB, ExitCode: jrDomain.ExitCode, ErrorMessage: jrDomain.ErrorMessage, CreatedAt: jrDomain.CreatedAt, UpdatedAt: jrDomain.UpdatedAt, StartedAt: jrDomain.StartedAt, FinishedAt: jrDomain.FinishedAt}), nil)
    }
}

// InternalFinishJudgeRun 仅内部/管理员调用：running -> 终态
func InternalFinishJudgeRun(judgeSvc *service.JudgeRunHTTPAdapter) gin.HandlerFunc {
    return func(c *gin.Context) {
        id := auth.GetIdentity(c)
        if id == nil || id.UserID == "guest" { respondError(c, http.StatusUnauthorized, errcode.CodeUnauthorized, errcode.Text(errcode.CodeUnauthorized)); return }
        runID := strings.TrimSpace(c.Param("id"))
        if runID == "" { respondError(c, http.StatusBadRequest, errcode.CodeInvalidID, "empty run id"); return }
        var req InternalFinishRequest
        if err := c.ShouldBindJSON(&req); err != nil { respondError(c, http.StatusBadRequest, errcode.CodeInvalidID, "invalid body"); return }
        // 早期校验：必须当前为 running
        jrExisting, err := judgeSvc.Get(c.Request.Context(), runID)
        if err != nil {
            respondError(c, http.StatusNotFound, errcode.CodeJudgeRunNotFound, errcode.Text(errcode.CodeJudgeRunNotFound))
            return
        }
        if jrExisting.Status != domain.JudgeRunStatusRunning {
            respondError(c, http.StatusBadRequest, errcode.CodeInvalidTransition, "cannot finish: not in running status")
            return
        }
        // 目标状态集合验证将由 service.Finish 再次严格校验
        jrDomain, err := judgeSvc.Service().Finish(c.Request.Context(), runID, req.Status, req.RuntimeMS, req.MemoryKB, req.ExitCode, req.ErrorMessage)
        if err != nil {
            if err == service.ErrJudgeRunInvalidStatus {
                respondError(c, http.StatusBadRequest, errcode.CodeInvalidStatus, err.Error())
                return
            }
            if err == repository.ErrJudgeRunNotFound {
                respondError(c, http.StatusNotFound, errcode.CodeJudgeRunNotFound, errcode.Text(errcode.CodeJudgeRunNotFound))
                return
            }
            if err == repository.ErrJudgeRunConflict {
                respondError(c, http.StatusConflict, errcode.CodeConflict, errcode.Text(errcode.CodeConflict))
                return
            }
            respondError(c, http.StatusBadRequest, errcode.CodeInvalidTransition, err.Error())
            return
        }
        respondOK(c, toJudgeRunResponse(service.JudgeRunDTO{ID: jrDomain.ID, SubmissionID: jrDomain.SubmissionID, Status: jrDomain.Status, JudgeVersion: jrDomain.JudgeVersion, RuntimeMS: jrDomain.RuntimeMS, MemoryKB: jrDomain.MemoryKB, ExitCode: jrDomain.ExitCode, ErrorMessage: jrDomain.ErrorMessage, CreatedAt: jrDomain.CreatedAt, UpdatedAt: jrDomain.UpdatedAt, StartedAt: jrDomain.StartedAt, FinishedAt: jrDomain.FinishedAt}), nil)
    }
}

