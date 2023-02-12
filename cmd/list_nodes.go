package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func list_nodes(src string) string {

	var sb strings.Builder

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		panic(err)
	}

	// Inspect the AST and print all identifiers and literals.
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		s := fmt.Sprintf("***** TYPE-----%t-----", n)
		if s != "" {
			sb.WriteString(fmt.Sprintf("%s:\t%s\n\n", fset.Position(n.Pos()), s))
		}
		return true
	})

	return sb.String()
}
