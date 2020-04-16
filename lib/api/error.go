package api

import (
	"fmt"
)

// Err is an error type for general api errors
type Err struct {
	s string
}

// Error returns error info
func (e Err) Error() string {
	return e.s
}

// NewErr creates a new Err
func NewErr(text string, args ...interface{}) error {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}

	return Err{s: text}
}

// ErrAuth is an error type for authentication errors
type ErrAuth struct {
	s string
}

// Error returns error info
func (e ErrAuth) Error() string {
	return e.s
}

// NewErrAuth creates a new ErrAuth
func NewErrAuth(text string, args ...interface{}) error {
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}

	return ErrAuth{s: text}
}
