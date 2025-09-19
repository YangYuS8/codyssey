package handler

import "github.com/gin-gonic/gin"

// Standard API response formats
// Success: { "data": <payload>, "meta": {..optional..}, "error": null }
// Error:   { "data": null, "error": { "code": "<CODE>", "message": "..." } }

type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

type SuccessResponse struct {
    Data any         `json:"data"`
    Meta any         `json:"meta,omitempty"`
    Err  *APIError   `json:"error"`
}

type ErrorResponse struct {
    Data any       `json:"data"`
    Err  *APIError `json:"error"`
}

func respondOK(c *gin.Context, data any, meta any) {
    c.JSON(200, SuccessResponse{Data: data, Meta: meta, Err: nil})
}

func respondCreated(c *gin.Context, data any) {
    c.JSON(201, SuccessResponse{Data: data, Err: nil})
}

func respondError(c *gin.Context, status int, code, message string) {
    c.JSON(status, ErrorResponse{Data: nil, Err: &APIError{Code: code, Message: message}})
}
