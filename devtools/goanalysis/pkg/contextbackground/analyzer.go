package contextbackground

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
	"/pkg/util/jwkutil/contextbackground.go:11:52",
	"/pkg/lib/config/contextbackground.go:11:52",
	"/pkg/lib/deps/contextbackground.go:10:28",
	"/pkg/lib/oauthrelyingparty/oauthrelyingpartyutil/contextbackground.go:11:52",
	"/cmd/authgear/background/contextbackground.go:8:34",
	"/cmd/authgear/main.go:46:9",
	"/cmd/portal/main.go:41:9",
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

func isVettedPos(pass *analysis.Pass, n ast.Node) bool {
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
						case "context.Background":
							isVetted := isVettedPos(pass, n)
							if !isVetted {
								pass.Reportf(n.Pos(), "Unvetted usage of context.Background is forbidden.")
							}
						case "context.TODO":
							isVetted := isVettedPos(pass, n)
							if !isVetted {
								pass.Reportf(n.Pos(), "Unvetted usage of context.TODO is forbidden.")
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
