package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	promhttp "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Global collectors (minimal set) – kept package-level for simple use.
var (
    reg *prometheus.Registry

    httpRequestsTotal *prometheus.CounterVec
    httpRequestDuration *prometheus.HistogramVec
    httpInFlight prometheus.Gauge

    submissionTransitions *prometheus.CounterVec
    judgeRunTransitions *prometheus.CounterVec
    judgeRunDuration *prometheus.HistogramVec
    submissionConflicts prometheus.Counter
    judgeRunConflicts prometheus.Counter
)

// Init initializes the metrics registry and registers collectors. Safe to call once.
func Init() {
    if reg != nil { return }
    reg = prometheus.NewRegistry()

    // Core HTTP metrics
    httpRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
        Namespace: "codyssey",
        Name:      "http_requests_total",
        Help:      "Total number of HTTP requests received.",
    }, []string{"method", "route", "status"})

    httpRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
        Namespace: "codyssey",
        Name:      "http_request_duration_seconds",
        Help:      "Histogram of HTTP request durations in seconds.",
        // Buckets tuned roughly for web API (ms to multiple seconds)
        Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
    }, []string{"method", "route"})

    httpInFlight = prometheus.NewGauge(prometheus.GaugeOpts{
        Namespace: "codyssey",
        Name:      "http_in_flight_requests",
        Help:      "Current number of in-flight HTTP requests.",
    })

    submissionTransitions = prometheus.NewCounterVec(prometheus.CounterOpts{
        Namespace: "codyssey",
        Name:      "submission_status_transitions_total",
        Help:      "Count of submission status transitions by from->to.",
    }, []string{"from", "to"})

    judgeRunTransitions = prometheus.NewCounterVec(prometheus.CounterOpts{
        Namespace: "codyssey",
        Name:      "judge_run_status_transitions_total",
        Help:      "Count of judge run status transitions by from->to.",
    }, []string{"from", "to"})

    judgeRunDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
        Namespace: "codyssey",
        Name:      "judge_run_duration_seconds",
        Help:      "Histogram of judge run (start->finish) durations in seconds.",
        Buckets:   []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10, 20},
    }, []string{"status"})

    submissionConflicts = prometheus.NewCounter(prometheus.CounterOpts{
        Namespace: "codyssey",
        Name:      "submission_conflicts_total",
        Help:      "Total number of submission status update conflicts (optimistic lock).",
    })
    judgeRunConflicts = prometheus.NewCounter(prometheus.CounterOpts{
        Namespace: "codyssey",
        Name:      "judge_run_conflicts_total",
        Help:      "Total number of judge run status update conflicts (optimistic lock).",
    })

    _ = reg.Register(httpRequestsTotal)
    _ = reg.Register(httpRequestDuration)
    _ = reg.Register(httpInFlight)
    _ = reg.Register(submissionTransitions)
    _ = reg.Register(judgeRunTransitions)
    _ = reg.Register(judgeRunDuration)
    _ = reg.Register(submissionConflicts)
    _ = reg.Register(judgeRunConflicts)
}

// Middleware instruments HTTP requests. Should be added high in the chain after recovery & trace.
func Middleware() gin.HandlerFunc {
    Init()
    return func(c *gin.Context) {
        start := time.Now()
        httpInFlight.Inc()
        defer httpInFlight.Dec()

        c.Next()

        route := c.FullPath()
        if route == "" { // for unmapped 404 etc.
            route = "unknown"
        }
        method := c.Request.Method
        status := c.Writer.Status()
        httpRequestsTotal.WithLabelValues(method, route, intToStr(status)).Inc()
        httpRequestDuration.WithLabelValues(method, route).Observe(time.Since(start).Seconds())
    }
}

// Handler exposes the /metrics endpoint.
func Handler() gin.HandlerFunc {
    Init()
    h := promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: true})
    return func(c *gin.Context) { h.ServeHTTP(c.Writer, c.Request) }
}

// ObserveSubmissionTransition increments the submission status transition counter.
func ObserveSubmissionTransition(from, to string) {
    if submissionTransitions == nil { return }
    submissionTransitions.WithLabelValues(from, to).Inc()
}

// ObserveJudgeRunTransition increments the judge run status transition counter.
func ObserveJudgeRunTransition(from, to string) {
    if judgeRunTransitions == nil { return }
    judgeRunTransitions.WithLabelValues(from, to).Inc()
}

// ObserveJudgeRunDuration 记录一次运行的总耗时（仅在 start & finish 时间都存在时）
func ObserveJudgeRunDuration(status string, startedAt, finishedAt *time.Time) {
    if judgeRunDuration == nil || startedAt == nil || finishedAt == nil { return }
    if finishedAt.Before(*startedAt) { return }
    judgeRunDuration.WithLabelValues(status).Observe(finishedAt.Sub(*startedAt).Seconds())
}

// IncSubmissionConflict increments submission conflict counter.
func IncSubmissionConflict() { if submissionConflicts != nil { submissionConflicts.Inc() } }

// IncJudgeRunConflict increments judge run conflict counter.
func IncJudgeRunConflict() { if judgeRunConflicts != nil { judgeRunConflicts.Inc() } }

// intToStr – small helper without importing strconv repeatedly.
func intToStr(i int) string {
    // hand-written fast path for common statuses; fallback minimal alloc.
    switch i {
    case 200: return "200"
    case 201: return "201"
    case 400: return "400"
    case 401: return "401"
    case 403: return "403"
    case 404: return "404"
    case 500: return "500"
    default:
        // inline simple conversion
        return fmtInt(i)
    }
}

func fmtInt(i int) string { // minimal int -> string (positive only for status codes)
    if i == 0 { return "0" }
    buf := [6]byte{} // status codes < 1000
    pos := len(buf)
    for i > 0 {
        pos--
        buf[pos] = byte('0' + i%10)
        i /= 10
    }
    return string(buf[pos:])
}
// end of file
