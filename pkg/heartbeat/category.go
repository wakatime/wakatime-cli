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

// UnmarshalJSON implements json.Unmarshaler interface.
func (c *Category) UnmarshalJSON(v []byte) error {
	switch strings.Trim(string(v), "\"") {
	case codingCategoryString:
		*c = CodingCategory
	case browsingCategoryString:
		*c = BrowsingCategory
	case buildingCategoryString:
		*c = BuildingCategory
	case codeReviewingCategoryString:
		*c = CodeReviewingCategory
	case debuggingCategoryString:
		*c = DebuggingCategory
	case designingCategoryString:
		*c = DesigningCategory
	case indexingCategoryString:
		*c = IndexingCategory
	case manualTestingCategoryString:
		*c = ManualTestingCategory
	case runningTestsCategoryString:
		*c = RunningTestsCategory
	case writingTestsCategoryString:
		*c = WritingTestsCategory
	default:
		return fmt.Errorf("unsupported category %q", v)
	}

	return nil
}

// MarshalJSON implements json.Marshaler interface.
func (c Category) MarshalJSON() ([]byte, error) {
	s := c.String()
	if s == "" {
		return nil, fmt.Errorf("unsupported category %v", c)
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
