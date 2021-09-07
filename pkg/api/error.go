package api

// Err represents a general api error.
type Err string

// Error method to implement error interface.
func (e Err) Error() string {
	return string(e)
}

// ErrAuth represents an authentication error.
type ErrAuth string

// Error method to implement error interface.
func (e ErrAuth) Error() string {
	return string(e)
}

// ErrBadRequest represents a 400 response from the API.
type ErrBadRequest string

// Error method to implement error interface.
func (e ErrBadRequest) Error() string {
	return string(e)
}
