package log

import (
	"io"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/term"
)

var WriterHookLevels = []logrus.Level{
	logrus.PanicLevel,
	logrus.FatalLevel,
	logrus.ErrorLevel,
	logrus.WarnLevel,
	logrus.InfoLevel,
	logrus.DebugLevel,
	logrus.TraceLevel,
}

func checkIfTerminal(w io.Writer) bool {
	switch v := w.(type) {
	case *os.File:
		return term.IsTerminal(int(v.Fd()))
	default:
		return false
	}
}

// WriterHook replaces logger.Out.
// Since logger.Out works with logger.Formatter,
// We have to replicate the behavior here.
// The reason for WriterHook to exist is to
// call our Ignore() to filter out unwanted entry.
type WriterHook struct {
	Writer    io.Writer
	Formatter logrus.Formatter
}

func NewWriterHook(w io.Writer) *WriterHook {
	colored := checkIfTerminal(w)
	return &WriterHook{
		Writer: w,
		Formatter: &logrus.TextFormatter{
			ForceColors: colored,
		},
	}
}

func (h *WriterHook) Levels() []logrus.Level {
	return WriterHookLevels
}

func (h *WriterHook) Fire(entry *logrus.Entry) error {
	if Ignore(entry) {
		return nil
	}

	bytes, err := h.Formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = h.Writer.Write(bytes)
	return err
}
