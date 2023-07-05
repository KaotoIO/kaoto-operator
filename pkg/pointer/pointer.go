package pointer

func Any[T any](value T) *T {
	return &value
}
