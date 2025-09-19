package repository

import (
	"context"
	"errors"
	"time"

	"github.com/YangYuS8/codyssey/backend/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrJudgeRunNotFound = errors.New("judge run not found")

// JudgeRunRepository 定义判题执行记录的持久化接口
// 状态转换：queued -> running -> (succeeded|failed|canceled)
// 不允许从终态回到非终态
// UpdateRunning: queued -> running（设置 started_at）
// UpdateFinished: running -> 终态（设置 finished_at、runtime/memory/exit_code/error_message）
type JudgeRunRepository interface {
    Create(ctx context.Context, jr domain.JudgeRun) error
    GetByID(ctx context.Context, id string) (domain.JudgeRun, error)
    ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]domain.JudgeRun, error)
    UpdateRunning(ctx context.Context, id string) error
    UpdateFinished(ctx context.Context, id string, status string, runtimeMS, memoryKB, exitCode int, errMsg string) error
}

// PG 实现

type PGJudgeRunRepository struct { pool *pgxpool.Pool }

func NewPGJudgeRunRepository(pool *pgxpool.Pool) *PGJudgeRunRepository { return &PGJudgeRunRepository{pool: pool} }

func (r *PGJudgeRunRepository) Create(ctx context.Context, jr domain.JudgeRun) error {
    if jr.ID == "" { jr.ID = uuid.New().String() }
    now := time.Now().UTC()
    if jr.CreatedAt.IsZero() { jr.CreatedAt = now }
    jr.UpdatedAt = now
    _, err := r.pool.Exec(ctx, `INSERT INTO judge_runs (id, submission_id, status, judge_version, runtime_ms, memory_kb, exit_code, error_message, started_at, finished_at, created_at, updated_at)
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
        jr.ID, jr.SubmissionID, jr.Status, jr.JudgeVersion, jr.RuntimeMS, jr.MemoryKB, jr.ExitCode, jr.ErrorMessage, jr.StartedAt, jr.FinishedAt, jr.CreatedAt, jr.UpdatedAt)
    return err
}

func (r *PGJudgeRunRepository) GetByID(ctx context.Context, id string) (domain.JudgeRun, error) {
    row := r.pool.QueryRow(ctx, `SELECT id, submission_id, status, judge_version, runtime_ms, memory_kb, exit_code, error_message, started_at, finished_at, created_at, updated_at FROM judge_runs WHERE id=$1`, id)
    var jr domain.JudgeRun
    if err := row.Scan(&jr.ID,&jr.SubmissionID,&jr.Status,&jr.JudgeVersion,&jr.RuntimeMS,&jr.MemoryKB,&jr.ExitCode,&jr.ErrorMessage,&jr.StartedAt,&jr.FinishedAt,&jr.CreatedAt,&jr.UpdatedAt); err != nil {
        if err.Error() == "no rows in result set" { return domain.JudgeRun{}, ErrJudgeRunNotFound }
        return domain.JudgeRun{}, err
    }
    return jr, nil
}

func (r *PGJudgeRunRepository) ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]domain.JudgeRun, error) {
    if limit <= 0 { limit = 20 }
    if offset < 0 { offset = 0 }
    rows, err := r.pool.Query(ctx, `SELECT id, submission_id, status, judge_version, runtime_ms, memory_kb, exit_code, error_message, started_at, finished_at, created_at, updated_at FROM judge_runs WHERE submission_id=$1 ORDER BY created_at ASC LIMIT $2 OFFSET $3`, submissionID, limit, offset)
    if err != nil { return nil, err }
    defer rows.Close()
    res := make([]domain.JudgeRun,0,limit)
    for rows.Next() {
        var jr domain.JudgeRun
        if err := rows.Scan(&jr.ID,&jr.SubmissionID,&jr.Status,&jr.JudgeVersion,&jr.RuntimeMS,&jr.MemoryKB,&jr.ExitCode,&jr.ErrorMessage,&jr.StartedAt,&jr.FinishedAt,&jr.CreatedAt,&jr.UpdatedAt); err != nil { return nil, err }
        res = append(res, jr)
    }
    return res, nil
}

func (r *PGJudgeRunRepository) UpdateRunning(ctx context.Context, id string) error {
    // 仅允许 queued -> running
    cmd, err := r.pool.Exec(ctx, `UPDATE judge_runs SET status='running', started_at=NOW(), updated_at=NOW() WHERE id=$1 AND status='queued'`, id)
    if err != nil { return err }
    if cmd.RowsAffected() == 0 { return ErrJudgeRunNotFound }
    return nil
}

func (r *PGJudgeRunRepository) UpdateFinished(ctx context.Context, id string, status string, runtimeMS, memoryKB, exitCode int, errMsg string) error {
    // 仅允许 running -> 终态
    switch status {
    case domain.JudgeRunStatusSucceeded, domain.JudgeRunStatusFailed, domain.JudgeRunStatusCanceled:
    default:
        return errors.New("invalid terminal status")
    }
    cmd, err := r.pool.Exec(ctx, `UPDATE judge_runs SET status=$1, runtime_ms=$2, memory_kb=$3, exit_code=$4, error_message=$5, finished_at=NOW(), updated_at=NOW() WHERE id=$6 AND status='running'`, status, runtimeMS, memoryKB, exitCode, errMsg, id)
    if err != nil { return err }
    if cmd.RowsAffected() == 0 { return ErrJudgeRunNotFound }
    return nil
}

// 内存实现（测试）

type MemoryJudgeRunRepository struct {
    list []domain.JudgeRun
}

func NewMemoryJudgeRunRepository() *MemoryJudgeRunRepository { return &MemoryJudgeRunRepository{list: make([]domain.JudgeRun,0,16)} }

func (m *MemoryJudgeRunRepository) Create(ctx context.Context, jr domain.JudgeRun) error {
    if jr.ID == "" { jr.ID = uuid.New().String() }
    now := time.Now().UTC()
    if jr.CreatedAt.IsZero() { jr.CreatedAt = now }
    jr.UpdatedAt = now
    m.list = append(m.list, jr)
    return nil
}

func (m *MemoryJudgeRunRepository) GetByID(ctx context.Context, id string) (domain.JudgeRun, error) {
    for _, jr := range m.list { if jr.ID == id { return jr, nil } }
    return domain.JudgeRun{}, ErrJudgeRunNotFound
}

func (m *MemoryJudgeRunRepository) ListBySubmission(ctx context.Context, submissionID string, limit, offset int) ([]domain.JudgeRun, error) {
    if limit <= 0 { limit = 20 }
    if offset < 0 { offset = 0 }
    filtered := make([]domain.JudgeRun,0)
    for _, jr := range m.list { if jr.SubmissionID == submissionID { filtered = append(filtered, jr) } }
    if offset >= len(filtered) { return []domain.JudgeRun{}, nil }
    end := offset + limit; if end > len(filtered) { end = len(filtered) }
    out := make([]domain.JudgeRun, end-offset)
    copy(out, filtered[offset:end])
    return out, nil
}

func (m *MemoryJudgeRunRepository) UpdateRunning(ctx context.Context, id string) error {
    for i, jr := range m.list {
        if jr.ID == id && jr.Status == domain.JudgeRunStatusQueued {
            now := time.Now().UTC(); m.list[i].Status = domain.JudgeRunStatusRunning; m.list[i].StartedAt = &now; m.list[i].UpdatedAt = now; return nil
        }
    }
    return ErrJudgeRunNotFound
}

func (m *MemoryJudgeRunRepository) UpdateFinished(ctx context.Context, id string, status string, runtimeMS, memoryKB, exitCode int, errMsg string) error {
    switch status {
    case domain.JudgeRunStatusSucceeded, domain.JudgeRunStatusFailed, domain.JudgeRunStatusCanceled:
    default:
        return errors.New("invalid terminal status")
    }
    for i, jr := range m.list {
        if jr.ID == id && jr.Status == domain.JudgeRunStatusRunning {
            now := time.Now().UTC()
            m.list[i].Status = status
            m.list[i].RuntimeMS = runtimeMS
            m.list[i].MemoryKB = memoryKB
            m.list[i].ExitCode = exitCode
            m.list[i].ErrorMessage = errMsg
            m.list[i].FinishedAt = &now
            m.list[i].UpdatedAt = now
            return nil
        }
    }
    return ErrJudgeRunNotFound
}
