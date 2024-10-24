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
	isTerminal := checkIfTerminal(w)
	return &WriterHook{
		Writer: w,
		Formatter: &logrus.TextFormatter{
			ForceColors: isTerminal,
			// Disable quote when it is terminal so that newlines are printed without being escaped.
			// See https://github.com/sirupsen/logrus/issues/608#issuecomment-745137306
			DisableQuote: isTerminal,
		},
	}
}

func (h *WriterHook) Levels() []logrus.Level {
	return WriterHookLevels
}

func (h *WriterHook) Fire(entry *logrus.Entry) error {
	// The writer hook is supposed to be placed before SkipLoggingHook.
	// So it does not skip logging any entry.
	bytes, err := h.Formatter.Format(entry)
	if err != nil {
		return err
	}

	_, err = h.Writer.Write(bytes)
	return err
}
