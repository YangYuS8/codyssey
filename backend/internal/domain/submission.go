package domain

import "time"

// Submission 基础模型（后续可加执行统计、判题详情等）
type Submission struct {
    ID        string    `json:"id"`
    UserID    string    `json:"user_id"`
    ProblemID string    `json:"problem_id"`
    Language  string    `json:"language"`
    Code      string    `json:"code"`
    Status    string    `json:"status"`
    RuntimeMS int       `json:"runtime_ms"`
    MemoryKB  int       `json:"memory_kb"`
    ErrorMessage string `json:"error_message"`
    Version   int       `json:"version"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
