package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/http/router"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
)

// TestMetrics_JudgeRunDuration 验证 finish 后 /metrics 暴露 judge_run_duration_seconds
func TestMetrics_JudgeRunDuration(t *testing.T) {
    os.Setenv("JWT_SECRET", "test-secret")
    subRepo := repository.NewMemorySubmissionRepository()
    jrRepo := repository.NewMemoryJudgeRunRepository()
    deps := router.Dependencies{SubmissionRepo: subRepo, JudgeRunRepo: jrRepo, Env: "test"}
    r := router.Setup(deps)
    srv := httptest.NewServer(r); defer srv.Close()

    // 创建 submission
    tok := makeToken(t, "test-secret", "u1", []string{auth.RoleSystemAdmin})
    subBody := map[string]string{"problem_id":"p1","language":"go","code":"print"}
    b,_ := json.Marshal(subBody)
    req,_ := http.NewRequest(http.MethodPost, srv.URL+"/submissions", bytes.NewReader(b))
    req.Header.Set("Authorization", "Bearer "+tok)
    req.Header.Set("Content-Type", "application/json")
    resp, err := http.DefaultClient.Do(req); require.NoError(t, err); require.Equal(t, http.StatusCreated, resp.StatusCode)
    var created struct{ Data domain.Submission `json:"data"` }
    _ = json.NewDecoder(resp.Body).Decode(&created); resp.Body.Close()

    // enqueue judge run
    runReq,_ := http.NewRequest(http.MethodPost, srv.URL+"/submissions/"+created.Data.ID+"/runs", nil)
    runReq.Header.Set("Authorization", "Bearer "+tok)
    runResp, err := http.DefaultClient.Do(runReq); require.NoError(t, err); require.Equal(t, http.StatusCreated, runResp.StatusCode)
    var runObj struct{ Data struct{ ID string `json:"id"` } `json:"data"` }
    _ = json.NewDecoder(runResp.Body).Decode(&runObj); runResp.Body.Close()

    // internal start
    startReq,_ := http.NewRequest(http.MethodPost, srv.URL+"/internal/judge-runs/"+runObj.Data.ID+"/start", nil)
    startReq.Header.Set("Authorization", "Bearer "+tok)
    startResp, err := http.DefaultClient.Do(startReq); require.NoError(t, err); require.Equal(t, http.StatusOK, startResp.StatusCode); startResp.Body.Close()

    // 模拟运行一些时间
    time.Sleep(10 * time.Millisecond)

    // internal finish
    finishPayload := []byte(`{"status":"succeeded","runtime_ms":5,"memory_kb":1,"exit_code":0}`)
    finReq,_ := http.NewRequest(http.MethodPost, srv.URL+"/internal/judge-runs/"+runObj.Data.ID+"/finish", bytes.NewReader(finishPayload))
    finReq.Header.Set("Authorization", "Bearer "+tok)
    finReq.Header.Set("Content-Type", "application/json")
    finResp, err := http.DefaultClient.Do(finReq); require.NoError(t, err); require.Equal(t, http.StatusOK, finResp.StatusCode); finResp.Body.Close()

    // 获取 metrics
    mResp, err := http.Get(srv.URL+"/metrics")
    require.NoError(t, err)
    raw, _ := io.ReadAll(mResp.Body); mResp.Body.Close()
    text := string(raw)
    require.True(t, strings.Contains(text, "judge_run_duration_seconds_bucket"), "metrics should contain histogram buckets, got:\n%s", text)
}
