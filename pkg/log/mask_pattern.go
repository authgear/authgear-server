package log

import (
	"regexp"
	"strings"
)

type MaskPattern interface {
	Mask(s string, mask string) string
}

type RegexMaskPattern struct {
	Pattern *regexp.Regexp
}

var _ MaskPattern = RegexMaskPattern{}

func NewRegexMaskPattern(expr string) RegexMaskPattern {
	return RegexMaskPattern{Pattern: regexp.MustCompile(expr)}
}

func (p RegexMaskPattern) Mask(s string, mask string) string {
	return p.Pattern.ReplaceAllString(s, mask)
}

type PlainMaskPattern struct {
	Pattern string
}

var _ MaskPattern = PlainMaskPattern{}

func NewPlainMaskPattern(s string) PlainMaskPattern {
	return PlainMaskPattern{Pattern: s}
}

func (p PlainMaskPattern) Mask(s string, mask string) string {
	return strings.ReplaceAll(s, p.Pattern, mask)
}
