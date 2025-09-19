package handler_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/http/router"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
)

func makeToken(t *testing.T, secret, userID string, roles []string) string {
    claims := jwt.MapClaims{
        "sub":   userID,
        "roles": roles,
        "perms": []string{},
        "exp":   time.Now().Add(30 * time.Minute).Unix(),
        "iat":   time.Now().Unix(),
    }
    tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    s, err := tok.SignedString([]byte(secret))
    require.NoError(t, err)
    return s
}

func TestSubmission_Create_Unauthorized(t *testing.T) {
    os.Setenv("JWT_SECRET", "test-secret")
    memSubRepo := repository.NewMemorySubmissionRepository()
    deps := router.Dependencies{SubmissionRepo: memSubRepo}
    r := router.Setup(deps)
    ts := httptest.NewServer(r); defer ts.Close()
    body := map[string]string{"problem_id":"p1","language":"go","code":"print(1)"}
    b,_ := json.Marshal(body)
    resp, err := http.Post(ts.URL+"/submissions", "application/json", bytes.NewReader(b))
    require.NoError(t, err)
    require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSubmission_Create_And_Get_Visibility(t *testing.T) {
    os.Setenv("JWT_SECRET", "test-secret")
    memSubRepo := repository.NewMemorySubmissionRepository()
    deps := router.Dependencies{SubmissionRepo: memSubRepo}
    r := router.Setup(deps)
    server := httptest.NewServer(r); defer server.Close()

    ownerToken := makeToken(t, "test-secret", "user-1", []string{auth.RoleStudent})
    createReq := map[string]string{"problem_id":"prob-1","language":"python","code":"print('hi')"}
    cb,_ := json.Marshal(createReq)
    req,_ := http.NewRequest(http.MethodPost, server.URL+"/submissions", bytes.NewReader(cb))
    req.Header.Set("Content-Type","application/json")
    req.Header.Set("Authorization", "Bearer "+ownerToken)
    resp, err := http.DefaultClient.Do(req)
    require.NoError(t, err)
    require.Equal(t, http.StatusCreated, resp.StatusCode)
    raw, _ := io.ReadAll(resp.Body)
    t.Logf("create raw=%s", string(raw))
    var created struct { Data domain.Submission `json:"data"` }
    _ = json.Unmarshal(raw, &created)
    require.NotEmpty(t, created.Data.ID)
    require.Equal(t, "print('hi')", created.Data.Code)

    getReq,_ := http.NewRequest(http.MethodGet, server.URL+"/submissions/"+created.Data.ID, nil)
    getReq.Header.Set("Authorization", "Bearer "+ownerToken)
    getResp, err := http.DefaultClient.Do(getReq)
    require.NoError(t, err)
    require.Equal(t, http.StatusOK, getResp.StatusCode)
    var gotOwner struct { Data domain.Submission `json:"data"` }
    _ = json.NewDecoder(getResp.Body).Decode(&gotOwner)
    require.Equal(t, "print('hi')", gotOwner.Data.Code)

    otherToken := makeToken(t, "test-secret", "user-2", []string{auth.RoleStudent})
    otherReq,_ := http.NewRequest(http.MethodGet, server.URL+"/submissions/"+created.Data.ID, nil)
    otherReq.Header.Set("Authorization", "Bearer "+otherToken)
    otherResp, err := http.DefaultClient.Do(otherReq)
    require.NoError(t, err)
    require.Equal(t, http.StatusOK, otherResp.StatusCode)
    var gotOther struct { Data domain.Submission `json:"data"` }
    _ = json.NewDecoder(otherResp.Body).Decode(&gotOther)
    require.Empty(t, gotOther.Data.Code)

    teacherToken := makeToken(t, "test-secret", "teacher-1", []string{auth.RoleTeacher})
    teacherReq,_ := http.NewRequest(http.MethodGet, server.URL+"/submissions/"+created.Data.ID, nil)
    teacherReq.Header.Set("Authorization", "Bearer "+teacherToken)
    teacherResp, err := http.DefaultClient.Do(teacherReq)
    require.NoError(t, err)
    require.Equal(t, http.StatusOK, teacherResp.StatusCode)
    var gotTeacher struct { Data domain.Submission `json:"data"` }
    _ = json.NewDecoder(teacherResp.Body).Decode(&gotTeacher)
    require.Equal(t, "print('hi')", gotTeacher.Data.Code)
}

