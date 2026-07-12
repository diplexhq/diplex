package utils

// Must panics on error, otherwise returns the value.
// Use for operations that must succeed (e.g. go.mod parsing, file opening).
func Must[T any](val T, err error) T {
	if err != nil {
		panic(err)
	}

	return val
}

// NoErr panics on error.
// Use inside goroutines and callbacks where graceful recovery is not possible.
func NoErr(err error) {
	if err != nil {
		panic(err)
	}
}
