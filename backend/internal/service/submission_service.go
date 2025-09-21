package service

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/metrics"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
	"github.com/google/uuid"
)

var (
    ErrSubmissionNotFound      = repository.ErrSubmissionNotFound
    ErrSubmissionConflict      = repository.ErrSubmissionConflict
    ErrEmptyCode               = errors.New("submission code empty")
    ErrLanguageRequired        = errors.New("language required")
    ErrInvalidStatus           = errors.New("invalid submission status")
    ErrInvalidStatusTransition = errors.New("invalid status transition")
)

// 判题状态常量
const (
    SubmissionStatusPending  = "pending"
    SubmissionStatusJudging  = "judging"
    SubmissionStatusAccepted = "accepted"
    SubmissionStatusWrongAns = "wrong_answer"
    SubmissionStatusError    = "error"
)

var allowedNext = map[string][]string{
    SubmissionStatusPending:  {SubmissionStatusJudging, SubmissionStatusAccepted, SubmissionStatusWrongAns, SubmissionStatusError},
    SubmissionStatusJudging:  {SubmissionStatusAccepted, SubmissionStatusWrongAns, SubmissionStatusError},
    SubmissionStatusAccepted: {},
    SubmissionStatusWrongAns: {},
    SubmissionStatusError:    {},
}

func isValidStatus(st string) bool {
    switch st {
    case SubmissionStatusPending, SubmissionStatusJudging, SubmissionStatusAccepted, SubmissionStatusWrongAns, SubmissionStatusError:
        return true
    default:
        return false
    }
}

type SubmissionRepo interface {
    Create(ctx context.Context, s domain.Submission) error
    GetByID(ctx context.Context, id string) (domain.Submission, error)
    UpdateStatus(ctx context.Context, id string, status string, expectedVersion int) error
    List(ctx context.Context, f repository.SubmissionFilter, limit, offset int) ([]domain.Submission, error)
    Count(ctx context.Context, f repository.SubmissionFilter) (int, error)
}

type SubmissionStatusLogRepo interface {
    Add(ctx context.Context, log domain.SubmissionStatusLog) error
    ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]domain.SubmissionStatusLog, error)
}

type SubmissionService struct {
    repo    SubmissionRepo
    logRepo SubmissionStatusLogRepo
}

func NewSubmissionService(repo SubmissionRepo, logRepo SubmissionStatusLogRepo) *SubmissionService { return &SubmissionService{repo: repo, logRepo: logRepo} }

func (s *SubmissionService) Create(ctx context.Context, userID, problemID, language, code string) (domain.Submission, error) {
    if strings.TrimSpace(code) == "" { return domain.Submission{}, ErrEmptyCode }
    if strings.TrimSpace(language) == "" { return domain.Submission{}, ErrLanguageRequired }
    // 代码长度限制（优先使用环境变量 MAX_SUBMISSION_CODE_BYTES；否则默认 128KB）
    maxBytes := 128 * 1024
    if v := os.Getenv("MAX_SUBMISSION_CODE_BYTES"); v != "" { if n, err := strconv.Atoi(v); err == nil && n > 0 { maxBytes = n } }
    if len(code) > maxBytes { return domain.Submission{}, errors.New("code too large") }
    sub := domain.Submission{ID: uuid.New().String(), UserID: userID, ProblemID: problemID, Language: language, Code: code, Status: SubmissionStatusPending, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(), Version: 1}
    if err := s.repo.Create(ctx, sub); err != nil { return domain.Submission{}, err }
    return sub, nil
}

func (s *SubmissionService) Get(ctx context.Context, id string) (domain.Submission, error) { return s.repo.GetByID(ctx, id) }

// UpdateStatus 带状态机校验 + 生成日志
func (s *SubmissionService) UpdateStatus(ctx context.Context, id string, newStatus string) (domain.Submission, error) {
    newStatus = strings.TrimSpace(newStatus)
    if !isValidStatus(newStatus) { return domain.Submission{}, ErrInvalidStatus }
    cur, err := s.repo.GetByID(ctx, id)
    if err != nil { return domain.Submission{}, err }
    fromStatus := cur.Status
    allowed := allowedNext[fromStatus]
    ok := false
    for _, a := range allowed { if a == newStatus { ok = true; break } }
    if !ok { return domain.Submission{}, ErrInvalidStatusTransition }
    if err := s.repo.UpdateStatus(ctx, id, newStatus, cur.Version); err != nil {
        if errors.Is(err, ErrSubmissionConflict) { metrics.IncSubmissionConflict() }
        return domain.Submission{}, err
    }
    metrics.ObserveSubmissionTransition(fromStatus, newStatus)
    cur.Status = newStatus
    cur.Version += 1
    cur.UpdatedAt = time.Now().UTC()
    if s.logRepo != nil { _ = s.logRepo.Add(ctx, domain.SubmissionStatusLog{SubmissionID: cur.ID, FromStatus: fromStatus, ToStatus: newStatus}) }
    return cur, nil
}

type SubmissionListFilter struct {
    UserID    string
    ProblemID string
    Status    string
    Limit     int
    Offset    int
}

func (s *SubmissionService) List(ctx context.Context, f SubmissionListFilter) ([]domain.Submission, error) {
    limit := f.Limit; offset := f.Offset
    if limit <= 0 { limit = 20 }
    return s.repo.List(ctx, repository.SubmissionFilter{UserID: f.UserID, ProblemID: f.ProblemID, Status: f.Status}, limit, offset)
}

// ListWithTotal 返回列表与符合过滤条件的总数（不受分页影响）。
func (s *SubmissionService) ListWithTotal(ctx context.Context, f SubmissionListFilter) ([]domain.Submission, int, error) {
    limit := f.Limit; offset := f.Offset
    if limit <= 0 { limit = 20 }
    filter := repository.SubmissionFilter{UserID: f.UserID, ProblemID: f.ProblemID, Status: f.Status}
    total, err := s.repo.Count(ctx, filter)
    if err != nil { return nil, 0, err }
    items, err := s.repo.List(ctx, filter, limit, offset)
    if err != nil { return nil, 0, err }
    return items, total, nil
}

func (s *SubmissionService) ListStatusLogs(ctx context.Context, submissionID string, limit, offset int) ([]domain.SubmissionStatusLog, error) {
    if s.logRepo == nil { return []domain.SubmissionStatusLog{}, nil }
    return s.logRepo.ListBySubmission(ctx, submissionID, limit, offset)
}
