package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"short-url/database"
	"short-url/middleware"
	"short-url/models"
	"short-url/utils"
)

func main() {
	//1.初始化数据库
	database.InitDB()
	db := database.DB

	//2.创建Gin引擎
	r := gin.New()

	//注册自定义中间件
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggerMiddleware())

	//3.注册路由
	//POST /shorten -生成短链 写入数据库
	r.POST("/shorten", func(c *gin.Context) {
		// 1解析请求体
		var req struct {
			URL string `json:"url" binding:"required"` //binding:"required"表示必须
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请提供有效的 url 字段",
			})
			return
		}
		fmt.Printf("收到URL: [%s] (长度: %d)\n", req.URL, len(req.URL))

		//先检查数据库,看这个URL是否已经生成过短链
		var existingURL models.URL
		result := db.Where("original_url = ?", req.URL).First(&existingURL)

		if result.Error == nil {
			//找到则返回已有的code
			c.JSON(http.StatusOK, gin.H{
				"code":         existingURL.Code,
				"original_url": existingURL.OriginalURL,
				"message":      "短链已存在",
			})
			return
		} else if result.Error != gorm.ErrRecordNotFound {
			//发生了其它数据库错误
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "查询失败"})
			return
		}

		//检查code是否已存在
		checkCodeExists := func(code string) bool {
			var count int64
			db.Model(&models.URL{}).Where("code = ?", code).Count(&count)
			return count > 0
		}

		//2生成短链码
		code := utils.GenerateShortCodeWithRetry(req.URL, checkCodeExists)

		//3创建URL记录
		url := models.URL{
			Code:        code,
			OriginalURL: req.URL,
			ClickCount:  0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		//4存入数据库
		if err := db.Create(&url).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "保存失败"})
			return
		}

		//5返回成功响应
		c.JSON(http.StatusOK, gin.H{
			"code":         code,
			"original_url": req.URL,
			"message":      "短链生成成功",
		})
	})

	//GET /:code - 跳转(从数据库查询)
	r.GET("/:code", func(c *gin.Context) {
		code := c.Param("code") //获取URL路径中的 :code参数

		//1从数据库查询
		var url models.URL
		result := db.Where("code = ?", code).First(&url)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "短链不存在",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "查询失败",
				})
			}
			return
		}

		//2异步更新点击统计(不阻塞跳转)
		go func() {
			//使用UpdateColumn跳过GORM钩子
			db.Model(&models.URL{}).Where("code = ?", code).
				UpdateColumn("click_count", gorm.Expr("click_count + 1"))
		}()

		//3 302临时重定向到原始URL
		c.Redirect(http.StatusFound, url.OriginalURL)

	})

	//GET /stats/:code -统计信息(从数据库查询)
	r.GET("/stats/:code", func(c *gin.Context) {
		code := c.Param("code")

		var url models.URL
		result := db.Where("code = ?", code).First(&url)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "短链不存在"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"code":         url.Code,
			"original_url": url.OriginalURL,
			"click_count":  url.ClickCount,
			"created_at":   url.CreatedAt,
		})
	})

	//3.启动服务,监听8080端口,后续使用viper抽离
	r.Run(":8080")

}
