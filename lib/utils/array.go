package utils

// Isset Check if index is present on given array
func Isset(arr []string, index int) bool {
	return len(arr) > index
}
