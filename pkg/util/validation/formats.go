package validation

import (
	"errors"
	"net/mail"
	"net/url"
	"path/filepath"

	jsonschemaformat "github.com/iawaknahc/jsonschema/pkg/jsonschema/format"

	"github.com/authgear/authgear-server/pkg/util/phone"
)

func init() {
	jsonschemaformat.DefaultChecker["phone"] = FormatPhone{}
	jsonschemaformat.DefaultChecker["email"] = FormatEmail{AllowName: false}
	jsonschemaformat.DefaultChecker["email-name-addr"] = FormatEmail{AllowName: true}
	jsonschemaformat.DefaultChecker["uri"] = FormatURI{}
}

// FormatPhone checks if input is a phone number in E.164 format.
// If the input is not a string or is an empty string, it is not an error.
// To enforce string or non-empty string, use other JSON schema constructs.
// This design allows this format to validate optional phone number.
type FormatPhone struct{}

func (f FormatPhone) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}
	return phone.EnsureE164(str)
}

// FormatEmail checks if input is an email address.
// If the input is not a string or is an empty string, it is not an error.
// To enforce string or non-empty string, use other JSON schema constructs.
// This design allows this format to validate optional email.
type FormatEmail struct {
	AllowName bool
}

func (f FormatEmail) CheckFormat(value interface{}) error {
	s, ok := value.(string)
	if !ok {
		return nil
	}

	addr, err := mail.ParseAddress(s)
	if err != nil {
		return err
	}

	if !f.AllowName && addr.Name != "" {
		return errors.New("input email must not have name")
	}

	return nil
}

// FormatURI checks if input is an absolute URI.
type FormatURI struct {
}

func (f FormatURI) CheckFormat(value interface{}) error {
	str, ok := value.(string)
	if !ok {
		return nil
	}

	u, err := url.Parse(str)
	if err != nil {
		return err
	}

	if u.Scheme == "" || u.Host == "" {
		return errors.New("input URL must be absolute")
	}
	p := u.EscapedPath()
	if p == "" {
		p = "/"
	}

	cleaned := filepath.Clean(p)
	if !filepath.IsAbs(p) || cleaned != p {
		return errors.New("input URL must be normalized")
	}

	return nil
}
