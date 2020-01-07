package validation

import (
	"errors"
	"net/mail"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"

	"github.com/skygeario/skygear-server/pkg/core/phone"
)

func addFormatChecker(name string, f gojsonschema.FormatChecker) {
	gojsonschema.FormatCheckers.Remove(name)
	gojsonschema.FormatCheckers.Add(name, f)
}

func init() {
	addFormatChecker("URLPathOnly", URL{
		URLVariant: URLVariantPathOnly,
	})
	addFormatChecker("URLFullOrPath", URL{
		URLVariant: URLVariantFullOrPath,
	})
	addFormatChecker("URLFullOnly", URL{
		URLVariant: URLVariantFullOrPath,
	})
	addFormatChecker("RelativeDirectoryPath", FilePath{
		Relative: true,
		File:     false,
	})
	addFormatChecker("RelativeFilePath", FilePath{
		Relative: true,
		File:     true,
	})
	addFormatChecker("phone", E164Phone{})
	addFormatChecker("email", Email{AllowName: false})
	addFormatChecker("NameEmailAddr", Email{AllowName: true})
}

type URLVariant int

const (
	URLVariantFullOnly URLVariant = iota
	URLVariantPathOnly
	URLVariantFullOrPath
)

type URL struct {
	URLVariant URLVariant
}

func (f URL) IsFormat(input interface{}) bool {
	return f.ValidateFormat(input) == nil
}

// nolint: gocyclo
func (f URL) ValidateFormat(input interface{}) error {
	str, ok := input.(string)
	if !ok {
		return nil
	}

	u, err := url.Parse(str)
	if err != nil {
		return err
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return errors.New("input URL must not have query/fragment")
	}

	p := ""
	switch f.URLVariant {
	case URLVariantFullOnly:
		if u.Scheme == "" || u.Host == "" {
			return errors.New("input URL must be absolute")
		}
		p = u.EscapedPath()
		if p == "" {
			p = "/"
		}
	case URLVariantPathOnly:
		if u.Scheme != "" || u.User != nil || u.Host != "" {
			return errors.New("input URL must be absolute path")
		}
		p = str
	case URLVariantFullOrPath:
		if u.Scheme != "" || u.User != nil || u.Host != "" {
			p = u.EscapedPath()
			if p == "" {
				p = "/"
			}
		} else {
			p = str
		}
	}

	cleaned := filepath.Clean(p)
	if !filepath.IsAbs(p) || cleaned != p {
		return errors.New("input URL must be normalized")
	}

	return nil
}

type FilePath struct {
	Relative bool
	File     bool
}

func (f FilePath) IsFormat(input interface{}) bool {
	return f.ValidateFormat(input) == nil
}

func (f FilePath) ValidateFormat(input interface{}) error {
	str, ok := input.(string)
	if !ok {
		return nil
	}

	abs := filepath.IsAbs(str)
	if f.Relative && abs {
		return errors.New("input must be a relative path")
	}
	if !f.Relative && !abs {
		return errors.New("input must be an absolute path")
	}

	trailingSlash := strings.HasSuffix(str, "/")
	if f.File && trailingSlash {
		return errors.New("input path must not have a trailing slash")
	}

	return nil
}

// E164Phone checks if input is a phone number in E.164 format.
// If the input is not a string or is an empty string, it is not an error.
// To enforce string or non-empty string, use other JSON schema constructs.
// This design allows this format to validate optional phone number.
type E164Phone struct{}

func (f E164Phone) IsFormat(input interface{}) bool {
	return f.ValidateFormat(input) == nil
}

func (f E164Phone) ValidateFormat(input interface{}) error {
	str, ok := input.(string)
	if !ok {
		return nil
	}
	return phone.EnsureE164(str)
}

// Email checks if input is an email address.
// If the input is not a string or is an empty string, it is not an error.
// To enforce string or non-empty string, use other JSON schema constructs.
// This design allows this format to validate optional email.
type Email struct {
	AllowName bool
}

func (f Email) IsFormat(input interface{}) bool {
	return f.ValidateFormat(input) == nil
}

func (f Email) ValidateFormat(input interface{}) error {
	s, ok := input.(string)
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

	ss := addr.String()
	// Remove <>
	if len(ss) >= 2 && ss[0] == '<' && ss[len(ss)-1] == '>' {
		ss = ss[1 : len(ss)-1]
	}
	if s != ss {
		return errors.New("input email must be normalized")
	}

	return nil
}
