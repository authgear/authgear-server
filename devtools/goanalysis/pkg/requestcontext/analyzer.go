package requestcontext

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var vettedPositions = []string{
	"/pkg/util/httproute/httproute.go:109:38",
}

var Analyzer = &analysis.Analyzer{
	Name: "requestcontext",
	Doc:  "requestcontext forbids (*net/http.Request).Context, except those locations explicitly hard-coded in this analyzer.",
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

func IsVettedPos(pass *analysis.Pass, n ast.Node) bool {
	position := pass.Fset.Position(n.Pos())

	// position is valid if Line > 0
	// See https://pkg.go.dev/go/token#Position
	if position.Line > 0 {
		b := slices.ContainsFunc(vettedPositions, func(s string) bool {
			return strings.HasSuffix(position.String(), s)
		})
		if b {
			return true
		}
	}

	return false
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	traverse := func(n ast.Node) {
		isTestFile := isATestFile(pass, n)

		if !isTestFile {
			if n, ok := n.(*ast.SelectorExpr); ok {
				if selObj := pass.TypesInfo.ObjectOf(n.Sel); selObj != nil {
					if f, ok := selObj.(*types.Func); ok {
						fullName := f.FullName()
						switch fullName {
						case "(*net/http.Request).Context":
							isVetted := IsVettedPos(pass, n)
							if !isVetted {
								pass.Reportf(n.Pos(), "Unvetted usage of request.Context is forbidden.")
							}
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
