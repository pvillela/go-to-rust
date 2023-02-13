package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
)

func nodeTree(src string) string {

	var sb strings.Builder

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		panic(err)
	}

	ast.Fprint(&sb, fset, f, ast.NotNilFilter)

	return sb.String()
}
