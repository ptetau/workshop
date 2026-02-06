package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInitScaffoldsArtifacts(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	args := []string{
		"--concept", "Order",
		"--field", "Order:Status:string",
		"--method", "Order:Approve",
		"--orchestrator", "CreateOrder",
		"--param", "CreateOrder:CustomerID:string",
		"--projection", "OrderSummary",
		"--query", "OrderSummary:OrderID:string",
		"--result", "OrderSummary:Status:string",
		"--route", "GET:/views/order-summary:OrderSummary",
		"--route", "POST:/orders:CreateOrder",
	}

	if err := runInit(args); err != nil {
		t.Fatalf("runInit failed: %v", err)
	}

	assertExists(t, "internal/domain/order/model.go")
	assertExists(t, "internal/adapters/storage/order/store.go")
	assertExists(t, "internal/application/orchestrators/create_order.go")
	assertExists(t, "internal/application/projections/order_summary.go")
	assertExists(t, "internal/adapters/http/routes.go")
	assertExists(t, ".scaffold/state.json")

	content := readFile(t, "internal/adapters/http/routes.go")
	if !strings.Contains(content, "handleGetViewsOrderSummaryOrderSummary") {
		t.Fatalf("expected GET handler in routes")
	}
	if !strings.Contains(content, "handlePostOrdersCreateOrder") {
		t.Fatalf("expected POST handler in routes")
	}
}

func TestInitAddsMigrationOnNewField(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	if err := runInit([]string{"--concept", "Order", "--field", "Order:Status:string"}); err != nil {
		t.Fatalf("runInit failed: %v", err)
	}
	if err := runInit([]string{"--concept", "Order", "--field", "Order:TotalCents:int"}); err != nil {
		t.Fatalf("runInit failed: %v", err)
	}

	createFiles, _ := filepath.Glob("internal/adapters/storage/migrations/*_create_order.sql")
	alterFiles, _ := filepath.Glob("internal/adapters/storage/migrations/*_alter_order.sql")
	if len(createFiles) == 0 {
		t.Fatalf("expected create migration")
	}
	if len(alterFiles) == 0 {
		t.Fatalf("expected alter migration")
	}
}

func TestRoutesRegenerateWhenGenerated(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	if err := runInit([]string{"--route", "GET:/views/order-summary:OrderSummary"}); err != nil {
		t.Fatalf("runInit failed: %v", err)
	}
	if err := runInit([]string{"--route", "GET:/views/order-detail:OrderDetail"}); err != nil {
		t.Fatalf("runInit failed: %v", err)
	}

	content := readFile(t, "internal/adapters/http/routes.go")
	if !strings.Contains(content, "handleGetViewsOrderDetailOrderDetail") {
		t.Fatalf("expected second route to be present")
	}
}

func TestScaffoldedAppCompiles(t *testing.T) {
	dir := t.TempDir()
	chdir(t, dir)

	if err := runInit([]string{"--concept", "Order", "--field", "Order:Status:string"}); err != nil {
		t.Fatalf("runInit failed: %v", err)
	}

	cmd := exec.Command("go", "test", "./...")
	cmd.Env = append(os.Environ(), "GOWORK=off")
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go test failed: %v\n%s", err, string(output))
	}
}

func chdir(t *testing.T, dir string) {
	t.Helper()
	old, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(old)
	})
}

func assertExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist: %v", path, err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}
