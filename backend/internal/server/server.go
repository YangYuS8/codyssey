package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/pressly/goose/v3"
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
	// 1. 连接数据库
	database, err := db.Connect(ctx, s.cfg.DB.ConnString())
	if err != nil { return err }
	s.db = database

	// 2. 执行 goose 迁移
	if err := s.runMigrations(); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}

	// 3. 初始化仓库 & 路由
	problemRepo := repository.NewPGProblemRepository(database.Pool)
	deps := router.Dependencies{
		ProblemRepo: problemRepo,
		HealthCheck: healthProbe{s: s},
		Version:     s.cfg.Version,
		Env:         s.cfg.Env,
	}
	r := router.Setup(deps)

	// 4. 启动 HTTP Server
	s.http = &http.Server{Addr: ":" + s.cfg.Port, Handler: r}
	go func() {
		s.logger.Info("http server starting", zap.String("addr", s.http.Addr))
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Fatal("http server error", zap.Error(err))
		}
	}()
	return nil
}

// runMigrations 使用 goose 执行 backend/migrations 下的所有 Up 迁移
func (s *Server) runMigrations() error {
	dir := filepath.Join("backend", "migrations")
	// 允许多次调用，goose 会记录版本
	goose.SetLogger(goose.NopLogger()) // 静默；我们用 zap 记录
	if err := goose.SetDialect("postgres"); err != nil { return err }
	// 使用现有连接获取 *sql.DB: goose 期望 database/sql，而我们是 pgxpool -> 暂时改为使用 connString 新开标准库连接
	// 简易方案：标准库打开（保持简单）；未来可换 pgx stdlib.
	dbstd, err := goose.OpenDBWithDriver("postgres", s.cfg.DB.ConnString())
	if err != nil { return err }
	defer dbstd.Close()
	if err := goose.Up(dbstd, dir); err != nil { return err }
	s.logger.Info("migrations applied", zap.String("dir", dir))
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
