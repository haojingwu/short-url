package database

import (
	"fmt"
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"short-url/models"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	//配置 DSN
	//格式: 用户名:密码@tcp(主机:端口)/数据库名?charset=utf8mb4&parseTime=True&loc=Local
	dsn := "root:root@tcp(127.0.0.1:3306)/short_url?charset=utf8mb4&parseTime=True&loc=Local"

	//连接数据库
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info), //开发环境打印SQL日志
	})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	//自动迁移(创建表 如果表不存在)
	err = DB.AutoMigrate(&models.URL{})
	if err != nil {
		log.Fatal("自动迁移失败:", err)
	}

	fmt.Println("数据库连接成功,表已经创建/更新")
}
