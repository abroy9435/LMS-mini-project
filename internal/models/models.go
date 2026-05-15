package models

import (
	"time"

	"github.com/google/uuid"
)

// Role defines the hierarchy levels (HOD, COE, Registrar)
type Role struct {
	ID             uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name           string    `gorm:"unique;not null" json:"name"`
	HierarchyLevel int       `gorm:"not null" json:"hierarchy_level"`
	CreatedAt      time.Time `json:"created_at"`
}

// Department defines the university academic/admin units
type Department struct {
	ID        uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name      string     `gorm:"unique;not null" json:"name"`
	HODUserID *uuid.UUID `gorm:"type:uuid" json:"hod_user_id,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// User is the core employee profile mapped to Supabase Auth
type User struct {
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey" json:"id"` // Matches auth.users
	EmployeeID   string     `gorm:"unique;not null" json:"employee_id"`
	FirstName    string     `gorm:"not null" json:"first_name"`
	LastName     string     `gorm:"not null" json:"last_name"`
	Email        string     `gorm:"unique;not null" json:"email"`
	DepartmentID *uuid.UUID `gorm:"type:uuid" json:"department_id,omitempty"`
	RoleID       uuid.UUID  `gorm:"type:uuid;not null" json:"role_id"`
	CreatedAt    time.Time  `json:"created_at"`
}

// LeaveType defines rules like Casual Leave vs Earned Leave
type LeaveType struct {
	ID             uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	Name           string    `gorm:"unique;not null" json:"name"`
	DefaultDays    float64   `gorm:"type:numeric(5,1);not null" json:"default_days"`
	RequiresRoleID uuid.UUID `gorm:"type:uuid;not null" json:"requires_role_id"`
	IsCarryForward bool      `gorm:"default:false" json:"is_carry_forward"`
}

// LeaveBalance is the ledger for each user's remaining quota
type LeaveBalance struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	LeaveTypeID   uuid.UUID `gorm:"type:uuid;not null" json:"leave_type_id"`
	Year          int       `gorm:"not null" json:"year"`
	AllocatedDays float64   `gorm:"type:numeric(5,1);not null" json:"allocated_days"`
	UsedDays      float64   `gorm:"type:numeric(5,1);default:0" json:"used_days"`
	RemainingDays float64   `gorm:"type:numeric(5,1);not null" json:"remaining_days"`
}

// LeaveRequest handles the actual application transactions
type LeaveRequest struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id"`
	UserID          uuid.UUID  `gorm:"type:uuid;not null" json:"user_id"`
	LeaveTypeID     uuid.UUID  `gorm:"type:uuid;not null" json:"leave_type_id"`
	StartDate       time.Time  `gorm:"type:date;not null" json:"start_date"`
	EndDate         time.Time  `gorm:"type:date;not null" json:"end_date"`
	CalculatedDays  float64    `gorm:"type:numeric(5,1);not null" json:"calculated_days"`
	Reason          string     `gorm:"not null" json:"reason"`
	Status          string     `gorm:"default:'PENDING'" json:"status"`
	ApproverUserID  *uuid.UUID `gorm:"type:uuid" json:"approver_user_id,omitempty"`
	ApproverRemarks string     `json:"approver_remarks,omitempty"`
	AppliedAt       time.Time  `json:"applied_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}
