package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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

// conflictMemorySubmissionRepo 用于模拟第一次状态更新成功、第二次返回冲突。
type conflictMemorySubmissionRepo struct {
    mu        sync.Mutex
    sub       domain.Submission
    baseStatus string
    updated   atomic.Int32 // 第一次成功后标记，其余返回冲突
}

func newConflictRepo(initial domain.Submission) *conflictMemorySubmissionRepo {
    return &conflictMemorySubmissionRepo{sub: initial, baseStatus: initial.Status}
}

func (m *conflictMemorySubmissionRepo) Create(_ context.Context, _ domain.Submission) error { return nil }

// 为兼容 service.SubmissionRepo 接口，形参使用 context.Context 但避免直接引入以减少样板
func (m *conflictMemorySubmissionRepo) GetByID(_ context.Context, id string) (domain.Submission, error) {
    if m.sub.ID != id { return domain.Submission{}, repository.ErrSubmissionNotFound }
    m.mu.Lock(); defer m.mu.Unlock()
    // 始终返回初始状态（pending），保证两个并发请求在服务层看到相同 fromStatus
    copy := m.sub
    copy.Status = m.baseStatus
    return copy, nil
}
func (m *conflictMemorySubmissionRepo) UpdateStatus(_ context.Context, id string, status string, expectedVersion int) error {
    m.mu.Lock(); defer m.mu.Unlock()
    if m.sub.ID != id { return repository.ErrSubmissionNotFound }
    if m.sub.Version != expectedVersion { return repository.ErrSubmissionConflict }
    if m.updated.Load() > 0 { return repository.ErrSubmissionConflict }
    // 第一次成功更新
    m.sub.Status = status
    m.sub.Version += 1
    m.sub.UpdatedAt = time.Now().UTC()
    m.updated.Add(1)
    return nil
}
func (m *conflictMemorySubmissionRepo) List(_ context.Context, _ repository.SubmissionFilter, _, _ int) ([]domain.Submission, error) { return []domain.Submission{m.sub}, nil }
func (m *conflictMemorySubmissionRepo) Count(_ context.Context, _ repository.SubmissionFilter) (int, error) { return 1, nil }

// TestSubmission_StatusUpdate_Conflict 验证并发状态更新产生 409 CONFLICT。
func TestSubmission_StatusUpdate_Conflict(t *testing.T) {
    os.Setenv("JWT_SECRET", "test-secret")
    // 初始 submission：pending
    sub := domain.Submission{ID: "sub-conflict-1", UserID: "u1", ProblemID: "p1", Language: "go", Code: "print", Status: "pending", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Version: 1}
    repo := newConflictRepo(sub)
    deps := router.Dependencies{SubmissionRepo: repo, Env: "test"}
    r := router.Setup(deps)
    srv := httptest.NewServer(r); defer srv.Close()

    // 学生角色默认没有 submission.update_status；这里通过 debug header 叠加权限（路由在 test env 使用 AttachDebugIdentity）
    token := makeToken(t, "test-secret", "u1", []string{auth.RoleStudent})

    // 构造两个并发 PATCH：pending -> judging
    body := map[string]string{"status":"judging"}
    payload, _ := json.Marshal(body)

    type result struct { code int; raw []byte }
    var wg sync.WaitGroup
    results := make([]result, 2)
    doReq := func(idx int) {
        defer wg.Done()
        req, _ := http.NewRequest(http.MethodPatch, srv.URL+"/submissions/"+sub.ID+"/status", bytes.NewReader(payload))
        req.Header.Set("Authorization", "Bearer "+token)
        req.Header.Set("Content-Type", "application/json")
    // 通过 X-Debug-Perms 为请求追加 submission.update_status 权限
    req.Header.Set("X-Debug-Perms", string(auth.PermSubmissionUpdateStatus))
    resp, err := http.DefaultClient.Do(req)
        require.NoError(t, err)
        defer resp.Body.Close()
        b, _ := io.ReadAll(resp.Body)
        results[idx] = result{code: resp.StatusCode, raw: b}
    }
    wg.Add(2)
    go doReq(0)
    go doReq(1)
    wg.Wait()

    // 断言：一个 200，一个 409
    var okCnt, conflictCnt int
    for _, r := range results { if r.code == http.StatusOK { okCnt++ } else if r.code == http.StatusConflict { conflictCnt++ } }
    require.Equal(t, 1, okCnt, "expect one success 200, results=%+v", results)
    require.Equal(t, 1, conflictCnt, "expect one conflict 409, results=%+v", results)
}
