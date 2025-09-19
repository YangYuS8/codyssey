package domain

import "time"

// User 领域模型（简化版，后续可扩展 email, password_hash 等）
type User struct {
    ID        string    `json:"id"`
    Username  string    `json:"username"`
    Roles     []string  `json:"roles"`
    CreatedAt time.Time `json:"created_at"`
    PasswordHash string `json:"-"`
}
