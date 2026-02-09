package listutil

import (
	"net/url"
	"strconv"
)

// PageParams carries pagination parameters parsed from a request.
type PageParams struct {
	Page    int // 1-indexed page number
	PerPage int // rows per page
}

// SortParams carries sorting parameters parsed from a request.
type SortParams struct {
	Sort string // column name
	Dir  string // "asc" or "desc"
}

// FilterParams carries search and filter parameters.
type FilterParams struct {
	Search  string            // free-text search query
	Filters map[string]string // exact-match filters (e.g. program=Adults)
}

// PageInfo carries pagination metadata for rendering.
type PageInfo struct {
	Page       int // current page (1-indexed)
	PerPage    int // rows per page
	Total      int // total matching rows
	TotalPages int // ceil(Total / PerPage)
}

// ListParams combines all list view parameters.
type ListParams struct {
	PageParams
	SortParams
	FilterParams
}

// DefaultPerPage is the default number of rows per page.
const DefaultPerPage = 20

// PerPageOptions are the allowed rows-per-page values.
var PerPageOptions = []int{10, 20, 50, 100, 200}

// ParsePageParams extracts page and per_page from URL query values.
// PRE: none
// POST: returns valid PageParams with defaults applied
func ParsePageParams(q url.Values) PageParams {
	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(q.Get("per_page"))
	if !isValidPerPage(perPage) {
		perPage = DefaultPerPage
	}
	return PageParams{Page: page, PerPage: perPage}
}

// ParseSortParams extracts sort and dir from URL query values.
// PRE: none
// POST: returns SortParams; Dir is always "asc" or "desc"
func ParseSortParams(q url.Values, allowedColumns []string) SortParams {
	sort := q.Get("sort")
	dir := q.Get("dir")

	if !isAllowedColumn(sort, allowedColumns) {
		sort = ""
	}
	if dir != "asc" && dir != "desc" {
		dir = "asc"
	}
	return SortParams{Sort: sort, Dir: dir}
}

// ParseFilterParams extracts search and named filters from URL query values.
// PRE: filterKeys lists the allowed filter parameter names
// POST: returns FilterParams with only recognised keys
func ParseFilterParams(q url.Values, filterKeys []string) FilterParams {
	fp := FilterParams{
		Search:  q.Get("q"),
		Filters: make(map[string]string),
	}
	for _, key := range filterKeys {
		if v := q.Get(key); v != "" {
			fp.Filters[key] = v
		}
	}
	return fp
}

// ParseListParams parses all list parameters from URL query values.
func ParseListParams(q url.Values, allowedSortCols []string, filterKeys []string) ListParams {
	return ListParams{
		PageParams:   ParsePageParams(q),
		SortParams:   ParseSortParams(q, allowedSortCols),
		FilterParams: ParseFilterParams(q, filterKeys),
	}
}

// NewPageInfo computes pagination metadata.
// PRE: total >= 0, perPage > 0, page >= 1
// POST: returns PageInfo with TotalPages computed; Page clamped to valid range
func NewPageInfo(page, perPage, total int) PageInfo {
	if perPage < 1 {
		perPage = DefaultPerPage
	}
	totalPages := (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}
	if page > totalPages {
		page = totalPages
	}
	if page < 1 {
		page = 1
	}
	return PageInfo{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// Offset returns the SQL OFFSET for the current page.
// PRE: PageInfo is valid
// POST: Returns (Page-1) * PerPage
func (p PageInfo) Offset() int {
	return (p.Page - 1) * p.PerPage
}

// StartRow returns the 1-indexed first row number on the current page.
// PRE: PageInfo is valid
// POST: Returns 0 if Total is 0, otherwise Offset+1
func (p PageInfo) StartRow() int {
	if p.Total == 0 {
		return 0
	}
	return p.Offset() + 1
}

// EndRow returns the 1-indexed last row number on the current page.
// PRE: PageInfo is valid
// POST: Returns min(Offset+PerPage, Total)
func (p PageInfo) EndRow() int {
	end := p.Offset() + p.PerPage
	if end > p.Total {
		end = p.Total
	}
	return end
}

// PageNumbers returns the page numbers to display in pagination controls.
// Shows at most 5 pages centered around the current page.
// PRE: PageInfo is valid
// POST: Returns slice of at most 5 page numbers centered on current page
func (p PageInfo) PageNumbers() []int {
	const maxButtons = 5
	start := p.Page - maxButtons/2
	if start < 1 {
		start = 1
	}
	end := start + maxButtons - 1
	if end > p.TotalPages {
		end = p.TotalPages
		start = end - maxButtons + 1
		if start < 1 {
			start = 1
		}
	}
	pages := make([]int, 0, end-start+1)
	for i := start; i <= end; i++ {
		pages = append(pages, i)
	}
	return pages
}

// ShowPagination returns true if pagination controls should be displayed.
// PRE: PageInfo is valid
// POST: Returns true if Total > PerPage
func (p PageInfo) ShowPagination() bool {
	return p.Total > p.PerPage
}

func isValidPerPage(n int) bool {
	for _, opt := range PerPageOptions {
		if n == opt {
			return true
		}
	}
	return false
}

func isAllowedColumn(col string, allowed []string) bool {
	for _, a := range allowed {
		if col == a {
			return true
		}
	}
	return false
}
