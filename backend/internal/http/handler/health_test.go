package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/YangYuS8/codyssey/backend/internal/http/handler"
	"github.com/gin-gonic/gin"
)

type fakeHealth struct{ up bool }
func (f fakeHealth) DBAlive() bool { return f.up }

func TestHealth_Up(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", handler.Health("v-test", "test", fakeHealth{up: true}))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK { t.Fatalf("expected 200 got %d", w.Code) }
	ct := w.Header().Get("Content-Type")
	if ct == "" { t.Errorf("missing content-type") }
}

func TestHealth_Down(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", handler.Health("v-test", "test", fakeHealth{up: false}))
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK { t.Fatalf("expected 200 got %d", w.Code) }
}
