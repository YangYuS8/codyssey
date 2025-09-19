-- +goose Up
CREATE TABLE IF NOT EXISTS problems (
    id UUID PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_problems_created_at ON problems(created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS problems;
