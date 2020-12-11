package deps

// State is a token parsing state.
type State int

const (
	// StateUnknown represents a unknown token parsing state.
	StateUnknown State = iota
	// StateImport means we are in import section during token parsing.
	StateImport
)
