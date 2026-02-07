package main

const httpTestTemplate = `package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

{{range .Tests}}
func {{.HandlerName}}(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		headers    map[string]string
		wantStatus int
		wantHeader map[string]string
	}{
		{
			name:       "{{.TestCaseName}}",
			method:     "{{.Method}}",
			path:       "{{.Path}}",
			// TODO: Add request body if needed
			wantStatus: http.StatusOK,
			// TODO: Add expected response headers
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// TODO: Setup mock stores
			// stores := &Stores{
			//     MemberStore: &mockMemberStore{},
			// }

			// TODO: Create request
			// req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			// for k, v := range tt.headers {
			//     req.Header.Set(k, v)
			// }

			// TODO: Create response recorder
			// rec := httptest.NewRecorder()

			// TODO: Call handler
			// handler := {{.HandlerFunc}}
			// handler(rec, req)

			// TODO: Assert response
			// if rec.Code != tt.wantStatus {
			//     t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			// }
		})
	}
}
{{end}}
`

const e2eTestTemplate = `package web

import (
	"context"
	"testing"
	"time"

	"github.com/chromedp/chromedp"
)

{{range .Tests}}
func TestE2E{{.FlowName}}(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping E2E test in short mode")
	}

	// TODO: Start test server
	// server, stores := testutil.NewTestServer(t)
	// defer server.Close()

	// Create browser context
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	// Set timeout
	ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// TODO: Navigate to {{.Path}}
	// TODO: Perform {{.Method}} action
	// TODO: Verify results

	err := chromedp.Run(ctx,
		// chromedp.Navigate(server.URL+"{{.Path}}"),
		// chromedp.WaitVisible("#some-element"),
		// TODO: Add more actions
	)
	if err != nil {
		t.Fatalf("E2E test failed: %v", err)
	}
}
{{end}}
`
