package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/http/handler"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
	"github.com/YangYuS8/codyssey/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type memoryRepo struct { items []domain.Problem }

func (m *memoryRepo) Create(ctx context.Context, p domain.Problem) error { m.items = append([]domain.Problem{p}, m.items...); return nil }
func (m *memoryRepo) List(ctx context.Context, limit, offset int) ([]domain.Problem, error) { return m.items, nil }
func (m *memoryRepo) GetByID(ctx context.Context, id uuid.UUID) (domain.Problem, error) {
	for _, it := range m.items { if it.ID == id { return it, nil } }
	return domain.Problem{}, repository.ErrNotFound
}
func (m *memoryRepo) Update(ctx context.Context, p domain.Problem) error {
	for i, it := range m.items { if it.ID == p.ID { m.items[i] = p; return nil } }
	return repository.ErrNotFound
}
func (m *memoryRepo) Delete(ctx context.Context, id uuid.UUID) error {
	for i, it := range m.items { if it.ID == id { m.items = append(m.items[:i], m.items[i+1:]...); return nil } }
	return repository.ErrNotFound
}

type noDB struct{}
func (noDB) DBAlive() bool { return true }

func setupProblemRouter(repo service.ProblemRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// 附加调试身份中间件
	r.Use(auth.AttachDebugIdentity())
	ps := service.NewProblemService(repo)
	r.POST("/problems", auth.Require(auth.PermProblemCreate), handler.CreateProblem(ps))
	r.GET("/problems", handler.ListProblems(ps))
	r.GET("/problems/:id", handler.GetProblem(ps))
	r.PUT("/problems/:id", auth.Require(auth.PermProblemUpdate), handler.UpdateProblem(ps))
	r.DELETE("/problems/:id", auth.Require(auth.PermProblemDelete), handler.DeleteProblem(ps))
	return r
}

func TestProblemCRUDLifecycle(t *testing.T) {
	repo := &memoryRepo{}
	r := setupProblemRouter(repo)

	// 先测试未授权创建（缺少 create 权限）
	wUnauth := httptest.NewRecorder()
	reqUnauth, _ := http.NewRequest(http.MethodPost, "/problems", bytes.NewReader([]byte(`{"title":"X","description":"Y"}`)))
	reqUnauth.Header.Set("Content-Type", "application/json")
	// 不设置 X-Debug-Perms -> 只有 read
	r.ServeHTTP(wUnauth, reqUnauth)
	if wUnauth.Code != http.StatusForbidden { t.Fatalf("unauthorized create expected 403 got %d", wUnauth.Code) }

	// ---- Create ----
	createBody := map[string]string{"title": "A Problem", "description": "Desc..."}
	cb, _ := json.Marshal(createBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/problems", bytes.NewReader(cb))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Debug-Perms", "problem.create,problem.update,problem.delete")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated { t.Fatalf("create expected 201 got %d", w.Code) }
	var createResp struct { Data domain.Problem; Error *struct{Code string} }
	if err := json.Unmarshal(w.Body.Bytes(), &createResp); err != nil { t.Fatalf("decode create: %v", err) }
	if createResp.Error != nil { t.Fatalf("unexpected error: %+v", createResp.Error) }
	pid := createResp.Data.ID

	// ---- Get ----
	wGet := httptest.NewRecorder()
	reqGet, _ := http.NewRequest(http.MethodGet, "/problems/"+pid.String(), nil)
	r.ServeHTTP(wGet, reqGet)
	if wGet.Code != http.StatusOK { t.Fatalf("get expected 200 got %d", wGet.Code) }

	// ---- Update ----
	newTitle := "Updated Title"
	upBody := map[string]string{"title": newTitle}
	ub, _ := json.Marshal(upBody)
	wUp := httptest.NewRecorder()
	reqUp, _ := http.NewRequest(http.MethodPut, "/problems/"+pid.String(), bytes.NewReader(ub))
	reqUp.Header.Set("Content-Type", "application/json")
	reqUp.Header.Set("X-Debug-Perms", "problem.update,problem.delete")
	r.ServeHTTP(wUp, reqUp)
	if wUp.Code != http.StatusOK { t.Fatalf("update expected 200 got %d", wUp.Code) }
	var upResp struct { Data domain.Problem }
	_ = json.Unmarshal(wUp.Body.Bytes(), &upResp)
	if upResp.Data.Title != newTitle { t.Fatalf("title not updated: %s", upResp.Data.Title) }

	// ---- List ----
	wList := httptest.NewRecorder()
	reqList, _ := http.NewRequest(http.MethodGet, "/problems?limit=10&offset=0", nil)
	r.ServeHTTP(wList, reqList)
	if wList.Code != http.StatusOK { t.Fatalf("list expected 200 got %d", wList.Code) }
	var listResp struct { Data []domain.Problem; Meta map[string]int }
	_ = json.Unmarshal(wList.Body.Bytes(), &listResp)
	if len(listResp.Data) == 0 { t.Fatalf("expected at least one problem in list") }

	// ---- Delete ----
	wDel := httptest.NewRecorder()
	reqDel, _ := http.NewRequest(http.MethodDelete, "/problems/"+pid.String(), nil)
	reqDel.Header.Set("X-Debug-Perms", "problem.delete")
	r.ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusOK { t.Fatalf("delete expected 200 got %d", wDel.Code) }

	// ---- Get NotFound ----
	wGet2 := httptest.NewRecorder()
	reqGet2, _ := http.NewRequest(http.MethodGet, "/problems/"+pid.String(), nil)
	r.ServeHTTP(wGet2, reqGet2)
	if wGet2.Code != http.StatusNotFound { t.Fatalf("expected 404 got %d", wGet2.Code) }
}
