package test

// Must is a helper function for functions with (value, error) signature
// returning the value and panicking if the error is not nil.
func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
