package models

import (
	"time"

	"github.com/google/uuid"
)

type LeaveRequest struct {
	ID              uuid.UUID  `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID          uuid.UUID  `gorm:"type:uuid;column:user_id;not null"`
	User            User       `gorm:"foreignKey:UserID"`
	LeaveTypeID     uuid.UUID  `gorm:"type:uuid;column:leave_type_id;not null"`
	LeaveType       LeaveType  `gorm:"foreignKey:LeaveTypeID"`
	StartDate       time.Time  `gorm:"type:date;column:start_date;not null"`
	EndDate         time.Time  `gorm:"type:date;column:end_date;not null"`
	CalculatedDays  float64    `gorm:"column:calculated_days;not null"` // Matches NUMERIC(5,1)
	Reason          string     `gorm:"type:text;column:reason;not null"`
	Status          string     `gorm:"type:varchar(20);column:status;default:'PENDING';not null"`
	ApproverUserID  *uuid.UUID `gorm:"type:uuid;column:approver_user_id"`
	Approver        *User      `gorm:"foreignKey:ApproverUserID"`
	ApproverRemarks string     `gorm:"type:text;column:approver_remarks"`
	AppliedAt       time.Time  `gorm:"column:applied_at;default:now()"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;default:now()"`
}
