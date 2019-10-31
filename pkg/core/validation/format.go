package validation

import (
	"net/url"
	"path/filepath"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

func addFormatChecker(name string, f gojsonschema.FormatChecker) {
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

// nolint: gocyclo
func (f URL) IsFormat(input interface{}) bool {
	str, ok := input.(string)
	if !ok {
		return false
	}
	if str == "" {
		return false
	}

	u, err := url.Parse(str)
	if err != nil {
		return false
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return false
	}

	p := ""
	switch f.URLVariant {
	case URLVariantFullOnly:
		if u.Scheme == "" || u.Host == "" {
			return false
		}
		p = u.EscapedPath()
		if p == "" {
			p = "/"
		}
	case URLVariantPathOnly:
		if u.Scheme != "" || u.User != nil || u.Host != "" {
			return false
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
		return false
	}

	return true
}

type FilePath struct {
	Relative bool
	File     bool
}

func (f FilePath) IsFormat(input interface{}) bool {
	str, ok := input.(string)
	if !ok {
		return false
	}

	if str == "" {
		return false
	}

	abs := filepath.IsAbs(str)
	if f.Relative && abs {
		return false
	}
	if !f.Relative && !abs {
		return false
	}

	trailingSlash := strings.HasSuffix(str, "/")
	if f.File && trailingSlash {
		return false
	}

	return true
}
