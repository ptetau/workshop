package main

import (
	"testing"
)

func TestBuildDesiredState_Metadata(t *testing.T) {
	// Normal Test: Verify method metadata (doc, pre, post, invariant) is correctly parsed
	state := buildDesiredState(
		multiFlag{"Concept"},                     // concepts
		multiFlag{},                              // fields
		multiFlag{"Concept:Method"},              // methods
		multiFlag{},                              // orchestrators
		multiFlag{},                              // params
		multiFlag{},                              // projections
		multiFlag{},                              // queries
		multiFlag{},                              // results
		multiFlag{},                              // routes
		multiFlag{},                              // descriptions
		multiFlag{"Concept:Method:x>0"},          // invariants
		multiFlag{},                              // fieldDocs
		multiFlag{"Concept:Method:Do something"}, // methodDocs
		multiFlag{"Concept:Method:x is valid"},   // pres
		multiFlag{"Concept:Method:x is processed"}, // posts
		multiFlag{}, // orchDocs
		multiFlag{}, // paramDocs
		multiFlag{}, // projDocs
		multiFlag{}, // queryDocs
		multiFlag{}, // resultDocs
	)

	c, ok := state.Concepts["Concept"]
	if !ok {
		t.Fatalf("Concept not found")
	}
	if len(c.Methods) != 1 {
		t.Fatalf("Expected 1 method, got %d", len(c.Methods))
	}
	m := c.Methods[0]
	if m.Name != "Method" {
		t.Errorf("Expected method Name 'Method', got '%s'", m.Name)
	}
	if m.Description != "Do something" {
		t.Errorf("Expected Description 'Do something', got '%s'", m.Description)
	}
	if m.PreCondition != "x is valid" {
		t.Errorf("Expected PreCondition 'x is valid', got '%s'", m.PreCondition)
	}
	if m.PostCondition != "x is processed" {
		t.Errorf("Expected PostCondition 'x is processed', got '%s'", m.PostCondition)
	}
	if m.Invariant != "x>0" {
		t.Errorf("Expected Invariant 'x>0', got '%s'", m.Invariant)
	}
}

func TestBuildDesiredState_PairwiseMetadata(t *testing.T) {
	// Pairwise-style Test: Check combinations of flags
	tests := []struct {
		name         string
		methodDocs   multiFlag
		pres         multiFlag
		posts        multiFlag
		expectedDesc string
		expectedPre  string
		expectedPost string
	}{
		{
			name:         "All Present",
			methodDocs:   multiFlag{"C:M:Desc"},
			pres:         multiFlag{"C:M:Pre"},
			posts:        multiFlag{"C:M:Post"},
			expectedDesc: "Desc", expectedPre: "Pre", expectedPost: "Post",
		},
		{
			name:         "Only Desc",
			methodDocs:   multiFlag{"C:M:Desc"},
			pres:         multiFlag{},
			posts:        multiFlag{},
			expectedDesc: "Desc", expectedPre: "", expectedPost: "",
		},
		{
			name:         "Only Pre",
			methodDocs:   multiFlag{},
			pres:         multiFlag{"C:M:Pre"},
			posts:        multiFlag{},
			expectedDesc: "", expectedPre: "Pre", expectedPost: "",
		},
		{
			name:         "Only Post",
			methodDocs:   multiFlag{},
			pres:         multiFlag{},
			posts:        multiFlag{"C:M:Post"},
			expectedDesc: "", expectedPre: "", expectedPost: "Post",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			state := buildDesiredState(
				multiFlag{"C"}, // concepts
				multiFlag{},
				multiFlag{"C:M"}, // methods
				multiFlag{}, multiFlag{}, multiFlag{}, multiFlag{}, multiFlag{}, multiFlag{}, multiFlag{}, multiFlag{},
				multiFlag{},
				tc.methodDocs,
				tc.pres,
				tc.posts,
				multiFlag{}, multiFlag{}, multiFlag{}, multiFlag{}, multiFlag{},
			)
			c := state.Concepts["C"]
			m := c.Methods[0]
			if m.Description != tc.expectedDesc {
				t.Errorf("Desc mismatch: want %q, got %q", tc.expectedDesc, m.Description)
			}
			if m.PreCondition != tc.expectedPre {
				t.Errorf("Pre mismatch: want %q, got %q", tc.expectedPre, m.PreCondition)
			}
			if m.PostCondition != tc.expectedPost {
				t.Errorf("Post mismatch: want %q, got %q", tc.expectedPost, m.PostCondition)
			}
		})
	}
}
