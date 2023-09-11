package utils

func MapIn[T comparable, V any](key T, m map[T]V) bool {
	_, ok := m[key]
	return ok
}
