package models

import (
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/gorm"
)

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Username string `json:"username" gorm:"uniqueIndex;type:varchar(191)"`
	Password string `json:"password" gorm:"type:varchar(255)"`
	Role     string `json:"role" gorm:"default:user"`
}

type Claims struct {
	UserID uint   `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func MigrateUser(db *gorm.DB) {
	db.AutoMigrate(&User{})
}
