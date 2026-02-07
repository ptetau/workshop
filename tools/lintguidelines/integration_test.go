package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestScaffoldedAppPassesLint(t *testing.T) {
	repoRoot := findRepoRoot(t)
	scaffoldDir := filepath.Join(repoRoot, "tools", "scaffold")
	workDir := t.TempDir()

	runGo(t, repoRoot, scaffoldDir, "run", ".", "init",
		"--root", workDir,
		"--concept", "Order",
		"--field", "Order:Status:string",
		"--method", "Order:Approve",
		"--post", "Order:Approve:Order is approved",
		"--orchestrator", "CreateOrder",
		"--param", "CreateOrder:CustomerID:string",
		"--projection", "OrderSummary",
		"--query", "OrderSummary:OrderID:string",
		"--result", "OrderSummary:Status:string",
		"--route", "GET:/views/order-summary:OrderSummary",
		"--route", "POST:/orders:CreateOrder",
	)

	runGo(t, repoRoot, repoRoot, "run", "./tools/lintguidelines", "--root", workDir, "--strict")
}

func TestScaffoldedAppWithMultipleConceptsPassesLint(t *testing.T) {
	repoRoot := findRepoRoot(t)
	scaffoldDir := filepath.Join(repoRoot, "tools", "scaffold")
	workDir := t.TempDir()

	runGo(t, repoRoot, scaffoldDir, "run", ".", "init",
		"--root", workDir,
		"--concept", "Order",
		"--field", "Order:Status:string",
		"--concept", "InventoryItem",
		"--field", "InventoryItem:SKU:string",
		"--orchestrator", "ReserveInventory",
		"--param", "ReserveInventory:InventoryItemID:string",
		"--projection", "InventoryStatus",
		"--query", "InventoryStatus:InventoryItemID:string",
		"--result", "InventoryStatus:OnHand:int",
		"--route", "GET:/views/inventory-status:InventoryStatus",
		"--route", "POST:/inventory/reserve:ReserveInventory",
	)

	runGo(t, repoRoot, repoRoot, "run", "./tools/lintguidelines", "--root", workDir, "--strict")
}

func runGo(t *testing.T, repoRoot, workDir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", args...)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), "GOWORK=off")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go %s failed: %v\n%s", args, err, string(output))
	}
}

func findRepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}
