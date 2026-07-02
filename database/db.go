package database

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"short-url/config"
	"short-url/models"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	cfg := config.Cfg

	dsn := cfg.Database.GetDSN()

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

	log.Println("✅ 数据库连接成功，表已创建/更新")
}
