package repository

import (
	"context"
	"sync"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/google/uuid"
)

type MemoryUserRepository struct {
    mu   sync.RWMutex
    list []domain.User
}

func NewMemoryUserRepository() *MemoryUserRepository { return &MemoryUserRepository{list: make([]domain.User,0,16)} }

func (m *MemoryUserRepository) Create(ctx context.Context, u domain.User) error {
    m.mu.Lock(); defer m.mu.Unlock()
    for _, existing := range m.list { if existing.Username == u.Username { return ErrUserDuplicate } }
    if u.ID == "" { u.ID = uuid.New().String() }
    if u.CreatedAt.IsZero() { u.CreatedAt = time.Now().UTC() }
    m.list = append(m.list, u)
    return nil
}

func (m *MemoryUserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
    m.mu.RLock(); defer m.mu.RUnlock()
    for _, u := range m.list { if u.ID == id { return u, nil } }
    return domain.User{}, ErrUserNotFound
}

func (m *MemoryUserRepository) GetByUsername(ctx context.Context, username string) (domain.User, error) {
    m.mu.RLock(); defer m.mu.RUnlock()
    for _, u := range m.list { if u.Username == username { return u, nil } }
    return domain.User{}, ErrUserNotFound
}

func (m *MemoryUserRepository) UpdateRoles(ctx context.Context, id string, roles []string) error {
    m.mu.Lock(); defer m.mu.Unlock()
    for i, u := range m.list { if u.ID == id { m.list[i].Roles = roles; return nil } }
    return ErrUserNotFound
}

func (m *MemoryUserRepository) Delete(ctx context.Context, id string) error {
    m.mu.Lock(); defer m.mu.Unlock()
    for i, u := range m.list { if u.ID == id { m.list = append(m.list[:i], m.list[i+1:]...); return nil } }
    return ErrUserNotFound
}

func (m *MemoryUserRepository) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
    m.mu.RLock(); defer m.mu.RUnlock()
    if limit <= 0 { limit = 20 }
    if offset >= len(m.list) { return []domain.User{}, nil }
    end := offset + limit; if end > len(m.list) { end = len(m.list) }
    res := make([]domain.User, end-offset)
    copy(res, m.list[offset:end])
    return res, nil
}
