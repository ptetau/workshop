package main

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestPairwiseRuleCombinations(t *testing.T) {
	rules := []string{
		"naming",
		"concept-coupling",
		"route-query",
		"storage-isolation",
	}

	combos := generateAllCombos(len(rules))
	for idx, combo := range combos {
		root := t.TempDir()
		setupBaseRepo(t, root)
		applyViolations(t, root, combo)

		violations, err := lint(root)
		if err != nil {
			t.Fatalf("combo %d lint failed: %v", idx, err)
		}

		for i, rule := range rules {
			if combo[i] {
				assertHasRule(t, violations, rule)
			} else {
				assertNotHasRule(t, violations, rule)
			}
		}
	}
}

func setupBaseRepo(t *testing.T, root string) {
	t.Helper()

	writeFile(t, filepath.Join(root, "internal/domain/order/model.go"), `package order

type Order struct {
	ID string
}
`)

	writeFile(t, filepath.Join(root, "internal/application/orchestrators/create_order.go"), `package orchestrators

import "context"

type CreateOrderInput struct{}

func ExecuteCreateOrder(ctx context.Context, input CreateOrderInput) error { return nil }
`)

	writeFile(t, filepath.Join(root, "internal/application/projections/order_summary.go"), `package projections

import "context"

type OrderSummaryQuery struct{}
type OrderSummaryResult struct{}

func QueryOrderSummary(ctx context.Context, query OrderSummaryQuery) (OrderSummaryResult, error) {
	return OrderSummaryResult{}, nil
}
`)

	writeFile(t, filepath.Join(root, "internal/adapters/http/routes.go"), `package web

import (
	"encoding/json"
	"net/http"

	"workshop/internal/application/orchestrators"
	"workshop/internal/application/projections"
)

func handleGetOrderSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	query := projections.OrderSummaryQuery{}
	result, _ := projections.QueryOrderSummary(r.Context(), query)
	_ = json.NewEncoder(w).Encode(result)
}

func handlePostOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_ = orchestrators.ExecuteCreateOrder(r.Context(), orchestrators.CreateOrderInput{})
	w.WriteHeader(http.StatusNoContent)
}
`)

	writeFile(t, filepath.Join(root, "internal/adapters/storage/order/store.go"), `package order

type Store interface {}
`)
}

func applyViolations(t *testing.T, root string, combo []bool) {
	t.Helper()

	if combo[0] || combo[1] {
		importLine := ""
		extra := ""
		if combo[1] {
			importLine = `import "workshop/internal/application/orchestrators"

`
			extra = `
var _ = orchestrators.ExecuteCreateOrder
`
		}
		field := "ID string"
		if combo[0] {
			field = "ID    string\n\tUsrID string"
		}
		writeFile(t, filepath.Join(root, "internal/domain/order/model.go"), `package order

`+importLine+`type Order struct {
	`+field+`
}

`+extra+``)
	}

	if combo[2] {
		writeFile(t, filepath.Join(root, "internal/adapters/http/routes.go"), `package web

import (
	"net/http"

	"workshop/internal/application/orchestrators"
)

func handleGetOrderSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	_ = orchestrators.ExecuteCreateOrder(r.Context(), orchestrators.CreateOrderInput{})
}
`)
	}

	if combo[3] {
		writeFile(t, filepath.Join(root, "internal/adapters/storage/billing/store.go"), `package billing

type Store interface {}
`)
	}
}

func generateAllCombos(params int) [][]bool {
	if params <= 0 {
		return nil
	}
	total := 1 << params
	rows := make([][]bool, 0, total)
	for mask := 0; mask < total; mask++ {
		row := make([]bool, params)
		for i := 0; i < params; i++ {
			row[i] = mask&(1<<i) != 0
		}
		rows = append(rows, row)
	}
	return rows
}

func assertNotHasRule(t *testing.T, violations []violation, rule string) {
	t.Helper()
	for _, v := range violations {
		if v.Rule == rule {
			t.Fatalf("expected no rule %s; got %s", rule, joinRules(violations))
		}
	}
}

func joinRules(violations []violation) string {
	var rules []string
	for _, v := range violations {
		rules = append(rules, v.Rule)
	}
	return strings.Join(rules, ", ")
}
