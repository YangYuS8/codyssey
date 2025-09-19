package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/your-org/codyssey/backend/internal/domain"
)

var ErrNotFound = errors.New("problem not found")

type ProblemRepository interface {
	Create(ctx context.Context, p domain.Problem) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Problem, error)
	Update(ctx context.Context, p domain.Problem) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]domain.Problem, error)
}

type PGProblemRepository struct {
	pool *pgxpool.Pool
}

func NewPGProblemRepository(pool *pgxpool.Pool) *PGProblemRepository {
	return &PGProblemRepository{pool: pool}
}

func (r *PGProblemRepository) Create(ctx context.Context, p domain.Problem) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO problems (id,title,description,created_at) VALUES ($1,$2,$3,$4)`, p.ID, p.Title, p.Description, p.CreatedAt)
	return err
}

func (r *PGProblemRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Problem, error) {
	row := r.pool.QueryRow(ctx, `SELECT id,title,description,created_at FROM problems WHERE id=$1`, id)
	var p domain.Problem
	var pid uuid.UUID
	if err := row.Scan(&pid, &p.Title, &p.Description, &p.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) { return domain.Problem{}, ErrNotFound }
		return domain.Problem{}, err
	}
	p.ID = pid
	return p, nil
}

func (r *PGProblemRepository) Update(ctx context.Context, p domain.Problem) error {
	cmd, err := r.pool.Exec(ctx, `UPDATE problems SET title=$1, description=$2 WHERE id=$3`, p.Title, p.Description, p.ID)
	if err != nil { return err }
	if cmd.RowsAffected() == 0 { return ErrNotFound }
	return nil
}

func (r *PGProblemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	cmd, err := r.pool.Exec(ctx, `DELETE FROM problems WHERE id=$1`, id)
	if err != nil { return err }
	if cmd.RowsAffected() == 0 { return ErrNotFound }
	return nil
}

func (r *PGProblemRepository) List(ctx context.Context, limit, offset int) ([]domain.Problem, error) {
	if limit <= 0 { limit = 20 }
	rows, err := r.pool.Query(ctx, `SELECT id,title,description,created_at FROM problems ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil { return nil, err }
	defer rows.Close()
	var res []domain.Problem
	for rows.Next() {
		var p domain.Problem
		var id uuid.UUID
		if err := rows.Scan(&id, &p.Title, &p.Description, &p.CreatedAt); err != nil { return nil, err }
		p.ID = id
		res = append(res, p)
	}
	return res, rows.Err()
}

// Migration helper (idempotent) - 可在初始化时调用
func EnsureSchema(ctx context.Context, conn pgx.Tx) error {
	_, err := conn.Exec(ctx, `CREATE TABLE IF NOT EXISTS problems(
		id UUID PRIMARY KEY,
		title TEXT NOT NULL,
		description TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
	)`)
	return err
}
