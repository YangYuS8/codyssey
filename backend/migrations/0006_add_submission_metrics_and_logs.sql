-- +goose Up
ALTER TABLE submissions
    ADD COLUMN IF NOT EXISTS runtime_ms INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS memory_kb INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS error_message TEXT,
    ADD COLUMN IF NOT EXISTS version INT NOT NULL DEFAULT 1;

CREATE TABLE IF NOT EXISTS submission_status_logs (
    id UUID PRIMARY KEY,
    submission_id UUID NOT NULL REFERENCES submissions(id) ON DELETE CASCADE,
    from_status TEXT NOT NULL,
    to_status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_submission_status_logs_submission ON submission_status_logs(submission_id);

-- +goose Down
DROP TABLE IF EXISTS submission_status_logs;
ALTER TABLE submissions
    DROP COLUMN IF EXISTS runtime_ms,
    DROP COLUMN IF EXISTS memory_kb,
    DROP COLUMN IF EXISTS error_message,
    DROP COLUMN IF EXISTS version;
