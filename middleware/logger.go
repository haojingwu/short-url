package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"short-url/logger"

	"go.uber.org/zap"
)

// LoggerMiddleware 记录每个请求的方法、路径、耗时 使用zap记录请求日志
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		//执行后续handler
		c.Next()

		//计算耗时
		duration := time.Since(start)

		//获取请求信息
		method := c.Request.Method
		path := c.Request.URL.Path
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		//如果有trace_id,从Context中取出
		traceID, _ := c.Get("trace_id")
		if traceID == nil {
			traceID = "N/A"
		}

		//用 zap 记录结构化日志
		logger.Info("HTTP 请求",
			zap.String("trace_id", traceID.(string)),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("Duration", duration),
			zap.String("client_ip", clientIP),
			zap.String("user_agent", userAgent),
		)
	}
}
