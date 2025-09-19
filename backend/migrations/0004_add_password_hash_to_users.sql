-- +goose Up
-- 为用户增加 password_hash 列（仅注册/认证使用，不返回给客户端）
ALTER TABLE users ADD COLUMN IF NOT EXISTS password_hash TEXT;
-- 对于旧用户可后续补齐（允许 NULL，注册时必须非空）

-- +goose Down
ALTER TABLE users DROP COLUMN IF EXISTS password_hash;
