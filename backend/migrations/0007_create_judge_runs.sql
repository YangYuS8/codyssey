-- +goose Up
CREATE TABLE IF NOT EXISTS judge_runs (
    id UUID PRIMARY KEY,
    submission_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'queued',
    judge_version TEXT NOT NULL DEFAULT 'v1',
    runtime_ms INT NOT NULL DEFAULT 0,
    memory_kb INT NOT NULL DEFAULT 0,
    exit_code INT NOT NULL DEFAULT 0,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_judge_runs_submission ON judge_runs(submission_id);
CREATE INDEX IF NOT EXISTS idx_judge_runs_status ON judge_runs(status);

-- +goose Down
DROP TABLE IF EXISTS judge_runs;
