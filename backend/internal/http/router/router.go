package router

import (
	"github.com/gin-gonic/gin"
	"github.com/your-org/codyssey/backend/internal/http/handler"
)

type Dependencies struct {
	ProblemRepo handler.ProblemRepo
	HealthCheck handler.HealthChecker
	Version     string
    Env         string
}

func Setup(dep Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", handler.Health(dep.Version, dep.Env, dep.HealthCheck))
	r.GET("/version", func(c *gin.Context) { c.JSON(200, gin.H{"version": dep.Version}) })

	if dep.ProblemRepo != nil {
		pr := dep.ProblemRepo
		r.GET("/problems", handler.ListProblems(pr))
		r.POST("/problems", handler.CreateProblem(pr))
		r.GET("/problems/:id", handler.GetProblem(pr))
		r.PUT("/problems/:id", handler.UpdateProblem(pr))
		r.DELETE("/problems/:id", handler.DeleteProblem(pr))
	}

	return r
}
