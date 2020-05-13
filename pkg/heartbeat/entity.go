package heartbeat

import (
	"fmt"
)

// EntityType defines the type of an entity.
type EntityType int

const (
	// FileType represents a file entity.
	FileType EntityType = iota
	// DomainType represents a domain entity.
	DomainType
	// AppType represents an app entity.
	AppType
)

const (
	fileTypeString   = "file"
	domainTypeString = "domain"
	appTypeString    = "app"
)

// UnmarshalJSON is a method to implement json.Unmarshaler interface.
func (t *EntityType) UnmarshalJSON(v []byte) error {
	switch string(v) {
	case `"` + fileTypeString + `"`:
		*t = FileType
	case `"` + domainTypeString + `"`:
		*t = DomainType
	case `"` + appTypeString + `"`:
		*t = AppType
	default:
		return fmt.Errorf("unsupported entity type: %q", v)
	}

	return nil
}

// MarshalJSON is a method to implement json.Marshaler interface.
func (t EntityType) MarshalJSON() ([]byte, error) {
	s := t.String()
	if s == "" {
		return nil, fmt.Errorf("unsupported entity type %v", t)
	}

	return []byte(`"` + s + `"`), nil
}

// String is a method to implement fmt.Stringer interface.
func (t EntityType) String() string {
	switch t {
	case FileType:
		return fileTypeString
	case DomainType:
		return domainTypeString
	case AppType:
		return appTypeString
	default:
		return ""
	}
}
