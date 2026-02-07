package main

import "testing"

func TestSymbolify(t *testing.T) {
	got := symbolify("order summary")
	if got != "OrderSummary" {
		t.Fatalf("expected OrderSummary, got %s", got)
	}
}

func TestBuildScaffoldArgs(t *testing.T) {
	g := graph{
		Concepts: []concept{
			{
				Name:        "Order",
				Description: "Represents a customer order",
				Fields: []field{
					{Name: "Status", Type: "string"},
				},
				Methods: []method{{Name: "Approve", Invariant: "Status must be pending"}},
			},
		},
		Orchestrators: []orchestrator{
			{
				Name: "CreateOrder",
				Params: []field{
					{Name: "CustomerID", Type: "string"},
				},
				Route: &route{Method: "POST", Path: "/orders", Target: "CreateOrder"},
			},
		},
		Projections: []projection{
			{
				Name: "OrderSummary",
				Query: []field{
					{Name: "OrderID", Type: "string"},
				},
				Result: []field{
					{Name: "Status", Type: "string"},
				},
				Route: &route{Method: "GET", Path: "/views/order-summary", Target: "OrderSummary"},
			},
		},
	}

	args := buildScaffoldArgs(g)
	if len(args) == 0 {
		t.Fatalf("expected scaffold args")
	}
	assertContains(t, args, "--concept")
	assertContains(t, args, "Order")
	assertContains(t, args, "--orchestrator")
	assertContains(t, args, "CreateOrder")
	assertContains(t, args, "--projection")
	assertContains(t, args, "OrderSummary")
	assertContains(t, args, "--route")
}

func assertContains(t *testing.T, args []string, value string) {
	t.Helper()
	for _, arg := range args {
		if arg == value {
			return
		}
	}
	t.Fatalf("expected args to contain %s", value)
}
