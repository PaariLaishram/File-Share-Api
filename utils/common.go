package utils

func CopyDataPtr[T string | bool | int](src *T) *T {
	if src == nil {
		return nil
	}
	v := *src
	return &v
}
