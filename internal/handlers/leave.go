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

	if calculatedDays == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Leave duration calculated to 0 days (e.g., applied only on weekends/holidays)"})
		return
	}

	// 6.5 Verify the user has enough balance in their ledger for the current year
	currentYear := time.Now().Year()
	var leaveBalance models.LeaveBalance

	if err := config.DB.Where("user_id = ? AND leave_type_id = ? AND year = ?", user.ID, req.LeaveTypeID, currentYear).First(&leaveBalance).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No leave balance allocation found for this year. Please contact HR."})
		return
	}

	if float64(calculatedDays) > leaveBalance.RemainingDays {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":          "Insufficient leave balance",
			"requested_days": calculatedDays,
			"remaining_days": leaveBalance.RemainingDays,
		})
		return
	}

	// 7. Build the structural database model object
	newLeave := models.LeaveRequest{
		ID:             uuid.New(),
		UserID:         user.ID,
		LeaveTypeID:    req.LeaveTypeID,
		StartDate:      start,
		EndDate:        end,
		CalculatedDays: float64(calculatedDays),
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

type UpdateLeaveStatusRequest struct {
	Status  string `json:"status" binding:"required"` // 'APPROVED' or 'REJECTED'
	Remarks string `json:"remarks"`
}

// UpdateLeaveStatus handles HOD/Admin approvals and balance deductions
func UpdateLeaveStatus(c *gin.Context) {
	leaveID := c.Param("id") // The UUID of the specific leave request in the URL
	var req UpdateLeaveStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	// 1. Validate the status
	if req.Status != "APPROVED" && req.Status != "REJECTED" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status must be APPROVED or REJECTED"})
		return
	}

	// 2. Identify the Approver (The HOD/Admin making the request)
	clerkID, _ := c.Get("user_id")
	var approver models.User
	if err := config.DB.Where("clerk_id = ?", clerkID.(string)).First(&approver).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Approver profile not found"})
		return
	}

	// 3. Start a Database Transaction (Crucial for data integrity)
	tx := config.DB.Begin()

	// 4. Fetch the existing Leave Request
	var leaveRequest models.LeaveRequest
	if err := tx.First(&leaveRequest, "id = ?", leaveID).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Leave request not found"})
		return
	}

	// 5. Check if it's already been processed
	if leaveRequest.Status != "PENDING" {
		tx.Rollback()
		c.JSON(http.StatusBadRequest, gin.H{"error": "This leave request has already been processed"})
		return
	}

	// 6. If APPROVED, deduct the balance
	if req.Status == "APPROVED" {
		currentYear := time.Now().Year()
		var balance models.LeaveBalance

		// Find the applicant's balance for this specific leave type
		if err := tx.Where("user_id = ? AND leave_type_id = ? AND year = ?", leaveRequest.UserID, leaveRequest.LeaveTypeID, currentYear).First(&balance).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not find applicant's leave balance record"})
			return
		}

		// Double-check they still have enough days (in case of concurrent approvals)
		if leaveRequest.CalculatedDays > balance.RemainingDays {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Applicant no longer has enough balance for this approval"})
			return
		}

		// Do the math: Subtract from remaining, add to used
		balance.RemainingDays -= leaveRequest.CalculatedDays
		balance.UsedDays += leaveRequest.CalculatedDays

		// Save the updated balance
		if err := tx.Save(&balance).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update leave balance ledger"})
			return
		}
	}

	// 7. Update the Leave Request record itself
	leaveRequest.Status = req.Status
	leaveRequest.ApproverUserID = &approver.ID
	leaveRequest.ApproverRemarks = req.Remarks
	leaveRequest.UpdatedAt = time.Now()

	if err := tx.Save(&leaveRequest).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update leave status"})
		return
	}

	// 8. Commit the transaction! (If we get here, both the deduction and status update were successful)
	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Leave request successfully " + req.Status,
		"status":  req.Status,
	})
}

// GetPendingLeaves fetches requests requiring approval by the current HOD
func GetPendingLeaves(c *gin.Context) {
	clerkID, _ := c.Get("user_id")
	var approver models.User

	if err := config.DB.Where("clerk_id = ?", clerkID.(string)).First(&approver).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Approver profile not found"})
		return
	}

	var pendingRequests []models.LeaveRequest

	// Preload the User and LeaveType data so the frontend can display names!
	query := config.DB.Preload("User").Preload("LeaveType").Where("leave_requests.status = ?", "PENDING")

	// If the approver belongs to a specific department, only show them requests from their faculty
	if approver.DepartmentID != nil {
		query = query.Joins("JOIN users ON users.id = leave_requests.user_id").
			Where("users.department_id = ?", approver.DepartmentID)
	}

	if err := query.Find(&pendingRequests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inbox"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": pendingRequests})
}

// GetMyLeaves fetches the personal leave history for the logged-in user
func GetMyLeaves(c *gin.Context) {
	clerkID, _ := c.Get("user_id")
	var user models.User
	config.DB.Where("clerk_id = ?", clerkID.(string)).First(&user)

	var myLeaves []models.LeaveRequest

	// Order by most recently applied first
	if err := config.DB.Preload("LeaveType").Preload("Approver").Where("user_id = ?", user.ID).Order("applied_at desc").Find(&myLeaves).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": myLeaves})
}

// GetMyBalances fetches the remaining leave quotas for the current year
func GetMyBalances(c *gin.Context) {
	clerkID, _ := c.Get("user_id")
	var user models.User
	config.DB.Where("clerk_id = ?", clerkID.(string)).First(&user)

	var balances []models.LeaveBalance
	currentYear := time.Now().Year()

	if err := config.DB.Where("user_id = ? AND year = ?", user.ID, currentYear).Find(&balances).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch balances"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": balances})
}
