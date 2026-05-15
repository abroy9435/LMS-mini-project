package main

import (
	"LMS-mini-project-backend/internal/config"
	"LMS-mini-project-backend/internal/handlers"
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Starting University LMS Backend...")

	// 1. Initialize Database Connection
	config.ConnectDatabase()

	// 2. Initialize Gin Router
	router := gin.Default()

	// 3. Setup Routes
	// Grouping our API versions is a great best practice
	api := router.Group("/api/v1")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "Server is running smoothly!"})
		})

		// The new connection!
		api.GET("/leave-types", handlers.GetLeaveTypes)
	}

	// 4. Start the server
	router.Run(":8080")
}
