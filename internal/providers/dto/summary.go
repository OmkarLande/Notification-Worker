package dto

import "time"

// MonthlySummaryData holds aggregated monthly statistics for a single user.
type MonthlySummaryData struct {
	UserID  int
	Month   time.Month
	Year    int
	Summary string
}
