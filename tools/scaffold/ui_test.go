package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestUiPairwise runs the scaffold tool with various combinations of UI elements
// to ensure Views and Forms are correctly generated.
func TestUiPairwise(t *testing.T) {
	// Pairwise dimensions:
	// 1. Projection with 0 fields
	// 2. Projection with simple fields
	// 3. Orchestrator with 0 params
	// 4. Orchestrator with simple params
	// 5. Orchestrator with mixed params (int/bool/string)

	cases := []struct {
		name          string
		args          []string
		expectedFiles []string
		checkContent  map[string][]string // file -> substrings
	}{
		{
			name: "SimpleProjection_View",
			args: []string{
				"init",
				"--module", "workshop",
				"--projection", "SimpleView",
				"--query", "SimpleView:ID:string",
				"--result", "SimpleView:ID:string",
				"--route", "GET:/view:SimpleView",
			},
			expectedFiles: []string{
				"internal/adapters/http/templates/layout.html",
				"internal/adapters/http/templates/simple_view.html",
			},
			checkContent: map[string][]string{
				"internal/adapters/http/templates/simple_view.html": {
					"<h1>SimpleView</h1>",
					"{{ .ID }}",
				},
			},
		},
		{
			name: "ComplexForm_Orchestrator",
			args: []string{
				"init",
				"--module", "workshop",
				"--orchestrator", "SubmitForm",
				"--param", "SubmitForm:Name:string",
				"--param", "SubmitForm:Age:int",
				"--param", "SubmitForm:IsActive:bool",
				"--route", "POST:/submit:SubmitForm",
			},
			expectedFiles: []string{
				"internal/adapters/http/templates/layout.html",
				"internal/adapters/http/templates/form_submit_form.html",
			},
			checkContent: map[string][]string{
				"internal/adapters/http/templates/form_submit_form.html": {
					"<form method=\"POST\">",
					"name=\"gorilla.csrf.Token\"",
					"type=\"text\"",     // Name
					"type=\"number\"",   // Age
					"type=\"checkbox\"", // IsActive
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Setup temp dir
			workDir, err := os.MkdirTemp("", "ui_pairwise_*")
			if err != nil {
				t.Fatalf("failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(workDir)

			// Run scaffold
			// We need to pass --root to ensure it runs in the temp dir
			args := append(tc.args, "--root", workDir)

			if tc.args[0] == "init" {
				// runInit expects args AFTER "init"
				// args is ["init", ..., "--root", dir]
				// So passing args[1:] gave ["...", "--root", dir]
				if err := runInit(args[1:]); err != nil {
					t.Fatalf("runInit failed: %v", err)
				}
			}

			// Verify Files
			for _, f := range tc.expectedFiles {
				path := filepath.Join(workDir, f)
				if _, err := os.Stat(path); os.IsNotExist(err) {
					t.Errorf("Expected file missing: %s", f)
				}
			}

			// Verify Content
			for f, substrings := range tc.checkContent {
				path := filepath.Join(workDir, f)
				content, err := os.ReadFile(path)
				if err != nil {
					t.Errorf("Failed to read file %s: %v", f, err)
					continue
				}
				sContent := string(content)
				for _, sub := range substrings {
					if !strings.Contains(sContent, sub) {
						t.Errorf("File %s missing substring: %s\nACTUAL CONTENT:\n%s\n", f, sub, sContent)
					}
				}
			}
		})
	}
}
