package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
)

// UpdateStructFields parses src, finds structName, and adds any missing fields.
// Returns the updated source code.
func UpdateStructFields(src []byte, structName string, newFields []field) ([]byte, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	found := false
	ast.Inspect(node, func(n ast.Node) bool {
		if found {
			return false
		}
		// Find TypeSpec
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok || typeSpec.Name.Name != structName {
			return true
		}
		// Find StructType
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return true
		}

		found = true

		// Map existing fields
		existing := map[string]bool{}
		if structType.Fields != nil {
			for _, f := range structType.Fields.List {
				for _, name := range f.Names {
					existing[name.Name] = true
				}
			}
		}

		// Add missing fields
		for _, nf := range newFields {
			if existing[nf.Name] {
				continue
			}

			// Create new AST field
			newField := &ast.Field{
				Names: []*ast.Ident{ast.NewIdent(nf.Name)},
				Type:  ast.NewIdent(nf.Type),
				// We don't verify tags here for simplicity, or we could add them?
				// Scaffold usually doesn't add tags to Domain models (clean arch),
				// but Orchestrator inputs might need json tags.
				// For now, simple Name Type.
			}

			// Add to list
			if structType.Fields == nil {
				structType.Fields = &ast.FieldList{}
			}
			structType.Fields.List = append(structType.Fields.List, newField)
		}
		return false
	})

	if !found {
		return nil, fmt.Errorf("struct %s not found in source", structName)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
