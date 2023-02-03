package output

import (
	"fmt"
)

// Output represents the output format.
type Output int

const (
	// TextOutput means output will be in text format. This is the default value.
	TextOutput Output = iota
	// JSONOutput means output will be in JSON format.
	JSONOutput
	// RawJSONOutput means output will be in raw JSON format.
	RawJSONOutput
)

const (
	textOutputString    = "text"
	jsonOutputString    = "json"
	jsonRawOutputString = "raw-json"
)

// Parse parses an output from a string.
func Parse(s string) (Output, error) {
	switch s {
	case textOutputString:
		return TextOutput, nil
	case jsonOutputString:
		return JSONOutput, nil
	case jsonRawOutputString:
		return RawJSONOutput, nil
	default:
		return TextOutput, fmt.Errorf("invalid output %q", s)
	}
}

// String returns the string representation of an output.
func (o Output) String() string {
	switch o {
	case TextOutput:
		return textOutputString
	case JSONOutput:
		return jsonOutputString
	case RawJSONOutput:
		return jsonRawOutputString
	default:
		return ""
	}
}
