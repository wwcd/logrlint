// Package logrlint defines an Analyzer that reports time package expressions that
package logrlint

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

var Analyzer = &analysis.Analyzer{
	Name: "logrlint",
	Doc:  "reports logr",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			ce, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			fn, _ := typeutil.Callee(pass.TypesInfo, ce).(*types.Func)
			if fn == nil {
				return true
			}
			var kvs []ast.Expr
			fname := strings.Replace(fn.FullName(), "(github.com/go-logr/logr.Logger)", "logr", -1)
			switch fname {
			case `logr.Info`:
				if len(ce.Args)%2 != 1 {
					pass.Reportf(ce.Pos(), "%s number of args must be even, %q", fname, render(pass.Fset, ce))
				}
				kvs = ce.Args[1:]
			case `logr.Error`:
				if len(ce.Args)%2 != 0 {
					pass.Reportf(ce.Pos(), "%s number of args must be odd, %q", fname, render(pass.Fset, ce))
				}
				kvs = ce.Args[2:]
			default:
				return true
			}

			for i := 0; i < len(kvs); i += 2 {
				switch typ := pass.TypesInfo.Types[kvs[i]].Type.(type) {
				case *types.Basic:
					switch typ.Kind() {
					case types.String:
						continue
					}
				}
				pass.Reportf(ce.Pos(), "%s type of %dth key is not string, %q", fname, i/2+1, render(pass.Fset, ce))
			}
			return true
		})
	}

	return nil, nil
}

// render returns the pretty-print of the given node
func render(fset *token.FileSet, x interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(&buf, fset, x); err != nil {
		panic(err)
	}
	return buf.String()
}
