package config

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type AppConfig struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	Log      LogConfig      `mapstructure:"log"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
	Compress   bool   `mapstructure:"compress"`
}

var Cfg *AppConfig

func InitConfig() {
	// 1. 初始化 Cfg（防止 nil pointer）
	Cfg = &AppConfig{}

	// 2. 设置配置文件名和类型
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	// 3. 添加配置搜索路径
	viper.AddConfigPath("./config")
	viper.AddConfigPath(".")

	// 4. 根据环境变量加载不同配置文件
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}

	if env == "prod" {
		viper.SetConfigName("config.prod")
		fmt.Println("📦 加载生产配置: config.prod.yaml")
	} else {
		fmt.Println("🛠️ 加载开发配置: config.yaml")
	}

	// 5. 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		// 如果配置文件不存在，给出更清晰的提示
		log.Printf("⚠️ 读取配置文件失败: %v\n", err)
		log.Printf("📁 当前工作目录: %s\n", getWorkingDir())
		log.Fatalf("❌ 请确保 config/config.yaml 或 config/config.prod.yaml 存在")
	}

	// 6. 支持环境变量覆盖
	viper.AutomaticEnv()
	viper.SetEnvPrefix("APP")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 7. 反序列化到结构体
	if err := viper.Unmarshal(Cfg); err != nil {
		log.Fatalf("❌ 解析配置失败: %v", err)
	}

	// 8. 打印配置摘要
	fmt.Println("✅ 配置加载成功")
	fmt.Printf("   服务端口: %s\n", Cfg.Server.Port)
	fmt.Printf("   数据库: %s:%s/%s\n", Cfg.Database.Host, Cfg.Database.Port, Cfg.Database.Name)
	fmt.Printf("   Redis: %s (DB %d)\n", Cfg.Redis.Addr, Cfg.Redis.DB)
}

func getWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "未知"
	}
	return dir
}

func (d *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.User, d.Password, d.Host, d.Port, d.Name)
}

func (r *RedisConfig) GetAddr() string {
	return r.Addr
}
