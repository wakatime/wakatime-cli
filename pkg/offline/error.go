package offline

import (
	"fmt"

	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"
)

// ErrOpenDB is an error returned when the database cannot be opened.
type ErrOpenDB struct {
	Err error
}

var _ wakaerror.Error = ErrOpenDB{}

// Error method to implement error interface.
func (e ErrOpenDB) Error() string {
	return e.Err.Error()
}

// Message method to implement wakaerror.Error interface.
func (e ErrOpenDB) Message() string {
	return fmt.Sprintf("failed to open db file: %s", e.Err)
}

// ExitCode method to implement wakaerror.Error interface.
func (ErrOpenDB) ExitCode() int {
	// Despite the error, we don't want to exit with an error code.
	return exitcode.Success
}

// SendDiagsOnErrors method to implement wakaerror.SendDiagsOnErrors interface.
func (ErrOpenDB) SendDiagsOnErrors() bool {
	return true
}

// ShouldLogError method to implement wakaerror.ShouldLogError interface.
func (ErrOpenDB) ShouldLogError() bool {
	return true
}
