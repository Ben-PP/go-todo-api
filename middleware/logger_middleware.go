package middleware

import (
	"go-todo/logging"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		logging.LogReq(
			c.ClientIP(),
			c.Request.Method,
			c.FullPath(),
			c.Request.UserAgent(),
			c.Writer.Status(),
		)
	}
}
