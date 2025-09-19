package router

import (
	"context"
	"os"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/http/handler"
	"github.com/YangYuS8/codyssey/backend/internal/http/middleware"
	"github.com/YangYuS8/codyssey/backend/internal/metrics"
	"github.com/YangYuS8/codyssey/backend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProblemRepo interface {
    Create(ctx context.Context, p domain.Problem) error
    GetByID(ctx context.Context, id uuid.UUID) (domain.Problem, error)
    Update(ctx context.Context, p domain.Problem) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, limit, offset int) ([]domain.Problem, error)
}

type Dependencies struct {
    ProblemRepo ProblemRepo
    UserRepo    service.UserRepo
    AuthService *auth.AuthService
    SubmissionRepo service.SubmissionRepo
    SubmissionStatusLogRepo service.SubmissionStatusLogRepo
    JudgeRunRepo service.JudgeRunRepo
    HealthCheck handler.HealthChecker
    Version     string
    Env         string
}

func Setup(dep Dependencies) *gin.Engine {
    r := gin.New()
    r.Use(gin.Logger(), gin.Recovery(), middleware.TraceID(), metrics.Middleware())
    // 依据 ENV 使用不同身份中间件（默认 development 下允许 debug 头）
    env := dep.Env
    if env == "" { env = os.Getenv("ENV") }
    if env == "development" || env == "test" {
        r.Use(auth.AttachDebugIdentity())
    } else {
        // JWT_SECRET 由 server 侧注入 env；这里直接读取
        r.Use(auth.StrictJWTAuth(os.Getenv("JWT_SECRET")))
    }

    r.GET("/health", handler.Health(dep.Version, dep.Env, dep.HealthCheck))
    r.GET("/metrics", metrics.Handler())
	r.GET("/version", func(c *gin.Context) { c.JSON(200, gin.H{"version": dep.Version}) })

    if dep.ProblemRepo != nil {
        ps := service.NewProblemService(dep.ProblemRepo)
        r.GET("/problems", handler.ListProblems(ps))
        r.POST("/problems", auth.Require(auth.PermProblemCreate), handler.CreateProblem(ps))
        r.GET("/problems/:id", handler.GetProblem(ps))
        r.PUT("/problems/:id", auth.Require(auth.PermProblemUpdate), handler.UpdateProblem(ps))
        r.DELETE("/problems/:id", auth.Require(auth.PermProblemDelete), handler.DeleteProblem(ps))
    }

    if dep.UserRepo != nil {
        us := service.NewUserService(dep.UserRepo)
        r.GET("/users", auth.Require(auth.PermUserList), handler.ListUsers(us))
        r.POST("/users", auth.Require(auth.PermUserCreate), handler.CreateUser(us))
        r.GET("/users/:id", auth.Require(auth.PermUserGet), handler.GetUser(us))
        r.PUT("/users/:id/roles", auth.Require(auth.PermUserUpdateRoles), handler.UpdateUserRoles(us))
        r.DELETE("/users/:id", auth.Require(auth.PermUserDelete), handler.DeleteUser(us))
    }

    if dep.AuthService != nil {
        ah := handler.NewAuthHandlers(dep.AuthService)
        r.POST("/auth/register", ah.Register)
        r.POST("/auth/login", ah.Login)
        r.POST("/auth/refresh", ah.Refresh)
    }

    if dep.SubmissionRepo != nil {
        ss := service.NewSubmissionService(dep.SubmissionRepo, dep.SubmissionStatusLogRepo)
        var jrAdapter *service.JudgeRunHTTPAdapter
        if dep.JudgeRunRepo != nil { jrAdapter = service.NewJudgeRunHTTPAdapter(service.NewJudgeRunService(dep.JudgeRunRepo)) }
        // 创建沿用 handler 内部校验登录，列表与单个获取加精细权限（list / get）
        r.POST("/submissions", handler.CreateSubmission(ss))
        r.GET("/submissions", auth.Require(auth.PermSubmissionList), handler.ListSubmissions(ss))
        r.GET("/submissions/:id", auth.Require(auth.PermSubmissionGet), handler.GetSubmission(ss))
        r.PATCH("/submissions/:id/status", auth.Require(auth.PermSubmissionUpdateStatus), handler.UpdateSubmissionStatus(ss))
        r.GET("/submissions/:id/logs", auth.Require(auth.PermSubmissionGet), handler.ListSubmissionStatusLogs(ss))
        if jrAdapter != nil {
            r.POST("/submissions/:id/runs", auth.Require(auth.PermJudgeRunEnqueue), handler.EnqueueJudgeRun(jrAdapter, ss))
            r.GET("/submissions/:id/runs", auth.Require(auth.PermJudgeRunList), handler.ListJudgeRuns(jrAdapter, ss))
            r.GET("/judge-runs/:id", auth.Require(auth.PermJudgeRunGet), handler.GetJudgeRun(jrAdapter, ss))
            // 内部判题执行控制（仅 system_admin: judge_run.manage）
            r.POST("/internal/judge-runs/:id/start", auth.Require(auth.PermJudgeRunManage), handler.InternalStartJudgeRun(jrAdapter))
            r.POST("/internal/judge-runs/:id/finish", auth.Require(auth.PermJudgeRunManage), handler.InternalFinishJudgeRun(jrAdapter))
        }
    }

	return r
}
