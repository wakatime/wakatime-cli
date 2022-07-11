package api

import (
	"encoding/base64"
	"errors"
	"fmt"
)

// BasicAuth contains authentication data.
type BasicAuth struct {
	User   string
	Secret string
}

// HeaderValue returns the value for Authorization header.
func (a BasicAuth) HeaderValue() (string, error) {
	if a.User == "" && a.Secret == "" {
		return "", errors.New("secret unset")
	}

	if a.Secret == "" {
		return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(
			[]byte(a.User+":"),
		)), nil
	}

	if a.User == "" {
		return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(
			[]byte(a.Secret),
		)), nil
	}

	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString(
		[]byte(fmt.Sprintf("%s:%s", a.User, a.Secret)),
	)), nil
}
