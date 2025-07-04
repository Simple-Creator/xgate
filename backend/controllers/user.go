package controllers

import (
	"jump-backend/config"
	"jump-backend/middleware"
	"jump-backend/models"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// 登录接口
func Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}
	var user models.User
	if err := config.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "用户不存在"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "密码错误"})
		return
	}
	claims := models.Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString(middleware.JWTKey)
	c.JSON(http.StatusOK, gin.H{
		"token":    tokenStr,
		"role":     user.Role,
		"username": user.Username,
	})
}

// Register a new user
func Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "无效的参数"})
		return
	}

	// 检查用户名是否已存在
	var existingUser models.User
	if err := config.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"msg": "用户名已存在"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "数据库错误", "error": err.Error()})
		return
	}

	// 确定用户角色
	var userCount int64
	if err := config.DB.Model(&models.User{}).Count(&userCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "数据库错误", "error": err.Error()})
		return
	}
	role := "user"
	if userCount == 0 {
		role = "admin"
	}

	// 创建用户
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := models.User{
		Username: req.Username,
		Password: string(hash),
		Role:     role,
	}
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "创建用户失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"msg": "注册成功，请登录"})
}

// 用户列表（仅管理员）
func ListUsers(c *gin.Context) {
	if !isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"msg": "无权限"})
		return
	}
	var users []models.User
	config.DB.Find(&users)
	for i := range users {
		users[i].Password = ""
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// 新增用户（仅管理员）
func AddUser(c *gin.Context) {
	if !isAdmin(c) {
		c.JSON(http.StatusForbidden, gin.H{"msg": "无权限"})
		return
	}
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Username == "" || req.Password == "" || req.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	user := models.User{Username: req.Username, Password: string(hash), Role: req.Role}
	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "用户名已存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "添加成功"})
}

// 删除用户（仅管理员）
func DeleteUser(c *gin.Context) {
	id := c.Param("id")
	if err := config.DB.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
}

// 修改用户（仅管理员）
func UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Role string `json:"role"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}
	if err := config.DB.Model(&models.User{}).Where("id = ?", id).Update("role", req.Role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "更新成功"})
}

// 管理员重置用户密码
func ResetPassword(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err := config.DB.Model(&models.User{}).Where("id = ?", id).Update("password", string(hash)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "重置失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "重置成功"})
}

// 用户修改自己密码
func ChangePassword(c *gin.Context) {
	id, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "无法获取用户信息"})
		return
	}

	var user models.User
	if err := config.DB.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"msg": "用户未找到"})
		return
	}

	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.OldPassword == "" || req.NewPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)) != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"msg": "原密码错误"})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err := config.DB.Model(&models.User{}).Where("id = ?", user.ID).Update("password", string(hash)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"msg": "修改失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "修改成功"})
}

// 工具函数：判断是否管理员
func isAdmin(c *gin.Context) bool {
	token := c.GetHeader("Authorization")
	if token == "" {
		return false
	}
	token = strings.TrimPrefix(token, "Bearer ")
	claims := &models.Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return middleware.JWTKey, nil
	})
	return err == nil && claims.Role == "admin"
}
