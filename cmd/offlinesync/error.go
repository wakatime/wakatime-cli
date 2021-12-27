package offlinesync

// ErrSyncDisabled represents when sync is disabled.
type ErrSyncDisabled string

// Error method to implement error interface.
func (e ErrSyncDisabled) Error() string {
	return string(e)
}
