package router

import (
	"context"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/http/handler"
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
    HealthCheck handler.HealthChecker
    Version     string
    Env         string
}

func Setup(dep Dependencies) *gin.Engine {
    r := gin.New()
    r.Use(gin.Logger(), gin.Recovery(), auth.AttachDebugIdentity())

	r.GET("/health", handler.Health(dep.Version, dep.Env, dep.HealthCheck))
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

	return r
}
