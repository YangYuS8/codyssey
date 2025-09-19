package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/http/router"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
)

func makeTokenWithPerms(t *testing.T, secret, userID string, roles []string, perms []string) string {
    claims := jwt.MapClaims{
        "sub":   userID,
        "roles": roles,
        "perms": perms,
        "exp":   time.Now().Add(30 * time.Minute).Unix(),
        "iat":   time.Now().Unix(),
    }
    tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    out, err := tok.SignedString([]byte(secret))
    require.NoError(t, err)
    return out
}

func createSubmission(t *testing.T, baseURL, token string, body map[string]string) string {
    b,_ := json.Marshal(body)
    req,_ := http.NewRequest(http.MethodPost, baseURL+"/submissions", bytes.NewReader(b))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")
    resp, err := http.DefaultClient.Do(req)
    require.NoError(t, err)
    require.Equal(t, http.StatusCreated, resp.StatusCode)
    var created struct { Data struct { ID string `json:"id"` } `json:"data"` }
    _ = json.NewDecoder(resp.Body).Decode(&created)
    return created.Data.ID
}

func patchStatus(t *testing.T, baseURL, token, id, status string) (*http.Response, []byte) {
    body := map[string]string{"status": status}
    b,_ := json.Marshal(body)
    req,_ := http.NewRequest(http.MethodPatch, baseURL+"/submissions/"+id+"/status", bytes.NewReader(b))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")
    resp, err := http.DefaultClient.Do(req)
    require.NoError(t, err)
    raw := new(bytes.Buffer)
    _, _ = raw.ReadFrom(resp.Body)
    return resp, raw.Bytes()
}

func TestSubmission_Status_Transitions(t *testing.T) {
    os.Setenv("JWT_SECRET", "test-secret")
    repo := repository.NewMemorySubmissionRepository()
    logRepo := repository.NewMemorySubmissionStatusLogRepository()
    deps := router.Dependencies{SubmissionRepo: repo, SubmissionStatusLogRepo: logRepo}
    r := router.Setup(deps)
    srv := httptest.NewServer(r); defer srv.Close()

    teacherToken := makeTokenWithPerms(t, "test-secret", "teacher-1", []string{auth.RoleTeacher}, nil)
    studentToken := makeTokenWithPerms(t, "test-secret", "stu-1", []string{auth.RoleStudent}, nil)

    subID := createSubmission(t, srv.URL, teacherToken, map[string]string{"problem_id":"p1","language":"go","code":"print()"})

    // 学生尝试更新状态 -> 403
    respForbidden, _ := patchStatus(t, srv.URL, studentToken, subID, "judging")
    require.Equal(t, http.StatusForbidden, respForbidden.StatusCode)

    // 合法流转 pending->judging ->accepted 并检查 version 递增
    resp1, body1 := patchStatus(t, srv.URL, teacherToken, subID, "judging")
    require.Equal(t, http.StatusOK, resp1.StatusCode, string(body1))
    // 读取 version1
    var env1 struct { Data struct { Version int `json:"version"` } `json:"data"` }
    _ = json.Unmarshal(body1, &env1)
    v1 := env1.Data.Version
    resp2, body2 := patchStatus(t, srv.URL, teacherToken, subID, "accepted")
    require.Equal(t, http.StatusOK, resp2.StatusCode, string(body2))
    var env2 struct { Data struct { Version int `json:"version"` } `json:"data"` }
    _ = json.Unmarshal(body2, &env2)
    require.Greater(t, env2.Data.Version, v1)

    // 非法状态值
    respInvalid, _ := patchStatus(t, srv.URL, teacherToken, subID, "weird")
    require.Equal(t, http.StatusBadRequest, respInvalid.StatusCode)

    // 已 accepted 再回到 judging -> 400
    respBack, _ := patchStatus(t, srv.URL, teacherToken, subID, "judging")
    require.Equal(t, http.StatusBadRequest, respBack.StatusCode)
}

func TestSubmission_StatusLogs(t *testing.T) {
    os.Setenv("JWT_SECRET", "test-secret")
    repo := repository.NewMemorySubmissionRepository()
    logRepo := repository.NewMemorySubmissionStatusLogRepository()
    deps := router.Dependencies{SubmissionRepo: repo, SubmissionStatusLogRepo: logRepo}
    r := router.Setup(deps)
    srv := httptest.NewServer(r); defer srv.Close()

    teacherToken := makeTokenWithPerms(t, "test-secret", "teacher-1", []string{auth.RoleTeacher}, nil)

    subID := createSubmission(t, srv.URL, teacherToken, map[string]string{"problem_id":"p1","language":"go","code":"print()"})

    // 直接从 pending -> accepted（允许的快速终态），再 -> error (不允许，应失败)
    respFast, _ := patchStatus(t, srv.URL, teacherToken, subID, "accepted")
    require.Equal(t, http.StatusOK, respFast.StatusCode)
    respInvalid, _ := patchStatus(t, srv.URL, teacherToken, subID, "error")
    require.Equal(t, http.StatusBadRequest, respInvalid.StatusCode)

    // 获取日志
    req,_ := http.NewRequest(http.MethodGet, srv.URL+"/submissions/"+subID+"/logs", nil)
    req.Header.Set("Authorization", "Bearer "+teacherToken)
    resp, err := http.DefaultClient.Do(req)
    require.NoError(t, err)
    require.Equal(t, http.StatusOK, resp.StatusCode)
    var logsEnv struct { Data []struct { From string `json:"from_status"`; To string `json:"to_status"` } `json:"data"` }
    _ = json.NewDecoder(resp.Body).Decode(&logsEnv)
    // 目前实现中日志只记录变更（pending->accepted）一条
    require.Len(t, logsEnv.Data, 1)
    if len(logsEnv.Data) == 1 {
        require.Equal(t, "pending", logsEnv.Data[0].From)
        require.Equal(t, "accepted", logsEnv.Data[0].To)
    }
}
