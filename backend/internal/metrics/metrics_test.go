package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// buildTestRouter creates a minimal gin engine with metrics middleware & endpoint.
func buildTestRouter() *gin.Engine {
    gin.SetMode(gin.TestMode)
    r := gin.New()
    r.Use(Middleware())
    r.GET("/ping", func(c *gin.Context) { c.String(200, "pong") })
    r.GET("/metrics", Handler())
    return r
}

func TestMetricsEndpoint_Basic(t *testing.T) {
    r := buildTestRouter()
    // hit one endpoint to generate metrics
    req := httptest.NewRequest(http.MethodGet, "/ping", nil)
    w := httptest.NewRecorder(); r.ServeHTTP(w, req)
    if w.Code != 200 { t.Fatalf("expected 200, got %d", w.Code) }

    // scrape metrics
    mreq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
    mw := httptest.NewRecorder(); r.ServeHTTP(mw, mreq)
    if mw.Code != 200 { t.Fatalf("/metrics expected 200, got %d", mw.Code) }
    body := mw.Body.String()
    // basic assertions: presence of our metric names
    expected := []string{
        "codyssey_http_requests_total",
        "codyssey_http_request_duration_seconds_bucket",
        "codyssey_http_in_flight_requests",
    }
    for _, e := range expected {
        if !strings.Contains(body, e) {
            t.Fatalf("expected metrics body to contain %s", e)
        }
    }
}
