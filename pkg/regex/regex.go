package regex

import (
	"fmt"
	"regexp"

	"github.com/wakatime/wakatime-cli/pkg/log"

	"github.com/dlclark/regexp2"
)

// Regex interface to use regexp.Regexp and regexp2.Regexp interchangeably.
type Regex interface {
	FindStringSubmatch(s string) []string
	MatchString(s string) bool
	String() string
}

// Compile compiles via standard regexp package. Upon failure, it will also
// attempt compilation via regexp2 package.
func Compile(s string) (Regex, error) {
	r, err := regexp.Compile(s)
	if err == nil {
		return r, nil
	}

	r2, err := regexp2.Compile(s, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to compile regex %q: %s", s, err)
	}

	return &regexp2Wrap{
		rgx: r2,
	}, nil
}

// MustCompile compiles via standard regexp package. Upon failure, it will also
// attempt compilation via regexp2 package.
// Will panic, if both compilation attempts failed.
func MustCompile(s string) Regex {
	r, err := Compile(s)
	if err != nil {
		panic(err)
	}

	return r
}

// regexp2Wrap is a wrapper around github.com/dlclark/regexp2.Regexp, which conforms
// to regexp.Regexp interface. Only supports a subset of methods.
type regexp2Wrap struct {
	rgx *regexp2.Regexp
}

// FindStringSubmatch returns a slice of strings holding the text of the leftmost
// match of the regular expression in s and the matches, if any, of its
// subexpressions, as defined by the 'Submatch' description in the package comment.
// A return value of nil indicates no match.
func (re *regexp2Wrap) FindStringSubmatch(s string) []string {
	m, err := re.rgx.FindStringMatch(s)
	if err != nil {
		log.Warnf("failed to find string match %q: %s", s, err)
		return nil
	}

	if m == nil {
		return nil
	}

	var result []string

	for _, g := range m.Groups() {
		for _, c := range g.Captures {
			result = append(result, c.String())
		}
	}

	return result
}

// MatchString reports whether the string s contains any match of the regular
// expression re.
func (re *regexp2Wrap) MatchString(s string) bool {
	matched, err := re.rgx.MatchString(s)
	if err != nil {
		log.Warnf("failed to match string %q: %s", s, err)
		return false
	}

	return matched
}

// String returns the source text used to compile the regular expression.
func (re *regexp2Wrap) String() string {
	return re.rgx.String()
}
