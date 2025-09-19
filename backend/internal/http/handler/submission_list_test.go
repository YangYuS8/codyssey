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

func makeTokenList(t *testing.T, secret, userID string, roles []string, perms []string) string {
    claims := jwt.MapClaims{
        "sub":   userID,
        "roles": roles,
        "perms": perms,
        "exp":   time.Now().Add(30 * time.Minute).Unix(),
        "iat":   time.Now().Unix(),
    }
    tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    s, err := tok.SignedString([]byte(secret))
    require.NoError(t, err)
    return s
}

func TestSubmission_List_Visibility_And_Filter(t *testing.T) {
    os.Setenv("JWT_SECRET", "test-secret")
    repo := repository.NewMemorySubmissionRepository()
    deps := router.Dependencies{SubmissionRepo: repo}
    r := router.Setup(deps)
    srv := httptest.NewServer(r); defer srv.Close()

    tokenA := makeTokenList(t, "test-secret", "userA", []string{auth.RoleStudent}, nil)
    tokenB := makeTokenList(t, "test-secret", "userB", []string{auth.RoleStudent}, nil)
    teacherToken := makeTokenList(t, "test-secret", "teacher1", []string{auth.RoleTeacher}, nil)

    create := func(tok, problemID, code string) {
        body := map[string]string{"problem_id": problemID, "language": "go", "code": code}
        b,_ := json.Marshal(body)
        req,_ := http.NewRequest(http.MethodPost, srv.URL+"/submissions", bytes.NewReader(b))
        req.Header.Set("Authorization", "Bearer "+tok)
        req.Header.Set("Content-Type", "application/json")
        resp, err := http.DefaultClient.Do(req)
        require.NoError(t, err)
        require.Equal(t, http.StatusCreated, resp.StatusCode)
    }

    create(tokenA, "p1", "code-a1")
    create(tokenA, "p2", "code-a2")
    create(tokenB, "p1", "code-b1")

    // 学生 A 列表：应返回 3 条，自己的两条保留 code，别人的清空
    reqA,_ := http.NewRequest(http.MethodGet, srv.URL+"/submissions?limit=10", nil)
    reqA.Header.Set("Authorization", "Bearer "+tokenA)
    respA, err := http.DefaultClient.Do(reqA)
    require.NoError(t, err)
    require.Equal(t, http.StatusOK, respA.StatusCode)
    var listA struct { Data []struct { UserID string `json:"user_id"`; Code string `json:"code"` } `json:"data"`; Meta struct { Total int `json:"total"` } `json:"meta"` }
    _ = json.NewDecoder(respA.Body).Decode(&listA)
    require.Len(t, listA.Data, 3)
    var ownWithCode, othersWithout int
    for _, it := range listA.Data {
        if it.UserID == "userA" {
            if it.Code != "" { ownWithCode++ }
        } else {
            if it.Code == "" { othersWithout++ }
        }
    }
    require.Equal(t, 2, ownWithCode)
    require.Equal(t, 1, othersWithout)
    require.Equal(t, 3, listA.Meta.Total)

    // teacher 列表：全部 3 条且都有 code
    reqT,_ := http.NewRequest(http.MethodGet, srv.URL+"/submissions", nil)
    reqT.Header.Set("Authorization", "Bearer "+teacherToken)
    respT, err := http.DefaultClient.Do(reqT)
    require.NoError(t, err)
    require.Equal(t, http.StatusOK, respT.StatusCode)
    var listT struct { Data []struct { Code string `json:"code"` } `json:"data"`; Meta struct { Total int `json:"total"` } `json:"meta"` }
    _ = json.NewDecoder(respT.Body).Decode(&listT)
    require.Len(t, listT.Data, 3)
    for _, it := range listT.Data { require.NotEmpty(t, it.Code) }
    require.Equal(t, 3, listT.Meta.Total)

    // 过滤 user_id=userB 只剩 1 条
    reqFilter,_ := http.NewRequest(http.MethodGet, srv.URL+"/submissions?user_id=userB", nil)
    reqFilter.Header.Set("Authorization", "Bearer "+teacherToken)
    respFilter, err := http.DefaultClient.Do(reqFilter)
    require.NoError(t, err)
    require.Equal(t, http.StatusOK, respFilter.StatusCode)
    var listFilter struct { Data []struct { UserID string `json:"user_id"` } `json:"data"`; Meta struct { Total int `json:"total"` } `json:"meta"` }
    _ = json.NewDecoder(respFilter.Body).Decode(&listFilter)
    require.Len(t, listFilter.Data, 1)
    require.Equal(t, "userB", listFilter.Data[0].UserID)
    require.Equal(t, 1, listFilter.Meta.Total)
}
