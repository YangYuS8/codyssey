package domain

import "time"

// SubmissionStatusLog 记录一次状态流转（from -> to）。
// 不保证所有历史都存在（若未来需要可加外部导入），但应用内调用应保持追加式。
type SubmissionStatusLog struct {
    ID           string    `json:"id"`
    SubmissionID string    `json:"submission_id"`
    FromStatus   string    `json:"from_status"`
    ToStatus     string    `json:"to_status"`
    CreatedAt    time.Time `json:"created_at"`
}
