package handler

import (
	"bytes"
	"context"
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

// Build a router with in-memory repos but without auto permissions; we inject identity manually per test.
func buildMatrixRouter() (*gin.Engine, *memorySubmissionRepo, *memoryJudgeRunRepo, *service.SubmissionService, *service.JudgeRunHTTPAdapter) {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    // identity injection middleware placeholder (overridden in each request via context value set)
    // We'll just set identity in request context using a simple middleware reading headers for perms.
    r.Use(func(c *gin.Context) {
        rawPerms := c.GetHeader("X-Test-Perms")
        id := &auth.Identity{UserID: c.GetHeader("X-Test-User"), Roles: []string{}, Permissions: map[auth.Permission]struct{}{}}
        if id.UserID == "" { id.UserID = "guest" }
        if rawPerms != "" {
            for _, p := range bytes.Split([]byte(rawPerms), []byte{','}) {
                if len(p) == 0 { continue }
                id.Permissions[auth.Permission(string(bytes.TrimSpace(p)))] = struct{}{}
            }
        }
        c.Set("__identity", id)
        c.Next()
    })

    // memory repos
    subRepo := newMemorySubmissionRepo()
    logRepo := &memoryStatusLogRepo{}
    subSvc := service.NewSubmissionService(subRepo, logRepo)

    jrRepo := newMemoryJudgeRunRepo()
    jrSvc := service.NewJudgeRunService(jrRepo)
    jrAdapter := service.NewJudgeRunHTTPAdapter(jrSvc)

    // fixtures: problem and submission owner user u1
    sub := domain.Submission{ID: "s1", UserID: "u1", ProblemID: "p1", Language: "go", Code: "print", Status: service.SubmissionStatusPending, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Version:1}
    _ = subRepo.Create(context.Background(), sub)

    // routes subset we want to test
    r.GET("/submissions/:id", GetSubmission(subSvc))
    r.POST("/submissions/:id/runs", auth.Require(auth.PermJudgeRunEnqueue), EnqueueJudgeRun(jrAdapter, subSvc))
    return r, subRepo, jrRepo, subSvc, jrAdapter
}

type matrixCase struct {
    name           string
    user           string
    perms          []auth.Permission
    path           string
    method         string
    expectedStatus int
}

func TestPermissionMatrix_Basic(t *testing.T) {
    r, _, _, _, _ := buildMatrixRouter()
    cases := []matrixCase{
        {name: "guest get submission no perm", user: "guest", perms: nil, method: http.MethodGet, path: "/submissions/s1", expectedStatus: http.StatusOK}, // code redaction handled elsewhere
        {name: "enqueue missing perm", user: "u1", perms: []auth.Permission{}, method: http.MethodPost, path: "/submissions/s1/runs", expectedStatus: http.StatusForbidden},
        {name: "enqueue with perm", user: "u1", perms: []auth.Permission{auth.PermJudgeRunEnqueue}, method: http.MethodPost, path: "/submissions/s1/runs", expectedStatus: http.StatusCreated},
        {name: "enqueue other user without perm", user: "u2", perms: []auth.Permission{}, method: http.MethodPost, path: "/submissions/s1/runs", expectedStatus: http.StatusForbidden},
    }

    for _, cse := range cases {
        t.Run(cse.name, func(t *testing.T) {
            var body *bytes.Reader
            if cse.method == http.MethodPost && cse.path == "/submissions/s1/runs" {
                body = bytes.NewReader(nil)
            } else { body = bytes.NewReader(nil) }
            req := httptest.NewRequest(cse.method, cse.path, body)
            if cse.user != "" { req.Header.Set("X-Test-User", cse.user) }
            if len(cse.perms) > 0 {
                // join perms
                arr := make([]byte, 0)
                for i, p := range cse.perms { if i>0 { arr = append(arr, ',') }; arr = append(arr, []byte(p)...)}
                req.Header.Set("X-Test-Perms", string(arr))
            }
            w := httptest.NewRecorder(); r.ServeHTTP(w, req)
            require.Equal(t, cse.expectedStatus, w.Code, w.Body.String())
        })
    }
}
