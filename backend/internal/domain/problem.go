package domain

import (
	"time"

	"github.com/google/uuid"
)

type Problem struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewProblem(title, description string) Problem {
	return Problem{
		ID:          uuid.New(),
		Title:       title,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}
}
