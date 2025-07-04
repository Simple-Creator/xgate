package controllers

import (
	"fmt"
	"jump-backend/config"
	"jump-backend/models"
	"net"
	"net/http"
	"sort"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ====== 连接缓存实现（按用户隔离，管理员缓存全部） ======
var (
	connCache      = make(map[uint][]models.Connection) // 普通用户缓存：userID -> 连接列表
	connCacheAdmin []models.Connection                  // 管理员缓存全部连接
	connCacheLock  sync.RWMutex                         // 读写锁保证并发安全
)

func getUserInfo(c *gin.Context) (uint, string) {
	id, _ := c.Get("user_id")
	role, _ := c.Get("role")
	return id.(uint), role.(string)
}

// 获取所有连接（带缓存）
func GetConnections(c *gin.Context) {
	userID, role := getUserInfo(c)
	var connections []models.Connection

	// 1. 优先查缓存
	connCacheLock.RLock()
	if role == "admin" && connCacheAdmin != nil {
		connections = connCacheAdmin
	} else if role != "admin" {
		if v, ok := connCache[userID]; ok {
			connections = v
		}
	}
	connCacheLock.RUnlock()

	if connections != nil {
		// 命中缓存，直接分组返回
		groupedConnections := make(map[string][]models.Connection)
		for _, conn := range connections {
			groupedConnections[conn.Group] = append(groupedConnections[conn.Group], conn)
		}
		type GroupedResult struct {
			GroupName   string              `json:"groupName"`
			Connections []models.Connection `json:"connections"`
		}
		var result []GroupedResult
		var sortedGroups []string
		for name := range groupedConnections {
			sortedGroups = append(sortedGroups, name)
		}
		sort.Strings(sortedGroups)
		for _, groupName := range sortedGroups {
			result = append(result, GroupedResult{
				GroupName:   groupName,
				Connections: groupedConnections[groupName],
			})
		}
		if result == nil {
			result = []GroupedResult{}
		}
		c.JSON(http.StatusOK, result)
		return
	}

	// 2. 未命中缓存，查数据库
	db := config.DB
	if role != "admin" {
		db = db.Where("user_id = ?", userID)
	}
	if err := db.Order("`group` asc, name asc").Find(&connections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch connections"})
		return
	}

	// 3. 写入缓存
	connCacheLock.Lock()
	if role == "admin" {
		connCacheAdmin = connections
	} else {
		connCache[userID] = connections
	}
	connCacheLock.Unlock()

	// 4. 分组返回（与原逻辑一致）
	groupedConnections := make(map[string][]models.Connection)
	for _, conn := range connections {
		groupedConnections[conn.Group] = append(groupedConnections[conn.Group], conn)
	}
	type GroupedResult struct {
		GroupName   string              `json:"groupName"`
		Connections []models.Connection `json:"connections"`
	}
	var result []GroupedResult
	var sortedGroups []string
	for name := range groupedConnections {
		sortedGroups = append(sortedGroups, name)
	}
	sort.Strings(sortedGroups)
	for _, groupName := range sortedGroups {
		result = append(result, GroupedResult{
			GroupName:   groupName,
			Connections: groupedConnections[groupName],
		})
	}
	if result == nil {
		result = []GroupedResult{}
	}
	c.JSON(http.StatusOK, result)
}

// 获取所有组
func GetGroups(c *gin.Context) {
	userID, role := getUserInfo(c)
	var groups []string
	db := config.DB.Model(&models.Connection{})
	if role != "admin" {
		db = db.Where("user_id = ?", userID)
	}
	if err := db.Distinct().Pluck("`group`", &groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups"})
		return
	}

	if groups == nil {
		groups = []string{}
	}
	c.JSON(http.StatusOK, groups)
}

// 新增连接（操作后自动清理缓存）
func AddConnection(c *gin.Context) {
	userID, _ := getUserInfo(c)
	var conn models.Connection
	if err := c.ShouldBindJSON(&conn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	conn.UserID = userID
	if err := config.DB.Create(&conn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create connection"})
		return
	}
	// 清理缓存（当前用户和管理员）
	connCacheLock.Lock()
	delete(connCache, userID)
	connCacheAdmin = nil
	connCacheLock.Unlock()
	c.JSON(http.StatusOK, conn)
}

// 更新连接（操作后自动清理缓存）
func UpdateConnection(c *gin.Context) {
	userID, role := getUserInfo(c)
	id := c.Param("id")
	var conn models.Connection
	db := config.DB
	if role != "admin" {
		db = db.Where("user_id = ?", userID)
	}
	if err := db.First(&conn, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	var updatedConn models.Connection
	if err := c.ShouldBindJSON(&updatedConn); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedConn.ID = conn.ID
	updatedConn.UserID = conn.UserID

	if err := config.DB.Save(&updatedConn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update connection"})
		return
	}
	// 清理缓存（当前用户和管理员）
	connCacheLock.Lock()
	delete(connCache, userID)
	connCacheAdmin = nil
	connCacheLock.Unlock()
	c.JSON(http.StatusOK, updatedConn)
}

// 删除连接（操作后自动清理缓存）
func DeleteConnection(c *gin.Context) {
	userID, role := getUserInfo(c)
	id := c.Param("id")
	var conn models.Connection
	db := config.DB
	if role != "admin" {
		db = db.Where("user_id = ?", userID)
	}
	if err := db.First(&conn, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Connection not found"})
		return
	}

	if err := config.DB.Delete(&conn).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete connection"})
		return
	}
	// 清理缓存（当前用户和管理员）
	connCacheLock.Lock()
	delete(connCache, userID)
	connCacheAdmin = nil
	connCacheLock.Unlock()
	c.JSON(http.StatusOK, gin.H{"message": "Connection deleted successfully"})
}

// 测试连接
func TestConnection(c *gin.Context) {
	var req models.Connection
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "参数错误"})
		return
	}

	address := req.Host
	if req.Port != 0 {
		address = address + ":" + fmt.Sprint(req.Port)
	}
	conn, err := net.DialTimeout("tcp", address, 3*time.Second)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"msg": "连接失败"})
		return
	}
	conn.Close()
	c.JSON(http.StatusOK, gin.H{"msg": "连接成功"})
}
