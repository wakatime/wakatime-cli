package diagnostic

import "fmt"

// Type is a type of diagnostic.
type Type int

const (
	// TypeUnknown designates an unknown type of diagnostic.
	TypeUnknown Type = iota
	// TypeError designates a Diagnostic for error.
	TypeError
	// TypeLogs designates a Diagnostic for logs.
	TypeLogs
	// TypeStack designates a Diagnostic for stack trace.
	TypeStack
)

// Diagnostic contains diagnostic info.
type Diagnostic struct {
	Type  Type
	Value string
}

// Error creates a new instance of Diagnostic of type TypeError.
func Error(err any) Diagnostic {
	return Diagnostic{
		Type:  TypeError,
		Value: fmt.Sprintf("%v", err),
	}
}

// Logs creates a new instance of Diagnostic of type TypeLogs.
func Logs(logs string) Diagnostic {
	return Diagnostic{
		Type:  TypeLogs,
		Value: logs,
	}
}

// Stack creates a new instance of Diagnostic of type TypeStack.
func Stack(stack string) Diagnostic {
	return Diagnostic{
		Type:  TypeStack,
		Value: stack,
	}
}
