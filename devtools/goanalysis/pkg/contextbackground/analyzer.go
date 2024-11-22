package contextbackground

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/authgear/authgear-server/devtools/goanalysis/pkg/util/vettedposutil"
)

func NewAnalyzer(pos *vettedposutil.VettedPositions) *analysis.Analyzer {
	run := func(pass *analysis.Pass) (interface{}, error) {
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
								isVetted := isVettedPos(pass, n, pos)
								if !isVetted {
									pass.Reportf(n.Pos(), "Unvetted usage of context.Background is forbidden.")
								}
							case "context.TODO":
								isVetted := isVettedPos(pass, n, pos)
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

	return &analysis.Analyzer{
		Name: "contextbackground",
		Doc:  "contextbackground forbids context.Background or context.TODO, except those locations explicitly hard-coded in this analyzer.",
		Run:  run,
		// See https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inspect
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
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

func isVettedPos(pass *analysis.Pass, n ast.Node, pos *vettedposutil.VettedPositions) bool {
	position := pass.Fset.Position(n.Pos())
	// position is valid if Line > 0
	// See https://pkg.go.dev/go/token#Position
	if position.Line > 0 {
		vetted := pos.CheckAndMarkUsed(pass.Analyzer.Name, position)
		if vetted {
			return true
		}
	}
	return false
}
