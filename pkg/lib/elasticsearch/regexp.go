package elasticsearch

import (
	"strings"
)

type regexpLiteral string

func (l regexpLiteral) Replace(old string, replacement string) regexpLiteral {
	return regexpLiteral(strings.ReplaceAll(string(l), old, replacement))
}

func (l regexpLiteral) String() string {
	return string(l)
}

// EscapeRegexp escapes literal according to
// https://www.elastic.co/guide/en/elasticsearch/reference/current/regexp-syntax.html
func EscapeRegexp(literal string) string {
	return regexpLiteral(literal).
		// We must first replace \ otherwise,
		// \ will be replaced twice.
		Replace(`\`, `\\`).
		Replace(`.`, `\.`).
		Replace(`?`, `\?`).
		Replace(`+`, `\+`).
		Replace(`*`, `\*`).
		Replace(`|`, `\|`).
		Replace(`{`, `\{`).
		Replace(`}`, `\}`).
		Replace(`[`, `\[`).
		Replace(`]`, `\]`).
		Replace(`(`, `\(`).
		Replace(`)`, `\)`).
		Replace(`"`, `\"`).
		Replace(`#`, `\#`).
		Replace(`@`, `\@`).
		Replace(`&`, `\&`).
		Replace(`<`, `\<`).
		Replace(`>`, `\>`).
		Replace(`~`, `\~`).
		String()
}
