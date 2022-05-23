//go:build authgearlite
// +build authgearlite

package messageformat

import (
	"fmt"
	templateparse "text/template/parse"

	"golang.org/x/text/language"
)

var TemplateRuntimeFuncName = ""

var TemplateRuntimeFunc = templateRuntimeFunc

var errPanic = fmt.Errorf("gomessageformat is not available in lite build")

func templateRuntimeFunc(typ string, args ...interface{}) interface{} {
	panic(errPanic)
}

func FormatTemplateParseTree(tag language.Tag, pattern string) (tree *templateparse.Tree, err error) {
	panic(errPanic)
}

func IsEmptyParseTree(tree *templateparse.Tree) bool {
	panic(errPanic)
}
