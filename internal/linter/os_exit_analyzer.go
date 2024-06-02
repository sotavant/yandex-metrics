package linter

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

var OsExitAnalyzer = &analysis.Analyzer{
	Name: "osexit_analyzer",
	Doc:  "check for os.exit in main function of main package",
	Run:  run,
}

const checkedPkgName = "main"
const checkedFncName = "main"

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

			fmt.Println(node.)

			return true
		})
	}
	return nil, nil
}
