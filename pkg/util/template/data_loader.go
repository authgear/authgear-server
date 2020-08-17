package template

import (
	"encoding/base64"
	"fmt"
	"strings"
	"unicode/utf8"
)

type DataLoader struct{}

func (l *DataLoader) Load(dataURI string) (templateContent string, err error) {
	if !strings.HasPrefix(dataURI, "data:base64,") {
		err = ErrInvalidDataURI
		return
	}
	encoded := strings.TrimPrefix(dataURI, "data:base64,")
	bytes, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		err = ErrInvalidDataURI
		return
	}

	if !utf8.Valid(bytes) {
		err = ErrInvalidUTF8
		return
	}

	templateContent = string(bytes)
	return
}

func DataURIWithContent(content string) string {
	return fmt.Sprintf("data:base64,%s", base64.StdEncoding.EncodeToString([]byte(content)))
}
