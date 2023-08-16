package exitchecker

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "check for os.Exit",
	Run:  run,
}

type Pass struct {
	Fset  *token.FileSet
	Files []*ast.File
	Pkg   *types.Package
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" {
			continue
		}

		var ParentFunc string
		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				ParentFunc = x.Name.Name

			case *ast.ExprStmt:
				if c, ok := x.X.(*ast.CallExpr); ok {
					if s, ok := c.Fun.(*ast.SelectorExpr); ok {
						if ParentFunc == "main" && s.X.(*ast.Ident).Name == "os" && s.Sel.Name == "Exit" {
							fmt.Printf("exitcheck: os.Exit() in %v\n", pass.Fset.Position(x.Pos()))
						}
					}
				}
			}
			return true
		})
	}
	return nil, nil
}
