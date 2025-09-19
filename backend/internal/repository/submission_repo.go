package repository

import (
	"context"
	"errors"
	"strconv"
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
    List(ctx context.Context, filter SubmissionFilter, limit, offset int) ([]domain.Submission, error)
}

// SubmissionFilter 用于列表过滤
type SubmissionFilter struct {
    UserID    string
    ProblemID string
    Status    string
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

func (r *PGSubmissionRepository) List(ctx context.Context, f SubmissionFilter, limit, offset int) ([]domain.Submission, error) {
    if limit <= 0 { limit = 20 }
    if offset < 0 { offset = 0 }
    clauses := []string{"1=1"}
    args := []any{}
    idx := 1
    if f.UserID != "" { clauses = append(clauses, "user_id=$"+itoa(idx)); args = append(args, f.UserID); idx++ }
    if f.ProblemID != "" { clauses = append(clauses, "problem_id=$"+itoa(idx)); args = append(args, f.ProblemID); idx++ }
    if f.Status != "" { clauses = append(clauses, "status=$"+itoa(idx)); args = append(args, f.Status); idx++ }
    // pagination
    args = append(args, limit, offset)
    q := `SELECT id, user_id, problem_id, language, code, status, created_at, updated_at FROM submissions WHERE ` + strings.Join(clauses, " AND ") + ` ORDER BY created_at DESC LIMIT $` + itoa(idx) + ` OFFSET $` + itoa(idx+1)
    rows, err := r.pool.Query(ctx, q, args...)
    if err != nil { return nil, err }
    defer rows.Close()
    res := make([]domain.Submission,0,limit)
    for rows.Next() {
        var s domain.Submission
        if err := rows.Scan(&s.ID,&s.UserID,&s.ProblemID,&s.Language,&s.Code,&s.Status,&s.CreatedAt,&s.UpdatedAt); err != nil { return nil, err }
        res = append(res, s)
    }
    return res, nil
}

// 简易 int->string (避免 strconv 每次导入)
func itoa(i int) string { return strconv.FormatInt(int64(i),10) }
