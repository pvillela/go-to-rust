package main

import (
	"fmt"
	"go/ast"

	// "go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"strings"
)

func parse_source(src string) string {

	var sb strings.Builder

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	astFile, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		panic(err)
	}

	// Inspect the AST and print all identifiers and literals.
	ast.Inspect(astFile, func(n ast.Node) bool {
		var s string
		ret := true
		switch x := n.(type) {
		case *ast.TypeSpec:
			// s = fmt.Sprintf("*** struct %v", x.Name.Name)
			// ret = false
			var sb strings.Builder
			ret = type_spec(&sb, x)
			s += sb.String()
		case *ast.FuncDecl:
			var sb strings.Builder
			ret = func_decl(fset, &sb, x)
			s += sb.String()
		case *ast.DeclStmt:
			var sb strings.Builder
			printer.Fprint(&sb, fset, x)
			s += sb.String()
			// case *ast.Ident:
			// 	s = "Ident -->" + x.Name
		}
		if s != "" {
			// sb.WriteString(fmt.Sprintf("%s:\t%s\n", fset.Position(n.Pos()), s))
			sb.WriteString(fmt.Sprintf("%s\n\n", s))
		}
		return ret
	})

	return sb.String()
}

func map_type(typ string) string {
	switch typ {
	case "int":
		return "i64"
	case "uint":
		return "u64"
	case "string":
		return "String"
	case "&{time Time}":
		return "DateTime"
	}
	return typ
}

func pubify_name(name string) (pub, newName string) {
	name1 := name[0:1]
	if strings.ToUpper(name1) == name1 {
		pub = "pub "
		name1 = strings.ToLower(name1)
		newName = name1 + name[1:]
	}
	return
}

func type_spec(sb *strings.Builder, node *ast.TypeSpec) bool {
	sb.WriteString(fmt.Sprintf("pub struct %v", node.Name.Name))
	return struct_type(sb, node.Type.(*ast.StructType))
}

func struct_type(sb *strings.Builder, node *ast.StructType) bool {
	sb.WriteString(" {\n")
	for _, field := range node.Fields.List {
		pub, name := pubify_name(field.Names[0].Name)
		typ := map_type(fmt.Sprint(field.Type))
		sb.WriteString(fmt.Sprintf("%v %v: %v,\n", pub, name, typ))
	}
	sb.WriteString("}\n")
	return false
}

func func_decl(fset *token.FileSet, sb *strings.Builder, node *ast.FuncDecl) bool {
	// fn name
	{
		pub, name := pubify_name(node.Name.Name)
		sb.WriteString(fmt.Sprintf("%vfn %v", pub, name))
	}

	// Parameters
	{
		sb.WriteString("(\n")
		for _, field := range node.Type.Params.List {
			name := field.Names[0].Name
			typ := map_type(fmt.Sprint(field.Type))
			sb.WriteString(fmt.Sprintf("%v: %v,\n", name, typ))
		}
		sb.WriteString(") ")
	}

	// Return type
	{
		sb.WriteString(" -> (")
		first := true
		for _, field := range node.Type.Results.List {
			typ := map_type(fmt.Sprint(field.Type))
			if first {
				first = false
			} else {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%v", typ))
		}
		sb.WriteString(") ")
	}

	// Body
	{
		// sb.WriteString(" {\n")
		// for _, stmt := range node.Body.List {
		// 	sb.WriteString(fmt.Sprint(stmt, "\n"))
		// }
		// sb.WriteString("}\n")

		printer.Fprint(sb, fset, node.Body)

		// err := format.Node(sb, fset, node.Body)
		// if err != nil {
		// 	panic(err)
		// }
	}

	return false
}
