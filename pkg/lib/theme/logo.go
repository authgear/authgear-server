package theme

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

// MigrateCreateCSSWithDefaultLogoHeight create a css file with default logo height for existing projects that does not have theme customization at all
func MigrateCreateCSSWithDefaultLogoHeight(selector string) (result []byte, err error) {
	cssRawStr := fmt.Sprintf("%v {}", selector)
	b := []byte(cssRawStr)
	r := bytes.NewReader(b)
	result, err = MigrateSetDefaultLogoHeight(r)
	return
}

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

	elements = setDefaultHeight(elements)
	var buf bytes.Buffer
	stringify(&buf, elements)

	result = buf.Bytes()
	return
}

func CheckLogoHeightDeclarationInSelector(cssString string, selector string) (bool, error) {
	return CheckDeclarationInSelector(cssString, selector, LogoHeightPropertyKey)
}

func CheckLogoHeightDeclarationInDarkThemeCSS(cssString string, selector string) (bool, error) {
	return CheckDeclarationInSelector(cssString, selector, LogoHeightPropertyKey)
}

var LightThemeCSSSelector string = ":root"
var DarkThemeCSSSelector string = ":root.dark"
var LogoHeightPropertyKey string = "--brand-logo__height"
var DefaultLogoHeight string = "40px"

func setDefaultHeight(elements []element) (out []element) {
	for _, el := range elements {
		switch v := el.(type) {
		case *ruleset:
			if v.Selector != LightThemeCSSSelector && v.Selector != DarkThemeCSSSelector {
				out = append(out, el)
				continue
			}
			// inside light/dark theme css selector
			if isLogoHeightSet(v) {
				out = append(out, el)
				continue
			}

			d := newLogoHeightDeclaration(DefaultLogoHeight)
			newEl := &ruleset{
				Selector:     v.Selector,
				Declarations: append(v.Declarations, d),
			}
			out = append(out, newEl)
		default:
			out = append(out, el)
		}
	}

	return
}

func isLogoHeightSet(rs *ruleset) bool {
	for _, d := range rs.Declarations {
		if d.Property == LogoHeightPropertyKey {
			return true
		}
	}

	return false
}

func newLogoHeightDeclaration(v string) *declaration {
	return &declaration{
		Property: LogoHeightPropertyKey,
		Value:    v,
	}
}
