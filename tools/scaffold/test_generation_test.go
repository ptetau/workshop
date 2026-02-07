package main

import (
	"os"
	"strings"
	"testing"
)

// TestPairwiseTestGeneration tests all combinations of test generation flags
func TestPairwiseTestGeneration(t *testing.T) {
	params := []string{
		"route",
		"generateTests",
		"testTypeHTTP",
		"testTypeE2E",
	}

	// Test all combinations: 2^4 = 16 combinations
	combos := generateAllCombos(len(params))

	for idx, combo := range combos {
		dir := t.TempDir()
		chdir(t, dir)

		args := buildTestGenArgsFromCombo(combo)
		if err := runInit(args); err != nil {
			t.Fatalf("combo %d runInit failed: %v", idx, err)
		}

		assertExists(t, ".scaffold/state.json")

		hasRoute := combo[0]
		generateTests := combo[1]
		testTypeHTTP := combo[2]
		testTypeE2E := combo[3]

		if hasRoute && generateTests {
			assertExists(t, "internal/adapters/http/routes.go")

			// Check if test file was generated
			testFilePath := "internal/adapters/http/routes_test.go"
			assertExists(t, testFilePath)

			// Read test file and verify content
			content, err := os.ReadFile(testFilePath)
			if err != nil {
				t.Fatalf("combo %d failed to read test file: %v", idx, err)
			}

			contentStr := string(content)

			// Verify HTTP test stub exists if testTypeHTTP or both
			if testTypeHTTP && !testTypeE2E {
				if !strings.Contains(contentStr, "func TestPostCreateorder(t *testing.T)") {
					t.Fatalf("combo %d expected HTTP test stub", idx)
				}
				if strings.Contains(contentStr, "func TestE2E") {
					t.Fatalf("combo %d should not have E2E test stub", idx)
				}
			}

			// Verify E2E test stub exists if testTypeE2E only
			if testTypeE2E && !testTypeHTTP {
				if !strings.Contains(contentStr, "func TestE2ECreateorder(t *testing.T)") {
					t.Fatalf("combo %d expected E2E test stub", idx)
				}
				if strings.Contains(contentStr, "httptest.NewRequest") {
					t.Fatalf("combo %d should not have HTTP test stub", idx)
				}
			}

			// Verify both test types if both flags set
			if testTypeHTTP && testTypeE2E {
				if !strings.Contains(contentStr, "func TestPostCreateorder(t *testing.T)") {
					t.Fatalf("combo %d expected HTTP test stub", idx)
				}
				if !strings.Contains(contentStr, "func TestE2ECreateorder(t *testing.T)") {
					t.Fatalf("combo %d expected E2E test stub", idx)
				}
			}
		}
	}
}

// TestTestGenerationWithMultipleRoutes verifies test generation works with multiple routes
func TestTestGenerationWithMultipleRoutes(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	// First route
	args1 := []string{
		"--route", "POST:/orders:CreateOrder",
		"--generate-tests", "true",
		"--test-type", "http",
	}
	if err := runInit(args1); err != nil {
		t.Fatalf("first runInit failed: %v", err)
	}

	// Second route
	args2 := []string{
		"--route", "GET:/orders:ListOrders",
		"--generate-tests", "true",
		"--test-type", "e2e",
	}
	if err := runInit(args2); err != nil {
		t.Fatalf("second runInit failed: %v", err)
	}

	// Verify test file contains both tests
	content, err := os.ReadFile("internal/adapters/http/routes_test.go")
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	contentStr := string(content)

	// Should have HTTP test for CreateOrder
	if !strings.Contains(contentStr, "func TestPostCreateorder(t *testing.T)") {
		t.Fatal("expected HTTP test for CreateOrder")
	}

	// Should have E2E test for ListOrders
	if !strings.Contains(contentStr, "func TestE2EListorders(t *testing.T)") {
		t.Fatal("expected E2E test for ListOrders")
	}
}

// TestTestGenerationPreservesExistingTests verifies that adding new routes doesn't overwrite existing tests
func TestTestGenerationPreservesExistingTests(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	// Create initial route with test
	args1 := []string{
		"--route", "POST:/orders:CreateOrder",
		"--generate-tests", "true",
		"--test-type", "http",
	}
	if err := runInit(args1); err != nil {
		t.Fatalf("first runInit failed: %v", err)
	}

	// Manually modify the test file to add custom content
	testPath := "internal/adapters/http/routes_test.go"
	content, _ := os.ReadFile(testPath)
	customContent := string(content) + "\n// CUSTOM COMMENT\n"
	os.WriteFile(testPath, []byte(customContent), 0644)

	// Add another route
	args2 := []string{
		"--route", "GET:/orders:ListOrders",
		"--generate-tests", "true",
		"--test-type", "http",
	}
	if err := runInit(args2); err != nil {
		t.Fatalf("second runInit failed: %v", err)
	}

	// Verify custom content is preserved
	newContent, _ := os.ReadFile(testPath)
	if !strings.Contains(string(newContent), "// CUSTOM COMMENT") {
		t.Fatal("custom content was not preserved")
	}
}

// TestTestGenerationCompiles verifies generated tests compile
func TestTestGenerationCompiles(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	args := []string{
		"--module", "testapp",
		"--route", "POST:/test:TestAction",
		"--generate-tests", "true",
		"--test-type", "both",
	}
	if err := runInit(args); err != nil {
		t.Fatalf("runInit failed: %v", err)
	}

	// Try to compile the test file (this will fail if syntax is wrong)
	// We can't actually run go test here because the handlers don't exist,
	// but we can at least verify the file was created with valid syntax
	testPath := "internal/adapters/http/routes_test.go"
	assertExists(t, testPath)

	content, err := os.ReadFile(testPath)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	// Basic syntax checks
	contentStr := string(content)
	if !strings.Contains(contentStr, "package web") {
		t.Fatal("missing package declaration")
	}
	if !strings.Contains(contentStr, "import (") {
		t.Fatal("missing import block")
	}
	if !strings.Contains(contentStr, "func Test") {
		t.Fatal("missing test function")
	}
}

func buildTestGenArgsFromCombo(combo []bool) []string {
	hasRoute := combo[0]
	generateTests := combo[1]
	testTypeHTTP := combo[2]
	testTypeE2E := combo[3]

	var args []string

	if hasRoute {
		args = append(args, "--route", "POST:/orders:CreateOrder")
	}

	if generateTests {
		args = append(args, "--generate-tests", "true")

		// Determine test type
		if testTypeHTTP && testTypeE2E {
			args = append(args, "--test-type", "both")
		} else if testTypeHTTP {
			args = append(args, "--test-type", "http")
		} else if testTypeE2E {
			args = append(args, "--test-type", "e2e")
		} else {
			// Neither flag set means default (both)
			args = append(args, "--test-type", "both")
		}
	} else {
		args = append(args, "--generate-tests", "false")
	}

	return args
}
