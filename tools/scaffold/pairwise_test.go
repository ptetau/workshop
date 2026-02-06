package main

import (
	"path/filepath"
	"testing"
)

func TestPairwiseFlagCombinations(t *testing.T) {
	params := []string{
		"concept",
		"field",
		"method",
		"orchestrator",
		"param",
		"projection",
		"query",
		"result",
		"route",
	}

	// Exhaustive coverage is stronger than pairwise and keeps behavior predictable.
	combos := generateAllCombos(len(params))
	if !allPairsCovered(combos, len(params)) {
		t.Fatalf("pairwise coverage not achieved")
	}

	for idx, combo := range combos {
		dir := t.TempDir()
		chdir(t, dir)

		args := buildArgsFromCombo(combo)
		if err := runInit(args); err != nil {
			t.Fatalf("combo %d runInit failed: %v", idx, err)
		}

		assertExists(t, ".scaffold/state.json")

		if combo[0] || combo[1] || combo[2] {
			assertExists(t, "internal/domain/order/model.go")
		}
		if combo[3] || combo[4] {
			assertExists(t, "internal/application/orchestrators/create_order.go")
		}
		if combo[5] || combo[6] || combo[7] {
			assertExists(t, "internal/application/projections/order_summary.go")
		}
		if combo[8] {
			assertExists(t, "internal/adapters/http/routes.go")
		}

		migrationFiles, _ := filepath.Glob("internal/adapters/storage/migrations/*_create_order.sql")
		if combo[1] && len(migrationFiles) == 0 {
			t.Fatalf("combo %d expected migration", idx)
		}
	}
}

func buildArgsFromCombo(combo []bool) []string {
	var args []string

	if combo[0] {
		args = append(args, "--concept", "Order")
	}
	if combo[1] {
		args = append(args, "--field", "Order:Status:string")
	}
	if combo[2] {
		args = append(args, "--method", "Order:Approve")
	}
	if combo[3] {
		args = append(args, "--orchestrator", "CreateOrder")
	}
	if combo[4] {
		args = append(args, "--param", "CreateOrder:CustomerID:string")
	}
	if combo[5] {
		args = append(args, "--projection", "OrderSummary")
	}
	if combo[6] {
		args = append(args, "--query", "OrderSummary:OrderID:string")
	}
	if combo[7] {
		args = append(args, "--result", "OrderSummary:Status:string")
	}
	if combo[8] {
		if combo[5] || combo[6] || combo[7] {
			args = append(args, "--route", "GET:/views/order-summary:OrderSummary")
		} else {
			args = append(args, "--route", "POST:/orders:CreateOrder")
		}
	}

	return args
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

func allPairsCovered(rows [][]bool, params int) bool {
	for i := 0; i < params; i++ {
		for j := i + 1; j < params; j++ {
			for _, vi := range []bool{false, true} {
				for _, vj := range []bool{false, true} {
					found := false
					for _, row := range rows {
						if row[i] == vi && row[j] == vj {
							found = true
							break
						}
					}
					if !found {
						return false
					}
				}
			}
		}
	}
	return true
}
