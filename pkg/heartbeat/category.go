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
	// CommunicatingCategory means user is currently chatting.
	CommunicatingCategory
	// DebuggingCategory means user is currently debugging.
	DebuggingCategory
	// DesigningCategory means user is currently designing.
	DesigningCategory
	// IndexingCategory means user is currently indexing.
	IndexingCategory
	// LearningCategory means user is currently learning.
	LearningCategory
	// ManualTestingCategory means user is currently manual testing.
	ManualTestingCategory
	// MeetingCategory means user is currently meeting.
	MeetingCategory
	// PlanningCategory means user is currently planning.
	PlanningCategory
	// ResearchingCategory means user is currently researching.
	ResearchingCategory
	// RunningTestsCategory means user is currently running tests.
	RunningTestsCategory
	// TranslatingCategory means user is currently translating.
	TranslatingCategory
	// WritingDocsCategory means user is currently writing docs.
	WritingDocsCategory
	// WritingTestsCategory means user is currently writing tests.
	WritingTestsCategory
)

const (
	browsingCategoryString      = "browsing"
	buildingCategoryString      = "building"
	codeReviewingCategoryString = "code reviewing"
	codingCategoryString        = "coding"
	communicatingCategoryString = "communicating"
	debuggingCategoryString     = "debugging"
	designingCategoryString     = "designing"
	indexingCategoryString      = "indexing"
	learningCategoryString      = "learning"
	manualTestingCategoryString = "manual testing"
	meetingCategoryString       = "meeting"
	planningCategoryString      = "planning"
	researchingCategoryString   = "researching"
	runningTestsCategoryString  = "running tests"
	translatingCategoryString   = "translating"
	writingDocsCategoryString   = "writing docs"
	writingTestsCategoryString  = "writing tests"
)

// ParseCategory parses a category from a string.
func ParseCategory(s string) (Category, error) {
	switch s {
	case browsingCategoryString:
		return BrowsingCategory, nil
	case buildingCategoryString:
		return BuildingCategory, nil
	case codeReviewingCategoryString:
		return CodeReviewingCategory, nil
	case codingCategoryString:
		return CodingCategory, nil
	case communicatingCategoryString:
		return CommunicatingCategory, nil
	case debuggingCategoryString:
		return DebuggingCategory, nil
	case designingCategoryString:
		return DesigningCategory, nil
	case indexingCategoryString:
		return IndexingCategory, nil
	case learningCategoryString:
		return LearningCategory, nil
	case manualTestingCategoryString:
		return ManualTestingCategory, nil
	case meetingCategoryString:
		return MeetingCategory, nil
	case planningCategoryString:
		return PlanningCategory, nil
	case researchingCategoryString:
		return ResearchingCategory, nil
	case runningTestsCategoryString:
		return RunningTestsCategory, nil
	case translatingCategoryString:
		return TranslatingCategory, nil
	case writingDocsCategoryString:
		return WritingDocsCategory, nil
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
	case BrowsingCategory:
		return browsingCategoryString
	case BuildingCategory:
		return buildingCategoryString
	case CodeReviewingCategory:
		return codeReviewingCategoryString
	case CodingCategory:
		return codingCategoryString
	case CommunicatingCategory:
		return communicatingCategoryString
	case DebuggingCategory:
		return debuggingCategoryString
	case DesigningCategory:
		return designingCategoryString
	case IndexingCategory:
		return indexingCategoryString
	case LearningCategory:
		return learningCategoryString
	case ManualTestingCategory:
		return manualTestingCategoryString
	case MeetingCategory:
		return meetingCategoryString
	case PlanningCategory:
		return planningCategoryString
	case ResearchingCategory:
		return researchingCategoryString
	case RunningTestsCategory:
		return runningTestsCategoryString
	case TranslatingCategory:
		return translatingCategoryString
	case WritingDocsCategory:
		return writingDocsCategoryString
	case WritingTestsCategory:
		return writingTestsCategoryString
	default:
		return ""
	}
}
