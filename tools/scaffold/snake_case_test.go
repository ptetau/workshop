package main

import "testing"

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ID", "id"},
		{"UserID", "user_id"},
		{"HTTPServer", "http_server"},
		{"URLPath", "url_path"},
		{"MemberID", "member_id"},
		{"Name", "name"},
		{"FirstName", "first_name"},
		{"HTTPSConnection", "https_connection"},
		{"APIKey", "api_key"},
		{"", ""},
		{"A", "a"},
		{"AB", "ab"},
		{"ABC", "abc"},
		{"ABCDef", "abc_def"},
		{"MyHTTPServer", "my_http_server"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toSnakeCase(tt.input)
			if result != tt.expected {
				t.Errorf("toSnakeCase(%q) = %q; want %q", tt.input, result, tt.expected)
			}
		})
	}
}
