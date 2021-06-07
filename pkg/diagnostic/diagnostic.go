package diagnostic

// Type is a type of diagnostic.
type Type int

const (
	// TypeUnknown designates an unknown type of diagnostic.
	TypeUnknown Type = iota
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
