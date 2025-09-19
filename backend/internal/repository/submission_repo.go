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
    Count(ctx context.Context, filter SubmissionFilter) (int, error)
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
    // version 初始为 1
    if s.Version == 0 { s.Version = 1 }
    _, err := r.pool.Exec(ctx, `INSERT INTO submissions (id, user_id, problem_id, language, code, status, runtime_ms, memory_kb, error_message, version, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
        s.ID, s.UserID, s.ProblemID, s.Language, s.Code, s.Status, s.RuntimeMS, s.MemoryKB, s.ErrorMessage, s.Version, s.CreatedAt, s.UpdatedAt)
    return err
}

func (r *PGSubmissionRepository) GetByID(ctx context.Context, id string) (domain.Submission, error) {
    row := r.pool.QueryRow(ctx, `SELECT id, user_id, problem_id, language, code, status, runtime_ms, memory_kb, error_message, version, created_at, updated_at FROM submissions WHERE id=$1`, id)
    var s domain.Submission
    if err := row.Scan(&s.ID, &s.UserID, &s.ProblemID, &s.Language, &s.Code, &s.Status, &s.RuntimeMS, &s.MemoryKB, &s.ErrorMessage, &s.Version, &s.CreatedAt, &s.UpdatedAt); err != nil {
        if strings.Contains(err.Error(), "no rows") { return domain.Submission{}, ErrSubmissionNotFound }
        return domain.Submission{}, err
    }
    return s, nil
}

func (r *PGSubmissionRepository) UpdateStatus(ctx context.Context, id string, status string) error {
    // 读取当前状态以支持条件更新（简单两步；可优化为单条 SQL 返回旧值）
    cur, err := r.GetByID(ctx, id)
    if err != nil { return err }
    cmd, err := r.pool.Exec(ctx, `UPDATE submissions SET status=$1, version=version+1, updated_at=NOW() WHERE id=$2 AND status=$3`, status, id, cur.Status)
    if err != nil { return err }
    if cmd.RowsAffected() == 0 { return errors.New("concurrent status change") }
    return nil
}

func (r *PGSubmissionRepository) List(ctx context.Context, f SubmissionFilter, limit, offset int) ([]domain.Submission, error) {
    if limit <= 0 { limit = 20 }
    if offset < 0 { offset = 0 }
    clauses := []string{"1=1"}
    args := []any{}
    // build dynamic placeholders using current len(args)+1 to avoid manual idx management
    if f.UserID != "" { clauses = append(clauses, "user_id=$"+itoa(len(args)+1)); args = append(args, f.UserID) }
    if f.ProblemID != "" { clauses = append(clauses, "problem_id=$"+itoa(len(args)+1)); args = append(args, f.ProblemID) }
    if f.Status != "" { clauses = append(clauses, "status=$"+itoa(len(args)+1)); args = append(args, f.Status) }
    limitPos := len(args) + 1
    offsetPos := len(args) + 2
    q := `SELECT id, user_id, problem_id, language, code, status, runtime_ms, memory_kb, error_message, version, created_at, updated_at FROM submissions WHERE ` + strings.Join(clauses, " AND ") + ` ORDER BY created_at DESC LIMIT $` + itoa(limitPos) + ` OFFSET $` + itoa(offsetPos)
    args = append(args, limit, offset)
    rows, err := r.pool.Query(ctx, q, args...)
    if err != nil { return nil, err }
    defer rows.Close()
    res := make([]domain.Submission,0,limit)
    for rows.Next() {
        var s domain.Submission
        if err := rows.Scan(&s.ID,&s.UserID,&s.ProblemID,&s.Language,&s.Code,&s.Status,&s.RuntimeMS,&s.MemoryKB,&s.ErrorMessage,&s.Version,&s.CreatedAt,&s.UpdatedAt); err != nil { return nil, err }
        res = append(res, s)
    }
    return res, nil
}

func (r *PGSubmissionRepository) Count(ctx context.Context, f SubmissionFilter) (int, error) {
    clauses := []string{"1=1"}
    args := []any{}
    if f.UserID != "" { clauses = append(clauses, "user_id=$"+itoa(len(args)+1)); args = append(args, f.UserID) }
    if f.ProblemID != "" { clauses = append(clauses, "problem_id=$"+itoa(len(args)+1)); args = append(args, f.ProblemID) }
    if f.Status != "" { clauses = append(clauses, "status=$"+itoa(len(args)+1)); args = append(args, f.Status) }
    q := `SELECT COUNT(*) FROM submissions WHERE ` + strings.Join(clauses, " AND ")
    row := r.pool.QueryRow(ctx, q, args...)
    var total int
    if err := row.Scan(&total); err != nil { return 0, err }
    return total, nil
}

// 简易 int->string (避免 strconv 每次导入)
func itoa(i int) string { return strconv.FormatInt(int64(i),10) }
