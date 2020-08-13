package template

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var ErrInvalidDataURI = errors.New("unvalid data URI")

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
		err = errors.New("expected content to be UTF-8 encoded")
		return
	}

	templateContent = string(bytes)
	return
}

func DataURIWithContent(content string) string {
	return fmt.Sprintf("data:base64,%s", base64.StdEncoding.EncodeToString([]byte(content)))
}
