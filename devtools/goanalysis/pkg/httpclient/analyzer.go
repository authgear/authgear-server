package httpclient

import (
	"go/ast"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name: "httpclient",
	Doc:  "httpclient restricts the way how a http.Client can be used.",
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

func isOfFilename(pass *analysis.Pass, n ast.Node, filename string) bool {
	position := pass.Fset.Position(n.Pos())
	// position is valid if Line > 0
	// See https://pkg.go.dev/go/token#Position
	if position.Line > 0 {
		return position.Filename == filename
	}
	return false
}

func isOfPackagePath(pass *analysis.Pass, packagePath string) bool {
	return pass.Pkg.Path() == packagePath
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspect := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	traverse := func(n ast.Node) {
		isTestFile := isATestFile(pass, n)
		isHttputil := isOfPackagePath(pass, "github.com/authgear/authgear-server/pkg/util/httputil")
		isExtClientDotGo := isOfFilename(pass, n, "ext_client.go")

		if !isTestFile && !isHttputil && !isExtClientDotGo {
			if n, ok := n.(*ast.CompositeLit); ok {
				if typ := pass.TypesInfo.TypeOf(n.Type); typ != nil {
					if typ.String() == "net/http.Client" {
						pass.Reportf(n.Pos(), "Constructing http.Client directly is forbidden. Use httputil.NewExternalClient instead.")
					}
				}
			}
		}

		if !isTestFile {
			if n, ok := n.(*ast.SelectorExpr); ok {
				if selObj := pass.TypesInfo.ObjectOf(n.Sel); selObj != nil {
					if v, ok := selObj.(*types.Var); ok {
						if v.String() == "var net/http.DefaultClient *net/http.Client" {
							pass.Reportf(n.Pos(), "Using http.DefaultClient is forbidden. Use httputil.NewExternalClient instead.")
						}
					}
				}
			}
		}

		if !isTestFile {
			if n, ok := n.(*ast.SelectorExpr); ok {
				if selObj := pass.TypesInfo.ObjectOf(n.Sel); selObj != nil {
					if f, ok := selObj.(*types.Func); ok {
						fullName := f.FullName()
						if fullName == "net/http.NewRequest" {
							pass.Reportf(n.Pos(), "Calling http.NewRequest is forbidden. Use http.NewRequestWithContext instead.")
						}
					}
				}
			}
		}

		if !isTestFile {
			if n, ok := n.(*ast.SelectorExpr); ok {
				if selObj := pass.TypesInfo.ObjectOf(n.Sel); selObj != nil {
					if f, ok := selObj.(*types.Func); ok {
						fullName := f.FullName()
						switch fullName {
						case "net/http.Get":
							fallthrough
						case "net/http.Head":
							fallthrough
						case "net/http.Post":
							fallthrough
						case "net/http.PostForm":
							pass.Reportf(n.Pos(), "context.Context is lost. Use http.Client.Do instead.")
						default:
							break
						}
					}
				}
			}
		}

		if !isTestFile {
			if n, ok := n.(*ast.SelectorExpr); ok {
				if selObj := pass.TypesInfo.ObjectOf(n.Sel); selObj != nil {
					if fun, ok := selObj.(*types.Func); ok {
						fullName := fun.FullName()
						switch fullName {
						case "(*net/http.Client).Get":
							fallthrough
						case "(*net/http.Client).Head":
							fallthrough
						case "(*net/http.Client).Post":
							fallthrough
						case "(*net/http.Client).PostForm":
							pass.Reportf(n.Pos(), "context.Context is lost. Use http.Client.Do instead.")
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
