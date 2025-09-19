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
	"github.com/YangYuS8/codyssey/backend/internal/repository"
	"github.com/YangYuS8/codyssey/backend/internal/service"
)

// 简单内存实现依赖：复用已有 memory submission repo（若存在）否则临时 mock

type memorySubmissionRepo struct { items map[string]domain.Submission }
func newMemorySubmissionRepo() *memorySubmissionRepo { return &memorySubmissionRepo{items: map[string]domain.Submission{}} }
func (m *memorySubmissionRepo) Create(ctx context.Context, s domain.Submission) error { m.items[s.ID] = s; return nil }
func (m *memorySubmissionRepo) GetByID(ctx context.Context, id string) (domain.Submission, error) { v, ok := m.items[id]; if !ok { return domain.Submission{}, service.ErrSubmissionNotFound }; return v, nil }
func (m *memorySubmissionRepo) UpdateStatus(ctx context.Context, id string, status string) error { v, ok := m.items[id]; if !ok { return service.ErrSubmissionNotFound }; v.Status = status; v.UpdatedAt = time.Now().UTC(); m.items[id] = v; return nil }
func (m *memorySubmissionRepo) List(ctx context.Context, f repository.SubmissionFilter, limit, offset int) ([]domain.Submission, error) {
    res := make([]domain.Submission,0)
    for _, v := range m.items {
        if f.UserID != "" && v.UserID != f.UserID { continue }
        if f.ProblemID != "" && v.ProblemID != f.ProblemID { continue }
        if f.Status != "" && v.Status != f.Status { continue }
        res = append(res, v)
    }
    return res, nil
}
func (m *memorySubmissionRepo) Count(ctx context.Context, f repository.SubmissionFilter) (int, error) { list, _ := m.List(ctx, f, 0, 0); return len(list), nil }

type memoryStatusLogRepo struct{ logs []domain.SubmissionStatusLog }
func (m *memoryStatusLogRepo) Add(ctx context.Context, l domain.SubmissionStatusLog) error { m.logs = append(m.logs, l); return nil }
func (m *memoryStatusLogRepo) ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]domain.SubmissionStatusLog, error) { out := []domain.SubmissionStatusLog{}; for _, l := range m.logs { if l.SubmissionID == submissionID { out = append(out, l) } }; return out, nil }

// memoryJudgeRunRepo 直接复用 service.JudgeRunRepo 接口需要的方法
type memoryJudgeRunRepo struct { items map[string]domain.JudgeRun }
func newMemoryJudgeRunRepo() *memoryJudgeRunRepo { return &memoryJudgeRunRepo{items: map[string]domain.JudgeRun{}} }
func (m *memoryJudgeRunRepo) Create(ctx context.Context, jr domain.JudgeRun) error { m.items[jr.ID] = jr; return nil }
func (m *memoryJudgeRunRepo) GetByID(ctx context.Context, id string) (domain.JudgeRun, error) { v, ok := m.items[id]; if !ok { return domain.JudgeRun{}, service.ErrJudgeRunNotFound }; return v, nil }
func (m *memoryJudgeRunRepo) ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]domain.JudgeRun, error) { out := []domain.JudgeRun{}; for _, v := range m.items { if v.SubmissionID == submissionID { out = append(out, v) } }; return out, nil }
func (m *memoryJudgeRunRepo) UpdateRunning(ctx context.Context, id string) error { v, ok := m.items[id]; if !ok { return service.ErrJudgeRunNotFound }; if v.Status != domain.JudgeRunStatusQueued { return service.ErrJudgeRunInvalidStatus }; now := time.Now().UTC(); v.Status = domain.JudgeRunStatusRunning; v.StartedAt = &now; v.UpdatedAt = now; m.items[id] = v; return nil }
func (m *memoryJudgeRunRepo) UpdateFinished(ctx context.Context, id string, status string, runtimeMS, memoryKB, exitCode int, errMsg string) error { v, ok := m.items[id]; if !ok { return service.ErrJudgeRunNotFound }; if v.Status != domain.JudgeRunStatusRunning { return service.ErrJudgeRunInvalidStatus }; now := time.Now().UTC(); v.Status = status; v.RuntimeMS = runtimeMS; v.MemoryKB = memoryKB; v.ExitCode = exitCode; v.ErrorMessage = errMsg; v.FinishedAt = &now; v.UpdatedAt = now; m.items[id] = v; return nil }

// helper 构建路由
// 构建测试路由，直接注入测试用身份（绕过 AttachDebugIdentity 里固定的 guest）
func buildJudgeRunTestRouter(userID string, subOwnedUser string) *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(func(c *gin.Context) {
        id := &auth.Identity{UserID: userID, Roles: []string{auth.RoleStudent}, Permissions: map[auth.Permission]struct{}{}}
        // handler 本身只检查 userID 与角色，这里补充权限以模拟完整身份
        id.Permissions[auth.PermJudgeRunEnqueue] = struct{}{}
        id.Permissions[auth.PermJudgeRunGet] = struct{}{}
        id.Permissions[auth.PermJudgeRunList] = struct{}{}
        id.Permissions[auth.PermSubmissionGet] = struct{}{}
        c.Set("__identity", id)
        c.Next()
    })

    subRepo := newMemorySubmissionRepo()
    // 预放一条 submission
    s := domain.Submission{ID: "sub1", UserID: subOwnedUser, ProblemID: "p1", Language: "go", Code: "print", Status: "pending", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Version:1}
    _ = subRepo.Create(context.Background(), s)
    logRepo := &memoryStatusLogRepo{}
    subSvc := service.NewSubmissionService(subRepo, logRepo)

    jrRepo := newMemoryJudgeRunRepo()
    jrSvc := service.NewJudgeRunService(jrRepo)
    adapter := service.NewJudgeRunHTTPAdapter(jrSvc)

    r.POST("/submissions/:id/runs", EnqueueJudgeRun(adapter, subSvc))
    r.GET("/submissions/:id/runs", ListJudgeRuns(adapter, subSvc))
    r.GET("/judge-runs/:id", GetJudgeRun(adapter, subSvc))
    return r
}

func TestJudgeRun_Enqueue_List_Get(t *testing.T) {
    r := buildJudgeRunTestRouter("u1", "u1")

    // enqueue
    body, _ := json.Marshal(map[string]string{"judge_version": "v1"})
    req := httptest.NewRequest(http.MethodPost, "/submissions/sub1/runs", bytes.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    require.Equal(t, http.StatusCreated, w.Code, w.Body.String())
    var created struct { Data JudgeRunResponse `json:"data"` }
    _ = json.Unmarshal(w.Body.Bytes(), &created)
    require.Equal(t, "queued", created.Data.Status)

    // list
    req2 := httptest.NewRequest(http.MethodGet, "/submissions/sub1/runs", nil)
    w2 := httptest.NewRecorder(); r.ServeHTTP(w2, req2)
    require.Equal(t, http.StatusOK, w2.Code)
    var list struct { Data []JudgeRunResponse `json:"data"` }
    _ = json.Unmarshal(w2.Body.Bytes(), &list)
    require.Len(t, list.Data, 1)

    // get
    runID := created.Data.ID
    req3 := httptest.NewRequest(http.MethodGet, "/judge-runs/"+runID, nil)
    w3 := httptest.NewRecorder(); r.ServeHTTP(w3, req3)
    require.Equal(t, http.StatusOK, w3.Code)
}

func TestJudgeRun_ForbiddenForNonOwner(t *testing.T) {
    owner := "owner1"
    r := buildJudgeRunTestRouter("u2", owner) // u2 不是 submission owner
    req := httptest.NewRequest(http.MethodPost, "/submissions/sub1/runs", nil)
    w := httptest.NewRecorder(); r.ServeHTTP(w, req)
    require.Equal(t, http.StatusForbidden, w.Code)
}
