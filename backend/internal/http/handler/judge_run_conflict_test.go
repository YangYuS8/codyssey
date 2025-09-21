package handler_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/http/router"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
)

// --- Start 冲突仓库 ---
type conflictStartRepo struct {
    mu      sync.Mutex
    run     domain.JudgeRun
    barrier chan struct{}
    gets    atomic.Int32
}
func newConflictStartRepo(run domain.JudgeRun) *conflictStartRepo { return &conflictStartRepo{run: run, barrier: make(chan struct{})} }
func (r *conflictStartRepo) Create(_ context.Context, jr domain.JudgeRun) error { r.mu.Lock(); defer r.mu.Unlock(); r.run = jr; return nil }
func (r *conflictStartRepo) GetByID(_ context.Context, id string) (domain.JudgeRun, error) {
    if r.run.ID != id { return domain.JudgeRun{}, repository.ErrJudgeRunNotFound }
    n := r.gets.Add(1)
    if n == 2 { close(r.barrier) }
    return r.run, nil
}
func (r *conflictStartRepo) ListBySubmission(_ context.Context, subID string, _, _ int) ([]domain.JudgeRun, error) { if r.run.SubmissionID == subID { return []domain.JudgeRun{r.run}, nil }; return []domain.JudgeRun{}, nil }
func (r *conflictStartRepo) UpdateRunning(_ context.Context, id string) error {
    if r.run.ID != id { return repository.ErrJudgeRunNotFound }
    <-r.barrier // 等待两个并发读取都完成
    r.mu.Lock(); defer r.mu.Unlock()
    if r.run.Status != domain.JudgeRunStatusQueued { return repository.ErrJudgeRunConflict }
    now := time.Now().UTC(); r.run.Status = domain.JudgeRunStatusRunning; r.run.StartedAt = &now; r.run.UpdatedAt = now; return nil
}
func (r *conflictStartRepo) UpdateFinished(_ context.Context, id string, status string, runtimeMS, memoryKB, exitCode int, errMsg string) error { return repository.ErrJudgeRunNotFound }

// --- Finish 冲突仓库 ---
type conflictFinishRepo struct {
    mu      sync.Mutex
    run     domain.JudgeRun
    barrier chan struct{}
    gets    atomic.Int32
}
func newConflictFinishRepo(run domain.JudgeRun) *conflictFinishRepo { return &conflictFinishRepo{run: run, barrier: make(chan struct{})} }
func (r *conflictFinishRepo) Create(_ context.Context, jr domain.JudgeRun) error { r.mu.Lock(); defer r.mu.Unlock(); r.run = jr; return nil }
func (r *conflictFinishRepo) GetByID(_ context.Context, id string) (domain.JudgeRun, error) {
    if r.run.ID != id { return domain.JudgeRun{}, repository.ErrJudgeRunNotFound }
    n := r.gets.Add(1)
    if n == 2 { close(r.barrier) }
    return r.run, nil
}
func (r *conflictFinishRepo) ListBySubmission(_ context.Context, subID string, _, _ int) ([]domain.JudgeRun, error) { if r.run.SubmissionID == subID { return []domain.JudgeRun{r.run}, nil }; return []domain.JudgeRun{}, nil }
func (r *conflictFinishRepo) UpdateRunning(_ context.Context, id string) error { return repository.ErrJudgeRunNotFound }
func (r *conflictFinishRepo) UpdateFinished(_ context.Context, id string, status string, runtimeMS, memoryKB, exitCode int, errMsg string) error {
    if r.run.ID != id { return repository.ErrJudgeRunNotFound }
    <-r.barrier
    r.mu.Lock(); defer r.mu.Unlock()
    if r.run.Status != domain.JudgeRunStatusRunning { return repository.ErrJudgeRunConflict }
    now := time.Now().UTC(); r.run.Status = status; r.run.RuntimeMS = runtimeMS; r.run.MemoryKB = memoryKB; r.run.ExitCode = exitCode; r.run.ErrorMessage = errMsg; r.run.FinishedAt = &now; r.run.UpdatedAt = now; return nil
}

// --- 测试：并发 Start 冲突 ---
func TestJudgeRun_Start_Conflict(t *testing.T) {
    os.Setenv("JWT_SECRET", "test-secret")
    jr := domain.JudgeRun{ID: "jr-start-1", SubmissionID: "sub-jr-1", Status: domain.JudgeRunStatusQueued, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
    repo := newConflictStartRepo(jr)
    deps := router.Dependencies{JudgeRunRepo: repo, SubmissionRepo: repository.NewMemorySubmissionRepository(), Env: "test"}
    // 放一条 submission 以通过 Enqueue 前置校验（直接创建）
    _ = deps.SubmissionRepo.Create(context.Background(), domain.Submission{ID: "sub-jr-1", UserID: "u1", ProblemID: "p", Language: "go", Code: "print", Status: "pending", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Version:1})
    r := router.Setup(deps)
    srv := httptest.NewServer(r); defer srv.Close()
    token := makeToken(t, "test-secret", "u1", []string{auth.RoleSystemAdmin}) // system_admin 拥有 manage 权限

    // 两个并发调用 internal start
    var wg sync.WaitGroup
    type res struct{ code int }
    results := make([]res,2)
    call := func(i int) {
        defer wg.Done()
        req, _ := http.NewRequest(http.MethodPost, srv.URL+"/internal/judge-runs/jr-start-1/start", nil)
        req.Header.Set("Authorization", "Bearer "+token)
        resp, err := http.DefaultClient.Do(req)
        require.NoError(t, err)
        results[i] = res{code: resp.StatusCode}
        resp.Body.Close()
    }
    wg.Add(2); go call(0); go call(1); wg.Wait()
    var okCnt, conflictCnt int
    for _, v := range results { if v.code == 200 { okCnt++ } else if v.code == 409 { conflictCnt++ } }
    require.Equal(t, 1, okCnt, "expect one 200")
    require.Equal(t, 1, conflictCnt, "expect one 409")
}

// --- 测试：并发 Finish 冲突 ---
func TestJudgeRun_Finish_Conflict(t *testing.T) {
    os.Setenv("JWT_SECRET", "test-secret")
    started := time.Now().Add(-2 * time.Second).UTC()
    jr := domain.JudgeRun{ID: "jr-finish-1", SubmissionID: "sub-jr-2", Status: domain.JudgeRunStatusRunning, StartedAt: &started, CreatedAt: started, UpdatedAt: started}
    repo := newConflictFinishRepo(jr)
    deps := router.Dependencies{JudgeRunRepo: repo, SubmissionRepo: repository.NewMemorySubmissionRepository(), Env: "test"}
    _ = deps.SubmissionRepo.Create(context.Background(), domain.Submission{ID: "sub-jr-2", UserID: "u1", ProblemID: "p", Language: "go", Code: "print", Status: "pending", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Version:1})
    r := router.Setup(deps)
    srv := httptest.NewServer(r); defer srv.Close()
    token := makeToken(t, "test-secret", "u1", []string{auth.RoleSystemAdmin})
    body := []byte(`{"status":"succeeded","runtime_ms":10,"memory_kb":1,"exit_code":0}`)
    var wg sync.WaitGroup
    type res struct{ code int }
    results := make([]res,2)
    call := func(i int) {
        defer wg.Done()
        req, _ := http.NewRequest(http.MethodPost, srv.URL+"/internal/judge-runs/jr-finish-1/finish", bytes.NewReader(body))
        req.Header.Set("Authorization", "Bearer "+token)
        req.Header.Set("Content-Type", "application/json")
        resp, err := http.DefaultClient.Do(req)
        require.NoError(t, err)
        results[i] = res{code: resp.StatusCode}
        resp.Body.Close()
    }
    wg.Add(2); go call(0); go call(1); wg.Wait()
    var okCnt, conflictCnt int
    for _, v := range results { if v.code == 200 { okCnt++ } else if v.code == 409 { conflictCnt++ } }
    require.Equal(t, 1, okCnt, "expect one 200")
    require.Equal(t, 1, conflictCnt, "expect one 409")
}
