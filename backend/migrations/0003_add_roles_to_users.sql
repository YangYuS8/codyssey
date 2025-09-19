-- +goose Up
-- 若 users 表不存在则创建（容错），并确保 roles 列存在
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- 添加 roles 列 (TEXT[]) （直接使用 IF NOT EXISTS 简化，避免 DO $$ 块带来的 goose/dialect 解析问题）
ALTER TABLE users ADD COLUMN IF NOT EXISTS roles TEXT[] NOT NULL DEFAULT ARRAY[]::TEXT[];
-- 不再创建 idx_users_username：因为 username 已 UNIQUE 自动有唯一约束索引。

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS roles;
