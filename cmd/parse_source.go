package main

import (
	"fmt"
	"go/ast"

	// "go/format"
	"go/parser"
	"go/printer"
	"go/token"
	"regexp"
	"strings"
)

// pass 1 rewrites identifiers in AST:
// * Preserve TypeSpec identifiers
// * Prefix FuncDecl Name identifiers with "P__" if exported, transform the rest of the identifier to snake_case.
// * Prefix Field Name identifiers with "P__" if exported, transform the rest of the identifier to snake_case.
// * Preserve Field Type identifiers
// * Transform all other identifiers to snake_case.
func pass1(n ast.Node) bool {
	switch x := n.(type) {
	case *ast.TypeSpec:
		// Leave Name as is and recurse
		if x.Type != nil {
			ast.Inspect(x.Type, pass1)
		}
		return false
	case *ast.FuncDecl:
		name := x.Name
		name.Name = camelToSnake(name.Name)
		if name.IsExported() {
			name.Name = "P__" + name.Name
		}
		if x.Recv != nil {
			ast.Inspect(x.Recv, pass1)
		}
		if x.Type != nil {
			ast.Inspect(x.Type, pass1)
		}
		if x.Body != nil {
			ast.Inspect(x.Body, pass1)
		}
		return false
	case *ast.Field:
		names := x.Names
		for _, name := range names {
			name.Name = camelToSnake(name.Name)
			if name.IsExported() {
				name.Name = "P__" + name.Name
			}
		}
		// Leave Type as is
		return false
	case *ast.ValueSpec: // same as for Field
		names := x.Names
		for _, name := range names {
			name.Name = camelToSnake(name.Name)
			if name.IsExported() {
				name.Name = "P__" + name.Name
			}
		}
		// Leave Type as is
		for _, value := range x.Values {
			ast.Inspect(value, pass1)
		}
		return false
	case *ast.Ident:
		x.Name = camelToSnake(x.Name)
		fmt.Println("***", x.Name)
		return true // doesn't matter
	}

	// Continue inspecting for all other node types.
	return true
}

// parseToRust parses a string containing Go source code and returns a string containing
// a rough equivalent Rust code.
func parseToRust(src string) string {
	var sb strings.Builder

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	astFile, err := parser.ParseFile(fset, "src.go", src, 0)
	if err != nil {
		panic(err)
	}

	// pass 2 generates code:
	// * Generate structs
	//   - Always make them pub
	//   - Process fields:
	//     . If the field begins with "P__", delete the prefix and generate "pub " in front of the field name
	// * Generate functions
	//   - If the FuncDecl name identifier begins with "P__", delete the prefix and generate "pub " in front of the field name
	// * Generate other nodes by simply printing them
	pass2 := func(n ast.Node) bool {
		var s string
		ret := true
		switch x := n.(type) {
		case *ast.TypeSpec:
			// s = fmt.Sprintf("*** struct %v", x.Name.Name)
			// ret = false
			var sb strings.Builder
			ret = type_spec(fset, &sb, x)
			s += sb.String()
		case *ast.FuncDecl:
			var sb strings.Builder
			ret = func_decl(fset, &sb, x)
			s += sb.String()
		case *ast.ValueSpec:
			var sb strings.Builder
			printer.Fprint(&sb, fset, x)
			s += sb.String()
			ret = false
		case *ast.DeclStmt:
			var sb strings.Builder
			printer.Fprint(&sb, fset, x)
			s += sb.String()
			// case *ast.Ident:
			// 	s = "Ident -->" + x.Name

			// case ast.Stmt:
			// 	var sb strings.Builder
			// 	printer.Fprint(&sb, fset, x)
			// 	s += sb.String()
			// 	ret = false
		}
		if s != "" {
			// sb.WriteString(fmt.Sprintf("%s:\t%s\n", fset.Position(n.Pos()), s))
			sb.WriteString(fmt.Sprintf("%s\n\n", s))
		}
		return ret
	}

	ast.Inspect(astFile, pass1)
	ast.Inspect(astFile, pass2)

	return sb.String()
}

// mapType maps Go types to Rust types
func mapType(typ string) string {
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

// pubifyName transforms Go names to Rust names, indicating when they are public
func pubifyName(name string) (pub, newName string) {
	name1 := name[0:1]
	if strings.ToUpper(name1) == name1 {
		pub = "pub "
		name1 = strings.ToLower(name1)
		newName = name1 + name[1:]
	}
	return
}

// type_spec transforms an *ast.TypeSpec node
func type_spec(fset *token.FileSet, sb *strings.Builder, node *ast.TypeSpec) bool {
	sb.WriteString(fmt.Sprintf("pub struct %v", node.Name.Name))
	typ, ok := node.Type.(*ast.StructType)

	if ok {
		struct_type(sb, typ)
	} else {
		var bodySb strings.Builder
		printer.Fprint(&bodySb, fset, node.Type)
		sb.Write([]byte(bodySb.String()))
	}
	return false
}

// struct_type transforms an *ast.StructType node
func struct_type(sb *strings.Builder, node *ast.StructType) bool {
	sb.WriteString(" {\n")
	for _, field := range node.Fields.List {
		pub, name := pubifyName(field.Names[0].Name)
		typ := mapType(fmt.Sprint(field.Type))
		sb.WriteString(fmt.Sprintf("%v %v: %v,\n", pub, name, typ))
	}
	sb.WriteString("}\n")
	return false
}

// func_decl transforms an *ast.FuncDecl node
func func_decl(fset *token.FileSet, sb *strings.Builder, node *ast.FuncDecl) bool {
	// fn name
	{
		pub, name := pubifyName(node.Name.Name)
		sb.WriteString(fmt.Sprintf("%vfn %v", pub, name))
	}

	// Parameters
	{
		sb.WriteString("(\n")
		for _, field := range node.Type.Params.List {
			name := field.Names[0].Name
			typ := mapType(fmt.Sprint(field.Type))
			sb.WriteString(fmt.Sprintf("%v: %v,\n", name, typ))
		}
		sb.WriteString(") ")
	}

	// Return type
	{
		sb.WriteString("-> ")
		first := true
		for _, field := range node.Type.Results.List {
			typ := mapType(fmt.Sprint(field.Type))
			if first {
				first = false
			} else {
				sb.WriteString(", ")
			}
			sb.WriteString(fmt.Sprintf("%v", typ))
		}
		sb.WriteString(" ")
	}

	// Body
	{
		var bodySb strings.Builder
		printer.Fprint(&bodySb, fset, node.Body)
		sb.Write([]byte(bodySb.String()))
	}

	return false
}

// Provide by ChatGPT
func camelToSnake(s string) string {
	// Use a regular expression to find all upper case characters in the string
	rx := regexp.MustCompile("[A-Z]")
	positions := rx.FindAllStringIndex(s, -1)

	// Iterate through the positions and add an underscore before each upper case character
	result := ""
	lastIndex := 0
	for _, pos := range positions {
		result += s[lastIndex:pos[0]] + "_" + strings.ToLower(s[pos[0]:pos[1]])
		lastIndex = pos[1]
	}
	result += s[lastIndex:]

	return strings.ToLower(result)
}
