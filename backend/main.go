package main

import (
	"log"
	"os"
	"repair-system/config"
	"repair-system/routes"
	"repair-system/utils"

	"github.com/gin-gonic/gin"
)

func main() {
	config.InitConfig()

	utils.InitDB()

	if err := os.MkdirAll(config.AppConfig.UploadDir, 0755); err != nil {
		log.Printf("Warning: Could not create upload directory: %v", err)
	}

	r := gin.Default()

	routes.SetupRoutes(r)

	port := config.AppConfig.Port
	log.Printf("Server starting on port %s...", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
