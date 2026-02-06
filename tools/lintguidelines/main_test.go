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
