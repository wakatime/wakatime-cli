package filter

// Err is a heartbeat filtering error, signaling to skip the heartbeat.
type Err string

// Error implements error interface.
func (e Err) Error() string {
	return string(e)
}
