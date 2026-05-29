package handlers

import (
	"net/http"
	"time"

	"LMS-mini-project-backend/internal/config"
	"LMS-mini-project-backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AssignRoleRequest struct {
	UserID   uuid.UUID `json:"user_id" binding:"required"`
	RoleName string    `json:"role_name" binding:"required"`
}

// AllocateYearlyLeaves distributes the default leave quotas to all users for the current year
func AllocateYearlyLeaves(c *gin.Context) {
	currentYear := time.Now().Year()

	var users []models.User
	var leaveTypes []models.LeaveType

	// 1. Fetch every user and every leave rule from the database
	if err := config.DB.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	if err := config.DB.Find(&leaveTypes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leave types"})
		return
	}

	allocationsCreated := 0

	// 2. Loop through every single user in the system
	for _, user := range users {
		// 3. Loop through every type of leave (CL, ML, EL, etc.)
		for _, leaveType := range leaveTypes {

			// Check if this specific allocation already exists (prevents duplicate data if run twice)
			var existingBalance models.LeaveBalance
			err := config.DB.Where("user_id = ? AND leave_type_id = ? AND year = ?", user.ID, leaveType.ID, currentYear).First(&existingBalance).Error

			if err != nil {
				// Record not found, so we create a fresh ledger for this year!
				newBalance := models.LeaveBalance{
					UserID:        user.ID,
					LeaveTypeID:   leaveType.ID,
					Year:          currentYear,
					AllocatedDays: leaveType.DefaultDays,
					RemainingDays: leaveType.DefaultDays, // Starts at maximum capacity
					UsedDays:      0.0,
				}

				if err := config.DB.Create(&newBalance).Error; err == nil {
					allocationsCreated++
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "Yearly leave allocation completed successfully!",
		"year":                currentYear,
		"allocations_created": allocationsCreated,
	})
}

// AssignRole allows a System Administrator to promote/demote users
func AssignRole(c *gin.Context) {
	var req AssignRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// 1. Verify the requested Role actually exists in the database
	var newRole models.Role
	if err := config.DB.Where("name = ?", req.RoleName).First(&newRole).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Role does not exist in the system"})
		return
	}

	// 2. Update the user's role_id
	if err := config.DB.Model(&models.User{}).Where("id = ?", req.UserID).Update("role_id", newRole.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "User role updated successfully",
		"new_role": req.RoleName,
	})
}
