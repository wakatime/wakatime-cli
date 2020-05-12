package summary

import (
	"time"
)

// Summary represents a summary of tracked working time.
type Summary struct {
	Category   string
	GrandTotal string
	Date       time.Time
}
