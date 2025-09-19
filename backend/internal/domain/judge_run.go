package domain

import "time"

// JudgeRun 表示一次判题执行实例（可重试、多评测策略、不同判题内核版本等）
// 设计：Submission 与 JudgeRun 1:N
// 状态机：queued -> running -> (succeeded|failed|canceled)
// 可扩展：加入 "timeout" / "aborted" 等终态
const (
    JudgeRunStatusQueued    = "queued"
    JudgeRunStatusRunning   = "running"
    JudgeRunStatusSucceeded = "succeeded"
    JudgeRunStatusFailed    = "failed"
    JudgeRunStatusCanceled  = "canceled"
)

// JudgeRun 对应 judge_runs 表
type JudgeRun struct {
    ID            string    `json:"id"`
    SubmissionID  string    `json:"submission_id"`
    Status        string    `json:"status"`
    JudgeVersion  string    `json:"judge_version"` // 判题内核版本（可用于回放/调试）
    RuntimeMS     int       `json:"runtime_ms"`    // 运行耗时（聚合或最终）
    MemoryKB      int       `json:"memory_kb"`     // 峰值内存
    ExitCode      int       `json:"exit_code"`     // 进程退出码
    ErrorMessage  string    `json:"error_message"` // 失败/取消原因
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
    StartedAt     *time.Time `json:"started_at,omitempty"`
    FinishedAt    *time.Time `json:"finished_at,omitempty"`
}
