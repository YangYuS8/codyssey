package service

import (
	"context"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/YangYuS8/codyssey/backend/internal/repository"
	"github.com/google/uuid"
)

type UserRepo interface {
    Create(ctx context.Context, u domain.User) error
    GetByID(ctx context.Context, id string) (domain.User, error)
    GetByUsername(ctx context.Context, username string) (domain.User, error)
    UpdateRoles(ctx context.Context, id string, roles []string) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit, offset int) ([]domain.User, error)
}

type UserService struct { repo UserRepo }

func NewUserService(r UserRepo) *UserService { return &UserService{repo: r} }

func (s *UserService) Create(ctx context.Context, username string, roles []string) (domain.User, error) {
    u := domain.User{ID: uuid.New().String(), Username: username, Roles: roles, CreatedAt: time.Now().UTC()}
    if err := s.repo.Create(ctx, u); err != nil { return domain.User{}, err }
    return u, nil
}

func (s *UserService) Get(ctx context.Context, id string) (domain.User, error) {
    return s.repo.GetByID(ctx, id)
}

func (s *UserService) GetByUsername(ctx context.Context, username string) (domain.User, error) {
    return s.repo.GetByUsername(ctx, username)
}

func (s *UserService) UpdateRoles(ctx context.Context, id string, roles []string) error {
    return s.repo.UpdateRoles(ctx, id, roles)
}

func (s *UserService) Delete(ctx context.Context, id string) error {
    return s.repo.Delete(ctx, id)
}

func (s *UserService) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
    return s.repo.List(ctx, limit, offset)
}

// 错误透传：由 handler 统一转换为响应格式
var (
    ErrUserNotFound   = repository.ErrUserNotFound
    ErrUserDuplicate  = repository.ErrUserDuplicate
)
