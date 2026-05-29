package middleware

import (
	"net/http"
	"os"
	"strings"

	"LMS-mini-project-backend/internal/config"
	"LMS-mini-project-backend/internal/models"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gin-gonic/gin"
)

// RequireAuth decodes incoming session tokens securely using the Clerk engine
func RequireAuth() gin.HandlerFunc {
	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			return
		}

		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{
			Token: parts[1],
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired Clerk token"})
			return
		}

		// Inject Clerk's unique user string identifier into context
		c.Set("user_id", claims.Subject)
		c.Next()
	}
}

// RequireRole checks the database to ensure the user has the specific required role(s)
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get the Clerk ID from the previous middleware
		clerkID, exists := c.Get("user_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized context"})
			return
		}

		// 2. Fetch the user and their associated Role from the database
		var user models.User
		if err := config.DB.Preload("Role").Where("clerk_id = ?", clerkID.(string)).First(&user).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User profile not found in database"})
			return
		}

		// 3. Verify if their DB role matches any of the allowed roles
		hasPermission := false
		for _, role := range allowedRoles {
			if user.Role.Name == role {
				hasPermission = true
				break
			}
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions to access this endpoint"})
			return
		}

		// Optional: Attach the full DB user to the context so your handlers don't have to query the DB again!
		c.Set("db_user", user)
		c.Next()
	}
}
