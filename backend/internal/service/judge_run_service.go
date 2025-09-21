package service

import (
	"context"
	"errors"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/metrics"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
	"github.com/google/uuid"
)

var (
    ErrJudgeRunNotFound      = repository.ErrJudgeRunNotFound
    ErrJudgeRunInvalidStatus = errors.New("invalid judge run status transition")
    ErrJudgeRunConflict      = repository.ErrJudgeRunConflict
)

// JudgeRunRepo 接口（与 repository.JudgeRunRepository 对齐方便测试替换）
type JudgeRunRepo interface {
    Create(ctx context.Context, jr domain.JudgeRun) error
    GetByID(ctx context.Context, id string) (domain.JudgeRun, error)
    ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]domain.JudgeRun, error)
    UpdateRunning(ctx context.Context, id string) error
    UpdateFinished(ctx context.Context, id string, status string, runtimeMS, memoryKB, exitCode int, errMsg string) error
}

type JudgeRunService struct { repo JudgeRunRepo }

func NewJudgeRunService(r JudgeRunRepo) *JudgeRunService { return &JudgeRunService{repo: r} }

// Enqueue 创建一个排队的 JudgeRun
func (s *JudgeRunService) Enqueue(ctx context.Context, submissionID string, judgeVersion string) (domain.JudgeRun, error) {
    jr := domain.JudgeRun{ID: uuid.New().String(), SubmissionID: submissionID, Status: domain.JudgeRunStatusQueued, JudgeVersion: judgeVersion, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
    if err := s.repo.Create(ctx, jr); err != nil { return domain.JudgeRun{}, err }
    metrics.ObserveJudgeRunTransition("", domain.JudgeRunStatusQueued)
    return jr, nil
}

// Start 将 queued 置为 running
func (s *JudgeRunService) Start(ctx context.Context, id string) (domain.JudgeRun, error) {
    if err := s.repo.UpdateRunning(ctx, id); err != nil {
        if errors.Is(err, repository.ErrJudgeRunConflict) { metrics.IncJudgeRunConflict() }
        return domain.JudgeRun{}, err
    }
    jr, err := s.repo.GetByID(ctx, id)
    if err == nil { metrics.ObserveJudgeRunTransition(domain.JudgeRunStatusQueued, domain.JudgeRunStatusRunning) }
    return jr, err
}

// Finish 将 running 置为终态（succeeded/failed/canceled），并写入指标
func (s *JudgeRunService) Finish(ctx context.Context, id string, status string, runtimeMS, memoryKB, exitCode int, errMsg string) (domain.JudgeRun, error) {
    switch status {
    case domain.JudgeRunStatusSucceeded, domain.JudgeRunStatusFailed, domain.JudgeRunStatusCanceled:
    default:
        return domain.JudgeRun{}, ErrJudgeRunInvalidStatus
    }
    if err := s.repo.UpdateFinished(ctx, id, status, runtimeMS, memoryKB, exitCode, errMsg); err != nil {
        if errors.Is(err, repository.ErrJudgeRunConflict) { metrics.IncJudgeRunConflict() }
        return domain.JudgeRun{}, err
    }
    jr, err := s.repo.GetByID(ctx, id)
    if err == nil {
        metrics.ObserveJudgeRunTransition(domain.JudgeRunStatusRunning, status)
        metrics.ObserveJudgeRunDuration(status, jr.StartedAt, jr.FinishedAt)
    }
    return jr, err
}

func (s *JudgeRunService) Get(ctx context.Context, id string) (domain.JudgeRun, error) { return s.repo.GetByID(ctx, id) }

func (s *JudgeRunService) ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]domain.JudgeRun, error) {
    return s.repo.ListBySubmission(ctx, submissionID, limit, offset)
}

// --- DTO & Adapter for HTTP layer ---
// 定义一个 DTO，避免 handler 直接依赖 domain 结构（未来可做字段裁剪或扩展）
type JudgeRunDTO struct {
    ID           string
    SubmissionID string
    Status       string
    JudgeVersion string
    RuntimeMS    int
    MemoryKB     int
    ExitCode     int
    ErrorMessage string
    CreatedAt    time.Time
    UpdatedAt    time.Time
    StartedAt    *time.Time
    FinishedAt   *time.Time
}

func toDTO(d domain.JudgeRun) JudgeRunDTO {
    return JudgeRunDTO{
        ID: d.ID, SubmissionID: d.SubmissionID, Status: d.Status, JudgeVersion: d.JudgeVersion,
        RuntimeMS: d.RuntimeMS, MemoryKB: d.MemoryKB, ExitCode: d.ExitCode, ErrorMessage: d.ErrorMessage,
        CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt, StartedAt: d.StartedAt, FinishedAt: d.FinishedAt,
    }
}

// HTTPAdapter 提供面向 handler 的包装（返回 DTO）
type JudgeRunHTTPAdapter struct { svc *JudgeRunService }

func NewJudgeRunHTTPAdapter(svc *JudgeRunService) *JudgeRunHTTPAdapter { return &JudgeRunHTTPAdapter{svc: svc} }

func (a *JudgeRunHTTPAdapter) Enqueue(ctx context.Context, submissionID, judgeVersion string) (JudgeRunDTO, error) {
    jr, err := a.svc.Enqueue(ctx, submissionID, judgeVersion)
    if err != nil { return JudgeRunDTO{}, err }
    return toDTO(jr), nil
}

func (a *JudgeRunHTTPAdapter) ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]JudgeRunDTO, error) {
    list, err := a.svc.ListBySubmission(ctx, submissionID, limit, offset)
    if err != nil { return nil, err }
    out := make([]JudgeRunDTO, 0, len(list))
    for _, it := range list { out = append(out, toDTO(it)) }
    return out, nil
}

func (a *JudgeRunHTTPAdapter) Get(ctx context.Context, id string) (JudgeRunDTO, error) {
    jr, err := a.svc.Get(ctx, id)
    if err != nil { return JudgeRunDTO{}, err }
    return toDTO(jr), nil
}

// Expose underlying service for internal handler usage (start/finish)
func (a *JudgeRunHTTPAdapter) Service() *JudgeRunService { return a.svc }
