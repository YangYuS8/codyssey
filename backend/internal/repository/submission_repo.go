package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSubmissionNotFound = errors.New("submission not found")

type SubmissionRepository interface {
    Create(ctx context.Context, s domain.Submission) error
    GetByID(ctx context.Context, id string) (domain.Submission, error)
    UpdateStatus(ctx context.Context, id string, status string) error
}

type PGSubmissionRepository struct { pool *pgxpool.Pool }

func NewPGSubmissionRepository(pool *pgxpool.Pool) *PGSubmissionRepository { return &PGSubmissionRepository{pool: pool} }

func (r *PGSubmissionRepository) Create(ctx context.Context, s domain.Submission) error {
    if s.ID == "" { s.ID = uuid.New().String() }
    now := time.Now().UTC()
    if s.CreatedAt.IsZero() { s.CreatedAt = now }
    s.UpdatedAt = now
    _, err := r.pool.Exec(ctx, `INSERT INTO submissions (id, user_id, problem_id, language, code, status, created_at, updated_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
        s.ID, s.UserID, s.ProblemID, s.Language, s.Code, s.Status, s.CreatedAt, s.UpdatedAt)
    return err
}

func (r *PGSubmissionRepository) GetByID(ctx context.Context, id string) (domain.Submission, error) {
    row := r.pool.QueryRow(ctx, `SELECT id, user_id, problem_id, language, code, status, created_at, updated_at FROM submissions WHERE id=$1`, id)
    var s domain.Submission
    if err := row.Scan(&s.ID, &s.UserID, &s.ProblemID, &s.Language, &s.Code, &s.Status, &s.CreatedAt, &s.UpdatedAt); err != nil {
        if strings.Contains(err.Error(), "no rows") { return domain.Submission{}, ErrSubmissionNotFound }
        return domain.Submission{}, err
    }
    return s, nil
}

func (r *PGSubmissionRepository) UpdateStatus(ctx context.Context, id string, status string) error {
    cmd, err := r.pool.Exec(ctx, `UPDATE submissions SET status=$1, updated_at=NOW() WHERE id=$2`, status, id)
    if err != nil { return err }
    if cmd.RowsAffected() == 0 { return ErrSubmissionNotFound }
    return nil
}
