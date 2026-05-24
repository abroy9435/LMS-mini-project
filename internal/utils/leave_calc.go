package utils

import (
	"LMS-mini-project-backend/internal/config"
	"LMS-mini-project-backend/internal/models"
	"time"
)

// FetchHolidaysFromDB reads all active gazetted holiday rows from your PostgreSQL table
func FetchHolidaysFromDB() map[string]bool {
	holidayMap := make(map[string]bool)
	var dbHolidays []models.Holiday

	// Query database where is_restricted = false (Gazetted university holidays)
	err := config.DB.Where("is_restricted = ?", false).Find(&dbHolidays).Error
	if err != nil {
		// Fallback safely to an empty map if database read fails
		return holidayMap
	}

	// Turn the slice into a map for quick O(1) performance lookup inside our loop
	for _, h := range dbHolidays {
		dateStr := h.HolidayDate.Format("2006-01-02")
		holidayMap[dateStr] = true
	}

	return holidayMap
}

// CalculateNetLeaveDays steps through dates and deducts matching rules
func CalculateNetLeaveDays(startDate time.Time, endDate time.Time, leaveTypeName string) float64 {
	if endDate.Before(startDate) {
		return 0.0
	}

	// Pull fresh records directly from your SQL table setup
	holidays := FetchHolidaysFromDB()
	var netDays float64 = 0.0

	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {

		// Rules condition: Medical Leave (ML) shouldn't skip weekends/holidays
		if leaveTypeName == "Medical Leave (ML)" {
			netDays++
			continue
		}

		// For standard Casual Leave (CL), filter out weekends
		weekday := d.Weekday()
		if weekday == time.Saturday || weekday == time.Sunday {
			continue
		}

		// Filter out holiday matches from your database seeding script
		dateStr := d.Format("2006-01-02")
		if holidays[dateStr] {
			continue
		}

		netDays++
	}

	return netDays
}
