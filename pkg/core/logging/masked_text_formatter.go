package logging

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

type MaskedTextFormatter struct {
	Inner    logrus.TextFormatter
	Patterns []MaskPattern
	Mask     string
}

func (f *MaskedTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	fields := make(logrus.Fields, len(entry.Data))
	for k, v := range entry.Data {
		s, ok := v.(string)
		if !ok {
			s = fmt.Sprint(v)
		}

		s = f.mask(s)
		fields[k] = s
	}
	entry.Data = fields
	entry.Message = f.mask(entry.Message)
	return f.Inner.Format(entry)
}

func (f *MaskedTextFormatter) mask(src string) (output string) {
	output = src
	for _, p := range f.Patterns {
		output = p.Mask(output, f.Mask)
	}
	return output
}
