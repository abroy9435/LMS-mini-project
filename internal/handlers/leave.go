package handlers

import (
	"net/http"
	"time"

	"LMS-mini-project-backend/internal/config"
	"LMS-mini-project-backend/internal/models"
	"LMS-mini-project-backend/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type SubmitLeaveRequest struct {
	LeaveTypeID uuid.UUID `json:"leave_type_id" binding:"required"`
	StartDate   string    `json:"start_date" binding:"required"` // Expecting YYYY-MM-DD
	EndDate     string    `json:"end_date" binding:"required"`   // Expecting YYYY-MM-DD
	Reason      string    `json:"reason" binding:"required"`
}

// ApplyForLeave handles the creation of a new transactional leave request
func ApplyForLeave(c *gin.Context) {
	var req SubmitLeaveRequest

	// 1. Validate the incoming JSON structure
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input: " + err.Error()})
		return
	}

	// 2. Parse string dates into Go time.Time objects
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start_date format. Use YYYY-MM-DD"})
		return
	}

	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end_date format. Use YYYY-MM-DD"})
		return
	}

	if end.Before(start) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "End date cannot be chronologically before start date"})
		return
	}

	// 3. Extract Clerk user_id securely from the JWT Context
	clerkID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized request"})
		return
	}

	// 4. Look up the Internal Database UUID associated with this Clerk ID
	var user models.User
	if err := config.DB.Where("clerk_id = ?", clerkID.(string)).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User profile not found in database. Please sync your profile first."})
		return
	}

	// 5. Look up the LeaveType in the DB to know its actual name (CL vs ML) for the calculation engine
	var leaveType models.LeaveType
	if err := config.DB.First(&leaveType, "id = ?", req.LeaveTypeID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Selected Leave Type does not exist"})
		return
	}

	// 6. Run the calculation engine (skipping weekends/holidays based on your DB configurations)
	calculatedDays := utils.CalculateNetLeaveDays(start, end, leaveType.Name)

	// 7. Build the structural database model object
	newLeave := models.LeaveRequest{
		ID:             uuid.New(),
		UserID:         user.ID, // <-- CRITICAL FIX: Using the internal DB UUID, not the Clerk string!
		LeaveTypeID:    req.LeaveTypeID,
		StartDate:      start,
		EndDate:        end,
		CalculatedDays: calculatedDays,
		Reason:         req.Reason,
		Status:         "PENDING", // Hardcoded starting status
	}

	// 8. Push straight into your transactional SQL table
	if err := config.DB.Create(&newLeave).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database tracking failure: " + err.Error()})
		return
	}

	// 9. Success output!
	c.JSON(http.StatusCreated, gin.H{
		"message":         "Leave application successfully submitted for routing approval!",
		"calculated_days": calculatedDays,
		"request_details": newLeave,
	})
}

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
