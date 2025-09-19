package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthResponse struct {
	Status  string `json:"status"`
	DB      string `json:"db"`
	Version string `json:"version"`
	Env     string `json:"env"`
}

type HealthChecker interface {
	DBAlive() bool
}

func Health(version, env string, hc HealthChecker) gin.HandlerFunc {
    return func(c *gin.Context) {
        resp := HealthResponse{Status: "ok", Version: version, Env: env}
		if hc != nil && hc.DBAlive() {
			resp.DB = "up"
		} else {
			resp.DB = "down"
		}
        c.JSON(http.StatusOK, resp)
    }
}
