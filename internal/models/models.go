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
	ID           uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	ClerkID      string     `gorm:"type:varchar(255);column:clerk_id;unique"`
	EmployeeID   string     `gorm:"type:varchar(50);column:employee_id;unique;not null"`
	FirstName    string     `gorm:"type:varchar(50);column:first_name;not null"`
	LastName     string     `gorm:"type:varchar(50);column:last_name;not null"`
	Email        string     `gorm:"type:varchar(255);column:email;unique;not null"`
	DepartmentID *uuid.UUID `gorm:"type:uuid;column:department_id"`
	RoleID       uuid.UUID  `gorm:"type:uuid;column:role_id;not null"`
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
