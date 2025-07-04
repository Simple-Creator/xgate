package models

import (
	"gorm.io/gorm"
)

type Connection struct {
	ID       uint   `gorm:"primaryKey" json:"id"`      // 主键
	Name     string `json:"name"`                      // 连接名称
	Host     string `json:"host"`                      // 主机地址
	Port     int    `json:"port"`                      // 端口
	Username string `json:"username"`                  // 用户名
	Password string `json:"password"`                  // 密码（加密存储）
	Group    string `json:"group" gorm:"default:默认分组"` // 分组
	UserID   uint   `json:"-"`                         //
}

func MigrateConnection(db *gorm.DB) {
	db.AutoMigrate(&Connection{})
}
