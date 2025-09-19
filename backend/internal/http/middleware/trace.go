package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const HeaderRequestID = "X-Request-ID"

// TraceID 生成或透传请求 ID，放入上下文
func TraceID() gin.HandlerFunc {
    return func(c *gin.Context) {
        rid := c.GetHeader(HeaderRequestID)
        if rid == "" {
            rid = uuid.New().String()
        }
        c.Writer.Header().Set(HeaderRequestID, rid)
        c.Set("request_id", rid)
        c.Next()
    }
}
