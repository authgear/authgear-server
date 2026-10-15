package htmlutil

import (
	"html"
	"html/template"
	"net/url"
	"strings"

	xhtml "golang.org/x/net/html"
)

// SanitizeLinksHTML keeps <a> elements whose href is http or https, and escapes everything else.
// The kept links open in a new tab.
func SanitizeLinksHTML(s string) template.HTML {
	var out strings.Builder
	z := xhtml.NewTokenizer(strings.NewReader(s))
	// openLinks counts the <a> tags kept, so that only their matching </a> is kept.
	openLinks := 0
	for {
		tt := z.Next()
		if tt == xhtml.ErrorToken {
			break
		}
		raw := string(z.Raw())
		switch tt {
		case xhtml.TextToken:
			// Raw text keeps the author's entities. Re-escape the decoded text instead.
			out.WriteString(html.EscapeString(string(z.Text())))
		case xhtml.StartTagToken:
			name, hasAttr := z.TagName()
			if string(name) == "a" {
				if href, ok := safeHref(z, hasAttr); ok {
					out.WriteString(`<a href="` + html.EscapeString(href) + `" target="_blank" rel="noopener noreferrer">`)
					openLinks++
					continue
				}
			}
			out.WriteString(html.EscapeString(raw))
		case xhtml.EndTagToken:
			name, _ := z.TagName()
			if string(name) == "a" && openLinks > 0 {
				openLinks--
				out.WriteString("</a>")
				continue
			}
			out.WriteString(html.EscapeString(raw))
		default:
			out.WriteString(html.EscapeString(raw))
		}
	}
	for ; openLinks > 0; openLinks-- {
		out.WriteString("</a>")
	}
	// The output is built from escaped text and the allowlisted <a> tag only.
	return template.HTML(out.String()) // nolint:gosec
}

func safeHref(z *xhtml.Tokenizer, hasAttr bool) (string, bool) {
	for hasAttr {
		var key, val []byte
		key, val, hasAttr = z.TagAttr()
		if string(key) != "href" {
			continue
		}
		u, err := url.Parse(strings.TrimSpace(string(val)))
		if err != nil {
			return "", false
		}
		scheme := strings.ToLower(u.Scheme)
		if scheme != "http" && scheme != "https" {
			return "", false
		}
		return strings.TrimSpace(string(val)), true
	}
	return "", false
}

// LinksToText returns the text of s with all tags removed.
func LinksToText(s string) string {
	var out strings.Builder
	z := xhtml.NewTokenizer(strings.NewReader(s))
	for {
		tt := z.Next()
		if tt == xhtml.ErrorToken {
			break
		}
		if tt == xhtml.TextToken {
			out.WriteString(string(z.Text()))
		}
	}
	return out.String()
}
