package listutil

import (
	"net/url"
	"testing"
)

// TestParsePageParams_Defaults verifies default page params when no query values provided.
func TestParsePageParams_Defaults(t *testing.T) {
	q := url.Values{}
	p := ParsePageParams(q)
	if p.Page != 1 {
		t.Errorf("expected page 1, got %d", p.Page)
	}
	if p.PerPage != DefaultPerPage {
		t.Errorf("expected per_page %d, got %d", DefaultPerPage, p.PerPage)
	}
}

// TestParsePageParams_Valid verifies correct parsing of valid page and per_page values.
func TestParsePageParams_Valid(t *testing.T) {
	q := url.Values{"page": {"3"}, "per_page": {"50"}}
	p := ParsePageParams(q)
	if p.Page != 3 {
		t.Errorf("expected page 3, got %d", p.Page)
	}
	if p.PerPage != 50 {
		t.Errorf("expected per_page 50, got %d", p.PerPage)
	}
}

// TestParsePageParams_InvalidPerPage verifies fallback to default for invalid per_page.
func TestParsePageParams_InvalidPerPage(t *testing.T) {
	q := url.Values{"per_page": {"25"}} // not in allowed list
	p := ParsePageParams(q)
	if p.PerPage != DefaultPerPage {
		t.Errorf("expected default per_page %d for invalid value, got %d", DefaultPerPage, p.PerPage)
	}
}

// TestParsePageParams_NegativePage verifies page is clamped to 1 for negative input.
func TestParsePageParams_NegativePage(t *testing.T) {
	q := url.Values{"page": {"-1"}}
	p := ParsePageParams(q)
	if p.Page != 1 {
		t.Errorf("expected page 1 for negative input, got %d", p.Page)
	}
}

// TestParseSortParams_Valid verifies correct parsing of sort column and direction.
func TestParseSortParams_Valid(t *testing.T) {
	q := url.Values{"sort": {"name"}, "dir": {"desc"}}
	s := ParseSortParams(q, []string{"name", "email"})
	if s.Sort != "name" {
		t.Errorf("expected sort=name, got %s", s.Sort)
	}
	if s.Dir != "desc" {
		t.Errorf("expected dir=desc, got %s", s.Dir)
	}
}

// TestParseSortParams_DisallowedColumn verifies disallowed sort columns are rejected.
func TestParseSortParams_DisallowedColumn(t *testing.T) {
	q := url.Values{"sort": {"password"}}
	s := ParseSortParams(q, []string{"name", "email"})
	if s.Sort != "" {
		t.Errorf("expected empty sort for disallowed column, got %s", s.Sort)
	}
}

// TestParseSortParams_InvalidDir verifies invalid direction defaults to asc.
func TestParseSortParams_InvalidDir(t *testing.T) {
	q := url.Values{"sort": {"name"}, "dir": {"DROP TABLE"}}
	s := ParseSortParams(q, []string{"name"})
	if s.Dir != "asc" {
		t.Errorf("expected dir=asc for invalid dir, got %s", s.Dir)
	}
}

// TestParseFilterParams verifies search and filter extraction from query values.
func TestParseFilterParams(t *testing.T) {
	q := url.Values{"q": {"smith"}, "program": {"Adults"}, "unknown": {"x"}}
	f := ParseFilterParams(q, []string{"program", "status"})
	if f.Search != "smith" {
		t.Errorf("expected search=smith, got %s", f.Search)
	}
	if f.Filters["program"] != "Adults" {
		t.Errorf("expected program=Adults, got %s", f.Filters["program"])
	}
	if _, ok := f.Filters["unknown"]; ok {
		t.Error("unexpected filter key 'unknown'")
	}
}

// TestNewPageInfo verifies pagination metadata computation.
func TestNewPageInfo(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		perPage    int
		total      int
		wantPages  int
		wantPage   int
		wantStart  int
		wantEnd    int
		wantOffset int
	}{
		{"basic", 1, 20, 85, 5, 1, 1, 20, 0},
		{"page2", 2, 20, 85, 5, 2, 21, 40, 20},
		{"lastPage", 5, 20, 85, 5, 5, 81, 85, 80},
		{"pageBeyondTotal", 10, 20, 85, 5, 5, 81, 85, 80},
		{"emptyList", 1, 20, 0, 1, 1, 0, 0, 0},
		{"exactFit", 1, 10, 10, 1, 1, 1, 10, 0},
		{"singleRow", 1, 20, 1, 1, 1, 1, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pi := NewPageInfo(tt.page, tt.perPage, tt.total)
			if pi.TotalPages != tt.wantPages {
				t.Errorf("TotalPages: got %d, want %d", pi.TotalPages, tt.wantPages)
			}
			if pi.Page != tt.wantPage {
				t.Errorf("Page: got %d, want %d", pi.Page, tt.wantPage)
			}
			if pi.StartRow() != tt.wantStart {
				t.Errorf("StartRow: got %d, want %d", pi.StartRow(), tt.wantStart)
			}
			if pi.EndRow() != tt.wantEnd {
				t.Errorf("EndRow: got %d, want %d", pi.EndRow(), tt.wantEnd)
			}
			if pi.Offset() != tt.wantOffset {
				t.Errorf("Offset: got %d, want %d", pi.Offset(), tt.wantOffset)
			}
		})
	}
}

// TestPageNumbers verifies page number window generation.
func TestPageNumbers(t *testing.T) {
	tests := []struct {
		name string
		page int
		tot  int
		want []int
	}{
		{"3pages_at1", 1, 3, []int{1, 2, 3}},
		{"10pages_at1", 1, 10, []int{1, 2, 3, 4, 5}},
		{"10pages_at5", 5, 10, []int{3, 4, 5, 6, 7}},
		{"10pages_at10", 10, 10, []int{6, 7, 8, 9, 10}},
		{"1page", 1, 1, []int{1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pi := NewPageInfo(tt.page, 20, tt.tot*20)
			got := pi.PageNumbers()
			if len(got) != len(tt.want) {
				t.Fatalf("PageNumbers length: got %d, want %d", len(got), len(tt.want))
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("PageNumbers[%d]: got %d, want %d", i, v, tt.want[i])
				}
			}
		})
	}
}

// TestShowPagination verifies pagination visibility logic.
func TestShowPagination(t *testing.T) {
	if NewPageInfo(1, 20, 20).ShowPagination() {
		t.Error("should not show pagination when total == perPage")
	}
	if !NewPageInfo(1, 20, 21).ShowPagination() {
		t.Error("should show pagination when total > perPage")
	}
}
