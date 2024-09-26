package theme

import (
	"bytes"
	"errors"
	"io"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

// MigrateSetDefaultLogoHeight set default logo heights for existing projects that does not have it set yet
func MigrateSetDefaultLogoHeight(r io.Reader) (result []byte, err error) {
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
