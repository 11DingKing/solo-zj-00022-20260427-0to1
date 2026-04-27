package routes

import (
	"net/http"
	"repair-system/controllers"
	"repair-system/middleware"
	"repair-system/models"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	r.Use(corsMiddleware())

	r.Static("/uploads", "./uploads")

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/login", controllers.Login)
		}

		api.Use(middleware.JWTAuth())
		{
			api.GET("/user/me", controllers.GetCurrentUser)
			api.GET("/technicians", controllers.GetTechnicians)
			api.GET("/devices", controllers.GetDevices)
			api.GET("/devices/:id", controllers.GetDeviceByID)

			api.POST("/images/upload", controllers.UploadImage)

			api.GET("/work-orders", controllers.GetWorkOrders)
			api.GET("/work-orders/:id", controllers.GetWorkOrderByID)

			employee := api.Group("")
			employee.Use(middleware.RequireRole(string(models.RoleEmployee)))
			{
				employee.POST("/work-orders", controllers.CreateWorkOrder)
				employee.PUT("/work-orders/:id/confirm", controllers.ConfirmWorkOrder)
				employee.PUT("/work-orders/:id/reject", controllers.RejectWorkOrder)
			}

			admin := api.Group("")
			admin.Use(middleware.RequireRole(string(models.RoleAdmin)))
			{
				admin.GET("/users", controllers.GetUsers)
				admin.GET("/users/:id", controllers.GetUserByID)
				admin.POST("/users", controllers.CreateUser)
				admin.PUT("/users/:id", controllers.UpdateUser)
				admin.DELETE("/users/:id", controllers.DeleteUser)
				admin.PUT("/users/:id/reset-password", controllers.ResetPassword)

				admin.POST("/devices", controllers.CreateDevice)
				admin.PUT("/devices/:id", controllers.UpdateDevice)
				admin.DELETE("/devices/:id", controllers.DeleteDevice)

				admin.PUT("/work-orders/:id/assign", controllers.AssignWorkOrder)

				admin.GET("/dashboard/stats", controllers.GetDashboardStats)
				admin.GET("/dashboard/fault-type-distribution", controllers.GetFaultTypeDistribution)
				admin.GET("/dashboard/30-days-trend", controllers.GetLast30DaysTrend)
				admin.GET("/dashboard/technician-ranking", controllers.GetTechnicianRanking)
			}

			technician := api.Group("")
			technician.Use(middleware.RequireRole(string(models.RoleTechnician)))
			{
				technician.PUT("/work-orders/:id/accept", controllers.AcceptWorkOrder)
				technician.PUT("/work-orders/:id/submit", controllers.SubmitRepair)
			}

			api.GET("/ping", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "pong"})
			})
		}
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
