package repository

import (
	"context"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SubmissionStatusLogRepository 日志仓库接口
type SubmissionStatusLogRepository interface {
    Add(ctx context.Context, log domain.SubmissionStatusLog) error
    ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]domain.SubmissionStatusLog, error)
}

// PG 实现
type PGSubmissionStatusLogRepository struct { pool *pgxpool.Pool }

func NewPGSubmissionStatusLogRepository(pool *pgxpool.Pool) *PGSubmissionStatusLogRepository { return &PGSubmissionStatusLogRepository{pool: pool} }

func (r *PGSubmissionStatusLogRepository) Add(ctx context.Context, l domain.SubmissionStatusLog) error {
    if l.ID == "" { l.ID = uuid.New().String() }
    if l.CreatedAt.IsZero() { l.CreatedAt = time.Now().UTC() }
    _, err := r.pool.Exec(ctx, `INSERT INTO submission_status_logs (id, submission_id, from_status, to_status, created_at) VALUES ($1,$2,$3,$4,$5)`,
        l.ID, l.SubmissionID, l.FromStatus, l.ToStatus, l.CreatedAt)
    return err
}

func (r *PGSubmissionStatusLogRepository) ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]domain.SubmissionStatusLog, error) {
    if limit <= 0 { limit = 20 }
    if offset < 0 { offset = 0 }
    rows, err := r.pool.Query(ctx, `SELECT id, submission_id, from_status, to_status, created_at FROM submission_status_logs WHERE submission_id=$1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`, submissionID, limit, offset)
    if err != nil { return nil, err }
    defer rows.Close()
    res := make([]domain.SubmissionStatusLog,0,limit)
    for rows.Next() {
        var l domain.SubmissionStatusLog
        if err := rows.Scan(&l.ID,&l.SubmissionID,&l.FromStatus,&l.ToStatus,&l.CreatedAt); err != nil { return nil, err }
        res = append(res, l)
    }
    return res, nil
}

// 内存实现（测试用）
type MemorySubmissionStatusLogRepository struct {
    list []domain.SubmissionStatusLog
}

func NewMemorySubmissionStatusLogRepository() *MemorySubmissionStatusLogRepository { return &MemorySubmissionStatusLogRepository{list: make([]domain.SubmissionStatusLog,0,16)} }

func (m *MemorySubmissionStatusLogRepository) Add(ctx context.Context, l domain.SubmissionStatusLog) error {
    if l.ID == "" { l.ID = uuid.New().String() }
    if l.CreatedAt.IsZero() { l.CreatedAt = time.Now().UTC() }
    m.list = append(m.list, l)
    return nil
}

func (m *MemorySubmissionStatusLogRepository) ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]domain.SubmissionStatusLog, error) {
    if limit <= 0 { limit = 20 }
    if offset < 0 { offset = 0 }
    filtered := make([]domain.SubmissionStatusLog,0)
    for _, l := range m.list { if l.SubmissionID == submissionID { filtered = append(filtered, l) } }
    if offset >= len(filtered) { return []domain.SubmissionStatusLog{}, nil }
    end := offset + limit; if end > len(filtered) { end = len(filtered) }
    out := make([]domain.SubmissionStatusLog, end-offset)
    copy(out, filtered[offset:end])
    return out, nil
}
