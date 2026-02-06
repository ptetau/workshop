package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPairwiseInterviewCombinations(t *testing.T) {
	params := []string{
		"concept",
		"field",
		"method",
		"orchestrator",
		"param",
		"projection",
		"queryResult",
		"route",
	}

	combos := generateAllCombos(len(params))
	for idx, combo := range combos {
		input := buildInterviewInput(combo)
		reader := strings.NewReader(input)
		var out bytes.Buffer
		iv := newInterviewer(reader, &out, 0.8)

		g, err := iv.run("prd")
		if err != nil {
			t.Fatalf("combo %d run failed: %v", idx, err)
		}

		assertComboGraph(t, idx, combo, g)
		args := buildScaffoldArgs(g)
		assertComboArgs(t, idx, combo, args)
	}
}

func TestDisambiguationYesMergesConcepts(t *testing.T) {
	input := strings.Join([]string{
		"Order",
		"done", // fields
		"done", // methods
		"Orderr",
		"yes",
		"done",
		"done", // orchestrators
		"done", // projections
	}, "\n")
	reader := strings.NewReader(input)
	var out bytes.Buffer
	iv := newInterviewer(reader, &out, 0.8)

	g, err := iv.run("prd")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(g.Concepts) != 1 {
		t.Fatalf("expected 1 concept, got %d", len(g.Concepts))
	}
}

func TestDisambiguationNoAddsConcept(t *testing.T) {
	input := strings.Join([]string{
		"Order",
		"done", // fields
		"done", // methods
		"Orderr",
		"no",
		"done", // fields
		"done", // methods
		"done",
		"done", // orchestrators
		"done", // projections
	}, "\n")
	reader := strings.NewReader(input)
	var out bytes.Buffer
	iv := newInterviewer(reader, &out, 0.8)

	g, err := iv.run("prd")
	if err != nil {
		t.Fatalf("run failed: %v", err)
	}
	if len(g.Concepts) != 2 {
		t.Fatalf("expected 2 concepts, got %d", len(g.Concepts))
	}
}

func buildInterviewInput(combo []bool) string {
	hasConcept := combo[0]
	hasField := combo[1]
	hasMethod := combo[2]
	hasOrchestrator := combo[3]
	hasParam := combo[4]
	hasProjection := combo[5]
	hasQueryResult := combo[6]
	hasRoute := combo[7]

	lines := []string{}

	// Concepts
	if hasConcept {
		lines = append(lines, "Order")
		if hasField {
			lines = append(lines, "Status", "string", "done")
		} else {
			lines = append(lines, "done")
		}
		if hasMethod {
			lines = append(lines, "Approve", "done")
		} else {
			lines = append(lines, "done")
		}
		lines = append(lines, "done")
	} else {
		lines = append(lines, "done")
	}

	// Orchestrators
	if hasOrchestrator {
		lines = append(lines, "Create Order")
		if hasParam {
			lines = append(lines, "CustomerID", "string", "done")
		} else {
			lines = append(lines, "done")
		}
		if hasRoute {
			lines = append(lines, "POST", "/orders")
		} else {
			lines = append(lines, "skip")
		}
		lines = append(lines, "done")
	} else {
		lines = append(lines, "done")
	}

	// Projections
	if hasProjection {
		lines = append(lines, "Order Summary")
		if hasQueryResult {
			lines = append(lines, "OrderID", "string", "done")
		} else {
			lines = append(lines, "done")
		}
		if hasQueryResult {
			lines = append(lines, "Status", "string", "done")
		} else {
			lines = append(lines, "done")
		}
		if hasRoute {
			lines = append(lines, "/views/order-summary")
		} else {
			lines = append(lines, "skip")
		}
		lines = append(lines, "done")
	} else {
		lines = append(lines, "done")
	}

	return strings.Join(lines, "\n")
}

func assertComboGraph(t *testing.T, idx int, combo []bool, g graph) {
	hasConcept := combo[0]
	hasField := combo[1]
	hasMethod := combo[2]
	hasOrchestrator := combo[3]
	hasParam := combo[4]
	hasProjection := combo[5]
	hasQueryResult := combo[6]
	hasRoute := combo[7]

	if hasConcept && len(g.Concepts) != 1 {
		t.Fatalf("combo %d expected concept", idx)
	}
	if !hasConcept && len(g.Concepts) != 0 {
		t.Fatalf("combo %d expected no concepts", idx)
	}
	if hasConcept {
		if hasField && len(g.Concepts[0].Fields) != 1 {
			t.Fatalf("combo %d expected field", idx)
		}
		if !hasField && len(g.Concepts[0].Fields) != 0 {
			t.Fatalf("combo %d expected no fields", idx)
		}
		if hasMethod && len(g.Concepts[0].Methods) != 1 {
			t.Fatalf("combo %d expected method", idx)
		}
		if !hasMethod && len(g.Concepts[0].Methods) != 0 {
			t.Fatalf("combo %d expected no methods", idx)
		}
	}

	if hasOrchestrator && len(g.Orchestrators) != 1 {
		t.Fatalf("combo %d expected orchestrator", idx)
	}
	if !hasOrchestrator && len(g.Orchestrators) != 0 {
		t.Fatalf("combo %d expected no orchestrators", idx)
	}
	if hasOrchestrator {
		if hasParam && len(g.Orchestrators[0].Params) != 1 {
			t.Fatalf("combo %d expected param", idx)
		}
		if !hasParam && len(g.Orchestrators[0].Params) != 0 {
			t.Fatalf("combo %d expected no params", idx)
		}
	}

	if hasProjection && len(g.Projections) != 1 {
		t.Fatalf("combo %d expected projection", idx)
	}
	if !hasProjection && len(g.Projections) != 0 {
		t.Fatalf("combo %d expected no projections", idx)
	}
	if hasProjection {
		if hasQueryResult && len(g.Projections[0].Query) != 1 {
			t.Fatalf("combo %d expected query field", idx)
		}
		if !hasQueryResult && len(g.Projections[0].Query) != 0 {
			t.Fatalf("combo %d expected no query fields", idx)
		}
		if hasQueryResult && len(g.Projections[0].Result) != 1 {
			t.Fatalf("combo %d expected result field", idx)
		}
		if !hasQueryResult && len(g.Projections[0].Result) != 0 {
			t.Fatalf("combo %d expected no result fields", idx)
		}
	}

	expectedRoutes := 0
	if hasRoute && hasOrchestrator {
		expectedRoutes++
	}
	if hasRoute && hasProjection {
		expectedRoutes++
	}
	if len(g.Routes) != expectedRoutes {
		t.Fatalf("combo %d expected %d routes, got %d", idx, expectedRoutes, len(g.Routes))
	}
}

func assertComboArgs(t *testing.T, idx int, combo []bool, args []string) {
	hasConcept := combo[0]
	hasField := combo[1]
	hasMethod := combo[2]
	hasOrchestrator := combo[3]
	hasParam := combo[4]
	hasProjection := combo[5]
	hasQueryResult := combo[6]
	hasRoute := combo[7]

	if hasConcept && !contains(args, "--concept") {
		t.Fatalf("combo %d expected --concept", idx)
	}
	if hasField && hasConcept && !contains(args, "--field") {
		t.Fatalf("combo %d expected --field", idx)
	}
	if hasMethod && hasConcept && !contains(args, "--method") {
		t.Fatalf("combo %d expected --method", idx)
	}
	if hasOrchestrator && !contains(args, "--orchestrator") {
		t.Fatalf("combo %d expected --orchestrator", idx)
	}
	if hasParam && hasOrchestrator && !contains(args, "--param") {
		t.Fatalf("combo %d expected --param", idx)
	}
	if hasProjection && !contains(args, "--projection") {
		t.Fatalf("combo %d expected --projection", idx)
	}
	if hasQueryResult && hasProjection && !contains(args, "--query") {
		t.Fatalf("combo %d expected --query", idx)
	}
	if hasQueryResult && hasProjection && !contains(args, "--result") {
		t.Fatalf("combo %d expected --result", idx)
	}
	if hasRoute && !contains(args, "--route") && (hasProjection || hasOrchestrator) {
		t.Fatalf("combo %d expected --route", idx)
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

func contains(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
