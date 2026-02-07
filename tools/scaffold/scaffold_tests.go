package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// scaffoldTests generates test files for routes
func scaffoldTests(moduleName string, routes []route, testType string, force bool) error {
	if testType == "" || testType == "none" {
		return nil
	}

	testPath := filepath.Join("internal/adapters/http", "routes_test.go")

	// Check if test file exists
	var existingSrc []byte
	var err error
	if _, err := os.Stat(testPath); err == nil {
		existingSrc, err = os.ReadFile(testPath)
		if err != nil {
			return err
		}
	}

	// Generate test content
	var testFuncs []string

	for _, r := range routes {
		// Generate HTTP test if requested
		if testType == "http" || testType == "both" {
			httpTest := GenerateHTTPTestStub(r.Method, r.Path, r.Target)
			testFuncs = append(testFuncs, httpTest)
		}

		// Generate E2E test if requested
		if testType == "e2e" || testType == "both" {
			e2eTest := GenerateE2ETestStub(r.Method, r.Path, r.Target)
			testFuncs = append(testFuncs, e2eTest)
		}
	}

	if len(testFuncs) == 0 {
		return nil
	}

	// If file exists, try to add functions
	if len(existingSrc) > 0 {
		updated := existingSrc
		for _, testFunc := range testFuncs {
			updated, err = AddTestFunction(updated, testFunc)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to add test function: %v\n", err)
				continue
			}
		}

		// Only write if something changed
		if string(updated) != string(existingSrc) {
			return os.WriteFile(testPath, updated, 0644)
		}
		return nil
	}

	// Create new test file
	content := generateTestFileHeader(moduleName) + strings.Join(testFuncs, "\n\n")
	return os.WriteFile(testPath, []byte(content), 0644)
}

func generateTestFileHeader(moduleName string) string {
	return fmt.Sprintf(`package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Auto-generated test stubs by scaffold
// TODO: Implement test logic

`)
}
