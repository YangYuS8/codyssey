package repository

import (
	"context"
	"sync"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/google/uuid"
)

type MemorySubmissionRepository struct {
    mu   sync.RWMutex
    list []domain.Submission
}

func NewMemorySubmissionRepository() *MemorySubmissionRepository { return &MemorySubmissionRepository{list: make([]domain.Submission,0,16)} }

func (m *MemorySubmissionRepository) Create(ctx context.Context, s domain.Submission) error {
    m.mu.Lock(); defer m.mu.Unlock()
    if s.ID == "" { s.ID = uuid.New().String() }
    now := time.Now().UTC()
    if s.CreatedAt.IsZero() { s.CreatedAt = now }
    s.UpdatedAt = now
    // 初始化版本号（与 PG 仓库保持一致，初始为 1）
    if s.Version == 0 { s.Version = 1 }
    // append copy
    m.list = append(m.list, s)
    return nil
}

func (m *MemorySubmissionRepository) GetByID(ctx context.Context, id string) (domain.Submission, error) {
    m.mu.RLock(); defer m.mu.RUnlock()
    for _, s := range m.list { if s.ID == id { return s, nil } }
    return domain.Submission{}, ErrSubmissionNotFound
}

func (m *MemorySubmissionRepository) UpdateStatus(ctx context.Context, id string, status string) error {
    m.mu.Lock(); defer m.mu.Unlock()
    for i, s := range m.list {
        if s.ID == id {
            m.list[i].Status = status
            m.list[i].Version += 1 // 状态更新也递增版本
            m.list[i].UpdatedAt = time.Now().UTC()
            return nil
        }
    }
    return ErrSubmissionNotFound
}

func (m *MemorySubmissionRepository) List(ctx context.Context, f SubmissionFilter, limit, offset int) ([]domain.Submission, error) {
    m.mu.RLock(); defer m.mu.RUnlock()
    if limit <= 0 { limit = 20 }
    if offset < 0 { offset = 0 }
    filtered := make([]domain.Submission,0,len(m.list))
    for _, s := range m.list {
        if f.UserID != "" && s.UserID != f.UserID { continue }
        if f.ProblemID != "" && s.ProblemID != f.ProblemID { continue }
        if f.Status != "" && s.Status != f.Status { continue }
        filtered = append(filtered, s)
    }
    if offset >= len(filtered) { return []domain.Submission{}, nil }
    end := offset + limit; if end > len(filtered) { end = len(filtered) }
    res := make([]domain.Submission, end-offset)
    copy(res, filtered[offset:end])
    return res, nil
}

func (m *MemorySubmissionRepository) Count(ctx context.Context, f SubmissionFilter) (int, error) {
    m.mu.RLock(); defer m.mu.RUnlock()
    total := 0
    for _, s := range m.list {
        if f.UserID != "" && s.UserID != f.UserID { continue }
        if f.ProblemID != "" && s.ProblemID != f.ProblemID { continue }
        if f.Status != "" && s.Status != f.Status { continue }
        total++
    }
    return total, nil
}
