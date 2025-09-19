package service

import (
	"context"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
	"github.com/google/uuid"
)

// ProblemRepo 抽象（接口收敛，方便后续装饰）
type ProblemRepo interface {
    Create(ctx context.Context, p domain.Problem) error
    GetByID(ctx context.Context, id uuid.UUID) (domain.Problem, error)
    Update(ctx context.Context, p domain.Problem) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, limit, offset int) ([]domain.Problem, error)
}

type ProblemService struct { repo ProblemRepo }

func NewProblemService(r ProblemRepo) *ProblemService { return &ProblemService{repo: r} }

func (s *ProblemService) Create(ctx context.Context, title, desc string) (domain.Problem, error) {
    p := domain.NewProblem(title, desc)
    if err := s.repo.Create(ctx, p); err != nil { return domain.Problem{}, err }
    return p, nil
}

func (s *ProblemService) Get(ctx context.Context, id uuid.UUID) (domain.Problem, error) {
    return s.repo.GetByID(ctx, id)
}

func (s *ProblemService) Update(ctx context.Context, id uuid.UUID, title *string, desc *string) (domain.Problem, error) {
    existing, err := s.repo.GetByID(ctx, id)
    if err != nil { return domain.Problem{}, err }
    if title != nil { existing.Title = *title }
    if desc != nil { existing.Description = *desc }
    if err := s.repo.Update(ctx, existing); err != nil { return domain.Problem{}, err }
    return existing, nil
}

func (s *ProblemService) Delete(ctx context.Context, id uuid.UUID) error {
    return s.repo.Delete(ctx, id)
}

func (s *ProblemService) List(ctx context.Context, limit, offset int) ([]domain.Problem, error) {
    if limit <= 0 { limit = 20 }
    if offset < 0 { offset = 0 }
    return s.repo.List(ctx, limit, offset)
}

// 错误透传，这里预留做 error wrapping / metrics
var ErrNotFound = repository.ErrNotFound
// end
