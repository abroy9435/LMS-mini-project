package handlers

import (
	"LMS-mini-project-backend/internal/config"

	"net/http"

	"LMS-mini-project-backend/internal/models"

	"github.com/gin-gonic/gin"
)

// GetLeaveTypes fetches all available leave rules from the database
func GetLeaveTypes(c *gin.Context) {
	var leaveTypes []models.LeaveType

	// Using our global DB connection to find all leave types
	result := config.DB.Find(&leaveTypes)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leave types"})
		return
	}

	// Send the data back to the frontend as JSON
	c.JSON(http.StatusOK, gin.H{
		"data": leaveTypes,
	})
}
