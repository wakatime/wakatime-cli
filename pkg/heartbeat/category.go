package heartbeat

import (
	"context"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/wakatime/wakatime-cli/pkg/log"
)

// Category represents a heartbeat category.
type Category int

const (
	// UndefinedCategory will default to Coding on the API server if no category is detected client side.
	UndefinedCategory Category = iota
	// CodingCategory means user is currently coding.
	CodingCategory
	// AdvisingCategory means user is currently adivising.
	AdvisingCategory
	// AICodingCategory means user is currently coding using an AI code gen tool.
	AICodingCategory
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
	// NotesCategory means user is currently taking notes.
	NotesCategory
	// PlanningCategory means user is currently planning.
	PlanningCategory
	// ResearchingCategory means user is currently researching.
	ResearchingCategory
	// RunningTestsCategory means user is currently running tests.
	RunningTestsCategory
	// SupportingCategory means user is doing customer support.
	SupportingCategory
	// TranslatingCategory means user is currently translating.
	TranslatingCategory
	// WritingDocsCategory means user is currently writing docs.
	WritingDocsCategory
	// WritingTestsCategory means user is currently writing tests.
	WritingTestsCategory
)

const (
	advisingCategoryString      = "advising"
	aiCodingCategoryString      = "ai coding"
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
	notesCategoryString         = "notes"
	planningCategoryString      = "planning"
	researchingCategoryString   = "researching"
	runningTestsCategoryString  = "running tests"
	supportingCategoryString    = "supporting"
	translatingCategoryString   = "translating"
	writingDocsCategoryString   = "writing docs"
	writingTestsCategoryString  = "writing tests"
)

// WithCategory initializes and returns a heartbeat handle option, which
// can be used in a heartbeat processing pipeline to detect an entity's category.
func WithCategory() HandleOption {
	return func(next Handle) Handle {
		return func(ctx context.Context, hh []Heartbeat) ([]Result, error) {
			logger := log.Extract(ctx)
			logger.Debugln("execute heartbeat category detection")

			for n, h := range hh {
				// remove category coding if it was set by the user to avoid sending it to the API
				if h.Category == codingCategoryString {
					hh[n].Category = ""
					continue
				}

				if h.Category != "" {
					logger.Debugf("heartbeat %s already has category %s, skipping detection", h.Entity, h.Category)
					continue
				}

				if h.EntityType != FileType {
					continue
				}

				hh[n].Category = DetectCategory(h).String()
			}

			return next(ctx, hh)
		}
	}
}

// DetectCategory accepts a heartbeat and detects it's category.
func DetectCategory(h Heartbeat) Category {
	file := strings.ToLower(h.Entity)

	if strings.HasSuffix(file, "_test.go") {
		return WritingTestsCategory
	}

	if slices.ContainsFunc([]string{"/tests/", "/test/", "/testdata/", "/spec/", "/specs/"}, func(s string) bool {
		return strings.Contains(file, s)
	}) {
		return WritingTestsCategory
	}

	var testFileRegex = regexp.MustCompile(`(?i).*\.(test|spec)\.[^.]+$`)
	if testFileRegex.MatchString(file) {
		return WritingTestsCategory
	}

	if strings.HasSuffix(file, ".md") || strings.HasSuffix(file, ".mdx") {
		return WritingDocsCategory
	}

	return UndefinedCategory
}

// ParseCategory parses a category from a string.
func ParseCategory(s string) (Category, error) {
	switch s {
	case advisingCategoryString:
		return AdvisingCategory, nil
	case aiCodingCategoryString:
		return AICodingCategory, nil
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
	case notesCategoryString:
		return NotesCategory, nil
	case planningCategoryString:
		return PlanningCategory, nil
	case researchingCategoryString:
		return ResearchingCategory, nil
	case runningTestsCategoryString:
		return RunningTestsCategory, nil
	case supportingCategoryString:
		return SupportingCategory, nil
	case translatingCategoryString:
		return TranslatingCategory, nil
	case "", "null":
		return UndefinedCategory, nil
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
	if c == UndefinedCategory {
		return []byte(`null`), nil
	}

	s := c.String()
	if s == "" {
		return nil, fmt.Errorf("invalid category %v", c)
	}

	return []byte(`"` + s + `"`), nil
}

// String implements fmt.Stringer interface.
func (c Category) String() string {
	switch c {
	case AdvisingCategory:
		return advisingCategoryString
	case AICodingCategory:
		return aiCodingCategoryString
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
	case NotesCategory:
		return notesCategoryString
	case PlanningCategory:
		return planningCategoryString
	case ResearchingCategory:
		return researchingCategoryString
	case RunningTestsCategory:
		return runningTestsCategoryString
	case SupportingCategory:
		return supportingCategoryString
	case TranslatingCategory:
		return translatingCategoryString
	case WritingDocsCategory:
		return writingDocsCategoryString
	case WritingTestsCategory:
		return writingTestsCategoryString
	case UndefinedCategory:
		fallthrough
	default:
		return ""
	}
}
