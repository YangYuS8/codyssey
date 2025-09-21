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
	"go.uber.org/zap/zapcore"

	"github.com/YangYuS8/codyssey/backend/internal/auth"
	"github.com/YangYuS8/codyssey/backend/internal/config"
	"github.com/YangYuS8/codyssey/backend/internal/db"
	"github.com/YangYuS8/codyssey/backend/internal/http/router"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
	_ "github.com/jackc/pgx/v5/stdlib" // register pgx driver for database/sql
	"github.com/pressly/goose/v3"
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
	if err := cfg.Validate(); err != nil { return nil, err }
	logger, err := buildLogger(cfg)
	if err != nil { return nil, err }
	return &Server{cfg: cfg, logger: logger}, nil
}

func buildLogger(cfg config.Config) (*zap.Logger, error) {
	var level zapcore.Level
	switch cfg.LogLevel {
	case "debug": level = zap.DebugLevel
	case "info": level = zap.InfoLevel
	case "warn": level = zap.WarnLevel
	case "error": level = zap.ErrorLevel
	default: level = zap.InfoLevel
	}
	encCfg := zap.NewProductionEncoderConfig()
	encCfg.TimeKey = "ts"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encCfg), zapcore.AddSync(os.Stdout), level)
	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel)), nil
}

func (s *Server) Start(ctx context.Context) error {
	// 启动模式摘要日志（在连接 DB 之前，便于早期排错）
	debugIdentityEnabled := (s.cfg.Env == "development" || s.cfg.Env == "test")
	s.logger.Info("server bootstrap",
		zap.String("env", s.cfg.Env),
		zap.String("version", s.cfg.Version),
		zap.Bool("auto_migrate", s.cfg.AutoMigrate),
		zap.Bool("debug_identity_enabled", debugIdentityEnabled),
		zap.Int("max_submission_code_bytes", s.cfg.MaxSubmissionCodeBytes),
		zap.Int("max_request_body_bytes", s.cfg.MaxRequestBodyBytes),
	)
	// 1. 连接数据库
	database, err := db.Connect(ctx, s.cfg.DB.ConnString())
	if err != nil { return err }
	s.db = database

	// 2. 条件执行 goose 迁移
	if s.cfg.AutoMigrate {
		if err := s.runMigrations(); err != nil {
			return fmt.Errorf("run migrations: %w", err)
		}
	} else {
		s.logger.Info("auto migrate disabled; skip applying migrations", zap.Bool("auto_migrate", s.cfg.AutoMigrate))
	}

	// 3. 初始化仓库 & 路由
	problemRepo := repository.NewPGProblemRepository(database.Pool)
	userRepo := repository.NewPGUserRepository(database.Pool)
	submissionRepo := repository.NewPGSubmissionRepository(database.Pool)
	judgeRunRepo := repository.NewPGJudgeRunRepository(database.Pool)
	statusLogRepo := repository.NewPGSubmissionStatusLogRepository(database.Pool)
	jwtMgr := auth.NewJWTManager(os.Getenv("JWT_SECRET"), 15*time.Minute, 7*24*time.Hour)
	authService := auth.NewAuthService(userRepo, jwtMgr)
	deps := router.Dependencies{
		ProblemRepo:            problemRepo,
		UserRepo:               userRepo,
		AuthService:            authService,
		SubmissionRepo:         submissionRepo,
		SubmissionStatusLogRepo: statusLogRepo,
		JudgeRunRepo:           judgeRunRepo,
		HealthCheck:            healthProbe{s: s},
		Version:                s.cfg.Version,
		Env:                    s.cfg.Env,
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
	// 运行时当前工作目录是在 backend (Makefile: cd backend && go run .)
	dir := filepath.Join("migrations")
	s.logger.Info("running migrations", zap.String("driver", "pgx"), zap.String("dir", dir))
	// 允许多次调用，goose 会记录版本
	goose.SetLogger(goose.NopLogger()) // 静默；我们用 zap 记录
	if err := goose.SetDialect("postgres"); err != nil { return err }
	// 使用现有连接获取 *sql.DB: goose 期望 database/sql，而我们是 pgxpool -> 暂时改为使用 connString 新开标准库连接
	// 简易方案：标准库打开（保持简单）；未来可换 pgx stdlib.
	dbstd, err := goose.OpenDBWithDriver("pgx", s.cfg.DB.ConnString())
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
