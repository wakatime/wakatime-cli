package offlinesync

import (
	"github.com/wakatime/wakatime-cli/pkg/exitcode"
	"github.com/wakatime/wakatime-cli/pkg/wakaerror"
)

// ErrSyncDisabled represents when sync is disabled.
type ErrSyncDisabled struct{}

var _ wakaerror.Error = ErrSyncDisabled{}

// Error method to implement error interface.
func (e ErrSyncDisabled) Error() string {
	return e.Message()
}

// ExitCode method to implement wakaerror.Error interface.
func (ErrSyncDisabled) ExitCode() int {
	return exitcode.Success
}

// Message method to implement wakaerror.Error interface.
func (ErrSyncDisabled) Message() string {
	return "sync offline activity is disabled"
}
