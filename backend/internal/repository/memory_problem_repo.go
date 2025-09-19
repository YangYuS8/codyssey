package repository

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/your-org/codyssey/backend/internal/domain"
)

type MemoryProblemRepository struct {
	mu   sync.RWMutex
	list []domain.Problem
}

func NewMemoryProblemRepository() *MemoryProblemRepository {
	return &MemoryProblemRepository{list: make([]domain.Problem, 0, 16)}
}

func (m *MemoryProblemRepository) Create(ctx context.Context, p domain.Problem) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.list = append([]domain.Problem{p}, m.list...)
	return nil
}

func (m *MemoryProblemRepository) List(ctx context.Context, limit, offset int) ([]domain.Problem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if offset >= len(m.list) {
		return []domain.Problem{}, nil
	}
	end := offset + limit
	if limit <= 0 { limit = 20; end = offset + limit }
	if end > len(m.list) { end = len(m.list) }
	res := make([]domain.Problem, end-offset)
	copy(res, m.list[offset:end])
	return res, nil
}

func (m *MemoryProblemRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Problem, error) {
	m.mu.RLock(); defer m.mu.RUnlock()
	for _, p := range m.list { if p.ID == id { return p, nil } }
	return domain.Problem{}, ErrNotFound
}

func (m *MemoryProblemRepository) Update(ctx context.Context, p domain.Problem) error {
	m.mu.Lock(); defer m.mu.Unlock()
	for i, item := range m.list { if item.ID == p.ID { m.list[i] = p; return nil } }
	return ErrNotFound
}

func (m *MemoryProblemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock(); defer m.mu.Unlock()
	for i, item := range m.list { if item.ID == id { m.list = append(m.list[:i], m.list[i+1:]...); return nil } }
	return ErrNotFound
}
