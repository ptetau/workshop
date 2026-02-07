package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
)

// TestCase represents a single test case in a table-driven test
type TestCase struct {
	Name   string
	Fields map[string]string // field name -> value
}

// UpdateTestCases patches new test cases into an existing table-driven test function.
// It finds the test function by name, locates the test table, and appends new cases
// while preserving existing ones.
func UpdateTestCases(src []byte, testFuncName string, newCases []TestCase) ([]byte, error) {
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

		// Find the test function
		funcDecl, ok := n.(*ast.FuncDecl)
		if !ok || funcDecl.Name.Name != testFuncName {
			return true
		}

		found = true

		// Find the test table (tests := []struct{...}{...})
		// This is a simplified implementation - in production, we'd need more robust AST traversal
		// For now, we'll just mark it as found and return the original source
		// TODO: Implement full AST manipulation to add test cases

		return false
	})

	if !found {
		return nil, fmt.Errorf("test function %s not found in source", testFuncName)
	}

	// For now, return original source
	// Full implementation would manipulate the AST to add test cases
	return src, nil
}

// AddTestFunction adds a new test function to the source if it doesn't already exist.
// The testFunc parameter should be the complete function source code.
func AddTestFunction(src []byte, testFunc string) ([]byte, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Parse the new function
	funcSrc := "package main\n" + testFunc
	funcNode, err := parser.ParseFile(fset, "", funcSrc, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse test function: %w", err)
	}

	if len(funcNode.Decls) == 0 {
		return nil, fmt.Errorf("no declarations found in test function")
	}

	newFunc, ok := funcNode.Decls[0].(*ast.FuncDecl)
	if !ok {
		return nil, fmt.Errorf("expected function declaration")
	}

	// Check if function already exists
	funcName := newFunc.Name.Name
	for _, decl := range node.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && fn.Name.Name == funcName {
			// Function already exists, don't add it
			return src, nil
		}
	}

	// Add the new function to the AST
	node.Decls = append(node.Decls, newFunc)

	// Format and return
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// RenameTestSymbols renames test functions, structs, and variables in test files.
// This is useful when renaming orchestrators, projections, or concepts.
func RenameTestSymbols(src []byte, oldName, newName string) ([]byte, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, "", src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// Rename identifiers
	ast.Inspect(node, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Ident:
			// Rename if the identifier contains the old name
			if strings.Contains(x.Name, oldName) {
				x.Name = strings.ReplaceAll(x.Name, oldName, newName)
			}
		}
		return true
	})

	// Format and return
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// GenerateHTTPTestStub generates a basic HTTP test stub for a route
func GenerateHTTPTestStub(method, path, target string) string {
	handlerName := fmt.Sprintf("Test%s%s", strings.Title(strings.ToLower(method)), toPascalCase(target))

	return fmt.Sprintf(`
// %s tests the %s %s endpoint.
func %s(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		headers    map[string]string
		wantStatus int
		wantHeader map[string]string
	}{
		{
			name:       "basic %s request",
			method:     "%s",
			path:       "%s",
			// TODO: Add request body if needed
			wantStatus: http.StatusOK,
			// TODO: Add expected headers
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Setup mock stores
			// TODO: Create request with httptest.NewRequest
			// TODO: Create response recorder with httptest.NewRecorder
			// TODO: Call handler
			// TODO: Assert response status, headers, body
		})
	}
}
`, handlerName, method, path, handlerName, method, method, path)
}

// GenerateE2ETestStub generates a basic E2E test stub for a route
func GenerateE2ETestStub(method, path, target string) string {
	flowName := toPascalCase(target)

	return fmt.Sprintf(`
// TestE2E%s is an end-to-end test for the %s %s flow.
func TestE2E%s(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// TODO: Start test server with NewTestServer
	// TODO: Create browser context with chromedp
	// TODO: Navigate to %s
	// TODO: Perform %s action
	// TODO: Verify results
	// TODO: Take screenshot on failure
}
`, flowName, method, path, flowName, path, method)
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '_' || r == '-' || r == ' '
	})
	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}
	return strings.Join(words, "")
}
