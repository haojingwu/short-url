package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 记录每个请求的方法、路径、耗时
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		//执行后续handler
		c.Next()

		//计算耗时
		duration := time.Since(start)

		//打印日志
		fmt.Printf(" [%s] %s %s -> %d (耗时: %v)\n",
			time.Now().Format("2026-01-02 15:04:05"),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration)
	}
}
