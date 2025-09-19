package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/service"
)

// 内部权限路由构建：赋予 manage 权限
func buildInternalJudgeRunRouter() (*gin.Engine, *memoryJudgeRunRepo, *service.SubmissionService, *service.JudgeRunHTTPAdapter) {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(func(c *gin.Context) {
        id := &auth.Identity{UserID: "admin", Roles: []string{auth.RoleSystemAdmin}, Permissions: map[auth.Permission]struct{}{}}
        id.Permissions[auth.PermJudgeRunManage] = struct{}{}
        id.Permissions[auth.PermJudgeRunGet] = struct{}{}
        id.Permissions[auth.PermSubmissionGet] = struct{}{}
        c.Set("__identity", id)
        c.Next()
    })

    // submission fixture
    subRepo := newMemorySubmissionRepo()
    sub := domain.Submission{ID: "subx", UserID: "u1", ProblemID: "p1", Language: "go", Code: "print", Status: service.SubmissionStatusPending, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Version:1}
    _ = subRepo.Create(context.Background(), sub)
    logRepo := &memoryStatusLogRepo{}
    subSvc := service.NewSubmissionService(subRepo, logRepo)

    jrRepo := newMemoryJudgeRunRepo()
    jrSvc := service.NewJudgeRunService(jrRepo)
    adapter := service.NewJudgeRunHTTPAdapter(jrSvc)

    // public enqueue to seed a queued run
    r.POST("/submissions/:id/runs", EnqueueJudgeRun(adapter, subSvc))
    // internal lifecycle
    r.POST("/internal/judge-runs/:id/start", InternalStartJudgeRun(adapter))
    r.POST("/internal/judge-runs/:id/finish", InternalFinishJudgeRun(adapter))
    r.GET("/judge-runs/:id", GetJudgeRun(adapter, subSvc))

    return r, jrRepo, subSvc, adapter
}

func TestInternalJudgeRunLifecycle_Success(t *testing.T) {
    r, _, _, _ := buildInternalJudgeRunRouter()

    // enqueue first
    req := httptest.NewRequest(http.MethodPost, "/submissions/subx/runs", nil)
    w := httptest.NewRecorder(); r.ServeHTTP(w, req)
    require.Equal(t, http.StatusCreated, w.Code)
    var created struct{ Data JudgeRunResponse `json:"data"` }
    _ = json.Unmarshal(w.Body.Bytes(), &created)

    // start
    startReq := httptest.NewRequest(http.MethodPost, "/internal/judge-runs/"+created.Data.ID+"/start", nil)
    w2 := httptest.NewRecorder(); r.ServeHTTP(w2, startReq)
    require.Equal(t, http.StatusOK, w2.Code, w2.Body.String())
    var started struct{ Data JudgeRunResponse `json:"data"` }
    _ = json.Unmarshal(w2.Body.Bytes(), &started)
    require.Equal(t, domain.JudgeRunStatusRunning, started.Data.Status)

    // finish succeeded
    finishBody, _ := json.Marshal(map[string]any{"status": domain.JudgeRunStatusSucceeded, "runtime_ms": 120, "memory_kb": 512, "exit_code": 0})
    finishReq := httptest.NewRequest(http.MethodPost, "/internal/judge-runs/"+created.Data.ID+"/finish", bytes.NewReader(finishBody))
    finishReq.Header.Set("Content-Type", "application/json")
    w3 := httptest.NewRecorder(); r.ServeHTTP(w3, finishReq)
    require.Equal(t, http.StatusOK, w3.Code, w3.Body.String())
    var finished struct{ Data JudgeRunResponse `json:"data"` }
    _ = json.Unmarshal(w3.Body.Bytes(), &finished)
    require.Equal(t, domain.JudgeRunStatusSucceeded, finished.Data.Status)
    require.NotNil(t, finished.Data.FinishedAt)
}

func TestInternalJudgeRun_InvalidTransition_StartTwice(t *testing.T) {
    r, _, _, _ := buildInternalJudgeRunRouter()
    // enqueue
    req := httptest.NewRequest(http.MethodPost, "/submissions/subx/runs", nil)
    w := httptest.NewRecorder(); r.ServeHTTP(w, req)
    var created struct{ Data JudgeRunResponse `json:"data"` }
    _ = json.Unmarshal(w.Body.Bytes(), &created)

    // first start
    startReq := httptest.NewRequest(http.MethodPost, "/internal/judge-runs/"+created.Data.ID+"/start", nil)
    w2 := httptest.NewRecorder(); r.ServeHTTP(w2, startReq)
    require.Equal(t, http.StatusOK, w2.Code)
    // second start should fail 400
    startReq2 := httptest.NewRequest(http.MethodPost, "/internal/judge-runs/"+created.Data.ID+"/start", nil)
    w3 := httptest.NewRecorder(); r.ServeHTTP(w3, startReq2)
    require.Equal(t, http.StatusBadRequest, w3.Code)
}

func TestInternalJudgeRun_InvalidStatus_FinishBadStatus(t *testing.T) {
    r, _, _, _ := buildInternalJudgeRunRouter()
    // enqueue
    req := httptest.NewRequest(http.MethodPost, "/submissions/subx/runs", nil)
    w := httptest.NewRecorder(); r.ServeHTTP(w, req)
    var created struct{ Data JudgeRunResponse `json:"data"` }
    _ = json.Unmarshal(w.Body.Bytes(), &created)
    // start
    startReq := httptest.NewRequest(http.MethodPost, "/internal/judge-runs/"+created.Data.ID+"/start", nil)
    w2 := httptest.NewRecorder(); r.ServeHTTP(w2, startReq)
    require.Equal(t, http.StatusOK, w2.Code)
    // finish with invalid status
    finishBody, _ := json.Marshal(map[string]any{"status": "weird", "runtime_ms": 1})
    finishReq := httptest.NewRequest(http.MethodPost, "/internal/judge-runs/"+created.Data.ID+"/finish", bytes.NewReader(finishBody))
    finishReq.Header.Set("Content-Type", "application/json")
    w3 := httptest.NewRecorder(); r.ServeHTTP(w3, finishReq)
    require.Equal(t, http.StatusBadRequest, w3.Code, w3.Body.String())
}

func TestInternalJudgeRun_NotFound(t *testing.T) {
    r, _, _, _ := buildInternalJudgeRunRouter()
    // start on non-existing id
    startReq := httptest.NewRequest(http.MethodPost, "/internal/judge-runs/nonexistent/start", nil)
    w := httptest.NewRecorder(); r.ServeHTTP(w, startReq)
    require.Equal(t, http.StatusNotFound, w.Code)
    // finish on non-existing id
    finishBody, _ := json.Marshal(map[string]any{"status": domain.JudgeRunStatusSucceeded})
    finishReq := httptest.NewRequest(http.MethodPost, "/internal/judge-runs/nonexistent/finish", bytes.NewReader(finishBody))
    finishReq.Header.Set("Content-Type", "application/json")
    w2 := httptest.NewRecorder(); r.ServeHTTP(w2, finishReq)
    require.Equal(t, http.StatusNotFound, w2.Code)
}
