package vettedpos

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"

	"github.com/authgear/authgear-server/devtools/goanalysis/pkg/util/vettedposutil"
)

func NewAnalyzer(pos *vettedposutil.VettedPositions, analyzers ...*analysis.Analyzer) *analysis.Analyzer {
	var requires []*analysis.Analyzer
	// See https://pkg.go.dev/golang.org/x/tools/go/analysis/passes/inspect
	requires = append(requires, inspect.Analyzer)
	requires = append(requires, analyzers...)

	run := func(pass *analysis.Pass) (interface{}, error) {
		inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
		traverse := func(n ast.Node) {
			if n, ok := n.(*ast.File); ok {
				pos.Report(pass, n)
			}
		}

		inspect.Preorder(nil, traverse)
		return nil, nil
	}

	return &analysis.Analyzer{
		Name:     "vettedpos",
		Doc:      "vettedpos reports any unused vetted positions.",
		Run:      run,
		Requires: requires,
	}
}
