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
	switch c {
	case FileType:
		return []byte(fileTypeString), nil
	case DomainType:
		return []byte(domainTypeString), nil
	case AppType:
		return []byte(appTypeString), nil
	default:
		return nil, fmt.Errorf("unsupported entity type %v", c)
	}
}
