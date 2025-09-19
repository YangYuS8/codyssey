package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
	"github.com/google/uuid"
)

var (
    ErrSubmissionNotFound = repository.ErrSubmissionNotFound
    ErrEmptyCode = errors.New("submission code empty")
    ErrLanguageRequired = errors.New("language required")
)

type SubmissionRepo interface {
    Create(ctx context.Context, s domain.Submission) error
    GetByID(ctx context.Context, id string) (domain.Submission, error)
    UpdateStatus(ctx context.Context, id string, status string) error
}

type SubmissionService struct { repo SubmissionRepo }

func NewSubmissionService(repo SubmissionRepo) *SubmissionService { return &SubmissionService{repo: repo} }

func (s *SubmissionService) Create(ctx context.Context, userID, problemID, language, code string) (domain.Submission, error) {
    if strings.TrimSpace(code) == "" { return domain.Submission{}, ErrEmptyCode }
    if strings.TrimSpace(language) == "" { return domain.Submission{}, ErrLanguageRequired }
    sub := domain.Submission{ID: uuid.New().String(), UserID: userID, ProblemID: problemID, Language: language, Code: code, Status: "pending", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
    if err := s.repo.Create(ctx, sub); err != nil { return domain.Submission{}, err }
    return sub, nil
}

func (s *SubmissionService) Get(ctx context.Context, id string) (domain.Submission, error) {
    return s.repo.GetByID(ctx, id)
}
