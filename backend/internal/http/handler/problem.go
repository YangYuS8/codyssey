package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"github.com/your-org/codyssey/backend/internal/domain"
	"github.com/your-org/codyssey/backend/internal/repository"
)

type ProblemRepo interface {
	Create(ctx context.Context, p domain.Problem) error
	GetByID(ctx context.Context, id uuid.UUID) (domain.Problem, error)
	Update(ctx context.Context, p domain.Problem) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, limit, offset int) ([]domain.Problem, error)
}

type ProblemCreateRequest struct {
	Title       string `json:"title" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"required,min=5"`
}

func CreateProblem(repo ProblemRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ProblemCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			respondError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		p := domain.NewProblem(req.Title, req.Description)
		if err := repo.Create(c.Request.Context(), p); err != nil {
			respondError(c, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
			return
		}
		respondCreated(c, p)
	}
}

func ListProblems(repo ProblemRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		items, err := repo.List(c.Request.Context(), limit, offset)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "LIST_FAILED", err.Error())
			return
		}
		meta := map[string]int{"limit": limit, "offset": offset, "count": len(items)}
		respondOK(c, items, meta)
	}
}

func GetProblem(repo ProblemRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil { respondError(c, http.StatusBadRequest, "INVALID_ID", "invalid uuid"); return }
		p, err := repo.GetByID(c.Request.Context(), id)
		if err != nil {
			if err == repository.ErrNotFound { respondError(c, http.StatusNotFound, "NOT_FOUND", "problem not found"); return }
			respondError(c, http.StatusInternalServerError, "GET_FAILED", err.Error()); return
		}
		respondOK(c, p, nil)
	}
}

type ProblemUpdateRequest struct {
	Title       *string `json:"title" binding:"omitempty,min=3,max=100"`
	Description *string `json:"description" binding:"omitempty,min=5"`
}

func UpdateProblem(repo ProblemRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil { respondError(c, http.StatusBadRequest, "INVALID_ID", "invalid uuid"); return }
		var req ProblemUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil { respondError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error()); return }
		// Load existing
		existing, err := repo.GetByID(c.Request.Context(), id)
		if err != nil {
			if err == repository.ErrNotFound { respondError(c, http.StatusNotFound, "NOT_FOUND", "problem not found"); return }
			respondError(c, http.StatusInternalServerError, "GET_FAILED", err.Error()); return
		}
		if req.Title != nil { existing.Title = *req.Title }
		if req.Description != nil { existing.Description = *req.Description }
		if err := repo.Update(c.Request.Context(), existing); err != nil {
			if err == repository.ErrNotFound { respondError(c, http.StatusNotFound, "NOT_FOUND", "problem not found"); return }
			respondError(c, http.StatusInternalServerError, "UPDATE_FAILED", err.Error()); return
		}
		respondOK(c, existing, nil)
	}
}

func DeleteProblem(repo ProblemRepo) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil { respondError(c, http.StatusBadRequest, "INVALID_ID", "invalid uuid"); return }
		if err := repo.Delete(c.Request.Context(), id); err != nil {
			if err == repository.ErrNotFound { respondError(c, http.StatusNotFound, "NOT_FOUND", "problem not found"); return }
			respondError(c, http.StatusInternalServerError, "DELETE_FAILED", err.Error()); return
		}
		respondOK(c, gin.H{"deleted": id.String()}, nil)
	}
}
