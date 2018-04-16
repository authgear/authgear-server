// This file is adapted from
// https://github.com/sirupsen/logrus/blob/778f2e7/text_formatter.go

package logging

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	krtext "github.com/kr/text"
	"github.com/sirupsen/logrus"
)

const defaultTimestampFormat = time.RFC3339

const (
	nocolor = 0
	red     = 31
	green   = 32
	yellow  = 33
	blue    = 36
	gray    = 37
)

var (
	baseTimestamp time.Time
	emptyFieldMap logrus.FieldMap
)

func init() {
	baseTimestamp = time.Now()
}

type ValueFormatter interface {
	Print(b *bytes.Buffer)
}

// TextFormatter formats logs into text
type TextFormatter struct {
	// Set to true to bypass checking for a TTY before outputting colors.
	ForceColors bool

	// Force disabling colors.
	DisableColors bool

	// Disable timestamp logging. useful when output is redirected to logging
	// system that already adds timestamps.
	DisableTimestamp bool

	// Enable logging the full timestamp when a TTY is attached instead of just
	// the time passed since beginning of execution.
	FullTimestamp bool

	// TimestampFormat to use for display when a full timestamp is printed
	TimestampFormat string

	// The fields are sorted by default for a consistent output. For applications
	// that log extremely frequently and don't use the JSON formatter this may not
	// be desired.
	DisableSorting bool

	// Disables the truncation of the level text to 4 characters.
	DisableLevelTruncation bool

	// QuoteEmptyFields will wrap empty fields in quotes if true
	QuoteEmptyFields bool

	// Whether the logger's out is to a terminal
	isTerminal bool

	sync.Once
}

func (f *TextFormatter) init(entry *logrus.Entry) {
	//if entry.Logger != nil {
	//	f.isTerminal = checkIfTerminal(entry.Logger.Out)
	//}
}

// Format renders a single log entry
func (f *TextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	keys := []string{}
	formattedKeys := []string{}
	for k, v := range entry.Data {
		if _, ok := v.(ValueFormatter); ok {
			formattedKeys = append(formattedKeys, k)
			continue
		}
		keys = append(keys, k)
	}

	if !f.DisableSorting {
		sort.Strings(keys)
	}
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	//prefixFieldClashes(entry.Data, emptyFieldMap)

	f.Do(func() { f.init(entry) })

	isColored := (f.ForceColors || f.isTerminal) && !f.DisableColors

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = defaultTimestampFormat
	}
	if isColored {
		f.printColored(b, entry, keys, timestampFormat)
	} else {
		if !f.DisableTimestamp {
			f.appendKeyValue(b, "time", entry.Time.Format(timestampFormat))
		}
		f.appendKeyValue(b, "level", entry.Level.String())
		if entry.Message != "" {
			f.appendKeyValue(b, "msg", entry.Message)
		}
		for _, key := range keys {
			f.appendKeyValue(b, key, entry.Data[key])
		}
	}

	b.WriteByte('\n')

	for _, key := range formattedKeys {
		if v, ok := entry.Data[key].(ValueFormatter); ok {
			vb := &bytes.Buffer{}
			v.Print(vb)

			b.WriteString(key)
			b.WriteByte(':')
			b.WriteByte('\n')
			b.WriteString(krtext.Indent(vb.String(), "    "))
			b.WriteByte('\n')
		}
	}

	return b.Bytes(), nil
}

func (f *TextFormatter) printColored(b *bytes.Buffer, entry *logrus.Entry, keys []string, timestampFormat string) {
	var levelColor int
	switch entry.Level {
	case logrus.DebugLevel:
		levelColor = gray
	case logrus.WarnLevel:
		levelColor = yellow
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColor = red
	default:
		levelColor = blue
	}

	levelText := strings.ToUpper(entry.Level.String())
	if !f.DisableLevelTruncation {
		levelText = levelText[0:4]
	}

	if f.DisableTimestamp {
		fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m %-44s ", levelColor, levelText, entry.Message)
	} else if !f.FullTimestamp {
		fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m[%04d] %-44s ", levelColor, levelText, int(entry.Time.Sub(baseTimestamp)/time.Second), entry.Message)
	} else {
		fmt.Fprintf(b, "\x1b[%dm%s\x1b[0m[%s] %-44s ", levelColor, levelText, entry.Time.Format(timestampFormat), entry.Message)
	}
	for _, k := range keys {
		v := entry.Data[k]
		fmt.Fprintf(b, " \x1b[%dm%s\x1b[0m=", levelColor, k)
		f.appendValue(b, v)
	}
}

// nolint: gocyclo
func (f *TextFormatter) needsQuoting(text string) bool {
	if f.QuoteEmptyFields && len(text) == 0 {
		return true
	}
	for _, ch := range text {
		if !((ch >= 'a' && ch <= 'z') ||
			(ch >= 'A' && ch <= 'Z') ||
			(ch >= '0' && ch <= '9') ||
			ch == '-' || ch == '.' || ch == '_' || ch == '/' || ch == '@' || ch == '^' || ch == '+') {
			return true
		}
	}
	return false
}

func (f *TextFormatter) appendKeyValue(b *bytes.Buffer, key string, value interface{}) {
	if b.Len() > 0 {
		b.WriteByte(' ')
	}
	b.WriteString(key)
	b.WriteByte('=')
	f.appendValue(b, value)
}

func (f *TextFormatter) appendValue(b *bytes.Buffer, value interface{}) {
	stringVal, ok := value.(string)
	if !ok {
		stringVal = fmt.Sprint(value)
	}

	if !f.needsQuoting(stringVal) {
		b.WriteString(stringVal)
	} else {
		b.WriteString(fmt.Sprintf("%q", stringVal))
	}
}
