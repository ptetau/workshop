package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLintFindsViolations(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "internal/domain/order/model.go"), `package order

import "workshop/internal/domain/customer"

type Order struct {
	UserID string
	UsrID  string
}

func (o *Order) Approve() {}
`)
	writeFile(t, filepath.Join(root, "internal/domain/customer/model.go"), `package customer

type Customer struct {
	ID string
}
`)
	writeFile(t, filepath.Join(root, "internal/adapters/http/routes.go"), `package web

import (
	"net/http"

	"workshop/internal/application/orchestrators"
)

func handleGetOrdersList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_ = orchestrators.ExecuteCreateOrder
}
`)
	writeFile(t, filepath.Join(root, "internal/adapters/storage/order/store.go"), `package storage

type Store interface {}
`)

	violations, err := lint(root)
	if err != nil {
		t.Fatalf("lint failed: %v", err)
	}

	assertHasRule(t, violations, "concept-coupling")
	assertHasRule(t, violations, "naming")
	assertHasRule(t, violations, "route-query")
	assertHasRule(t, violations, "storage-isolation")
}

// TestFeatureFlag_HandlerFileWithoutGate verifies that a handlers_*.go file
// defining handler functions without requireFeatureAPI/requireFeaturePage is flagged.
func TestFeatureFlag_HandlerFileWithoutGate(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "internal/adapters/http/handlers_widget.go"), `package web

import "net/http"

func handleWidgets(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
`)

	violations, err := lint(root)
	if err != nil {
		t.Fatalf("lint failed: %v", err)
	}
	assertHasRule(t, violations, "feature-flag")
}

// TestFeatureFlag_HandlerFileWithGate verifies that a handlers_*.go file
// that calls requireFeatureAPI is NOT flagged.
func TestFeatureFlag_HandlerFileWithGate(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "internal/adapters/http/handlers_widget.go"), `package web

import "net/http"

func handleWidgets(w http.ResponseWriter, r *http.Request) {
	if !requireFeatureAPI(w, r, sess, "widget") {
		return
	}
	w.WriteHeader(http.StatusOK)
}
`)

	violations, err := lint(root)
	if err != nil {
		t.Fatalf("lint failed: %v", err)
	}
	for _, v := range violations {
		if v.Rule == "feature-flag" {
			t.Fatalf("unexpected feature-flag violation: %s", v.Message)
		}
	}
}

// TestFeatureFlag_TestFilesExcluded verifies that handlers_*_test.go files
// are not checked for feature gating.
func TestFeatureFlag_TestFilesExcluded(t *testing.T) {
	root := t.TempDir()

	writeFile(t, filepath.Join(root, "internal/adapters/http/handlers_widget_test.go"), `package web

import "net/http"

func handleWidgetsTest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
`)

	violations, err := lint(root)
	if err != nil {
		t.Fatalf("lint failed: %v", err)
	}
	for _, v := range violations {
		if v.Rule == "feature-flag" {
			t.Fatalf("test files should not trigger feature-flag rule: %s", v.Message)
		}
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func assertHasRule(t *testing.T, violations []violation, rule string) {
	t.Helper()
	for _, v := range violations {
		if v.Rule == rule {
			return
		}
	}
	var rules []string
	for _, v := range violations {
		rules = append(rules, v.Rule)
	}
	t.Fatalf("expected rule %s; got %s", rule, strings.Join(rules, ", "))
}
