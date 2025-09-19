package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var version = "0.1.0"

func loadEnv() {
	_ = godotenv.Load()
}

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": version})
	})

	// Problem list stub
	r.GET("/problems", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"data": []string{}})
	})

	return r
}

func main() {
	loadEnv()
	port := os.Getenv("GO_BACKEND_PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Go backend listening on %s", addr)
	if err := setupRouter().Run(addr); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
