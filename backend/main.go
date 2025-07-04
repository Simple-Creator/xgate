package main

import (
	"xgate-backend/config"
	"xgate-backend/middleware"
	"xgate-backend/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化配置和数据库
	config.LoadConfig() // 先加载配置
	config.InitDB()     // 再初始化数据库

	r := gin.Default()
	r.Use(middleware.CORSMiddleware())

	r.MaxMultipartMemory = 100 << 20 // 100 MiB，支持大文件上传

	// 注册路由
	routes.RegisterRoutes(r)

	r.Run(":8080") // 默认监听8080端口
}
