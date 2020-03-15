package utils

//IsValidEntityType Return true if entity type is valid, otherwise false
func IsValidEntityType(entityType string) bool {
	switch entityType {
	case
		"file",
		"domain",
		"app":
		return true
	}
	return false
}

//IsValidCategory Return true if category is valid, otherwise false
func IsValidCategory(category string) bool {
	switch category {
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
