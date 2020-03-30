package heartbeat

import (
	"fmt"
)

// EntityType defines the type of an entity
type EntityType int

const (
	// FileType represents a file entity
	FileType EntityType = iota
	// DomainType represents a domain entity
	DomainType
	// AppType represents an app entity
	AppType
)

const (
	fileTypeString   = "file"
	domainTypeString = "domain"
	appTypeString    = "app"
)

// Unmarshal is a method to implement json.Unmarshaler interface
func (t *EntityType) UnmarshalJSON(v []byte) error {
	switch string(v) {
	case fileTypeString:
		*t = FileType
	case domainTypeString:
		*t = DomainType
	case appTypeString:
		*t = AppType
	default:
		return fmt.Errorf("unsupported entity type: %q", v)
	}

	return nil
}

// MarshalJSON is a method to implement json.Marshaler interface
func (c EntityType) MarshalJSON() ([]byte, error) {
	s := c.String()
	if s == "" {
		return nil, fmt.Errorf("unsupported entity type %v", c)
	}

	return []byte(s), nil
}

// String is a method to implement fmt.Stringer interface
func (c EntityType) String() string {
	switch c {
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
