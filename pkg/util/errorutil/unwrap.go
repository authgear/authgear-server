package errorutil

// Unwrap unwraps err. It DOES NOT use errors.Unwrap and handles errors.Join.
func Unwrap(err error, visitor func(err error)) {
	if err == nil {
		return
	}
	switch e := err.(type) {
	case interface{ Unwrap() error }:
		// Visit error that wraps another error.
		visitor(err)
		err = e.Unwrap()
		Unwrap(err, visitor)
	case interface{ Unwrap() []error }:
		// Do not visit the error returned by errors.Join.
		// It is because logically, it is not an error.
		for _, err := range e.Unwrap() {
			Unwrap(err, visitor)
		}
	default:
		// Visit any ordinary error.
		visitor(err)
	}
}
