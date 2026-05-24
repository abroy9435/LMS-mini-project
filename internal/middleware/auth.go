package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/jwt"
	"github.com/gin-gonic/gin"
)

// RequireAuth decodes incoming session tokens securely using the Clerk engine
func RequireAuth() gin.HandlerFunc {
	// Inject the secret key from your environment configuration into the Clerk client
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
		tokenString := parts[1]

		// Verify the incoming session JWT token seamlessly using Clerk's validation rules
		claims, err := jwt.Verify(c.Request.Context(), &jwt.VerifyParams{
			Token: tokenString,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired Clerk token: " + err.Error()})
			return
		}

		// Inject Clerk's unique user string identifier (e.g., "user_2xxxx") into context
		c.Set("user_id", claims.Subject)

		c.Next()
	}
}
