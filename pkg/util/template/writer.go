package template

import (
	"io"
)

// MaxTemplateSize is 1MiB.
const MaxTemplateSize = 1024 * 1024 * 1

type LimitWriter struct {
	// Writer is the underlying writer.
	Writer io.Writer
	// RemainingQuota is the remaining quota in bytes.
	RemainingQuota int64
}

func NewLimitWriter(w io.Writer) *LimitWriter {
	return &LimitWriter{
		Writer:         w,
		RemainingQuota: MaxTemplateSize,
	}
}

func (w *LimitWriter) Write(p []byte) (n int, err error) {
	if w.RemainingQuota-int64(len(p)) <= 0 {
		return 0, ErrLimitReached
	}

	n, err = w.Writer.Write(p)
	w.RemainingQuota -= int64(n)

	return
}
