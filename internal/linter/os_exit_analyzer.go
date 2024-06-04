package linter

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexit_analyzer",
	Doc:  "check for os.exit in main function of main package",
	Run:  run,
}

const checkedPkgName = "main"
const checkedFncName = "main"
const checkedCallExpr = "os.Exit"
const checkCallArg = "0"

func run(pass *analysis.Pass) (interface{}, error) {
	//var lastFunc *ast.FuncDecl
	var isMainFunc = false
	//var startLine, endLine int

	for _, file := range pass.Files {
		isMainFunc = false
		// функцией ast.Inspect проходим по всем узлам AST
		ast.Inspect(file, func(node ast.Node) bool {
			if file.Name.String() != checkedPkgName {
				return true
			}

			if funcDecl, ok := node.(*ast.FuncDecl); ok {
				if funcDecl.Name.String() == checkedFncName {
					isMainFunc = true
				} else {
					isMainFunc = false
				}
			}

			if !isMainFunc {
				return true
			}

			switch n := node.(type) {
			case *ast.CallExpr:
				if _, ok := n.Fun.(*ast.SelectorExpr); !ok {
					return true
				}

				X := n.Fun.(*ast.SelectorExpr).X
				if _, ok := X.(*ast.Ident); !ok {
					return true
				}

				x := n.Fun.(*ast.SelectorExpr).X.(*ast.Ident).Name
				sel := n.Fun.(*ast.SelectorExpr).Sel.Name
				expr := x + "." + sel
				if expr != checkedCallExpr {
					return true
				}

				if n.Args[0].(*ast.BasicLit).Value == checkCallArg {
					pass.Reportf(n.Fun.Pos(), "bad expression for use")
				}
			}

			return true
		})
	}
	return nil, nil
}
