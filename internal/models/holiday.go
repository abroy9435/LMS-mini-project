package models

import (
	"time"

	"github.com/google/uuid"
)

type Holiday struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	HolidayDate  time.Time `gorm:"type:date;column:holiday_date;unique;not null"`
	Name         string    `gorm:"type:varchar(100);column:name;not null"`
	IsRestricted bool      `gorm:"column:is_restricted;default:false"`
}
