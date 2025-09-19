package service

import (
	"context"
	"errors"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
	"github.com/google/uuid"
)

var (
    ErrJudgeRunNotFound      = repository.ErrJudgeRunNotFound
    ErrJudgeRunInvalidStatus = errors.New("invalid judge run status transition")
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
    return jr, nil
}

// Start 将 queued 置为 running
func (s *JudgeRunService) Start(ctx context.Context, id string) (domain.JudgeRun, error) {
    if err := s.repo.UpdateRunning(ctx, id); err != nil { return domain.JudgeRun{}, err }
    return s.repo.GetByID(ctx, id)
}

// Finish 将 running 置为终态（succeeded/failed/canceled），并写入指标
func (s *JudgeRunService) Finish(ctx context.Context, id string, status string, runtimeMS, memoryKB, exitCode int, errMsg string) (domain.JudgeRun, error) {
    switch status {
    case domain.JudgeRunStatusSucceeded, domain.JudgeRunStatusFailed, domain.JudgeRunStatusCanceled:
    default:
        return domain.JudgeRun{}, ErrJudgeRunInvalidStatus
    }
    if err := s.repo.UpdateFinished(ctx, id, status, runtimeMS, memoryKB, exitCode, errMsg); err != nil { return domain.JudgeRun{}, err }
    return s.repo.GetByID(ctx, id)
}

func (s *JudgeRunService) Get(ctx context.Context, id string) (domain.JudgeRun, error) { return s.repo.GetByID(ctx, id) }

func (s *JudgeRunService) ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]domain.JudgeRun, error) {
    return s.repo.ListBySubmission(ctx, submissionID, limit, offset)
}
