package routes

import (
	"jump-backend/controllers"
	"jump-backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		// 认证路由，不需要JWT中间件
		auth := api.Group("/")
		{
			auth.POST("/login", controllers.Login)
			auth.POST("/register", controllers.Register)
		}

		// 受保护的路由，需要JWT中间件
		protected := api.Group("/")
		protected.Use(middleware.JWTMiddleware())
		{
			protected.GET("/groups", controllers.GetGroups)

			protected.GET("/connections", controllers.GetConnections)
			protected.POST("/connections", controllers.AddConnection)
			protected.PUT("/connections/:id", controllers.UpdateConnection)
			protected.DELETE("/connections/:id", controllers.DeleteConnection)
			protected.POST("/connections/test", controllers.TestConnection)

			protected.GET("/files/:id", controllers.ListFiles)
			protected.POST("/files/:id/upload", controllers.UploadFile)
			protected.GET("/files/:id/download", controllers.DownloadFile)
			protected.DELETE("/files/:id", controllers.DeleteFile)
			protected.PUT("/files/:id/rename", controllers.RenameFile)
			protected.PUT("/files/:id/edit", controllers.EditFile)
			protected.GET("/files/:id/home", controllers.GetHomeDir)
			protected.GET("/files/:id/read", controllers.ReadFile)

			protected.GET("/terminal/:id", controllers.TerminalWS)

			protected.PUT("/users/changepwd", controllers.ChangePassword)

			// 管理员专属路由
			admin := protected.Group("/users")
			admin.Use(middleware.AdminRequired())
			{
				admin.GET("/", controllers.ListUsers)
				admin.POST("/", controllers.AddUser)
				admin.DELETE("/:id", controllers.DeleteUser)
				admin.PUT("/:id", controllers.UpdateUser)
				admin.PUT("/:id/resetpwd", controllers.ResetPassword)
			}
		}
	}
}
