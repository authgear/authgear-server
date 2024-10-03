package theme

import (
	"bytes"
	"errors"
	"io"
	"strings"

	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/css"
)

type indentation string

func (i indentation) Next() indentation {
	if i == "" {
		return indentation("  ")
	}
	return indentation(string(i) + string(i))
}

type element interface {
	Stringify(buf *bytes.Buffer, indent indentation)
}

type declaration struct {
	Property string
	Value    string
}

func (d *declaration) Stringify(buf *bytes.Buffer, indent indentation) {
	buf.Write([]byte(indent))
	buf.Write([]byte(d.Property))
	buf.Write([]byte(": "))
	buf.Write([]byte(d.Value))
	buf.Write([]byte(";\n"))
}

type ruleset struct {
	Selector     string
	Declarations []*declaration
}

func (r *ruleset) Stringify(buf *bytes.Buffer, indent indentation) {
	buf.Write([]byte(indent))
	buf.Write([]byte(r.Selector))
	buf.Write([]byte(" {\n"))
	nextIndent := indent.Next()
	for _, decl := range r.Declarations {
		decl.Stringify(buf, nextIndent)
	}
	buf.Write([]byte(indent))
	buf.Write([]byte("}\n"))
}

type atrule struct {
	Identifier string
	Value      string
	Rulesets   []*ruleset
}

func (r *atrule) Stringify(buf *bytes.Buffer, indent indentation) {
	buf.Write([]byte(indent))
	buf.Write([]byte(r.Identifier))
	buf.Write([]byte(" "))
	buf.Write([]byte(r.Value))
	buf.Write([]byte(" {\n"))
	nextIndent := indent.Next()
	for _, ruleset := range r.Rulesets {
		ruleset.Stringify(buf, nextIndent)
	}
	buf.Write([]byte(indent))
	buf.Write([]byte("}\n"))
}

func parseAtrule(p *css.Parser, a *atrule) (err error) {
	for {
		gt, _, _ := p.Next()
		switch gt {
		case css.ErrorGrammar:
			err = p.Err()
			return
		case css.EndAtRuleGrammar:
			return
		case css.BeginRulesetGrammar:
			r := &ruleset{
				Selector: collectTokensAsString(p.Values()),
			}
			err = parseRuleset(p, r)
			if err != nil {
				return
			}
			a.Rulesets = append(a.Rulesets, r)
		default:
			// Ignore everything we do not recognize.
			continue
		}
	}
}

func parseRuleset(p *css.Parser, r *ruleset) (err error) {
	for {
		gt, _, data := p.Next()
		switch gt {
		case css.ErrorGrammar:
			err = p.Err()
			return
		case css.EndRulesetGrammar:
			return
		case css.DeclarationGrammar:
			decl := &declaration{
				Property: string(data),
				Value:    collectTokensAsString(p.Values()),
			}
			r.Declarations = append(r.Declarations, decl)
		case css.CustomPropertyGrammar:
			// The tokens looks like [CustomPropertyValue(" value")]
			// So we have to trim the spaces.
			decl := &declaration{
				Property: string(data),
				Value:    strings.TrimSpace(collectTokensAsString(p.Values())),
			}
			r.Declarations = append(r.Declarations, decl)
		default:
			// Ignore everything we do not recognize.
			continue
		}
	}
}

func parseElement(p *css.Parser) (element element, err error) {
	for {
		gt, _, data := p.Next()
		switch gt {
		case css.ErrorGrammar:
			err = p.Err()
			return
		case css.BeginAtRuleGrammar:
			a := &atrule{
				Identifier: string(data),
				Value:      collectTokensAsString(p.Values()),
			}
			err = parseAtrule(p, a)
			if err != nil {
				return
			}
			element = a
			return
		case css.BeginRulesetGrammar:
			r := &ruleset{
				Selector: collectTokensAsString(p.Values()),
			}
			err = parseRuleset(p, r)
			if err != nil {
				return
			}
			element = r
			return
		default:
			// Ignore everything we do not recognize.
			continue
		}
	}
}

func collectTokensAsString(tokens []css.Token) string {
	var buf bytes.Buffer
	for _, token := range tokens {
		buf.Write(token.Data)
	}
	return buf.String()
}

func stringify(buf *bytes.Buffer, elements []element) {
	for _, element := range elements {
		var indent indentation
		element.Stringify(buf, indent)
	}
}

func parseCSSRawString(cssStr string) ([]element, error) {
	b := []byte(cssStr)
	r := bytes.NewReader(b)
	p := css.NewParser(parse.NewInput(r), false)
	var elements []element
	for {
		var el element
		el, err := parseElement(p)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, err
		}
		elements = append(elements, el)
	}
	return elements, nil
}

// CheckDeclarationInSelector checks if the declaration is in the selector of provided css
func CheckDeclarationInSelector(cssString string, selector string, declarationProperty string) (bool, error) {
	elements, err := parseCSSRawString(cssString)
	if err != nil {
		return false, err
	}

	for _, el := range elements {
		switch v := el.(type) {
		case *ruleset:
			if v.Selector == selector {
				for _, d := range v.Declarations {
					if d.Property == declarationProperty {
						return true, nil
					}
				}
			}
		}
	}
	return false, nil
}

// Add declaration in selector if not present already. If added, then added is true.
func AddDeclarationInSelectorIfNotPresentAlready(cssString string, selector string, declaration declaration) (newCSS string, added bool) {
	alreadyPresent, err := CheckDeclarationInSelector(cssString, selector, declaration.Property)
	if err != nil {
		return cssString, false
	}
	if alreadyPresent {
		return cssString, false
	}

	elements, err := parseCSSRawString(cssString)
	if err != nil {
		return cssString, false
	}

	var out []element
	for _, el := range elements {
		switch v := el.(type) {
		case *ruleset:
			if v.Selector != selector {
				out = append(out, el)
				continue
			}
			// inside target selector

			// we know that this ruleset does not have target declaration set yet
			// so we just add it
			d := &declaration
			newEl := &ruleset{
				Selector:     v.Selector,
				Declarations: append(v.Declarations, d),
			}
			out = append(out, newEl)
		default:
			out = append(out, el)
		}
	}

	var buf bytes.Buffer
	stringify(&buf, out)

	return buf.String(), true
}
