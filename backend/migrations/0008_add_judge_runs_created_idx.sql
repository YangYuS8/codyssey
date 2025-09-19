-- +goose Up
-- 为加速按 submission 最近 run 查询，添加复合索引（倒序可使用 DESC 避免排序开销）
CREATE INDEX IF NOT EXISTS idx_judge_runs_submission_created_desc ON judge_runs (submission_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_judge_runs_submission_created_desc;
