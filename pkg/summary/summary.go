package summary

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Category represents tracked working time of a specific category.
type Category struct {
	Category string
	Total    string
}

// Summary represents the tracked working time for a single day.
type Summary struct {
	Date       time.Time
	Total      string
	ByCategory []Category
}
