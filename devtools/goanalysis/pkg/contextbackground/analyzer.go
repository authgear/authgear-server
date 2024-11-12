package contextbackground

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var vettedFilenames = []string{
	"/pkg/util/jwkutil/contextbackground.go",
	"/pkg/lib/config/contextbackground.go",
	"/pkg/lib/deps/contextbackground.go",
	"/cmd/authgear/background/contextbackground.go",
}

var Analyzer = &analysis.Analyzer{
	Name: "contextbackground",
	Doc:  "contextbackground forbids context.Background or context.TODO, except those locations explicitly hard-coded in this analyzer.",
	Run:  run,
	// See https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inspect
	Requires: []*analysis.Analyzer{inspect.Analyzer},
}

func isATestFile(pass *analysis.Pass, n ast.Node) bool {
	position := pass.Fset.Position(n.Pos())
	// position is valid if Line > 0
	// See https://pkg.go.dev/go/token#Position
	if position.Line > 0 {
		if strings.HasSuffix(position.Filename, "_test.go") {
			return true
		}
	}
	return false
}

func isVettedFile(pass *analysis.Pass, n ast.Node) bool {
	position := pass.Fset.Position(n.Pos())
	// position is valid if Line > 0
	// See https://pkg.go.dev/go/token#Position
	if position.Line > 0 {
		for _, suffix := range vettedFilenames {
			if strings.HasSuffix(position.Filename, suffix) {
				return true
			}
		}
	}
	return false
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	traverse := func(n ast.Node) {
		isTestFile := isATestFile(pass, n)
		isVettedFile := isVettedFile(pass, n)

		if !isTestFile && !isVettedFile {
			if n, ok := n.(*ast.SelectorExpr); ok {
				if selObj := pass.TypesInfo.ObjectOf(n.Sel); selObj != nil {
					if f, ok := selObj.(*types.Func); ok {
						fullName := f.FullName()
						switch fullName {
						case "context.Background":
							pass.Reportf(n.Pos(), "Unvetted usage of context.Background is forbidden.")
						case "context.TODO":
							pass.Reportf(n.Pos(), "Unvetted usage of context.TODO is forbidden.")
						default:
							break
						}
					}
				}
			}
		}
	}

	inspect.Preorder(nil, traverse)
	return nil, nil
}
