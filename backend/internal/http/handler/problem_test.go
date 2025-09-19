package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/codyssey/backend/internal/domain"
	"github.com/your-org/codyssey/backend/internal/http/handler"
	"github.com/your-org/codyssey/backend/internal/repository"
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

func setupProblemRouter(repo handler.ProblemRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/problems", handler.CreateProblem(repo))
	r.GET("/problems", handler.ListProblems(repo))
	r.GET("/problems/:id", handler.GetProblem(repo))
	r.PUT("/problems/:id", handler.UpdateProblem(repo))
	r.DELETE("/problems/:id", handler.DeleteProblem(repo))
	return r
}

func TestProblemCRUDLifecycle(t *testing.T) {
	repo := &memoryRepo{}
	r := setupProblemRouter(repo)

	// ---- Create ----
	createBody := map[string]string{"title": "A Problem", "description": "Desc..."}
	cb, _ := json.Marshal(createBody)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/problems", bytes.NewReader(cb))
	req.Header.Set("Content-Type", "application/json")
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
	r.ServeHTTP(wDel, reqDel)
	if wDel.Code != http.StatusOK { t.Fatalf("delete expected 200 got %d", wDel.Code) }

	// ---- Get NotFound ----
	wGet2 := httptest.NewRecorder()
	reqGet2, _ := http.NewRequest(http.MethodGet, "/problems/"+pid.String(), nil)
	r.ServeHTTP(wGet2, reqGet2)
	if wGet2.Code != http.StatusNotFound { t.Fatalf("expected 404 got %d", wGet2.Code) }
}
