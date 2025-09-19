package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/http/handler"
	"github.com/YangYuS8/codyssey/backend/internal/service"
	"github.com/gin-gonic/gin"
)

type memUserRepo struct { items []domain.User }

func (m *memUserRepo) Create(_ context.Context, u domain.User) error { m.items = append([]domain.User{u}, m.items...); return nil }
func (m *memUserRepo) GetByID(_ context.Context, id string) (domain.User, error) {
    for _, it := range m.items { if it.ID == id { return it, nil } }
    return domain.User{}, service.ErrUserNotFound
}
func (m *memUserRepo) GetByUsername(_ context.Context, username string) (domain.User, error) {
    for _, it := range m.items { if it.Username == username { return it, nil } }
    return domain.User{}, service.ErrUserNotFound
}
func (m *memUserRepo) UpdateRoles(_ context.Context, id string, roles []string) error {
    for i, it := range m.items { if it.ID == id { m.items[i].Roles = roles; return nil } }
    return service.ErrUserNotFound
}
func (m *memUserRepo) Delete(_ context.Context, id string) error {
    for i, it := range m.items { if it.ID == id { m.items = append(m.items[:i], m.items[i+1:]...); return nil } }
    return service.ErrUserNotFound
}
func (m *memUserRepo) List(_ context.Context, limit, offset int) ([]domain.User, error) { return m.items, nil }

// adapt interface expected by service (context.Context used but tests ignore)
// Provide wrappers with context.Context signature matching service.UserRepo
// But our service expects context.Context; test repo uses any to keep simple -> we'll adjust service to accept context.Context so this compiles.

func setupUserRouter(repo service.UserRepo) *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(auth.AttachDebugIdentity())
    us := service.NewUserService(repo)
    r.POST("/users", auth.Require(auth.PermUserCreate), handler.CreateUser(us))
    r.GET("/users", auth.Require(auth.PermUserList), handler.ListUsers(us))
    r.GET("/users/:id", auth.Require(auth.PermUserGet), handler.GetUser(us))
    r.PUT("/users/:id/roles", auth.Require(auth.PermUserUpdateRoles), handler.UpdateUserRoles(us))
    r.DELETE("/users/:id", auth.Require(auth.PermUserDelete), handler.DeleteUser(us))
    return r
}

func TestUserCRUDLifecycle(t *testing.T) {
    repo := &memUserRepo{}
    r := setupUserRouter(repo)

    // create without permission
    wNo := httptest.NewRecorder()
    reqNo, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader([]byte(`{"username":"alice"}`)))
    r.ServeHTTP(wNo, reqNo)
    if wNo.Code != http.StatusForbidden { t.Fatalf("expected 403 got %d", wNo.Code) }

    // create
    body := map[string]any{"username": "alice", "roles": []string{"student"}}
    b, _ := json.Marshal(body)
    w := httptest.NewRecorder()
    req, _ := http.NewRequest(http.MethodPost, "/users", bytes.NewReader(b))
    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("X-Debug-Perms", "user.create,user.list,user.get,user.update_roles,user.delete")
    r.ServeHTTP(w, req)
    if w.Code != http.StatusCreated { t.Fatalf("create expected 201 got %d body=%s", w.Code, w.Body.String()) }
    var created struct { Data domain.User }
    _ = json.Unmarshal(w.Body.Bytes(), &created)
    if created.Data.Username != "alice" { t.Fatalf("username mismatch") }
    if created.Data.ID == "" { t.Fatalf("expected id generated") }
    uid := created.Data.ID

    // list
    wList := httptest.NewRecorder()
    reqList, _ := http.NewRequest(http.MethodGet, "/users", nil)
    reqList.Header.Set("X-Debug-Perms", "user.list")
    r.ServeHTTP(wList, reqList)
    if wList.Code != http.StatusOK { t.Fatalf("list expected 200 got %d", wList.Code) }

    // get
    wGet := httptest.NewRecorder()
    reqGet, _ := http.NewRequest(http.MethodGet, "/users/"+uid, nil)
    reqGet.Header.Set("X-Debug-Perms", "user.get")
    r.ServeHTTP(wGet, reqGet)
    if wGet.Code != http.StatusOK { t.Fatalf("get expected 200 got %d", wGet.Code) }

    // update roles
    wUp := httptest.NewRecorder()
    upBody, _ := json.Marshal(map[string]any{"roles": []string{"student","contestant"}})
    reqUp, _ := http.NewRequest(http.MethodPut, "/users/"+uid+"/roles", bytes.NewReader(upBody))
    reqUp.Header.Set("Content-Type", "application/json")
    reqUp.Header.Set("X-Debug-Perms", "user.update_roles")
    r.ServeHTTP(wUp, reqUp)
    if wUp.Code != http.StatusOK { t.Fatalf("update roles expected 200 got %d", wUp.Code) }

    // delete
    wDel := httptest.NewRecorder()
    reqDel, _ := http.NewRequest(http.MethodDelete, "/users/"+uid, nil)
    reqDel.Header.Set("X-Debug-Perms", "user.delete")
    r.ServeHTTP(wDel, reqDel)
    if wDel.Code != http.StatusOK { t.Fatalf("delete expected 200 got %d", wDel.Code) }

    // get not found
    wGet2 := httptest.NewRecorder()
    reqGet2, _ := http.NewRequest(http.MethodGet, "/users/"+uid, nil)
    reqGet2.Header.Set("X-Debug-Perms", "user.get")
    r.ServeHTTP(wGet2, reqGet2)
    if wGet2.Code != http.StatusNotFound { t.Fatalf("expected 404 got %d", wGet2.Code) }

    _ = time.Now() // silence imported time if unused later
}
