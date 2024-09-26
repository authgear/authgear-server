package theme

import (
	"bytes"
	"errors"
	"io"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

// MigrateMediaQueryToClassBased migrates media query dark theme to class-based dark theme.
func MigrateMediaQueryToClassBased(r io.Reader) (result []byte, err error) {
	p := css.NewParser(parse.NewInput(r), false)

	var elements []element
	for {
		var el element
		el, err = parseElement(p)
		if errors.Is(err, io.EOF) {
			err = nil
			break
		}
		if err != nil {
			return
		}
		elements = append(elements, el)
	}

	elements = transform(elements)
	var buf bytes.Buffer
	stringify(&buf, elements)

	result = buf.Bytes()
	return
}

func transform(elements []element) (out []element) {
	for _, el := range elements {
		switch v := el.(type) {
		case *atrule:
			if v.Identifier == "@media" && v.Value == "(prefers-color-scheme:dark)" {
				// Remove this at rule
				for _, ruleset := range v.Rulesets {
					if ruleset.Selector == ":root" {
						ruleset.Selector = ":root.dark"
					}
					out = append(out, ruleset)
				}
			}
		default:
			out = append(out, el)
		}
	}

	return
}
