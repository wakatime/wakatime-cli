package utils

//IsValidEntityType Return true if entity type is valid, otherwise false
func IsValidEntityType(et string) bool {
	switch et {
	case
		"file",
		"domain",
		"app":
		return true
	}
	return false
}

//IsValidCategory Return true if category is valid, otherwise false
func IsValidCategory(c string) bool {
	switch c {
	case
		"coding",
		"building",
		"indexing",
		"debugging",
		"running tests",
		"manual testing",
		"writing tests",
		"browsing",
		"code reviewing",
		"designing":
		return true
	}
	return false
}
