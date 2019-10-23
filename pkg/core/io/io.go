package io

import (
	"bytes"
)

// BytesReaderCloser is bytes.Reader with a Close method.
type BytesReaderCloser struct {
	*bytes.Reader
}

// Close implements Closer.
func (r *BytesReaderCloser) Close() error {
	return nil
}
