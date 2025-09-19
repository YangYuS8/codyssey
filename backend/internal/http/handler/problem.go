package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/your-org/codyssey/backend/internal/repository"
	"github.com/your-org/codyssey/backend/internal/service"
)

// 使用 service 层抽象，避免 handler 直接操作仓储
type ProblemService interface {
    Create(ctx any, title, desc string) (any, error)
    Get(ctx any, id uuid.UUID) (any, error)
    Update(ctx any, id uuid.UUID, title *string, desc *string) (any, error)
    Delete(ctx any, id uuid.UUID) error
    List(ctx any, limit, offset int) ([]any, error)
}

type ProblemCreateRequest struct {
	Title       string `json:"title" binding:"required,min=3,max=100"`
	Description string `json:"description" binding:"required,min=5"`
}

func CreateProblem(s *service.ProblemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ProblemCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			respondError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
			return
		}
		p, err := s.Create(c.Request.Context(), req.Title, req.Description)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "CREATE_FAILED", err.Error())
			return
		}
		respondCreated(c, p)
	}
}

func ListProblems(s *service.ProblemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
		items, err := s.List(c.Request.Context(), limit, offset)
		if err != nil {
			respondError(c, http.StatusInternalServerError, "LIST_FAILED", err.Error())
			return
		}
		meta := map[string]int{"limit": limit, "offset": offset, "count": len(items)}
		respondOK(c, items, meta)
	}
}

func GetProblem(s *service.ProblemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Param("id")
		id, err := uuid.Parse(idStr)
		if err != nil { respondError(c, http.StatusBadRequest, "INVALID_ID", "invalid uuid"); return }
		p, err := s.Get(c.Request.Context(), id)
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

func UpdateProblem(s *service.ProblemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil { respondError(c, http.StatusBadRequest, "INVALID_ID", "invalid uuid"); return }
		var req ProblemUpdateRequest
		if err := c.ShouldBindJSON(&req); err != nil { respondError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error()); return }
		updated, err := s.Update(c.Request.Context(), id, req.Title, req.Description)
		if err != nil {
			if err == repository.ErrNotFound { respondError(c, http.StatusNotFound, "NOT_FOUND", "problem not found"); return }
			respondError(c, http.StatusInternalServerError, "UPDATE_FAILED", err.Error()); return
		}
		respondOK(c, updated, nil)
	}
}

func DeleteProblem(s *service.ProblemService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := uuid.Parse(c.Param("id"))
		if err != nil { respondError(c, http.StatusBadRequest, "INVALID_ID", "invalid uuid"); return }
		if err := s.Delete(c.Request.Context(), id); err != nil {
			if err == repository.ErrNotFound { respondError(c, http.StatusNotFound, "NOT_FOUND", "problem not found"); return }
			respondError(c, http.StatusInternalServerError, "DELETE_FAILED", err.Error()); return
		}
		respondOK(c, gin.H{"deleted": id.String()}, nil)
	}
}
