package heartbeat

import (
	"fmt"
	"strings"
)

// Category represents a heartbeat category.
type Category int

const (
	// CodingCategory means user is currently coding. This is the default value.
	CodingCategory Category = iota
	// BrowsingCategory means user is currently browsing.
	BrowsingCategory
	// BuildingCategory means user is currently building.
	BuildingCategory
	// CodeReviewingCategory means user is currently reviewing code.
	CodeReviewingCategory
	// DebuggingCategory means user is currently debugging.
	DebuggingCategory
	// DesigningCategory means user is currently designing.
	DesigningCategory
	// IndexingCategory means user is currently indexing.
	IndexingCategory
	// ManualTestingCategory means user is currently manual testing.
	ManualTestingCategory
	// RunningTestsCategory means user is currently running tests.
	RunningTestsCategory
	// WritingTestsCategory means user is currently writing tests.
	WritingTestsCategory
)

const (
	codingCategoryString        = "coding"
	browsingCategoryString      = "browsing"
	buildingCategoryString      = "building"
	codeReviewingCategoryString = "code reviewing"
	debuggingCategoryString     = "debugging"
	designingCategoryString     = "designing"
	indexingCategoryString      = "indexing"
	manualTestingCategoryString = "manual testing"
	runningTestsCategoryString  = "running tests"
	writingTestsCategoryString  = "writing tests"
)

// ParseCategory parses a category from a string.
func ParseCategory(s string) (Category, error) {
	switch s {
	case codingCategoryString:
		return CodingCategory, nil
	case browsingCategoryString:
		return BrowsingCategory, nil
	case buildingCategoryString:
		return BuildingCategory, nil
	case codeReviewingCategoryString:
		return CodeReviewingCategory, nil
	case debuggingCategoryString:
		return DebuggingCategory, nil
	case designingCategoryString:
		return DesigningCategory, nil
	case indexingCategoryString:
		return IndexingCategory, nil
	case manualTestingCategoryString:
		return ManualTestingCategory, nil
	case runningTestsCategoryString:
		return RunningTestsCategory, nil
	case writingTestsCategoryString:
		return WritingTestsCategory, nil
	default:
		return 0, fmt.Errorf("invalid category %q", s)
	}
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (c *Category) UnmarshalJSON(v []byte) error {
	trimmed := strings.Trim(string(v), "\"")

	category, err := ParseCategory(trimmed)
	if err != nil {
		return err
	}

	*c = category

	return nil
}

// MarshalJSON implements json.Marshaler interface.
func (c Category) MarshalJSON() ([]byte, error) {
	s := c.String()
	if s == "" {
		return nil, fmt.Errorf("invalid category %v", c)
	}

	return []byte(`"` + s + `"`), nil
}

// String implements fmt.Stringer interface.
func (c Category) String() string {
	switch c {
	case CodingCategory:
		return codingCategoryString
	case BrowsingCategory:
		return browsingCategoryString
	case BuildingCategory:
		return buildingCategoryString
	case CodeReviewingCategory:
		return codeReviewingCategoryString
	case DebuggingCategory:
		return debuggingCategoryString
	case DesigningCategory:
		return designingCategoryString
	case IndexingCategory:
		return indexingCategoryString
	case ManualTestingCategory:
		return manualTestingCategoryString
	case RunningTestsCategory:
		return runningTestsCategoryString
	case WritingTestsCategory:
		return writingTestsCategoryString
	default:
		return ""
	}
}
