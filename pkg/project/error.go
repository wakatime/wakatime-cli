package project

// ErrProject handles a custom error when finding for project and branch names.
type ErrProject string

// Error implements error interface.
func (e ErrProject) Error() string {
	return string(e)
}
