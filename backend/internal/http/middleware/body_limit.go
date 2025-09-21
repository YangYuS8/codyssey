package middleware

import (
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// BodyLimit 限制请求体字节数，超出直接 413。
func BodyLimit(maxBytes int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if maxBytes <= 0 { c.Next(); return }
		// 如果已知 Content-Length 且超过直接拒绝
		if clStr := c.GetHeader("Content-Length"); clStr != "" {
			if cl, err := strconv.Atoi(clStr); err == nil && cl > maxBytes {
				c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{"data": nil, "error": gin.H{"code": "PAYLOAD_TOO_LARGE", "message": "request body too large"}})
				return
			}
		}
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, int64(maxBytes))
		// 读取阶段如果超限，MaxBytesReader 会返回错误，交由后续绑定时报错；这里设置自定义错误处理
		c.Next()
		if err := c.Errors.Last(); err != nil {
			// 如果是 io.EOF 正常，不处理
		}
	}
}

// ReadAndCheckSize 读取 code 并判定长度；供 handler 或 service 使用。
func ReadAndCheckSize(data string, max int) error {
	if max > 0 && len(data) > max { return io.ErrShortWrite } // 复用错误类型表明截断/超限
	return nil
}
