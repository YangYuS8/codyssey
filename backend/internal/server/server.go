package server

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5"
	"github.com/your-org/codyssey/backend/internal/config"
	"github.com/your-org/codyssey/backend/internal/db"
	"github.com/your-org/codyssey/backend/internal/http/router"
	"github.com/your-org/codyssey/backend/internal/repository"
)

type Server struct {
	cfg    config.Config
	logger *zap.Logger
	http   *http.Server
	db     *db.Database
}

type healthProbe struct { s *Server }
func (h healthProbe) DBAlive() bool { return h.s != nil && h.s.db != nil }

func New(cfg config.Config) (*Server, error) {
	logger, err := zap.NewDevelopment()
	if err != nil { return nil, err }
	return &Server{cfg: cfg, logger: logger}, nil
}

func (s *Server) Start(ctx context.Context) error {
	// DB 连接（失败则直接返回）
	database, err := db.Connect(ctx, s.cfg.DB.ConnString())
	if err != nil {
		return err
	}
	s.db = database

	// 迁移 (problems 表) - 使用事务
	tx, err := s.db.Pool.BeginTx(ctx, pgx.TxOptions{})
	if err == nil {
		if migErr := repository.EnsureSchema(ctx, tx); migErr != nil {
			s.logger.Warn("migration failed", zap.Error(migErr))
			_ = tx.Rollback(ctx)
		} else {
			_ = tx.Commit(ctx)
		}
	} else {
		s.logger.Warn("begin tx for migration failed", zap.Error(err))
	}
	// Repo 实例
	problemRepo := repository.NewPGProblemRepository(database.Pool)

	deps := router.Dependencies{
        ProblemRepo: problemRepo,
        HealthCheck: healthProbe{s: s},
        Version:     s.cfg.Version,
        Env:         s.cfg.Env,
    }
	r := router.Setup(deps)

	s.http = &http.Server{Addr: ":" + s.cfg.Port, Handler: r}
	go func() {
		s.logger.Info("http server starting", zap.String("addr", s.http.Addr))
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal("http server error", zap.Error(err))
		}
	}()
	return nil
}

func (s *Server) WaitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	s.logger.Info("shutdown signal received")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if s.http != nil {
		_ = s.http.Shutdown(ctx)
	}
	if s.db != nil { s.db.Close() }
	_ = s.logger.Sync()
}
