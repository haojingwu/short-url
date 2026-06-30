package database

import (
	"fmt"
	"log"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"short-url/models"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	// 从环境变量读取配置
	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1" // 默认本地
	}
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}
	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "123456"
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "short_url"
	}

	// 构建 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		user, password, host, port, dbName)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	err = DB.AutoMigrate(&models.URL{})
	if err != nil {
		log.Fatal("自动迁移失败:", err)
	}

	fmt.Println("✅ 数据库连接成功，表已创建/更新")
}
