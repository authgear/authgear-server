package errors

type errorSecondary struct {
	inner     error
	secondary error
}

func WithSecondaryError(err, serr error) error {
	if err == nil || serr == nil {
		return err
	}
	return &errorSecondary{inner: err, secondary: serr}
}

func (e *errorSecondary) Error() string { return e.inner.Error() }
func (e *errorSecondary) Unwrap() error { return e.inner }
func (e *errorSecondary) FillDetails(d Details) {
	CollectDetails(e.secondary, d)
}
