package handlers

import (
	"net/http"

	"LMS-mini-project-backend/internal/config"
	"LMS-mini-project-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// We only need this one struct now to build their local profile
type CreateProfileRequest struct {
	Email        string     `json:"email" binding:"required,email"`
	FirstName    string     `json:"first_name" binding:"required"`
	LastName     string     `json:"last_name" binding:"required"`
	EmployeeID   string     `json:"employee_id" binding:"required"`
	RoleID       uuid.UUID  `json:"role_id" binding:"required"`
	DepartmentID *uuid.UUID `json:"department_id"` // Optional
}

// CreateProfile maps the Clerk User ID to a local University database record
func CreateProfile(c *gin.Context) {
	var req CreateProfileRequest

	// 1. Bind the incoming JSON profile data
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload: " + err.Error()})
		return
	}

	// 2. Extract the secure Clerk ID placed here by our new middleware
	clerkID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized request"})
		return
	}

	// 3. Create the User profile in our local PostgreSQL 'users' table
	// We generate a new standard UUID for the database primary key, but link the Clerk string to it
	newUser := models.User{
		ID:           uuid.New(),
		ClerkID:      clerkID.(string),
		EmployeeID:   req.EmployeeID,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Email:        req.Email,
		RoleID:       req.RoleID,
		DepartmentID: req.DepartmentID,
	}

	// 4. Save to database
	if result := config.DB.Create(&newUser); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create local user profile: " + result.Error.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "University profile synchronized successfully!",
		"user":    newUser,
	})
}

// GetMe fetches the user profile using the Clerk ID
func GetMe(c *gin.Context) {
	clerkID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found in context"})
		return
	}

	var user models.User

	// Notice we query by clerk_id now, not id
	if err := config.DB.Where("clerk_id = ?", clerkID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User profile not found in database"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}
