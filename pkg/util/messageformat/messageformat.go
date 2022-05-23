//go:build !authgearlite
// +build !authgearlite

package messageformat

import (
	templateparse "text/template/parse"

	"golang.org/x/text/language"

	gomessageformat "github.com/iawaknahc/gomessageformat"
)

var TemplateRuntimeFuncName = gomessageformat.TemplateRuntimeFuncName

var TemplateRuntimeFunc = gomessageformat.TemplateRuntimeFunc

func FormatTemplateParseTree(tag language.Tag, pattern string) (tree *templateparse.Tree, err error) {
	return gomessageformat.FormatTemplateParseTree(tag, pattern)
}

func IsEmptyParseTree(tree *templateparse.Tree) bool {
	return gomessageformat.IsEmptyParseTree(tree)
}
