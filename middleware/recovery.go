package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

// RecoveryMiddleware 捕获panic,返回500,程序不崩溃
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				//打印堆栈信息
				fmt.Printf("panic 发送: %v\n", err)
				fmt.Printf("堆栈信息: \n%s\n", debug.Stack())

				//返回500错误
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "服务器内部错误",
				})
			}
		}()
		c.Next() //继续执行后续handler
	}
}
