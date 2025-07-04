package config

import (
	"fmt"
	"jump-backend/models"
	"log"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

// Config struct holds all configuration for the application.
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
	Path     string `mapstructure:"path"` // For SQLite
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

// C is the global configuration object.
var C Config

// LoadConfig loads config from file or environment variables.
func LoadConfig() {
	// 1. 设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("jwt.secret", "a_very_secret_key_for_local_dev")
	viper.SetDefault("database.type", "sqlite")
	viper.SetDefault("database.path", "xgate.db")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.user", "root")
	viper.SetDefault("database.password", "")
	viper.SetDefault("database.name", "xgate")

	// 2. 设置Viper从环境变量读取
	viper.AutomaticEnv()
	viper.SetEnvPrefix("XGATE")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// 3. 将所有配置解析到C结构体
	// 优先级: 环境变量 > 默认值
	if err := viper.Unmarshal(&C); err != nil {
		log.Fatalf("无法解析配置: %s", err)
	}
}

func InitDB() {
	var err error
	var dsn string

	switch C.Database.Type {
	case "mysql":
		dsn = fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			C.Database.User,
			C.Database.Password,
			C.Database.Host,
			C.Database.Port,
			C.Database.Name,
		)
		DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("无法连接到数据库: %v. DSN: %s", err, dsn)
		}

	case "sqlite":
		dsn = C.Database.Path
		DB, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("无法连接到数据库: %v. DSN: %s", err, dsn)
		}
	default:
		log.Fatalf("不支持的数据库类型: %s", C.Database.Type)
	}

	// 测试数据库连接
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("获取数据库实例失败: %v", err)
	}

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("数据库Ping失败: %v", err)
	}

	DB.AutoMigrate(&models.User{}, &models.Connection{})
}
