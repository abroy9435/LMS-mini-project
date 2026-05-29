package main

import (
	"LMS-mini-project-backend/internal/config"
	"LMS-mini-project-backend/internal/handlers"
	"LMS-mini-project-backend/internal/middleware" // Import the middleware
	"fmt"
	"os"
	"time"

	"github.com/gin-contrib/cors" // 1. Import the Gin CORS package
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("Starting University LMS Backend...")

	config.ConnectDatabase()
	router := gin.Default()

	// 2. Configure and attach the CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Allow your React Vite app
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour, // Cache the preflight request for 12 hours
	}))

	// PUBLIC ROUTES (No token required)
	public := router.Group("/api/v1")
	{
		public.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "Server is running smoothly!"})
		})

		public.POST("/webhooks/clerk", handlers.HandleClerkWebhook)
	}

	// PRIVATE ROUTES (Requires a valid Supabase JWT)
	private := router.Group("/api/v1")
	private.Use(middleware.RequireAuth()) // Apply the lock here!
	{
		private.POST("/profile", handlers.CreateProfile)
		private.GET("/leave-types", handlers.GetLeaveTypes)
		private.GET("/me", handlers.GetMe)
		private.POST("/leaves", middleware.RequireAuth(), handlers.ApplyForLeave)
		private.PUT("/leaves/:id/status", middleware.RequireAuth(), middleware.RequireRole("APPROVER", "ADMIN"), handlers.UpdateLeaveStatus)
		private.GET("/leaves/pending", handlers.GetPendingLeaves)
		private.GET("/leaves/me", handlers.GetMyLeaves)
		private.GET("/balances/me", handlers.GetMyBalances)
		private.POST("/admin/allocate-leaves", handlers.AllocateYearlyLeaves)
		private.PUT("/admin/assign-role", middleware.RequireAuth(), middleware.RequireRole("ADMIN"), handlers.AssignRole)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "7860"
	}

	router.Run("0.0.0.0:" + port)
}
