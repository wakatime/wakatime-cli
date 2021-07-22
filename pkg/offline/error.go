package offline

// ErrOfflineEnqueue represents an enqueue to offline db error.
type ErrOfflineEnqueue string

// Error method to implement error interface.
func (e ErrOfflineEnqueue) Error() string {
	return string(e)
}
