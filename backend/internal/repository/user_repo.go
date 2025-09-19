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

var ErrUserNotFound = errors.New("user not found")
var ErrUserDuplicate = errors.New("username already exists")

type UserRepository interface {
    Create(ctx context.Context, u domain.User) error
    GetByID(ctx context.Context, id string) (domain.User, error)
    GetByUsername(ctx context.Context, username string) (domain.User, error)
    UpdateRoles(ctx context.Context, id string, roles []string) error
    Delete(ctx context.Context, id string) error
    List(ctx context.Context, limit, offset int) ([]domain.User, error)
}

type PGUserRepository struct { pool *pgxpool.Pool }

func NewPGUserRepository(pool *pgxpool.Pool) *PGUserRepository { return &PGUserRepository{pool: pool} }

func (r *PGUserRepository) Create(ctx context.Context, u domain.User) error {
    if u.ID == "" { u.ID = uuid.New().String() }
    if u.CreatedAt.IsZero() { u.CreatedAt = time.Now().UTC() }
    _, err := r.pool.Exec(ctx, `INSERT INTO users (id, username, roles, created_at) VALUES ($1,$2,$3,$4)`,
        u.ID, u.Username, u.Roles, u.CreatedAt)
    if err != nil {
        if strings.Contains(strings.ToLower(err.Error()), "unique") { return ErrUserDuplicate }
        return err
    }
    return nil
}

func (r *PGUserRepository) GetByID(ctx context.Context, id string) (domain.User, error) {
    row := r.pool.QueryRow(ctx, `SELECT id, username, roles, created_at FROM users WHERE id=$1`, id)
    var u domain.User
    if err := row.Scan(&u.ID, &u.Username, &u.Roles, &u.CreatedAt); err != nil {
        if strings.Contains(err.Error(), "no rows") { return domain.User{}, ErrUserNotFound }
        return domain.User{}, err
    }
    return u, nil
}

func (r *PGUserRepository) GetByUsername(ctx context.Context, username string) (domain.User, error) {
    row := r.pool.QueryRow(ctx, `SELECT id, username, roles, created_at FROM users WHERE username=$1`, username)
    var u domain.User
    if err := row.Scan(&u.ID, &u.Username, &u.Roles, &u.CreatedAt); err != nil {
        if strings.Contains(err.Error(), "no rows") { return domain.User{}, ErrUserNotFound }
        return domain.User{}, err
    }
    return u, nil
}

func (r *PGUserRepository) UpdateRoles(ctx context.Context, id string, roles []string) error {
    cmd, err := r.pool.Exec(ctx, `UPDATE users SET roles=$1 WHERE id=$2`, roles, id)
    if err != nil { return err }
    if cmd.RowsAffected() == 0 { return ErrUserNotFound }
    return nil
}

func (r *PGUserRepository) Delete(ctx context.Context, id string) error {
    cmd, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
    if err != nil { return err }
    if cmd.RowsAffected() == 0 { return ErrUserNotFound }
    return nil
}

func (r *PGUserRepository) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
    if limit <= 0 { limit = 20 }
    rows, err := r.pool.Query(ctx, `SELECT id, username, roles, created_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
    if err != nil { return nil, err }
    defer rows.Close()
    res := make([]domain.User, 0)
    for rows.Next() {
        var u domain.User
        if err := rows.Scan(&u.ID, &u.Username, &u.Roles, &u.CreatedAt); err != nil { return nil, err }
        res = append(res, u)
    }
    return res, rows.Err()
}
