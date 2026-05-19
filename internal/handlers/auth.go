package handlers

import (
	"net/http"
	"os"

	"LMS-mini-project-backend/internal/config"
	"LMS-mini-project-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/nedpals/supabase-go"
)

// Helper function to initialize the Supabase client
func getSupabaseClient() *supabase.Client {
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_ANON_KEY")
	return supabase.CreateClient(supabaseURL, supabaseKey)
}

// Structs to define what JSON we expect from the frontend
type RegisterRequest struct {
	Email        string     `json:"email" binding:"required,email"`
	Password     string     `json:"password" binding:"required,min=6"`
	FirstName    string     `json:"first_name" binding:"required"`
	LastName     string     `json:"last_name" binding:"required"`
	EmployeeID   string     `json:"employee_id" binding:"required"`
	RoleID       uuid.UUID  `json:"role_id" binding:"required"`
	DepartmentID *uuid.UUID `json:"department_id"` // Optional
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

// Register creates the auth user AND the local database profile
func Register(c *gin.Context) {
	var req RegisterRequest

	// 1. Bind the incoming JSON to our struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	sb := getSupabaseClient()

	// 2. Register the user in Supabase Auth
	authUser, err := sb.Auth.SignUp(c.Request.Context(), supabase.UserCredentials{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Supabase Auth Failed: " + err.Error()})
		return
	}

	// 3. Parse the UUID Supabase just generated
	parsedUUID, _ := uuid.Parse(authUser.ID)

	// 4. Create the User profile in our local PostgreSQL 'users' table
	newUser := models.User{
		ID:           parsedUUID, // This links our table exactly to Supabase auth.users
		EmployeeID:   req.EmployeeID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		RoleID:       req.RoleID,
		DepartmentID: req.DepartmentID,
	}

	// 5. Save to database
	if result := config.DB.Create(&newUser); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create local user profile: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully!",
		"user":    newUser,
	})
}

// Login verifies credentials and returns the JWT
func Login(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	sb := getSupabaseClient()

	// Sign in using Supabase
	authData, err := sb.Auth.SignIn(c.Request.Context(), supabase.UserCredentials{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid email or password"})
		return
	}

	// Return the JWT Token to the frontend!
	c.JSON(http.StatusOK, gin.H{
		"message":      "Login successful",
		"access_token": authData.AccessToken,
		"user_id":      authData.User.ID,
	})
}

func ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	redirectURL := os.Getenv("FRONTEND_URL")
	if redirectURL == "" {
		redirectURL = "http://localhost:3000/reset-password"
	}

	sb := getSupabaseClient()

	err := sb.Auth.ResetPasswordForEmail(c.Request.Context(), req.Email, redirectURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send reset email: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset email sent. Please check your inbox.",
	})
}

func GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	var user models.User

	if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User profile not found in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
