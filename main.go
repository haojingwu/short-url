package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"short-url/cache"
	"short-url/database"
	"short-url/logger"
	"short-url/middleware"
	"short-url/models"
	"short-url/utils"
)

func main() {
	//1.初始化日志
	logger.InitLogger()
	defer logger.Sync()

	//2.初始化数据库
	database.InitDB()
	db := database.DB

	//3.初始化 Redis
	cache.InitRedis()

	//4.创建Gin引擎
	r := gin.New()

	//注册自定义中间件
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.LoggerMiddleware())

	//注册路由
	//POST /shorten -生成短链 写入数据库(写入时清除缓存)
	r.POST("/shorten", func(c *gin.Context) {
		// 解析请求体
		var req struct {
			URL string `json:"url" binding:"required"` //binding:"required"表示必须
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			logger.Warn("请求参数错误", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "请提供有效的 url 字段",
			})
			return
		}
		logger.Info("收到生成请求短链请求", zap.String("url", req.URL))
		//幂等性检查
		//先检查数据库,看这个URL是否已经生成过短链
		var existingURL models.URL
		result := db.Where("original_url = ?", req.URL).First(&existingURL)

		if result.Error == nil {
			logger.Info("短链已存在(幂等返回)",
				zap.String("code", existingURL.Code),
				zap.String("url", req.URL),
			)
			//找到则返回已有的code
			c.JSON(http.StatusOK, gin.H{
				"code":         existingURL.Code,
				"original_url": existingURL.OriginalURL,
				"message":      "短链已存在(幂等返回)",
			})
			return
		} else if result.Error != gorm.ErrRecordNotFound {
			//发生了其它数据库错误
			logger.Error("查询数据库失败", zap.Error(result.Error))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "查询失败"})
			return
		}
		//生成短链码
		//检查code是否已存在
		checkCodeExists := func(code string) bool {
			var count int64
			db.Model(&models.URL{}).Where("code = ?", code).Count(&count)
			return count > 0
		}

		code := utils.GenerateShortCodeWithRetry(req.URL, checkCodeExists)

		//创建URL记录
		url := models.URL{
			Code:        code,
			OriginalURL: req.URL,
			ClickCount:  0,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		//存入数据库
		if err := db.Create(&url).Error; err != nil {
			logger.Error("保存数据库失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "保存失败"})
			return
		}

		//删除缓存(保证数据一致性)
		ctx := c.Request.Context()
		cacheKey := "short:" + code
		if err := cache.DeleteCache(ctx, cacheKey); err != nil {
			logger.Warn("删除缓存失败", zap.Error(err), zap.String("key", cacheKey))
		}

		//返回成功响应
		logger.Info("短链生成成功",
			zap.String("code", code),
			zap.String("url", req.URL),
		)
		c.JSON(http.StatusOK, gin.H{
			"code":         code,
			"original_url": req.URL,
			"message":      "短链生成成功",
		})
	})

	//GET /:code - 跳转(从数据库查询)
	r.GET("/:code", func(c *gin.Context) {
		code := c.Param("code") //获取URL路径中的 :code参数
		ctx := c.Request.Context()
		cacheKey := "short:" + code

		//查Redis缓存
		cached, err := cache.GetCache(ctx, cacheKey)
		if err == nil {
			//缓存命中 反序列化JSON得到URL对象
			var url models.URL
			if err := json.Unmarshal([]byte(cached), &url); err == nil {
				//异步更新点击统计(不阻塞跳转
				go safeIncrementClick(db, code)

				logger.Info("缓存命中",
					zap.String("code", code),
					zap.String("url", url.OriginalURL))
				c.Redirect(http.StatusFound, url.OriginalURL)
				return
			}
		}
		//缓存未命中 查数据库
		var url models.URL
		result := db.Where("code = ?", code).First(&url)

		if result.Error != nil {
			if result.Error == gorm.ErrRecordNotFound {
				logger.Warn("短链不存在", zap.String("code", code))
				c.JSON(http.StatusNotFound, gin.H{
					"error": "短链不存在",
				})
			} else {
				logger.Error("查询数据库失败", zap.Error(result.Error))
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "查询失败",
				})
			}
			return
		}

		//回写Redis缓存(TTL1小时)
		urlJSON, _ := json.Marshal(url)

		if err := cache.SetCache(ctx, cacheKey, string(urlJSON), 1*time.Hour); err != nil {
			//缓存写入失败不影响主流程
			logger.Warn("写入缓存失败", zap.Error(err), zap.String("key", cacheKey))
		}

		//2异步更新点击统计(不阻塞跳转)
		go safeIncrementClick(db, code)

		logger.Info("缓存未名字, 查 DB 后回写",
			zap.String("code", code),
			zap.String("url", url.OriginalURL),
		)

		//3 302临时重定向到原始URL
		c.Redirect(http.StatusFound, url.OriginalURL)

	})

	//GET /stats/:code -统计信息(从数据库查询,不缓存)
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

		logger.Info("查询统计",
			zap.String("code", code),
			zap.Int("click_count", url.ClickCount),
		)

		c.JSON(http.StatusOK, gin.H{
			"code":         url.Code,
			"original_url": url.OriginalURL,
			"click_count":  url.ClickCount,
			"created_at":   url.CreatedAt,
		})
	})

	logger.Info("🚀 服务启动成功", zap.String("port", "8080"))

	if err := r.Run(":8080"); err != nil {
		logger.Fatal("服务启动失败", zap.Error(err))
	}

}

// 安全异步更新统计(带recover防止panic)
func safeIncrementClick(db *gorm.DB, code string) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("异步更新统计 panic",
				zap.Any("error", err),
				zap.String("code", code),
			)
		}
	}()

	if err := db.Model(&models.URL{}).
		Where("code = ?", code).
		UpdateColumn("click_count", gorm.Expr("click_count + 1")).Error; err != nil {
		logger.Error("异步更新统计失败",
			zap.Error(err),
			zap.String("code", code),
		)
	}
}
