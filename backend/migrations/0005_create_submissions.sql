-- +goose Up
CREATE TABLE IF NOT EXISTS submissions (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    problem_id UUID NOT NULL,
    language TEXT NOT NULL,
    code TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_submissions_user ON submissions(user_id);
CREATE INDEX IF NOT EXISTS idx_submissions_problem ON submissions(problem_id);

-- +goose Down
DROP TABLE IF EXISTS submissions;
