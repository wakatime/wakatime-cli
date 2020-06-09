package project

// Err handles a custom error when finding for project and branch names.
type Err string

// Error implements error interface.
func (e Err) Error() string {
	return string(e)
}
