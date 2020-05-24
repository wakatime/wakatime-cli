package heartbeat

// Heartbeat is a structure representing activity for a user on a some entity.
type Heartbeat struct {
	Branch         *string    `json:"branch"`
	Category       Category   `json:"category"`
	CursorPosition *int       `json:"cursorpos"`
	Dependencies   []string   `json:"dependencies"`
	Entity         string     `json:"entity"`
	EntityType     EntityType `json:"type"`
	IsWrite        *bool      `json:"is_write"`
	Language       *string    `json:"language"`
	LineNumber     *int       `json:"lineno"`
	Lines          *int       `json:"lines"`
	Project        *string    `json:"project"`
	Time           float64    `json:"time"`
	UserAgent      string     `json:"user_agent"`
}

// Bool returns a pointer to the bool value passed in.
func Bool(v bool) *bool {
	return &v
}

// Int returns a pointer to the int value passed in.
func Int(v int) *int {
	return &v
}

// String returns a pointer to the string value passed in.
func String(v string) *string {
	return &v
}

// Result represents a response from the wakatime api.
type Result struct {
	Errors    []string
	Status    int
	Heartbeat Heartbeat
}
