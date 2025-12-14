package models

func SanitizeData[T int | float64 | string](data *T) T {
	var default_value T
	if data == nil {
		return default_value
	}
	return *data
}

func SanitizeString(data *string) string {
	if data == nil {
		return ""
	}
	return *data
}

func SanitizeBoolean(data *bool) bool {
	if data == nil {
		return false
	}
	return *data
}
