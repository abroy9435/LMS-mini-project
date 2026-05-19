package main

import (
	"LMS-mini-project-backend/internal/config"
	"LMS-mini-project-backend/internal/handlers"
	"LMS-mini-project-backend/internal/middleware" // Import the middleware
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Starting University LMS Backend...")

	config.ConnectDatabase()
	router := gin.Default()

	// PUBLIC ROUTES (No token required)
	public := router.Group("/api/v1")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "Server is running smoothly!"})
		})
		public.POST("/register", handlers.Register)
		public.POST("/login", handlers.Login)
		public.POST("/forgot-password", handlers.ForgotPassword)
	}

	// PRIVATE ROUTES (Requires a valid Supabase JWT)
	private := router.Group("/api/v1")
	private.Use(middleware.RequireAuth()) // Apply the lock here!
	{
		// This route is now protected!
		private.GET("/leave-types", handlers.GetLeaveTypes)
		private.GET("/me", handlers.GetMe)
	}

	router.Run(":8080")
}
