package main

import (
	"context"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/your-org/codyssey/backend/internal/config"
	"github.com/your-org/codyssey/backend/internal/server"
)

// buildVersion 通过 -ldflags "-X main.buildVersion=xxxx" 注入
var buildVersion = "dev"

func main() {
    _ = godotenv.Load()
    cfg := config.Load()
    cfg.Version = buildVersion

    svc, err := server.New(cfg)
    if err != nil {
        log.Fatalf("init server: %v", err)
    }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    if err := svc.Start(ctx); err != nil {
        log.Fatalf("start server: %v", err)
    }
    svc.WaitForShutdown()
}
