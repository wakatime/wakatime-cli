package heartbeat

import (
	"fmt"
	"strings"
)

// EntityType defines the type of an entity.
type EntityType int

const (
	// FileType represents a file entity.
	FileType EntityType = iota
	// DomainType represents a domain entity.
	DomainType
	// URLType represents a url entity without the url params.
	URLType
	// EventType represents a meeting or calendar event.
	EventType
	// AppType represents an app entity.
	AppType
)

const (
	fileTypeString   = "file"
	domainTypeString = "domain"
	urlTypeString    = "url"
	eventTypeString  = "event"
	appTypeString    = "app"
)

// ParseEntityType parses an entity type from a string.
func ParseEntityType(s string) (EntityType, error) {
	switch s {
	case fileTypeString:
		return FileType, nil
	case domainTypeString:
		return DomainType, nil
	case urlTypeString:
		return URLType, nil
	case eventTypeString:
		return EventType, nil
	case appTypeString:
		return AppType, nil
	default:
		return 0, fmt.Errorf("invalid entity type %q", s)
	}
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (t *EntityType) UnmarshalJSON(v []byte) error {
	trimmed := strings.Trim(string(v), "\"")

	entityType, err := ParseEntityType(trimmed)
	if err != nil {
		return err
	}

	*t = entityType

	return nil
}

// MarshalJSON implements json.Marshaler interface.
func (t EntityType) MarshalJSON() ([]byte, error) {
	s := t.String()
	if s == "" {
		return nil, fmt.Errorf("invalid entity type %v", t)
	}

	return []byte(`"` + s + `"`), nil
}

// String implements fmt.Stringer interface.
func (t EntityType) String() string {
	switch t {
	case FileType:
		return fileTypeString
	case DomainType:
		return domainTypeString
	case URLType:
		return urlTypeString
	case EventType:
		return eventTypeString
	case AppType:
		return appTypeString
	default:
		return ""
	}
}
