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

// ErrRequest represents a request failure, where no response was received from the api.
type ErrRequest string

// Error method to implement error interface.
func (e ErrRequest) Error() string {
	return string(e)
}
