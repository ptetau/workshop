package web

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	attendanceDomain "workshop/internal/domain/attendance"
	injuryDomain "workshop/internal/domain/injury"
	memberDomain "workshop/internal/domain/member"
	waiverDomain "workshop/internal/domain/waiver"

	attendanceStore "workshop/internal/adapters/storage/attendance"
	injuryStore "workshop/internal/adapters/storage/injury"
	memberStore "workshop/internal/adapters/storage/member"
	waiverStore "workshop/internal/adapters/storage/waiver"
)

// Mock implementations for testing
type mockMemberStore struct {
	members map[string]memberDomain.Member
}

// GetByID implements the member store interface for testing.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (m *mockMemberStore) GetByID(ctx context.Context, id string) (memberDomain.Member, error) {
	if mem, ok := m.members[id]; ok {
		return mem, nil
	}
	return memberDomain.Member{}, sql.ErrNoRows
}

// GetByEmail implements the member store interface for testing.
// PRE: email is non-empty
// POST: Returns the entity or an error if not found
func (m *mockMemberStore) GetByEmail(ctx context.Context, email string) (memberDomain.Member, error) {
	for _, mem := range m.members {
		if mem.Email == email {
			return mem, nil
		}
	}
	return memberDomain.Member{}, sql.ErrNoRows
}

// Save implements the member store interface for testing.
// PRE: entity has been validated
// POST: Entity is persisted
func (m *mockMemberStore) Save(ctx context.Context, mem memberDomain.Member) error {
	if m.members == nil {
		m.members = make(map[string]memberDomain.Member)
	}
	m.members[mem.ID] = mem
	return nil
}

// Delete implements the member store interface for testing.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (m *mockMemberStore) Delete(ctx context.Context, id string) error {
	delete(m.members, id)
	return nil
}

// SearchByName implements the member store interface for testing.
// PRE: query is non-empty
// POST: Returns matching members
func (m *mockMemberStore) SearchByName(ctx context.Context, query string, limit int) ([]memberDomain.Member, error) {
	var list []memberDomain.Member
	for _, mem := range m.members {
		if len(list) >= limit {
			break
		}
		list = append(list, mem)
	}
	return list, nil
}

// List implements the member store interface for testing.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (m *mockMemberStore) List(ctx context.Context, filter memberStore.ListFilter) ([]memberDomain.Member, error) {
	var list []memberDomain.Member
	for _, mem := range m.members {
		list = append(list, mem)
	}
	return list, nil
}

// Count implements the member store interface for testing.
// PRE: filter has valid parameters
// POST: Returns count of matching entities
func (m *mockMemberStore) Count(ctx context.Context, filter memberStore.ListFilter) (int, error) {
	return len(m.members), nil
}

type mockAttendanceStore struct {
	attendances map[string]attendanceDomain.Attendance
}

// Delete implements the attendance store interface for testing.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (m *mockAttendanceStore) Delete(ctx context.Context, id string) error {
	delete(m.attendances, id)
	return nil
}

// GetByID implements the attendance store interface for testing.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (m *mockAttendanceStore) GetByID(ctx context.Context, id string) (attendanceDomain.Attendance, error) {
	if a, ok := m.attendances[id]; ok {
		return a, nil
	}
	return attendanceDomain.Attendance{}, nil
}

// Save implements the attendance store interface for testing.
// PRE: entity has been validated
// POST: Entity is persisted
func (m *mockAttendanceStore) Save(ctx context.Context, a attendanceDomain.Attendance) error {
	if m.attendances == nil {
		m.attendances = make(map[string]attendanceDomain.Attendance)
	}
	m.attendances[a.ID] = a
	return nil
}

// ListByMemberID implements the attendance store interface for testing.
// PRE: memberID is non-empty
// POST: Returns records for the given member
func (m *mockAttendanceStore) ListByMemberID(ctx context.Context, memberID string) ([]attendanceDomain.Attendance, error) {
	var list []attendanceDomain.Attendance
	for _, a := range m.attendances {
		if a.MemberID == memberID {
			list = append(list, a)
		}
	}
	return list, nil
}

// List implements the attendance store interface for testing.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (m *mockAttendanceStore) List(ctx context.Context, filter attendanceStore.ListFilter) ([]attendanceDomain.Attendance, error) {
	var list []attendanceDomain.Attendance
	for _, a := range m.attendances {
		list = append(list, a)
	}
	return list, nil
}

type mockInjuryStore struct {
	injuries map[string]injuryDomain.Injury
}

// Delete implements the injury store interface for testing.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (m *mockInjuryStore) Delete(ctx context.Context, id string) error {
	delete(m.injuries, id)
	return nil
}

// GetByID implements the injury store interface for testing.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (m *mockInjuryStore) GetByID(ctx context.Context, id string) (injuryDomain.Injury, error) {
	if i, ok := m.injuries[id]; ok {
		return i, nil
	}
	return injuryDomain.Injury{}, nil
}

// Save implements the injury store interface for testing.
// PRE: entity has been validated
// POST: Entity is persisted
func (m *mockInjuryStore) Save(ctx context.Context, i injuryDomain.Injury) error {
	if m.injuries == nil {
		m.injuries = make(map[string]injuryDomain.Injury)
	}
	m.injuries[i.ID] = i
	return nil
}

// List implements the injury store interface for testing.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (m *mockInjuryStore) List(ctx context.Context, filter injuryStore.ListFilter) ([]injuryDomain.Injury, error) {
	var list []injuryDomain.Injury
	for _, i := range m.injuries {
		list = append(list, i)
	}
	return list, nil
}

type mockWaiverStore struct {
	waivers map[string]waiverDomain.Waiver
}

// Delete implements the waiver store interface for testing.
// PRE: id is non-empty
// POST: Entity with given id is removed
func (m *mockWaiverStore) Delete(ctx context.Context, id string) error {
	delete(m.waivers, id)
	return nil
}

// GetByID implements the waiver store interface for testing.
// PRE: id is non-empty
// POST: Returns the entity or an error if not found
func (m *mockWaiverStore) GetByID(ctx context.Context, id string) (waiverDomain.Waiver, error) {
	if w, ok := m.waivers[id]; ok {
		return w, nil
	}
	return waiverDomain.Waiver{}, nil
}

// Save implements the waiver store interface for testing.
// PRE: entity has been validated
// POST: Entity is persisted
func (m *mockWaiverStore) Save(ctx context.Context, w waiverDomain.Waiver) error {
	if m.waivers == nil {
		m.waivers = make(map[string]waiverDomain.Waiver)
	}
	m.waivers[w.ID] = w
	return nil
}

// List implements the waiver store interface for testing.
// PRE: filter has valid parameters
// POST: Returns matching entities
func (m *mockWaiverStore) List(ctx context.Context, filter waiverStore.ListFilter) ([]waiverDomain.Waiver, error) {
	var list []waiverDomain.Waiver
	for _, w := range m.waivers {
		list = append(list, w)
	}
	return list, nil
}

// setupTestStores creates mock stores for testing
func setupTestStores(t *testing.T) {
	t.Helper()

	// Create a custom Stores struct with mocks
	// We need to use interface-based approach since Stores expects concrete types
	// For now, we'll set stores to nil and skip tests that need it
	stores = nil
}

// TestPostRegistermember tests the POST register member endpoint.
func TestPostRegistermember(t *testing.T) {
	tests := []struct {
		name         string
		formData     url.Values
		wantStatus   int
		wantRedirect string
		checkMember  bool
	}{
		{
			name: "valid registration with adults program",
			formData: url.Values{
				"Name":    []string{"John Doe"},
				"Email":   []string{"john@example.com"},
				"Program": []string{"adults"},
			},
			wantStatus:   http.StatusSeeOther,
			wantRedirect: "/",
			checkMember:  true,
		},
		{
			name: "valid registration with kids program",
			formData: url.Values{
				"Name":    []string{"Jane Smith"},
				"Email":   []string{"jane@example.com"},
				"Program": []string{"kids"},
			},
			wantStatus:   http.StatusSeeOther,
			wantRedirect: "/",
			checkMember:  true,
		},
		{
			name: "missing name",
			formData: url.Values{
				"Email":   []string{"john@example.com"},
				"Program": []string{"adults"},
			},
			wantStatus:  http.StatusInternalServerError,
			checkMember: false,
		},
		{
			name: "missing email",
			formData: url.Values{
				"Name":    []string{"John Doe"},
				"Program": []string{"adults"},
			},
			wantStatus:  http.StatusInternalServerError,
			checkMember: false,
		},
		{
			name: "invalid program",
			formData: url.Values{
				"Name":    []string{"John Doe"},
				"Email":   []string{"john@example.com"},
				"Program": []string{"invalid"},
			},
			wantStatus:  http.StatusInternalServerError,
			checkMember: false,
		},
		{
			name: "missing program",
			formData: url.Values{
				"Name":  []string{"John Doe"},
				"Email": []string{"john@example.com"},
			},
			wantStatus:  http.StatusInternalServerError,
			checkMember: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock stores
			mockMember := &mockMemberStore{members: make(map[string]memberDomain.Member)}
			stores = &Stores{
				MemberStore:     mockMember,
				AttendanceStore: &mockAttendanceStore{attendances: make(map[string]attendanceDomain.Attendance)},
				InjuryStore:     &mockInjuryStore{injuries: make(map[string]injuryDomain.Injury)},
				WaiverStore:     &mockWaiverStore{waivers: make(map[string]waiverDomain.Waiver)},
			}

			// Create request
			req := httptest.NewRequest("POST", "/members", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Accept", "text/html")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Call handler
			handleMembers(rec, req)

			// Assert response status
			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d. Body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			// Assert redirect location if expected
			if tt.wantRedirect != "" {
				location := rec.Header().Get("Location")
				if location != tt.wantRedirect {
					t.Errorf("got redirect %q, want %q", location, tt.wantRedirect)
				}
			}

			// Verify member was created if successful
			if tt.checkMember && tt.wantStatus == http.StatusSeeOther {
				ctx := context.Background()
				members, err := mockMember.List(ctx, memberStore.ListFilter{Limit: 100})
				if err != nil {
					t.Fatalf("failed to list members: %v", err)
				}

				if len(members) != 1 {
					t.Errorf("expected 1 member, got %d", len(members))
				}

				if len(members) > 0 {
					member := members[0]
					if member.Name != tt.formData.Get("Name") {
						t.Errorf("got name %q, want %q", member.Name, tt.formData.Get("Name"))
					}
					if member.Email != tt.formData.Get("Email") {
						t.Errorf("got email %q, want %q", member.Email, tt.formData.Get("Email"))
					}
					if member.Program != tt.formData.Get("Program") {
						t.Errorf("got program %q, want %q", member.Program, tt.formData.Get("Program"))
					}
					if member.Status != "active" {
						t.Errorf("got status %q, want %q", member.Status, "active")
					}
				}
			}
		})
	}
}

// TestGetGetmemberlist tests the GET member list endpoint.
func TestGetGetmemberlist(t *testing.T) {
	tests := []struct {
		name         string
		setupMembers []memberDomain.Member
		wantStatus   int
		wantCount    int
	}{
		{
			name: "list with multiple members",
			setupMembers: []memberDomain.Member{
				{
					ID:        "member-1",
					Name:      "John Doe",
					Email:     "john@example.com",
					Program:   "adults",
					Fee:       150,
					Frequency: "monthly",
					Status:    "active",
				},
				{
					ID:        "member-2",
					Name:      "Jane Smith",
					Email:     "jane@example.com",
					Program:   "kids",
					Fee:       100,
					Frequency: "monthly",
					Status:    "active",
				},
			},
			wantStatus: http.StatusOK,
			wantCount:  2,
		},
		{
			name:         "empty list",
			setupMembers: []memberDomain.Member{},
			wantStatus:   http.StatusOK,
			wantCount:    0,
		},
		{
			name: "single member",
			setupMembers: []memberDomain.Member{
				{
					ID:        "member-1",
					Name:      "John Doe",
					Email:     "john@example.com",
					Program:   "adults",
					Fee:       150,
					Frequency: "monthly",
					Status:    "active",
				},
			},
			wantStatus: http.StatusOK,
			wantCount:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock stores
			mockMember := &mockMemberStore{members: make(map[string]memberDomain.Member)}
			stores = &Stores{
				MemberStore:     mockMember,
				AttendanceStore: &mockAttendanceStore{attendances: make(map[string]attendanceDomain.Attendance)},
				InjuryStore:     &mockInjuryStore{injuries: make(map[string]injuryDomain.Injury)},
				WaiverStore:     &mockWaiverStore{waivers: make(map[string]waiverDomain.Waiver)},
			}

			// Insert test members
			ctx := context.Background()
			for _, member := range tt.setupMembers {
				if err := mockMember.Save(ctx, member); err != nil {
					t.Fatalf("failed to save test member: %v", err)
				}
			}

			// Create request
			req := httptest.NewRequest("GET", "/members", nil)
			req.Header.Set("Accept", "application/json")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Call handler
			handleMembers(rec, req)

			// Assert response status
			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}

			// Verify response is JSON
			contentType := rec.Header().Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Errorf("expected JSON content type, got %q", contentType)
			}

			// Parse and verify response body
			var result struct {
				Members  []json.RawMessage `json:"Members"`
				PageInfo json.RawMessage   `json:"PageInfo"`
			}
			if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if len(result.Members) != tt.wantCount {
				t.Errorf("got %d members, want %d", len(result.Members), tt.wantCount)
			}
		})
	}
}

// TestPostCheckinmember tests the POST check-in member endpoint.
func TestPostCheckinmember(t *testing.T) {
	tests := []struct {
		name        string
		setupMember *memberDomain.Member
		formData    url.Values
		wantStatus  int
	}{
		{
			name: "valid check-in for existing member",
			setupMember: &memberDomain.Member{
				ID:        "member-123",
				Name:      "John Doe",
				Email:     "john@example.com",
				Program:   "adults",
				Fee:       150,
				Frequency: "monthly",
				Status:    "active",
			},
			formData: url.Values{
				"MemberID": []string{"member-123"},
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name:        "missing member ID",
			setupMember: nil,
			formData:    url.Values{},
			wantStatus:  http.StatusInternalServerError,
		},
		{
			name:        "empty member ID",
			setupMember: nil,
			formData: url.Values{
				"MemberID": []string{""},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock stores
			mockMember := &mockMemberStore{members: make(map[string]memberDomain.Member)}
			mockAttendance := &mockAttendanceStore{attendances: make(map[string]attendanceDomain.Attendance)}
			stores = &Stores{
				MemberStore:     mockMember,
				AttendanceStore: mockAttendance,
				InjuryStore:     &mockInjuryStore{injuries: make(map[string]injuryDomain.Injury)},
				WaiverStore:     &mockWaiverStore{waivers: make(map[string]waiverDomain.Waiver)},
			}

			// Setup member if provided
			if tt.setupMember != nil {
				ctx := context.Background()
				if err := mockMember.Save(ctx, *tt.setupMember); err != nil {
					t.Fatalf("failed to save test member: %v", err)
				}
			}

			// Create request
			req := httptest.NewRequest("POST", "/checkin", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Accept", "text/html")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Call handler
			handlePostCheckinCheckInMember(rec, req)

			// Assert response status
			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d. Body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}

			// Verify redirect on success
			if tt.wantStatus == http.StatusSeeOther {
				location := rec.Header().Get("Location")
				if location != "/" {
					t.Errorf("got redirect %q, want %q", location, "/")
				}
			}
		})
	}
}

// TestGetGetattendancetoday tests the GET attendance today endpoint.
func TestGetGetattendancetoday(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus int
	}{
		{
			name:       "get today's attendance",
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock stores
			stores = &Stores{
				MemberStore:     &mockMemberStore{members: make(map[string]memberDomain.Member)},
				AttendanceStore: &mockAttendanceStore{attendances: make(map[string]attendanceDomain.Attendance)},
				InjuryStore:     &mockInjuryStore{injuries: make(map[string]injuryDomain.Injury)},
				WaiverStore:     &mockWaiverStore{waivers: make(map[string]waiverDomain.Waiver)},
			}

			// Create request
			req := httptest.NewRequest("GET", "/attendance", nil)
			req.Header.Set("Accept", "application/json")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Call handler
			handleGetAttendanceGetAttendanceToday(rec, req)

			// Assert response status
			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", rec.Code, tt.wantStatus)
			}

			// Verify response is JSON
			contentType := rec.Header().Get("Content-Type")
			if !strings.Contains(contentType, "application/json") {
				t.Errorf("expected JSON content type, got %q", contentType)
			}
		})
	}
}

// TestPostReportinjury tests the POST report injury endpoint.
func TestPostReportinjury(t *testing.T) {
	tests := []struct {
		name        string
		setupMember *memberDomain.Member
		formData    url.Values
		wantStatus  int
	}{
		{
			name: "valid injury report",
			setupMember: &memberDomain.Member{
				ID:        "member-123",
				Name:      "John Doe",
				Email:     "john@example.com",
				Program:   "adults",
				Fee:       150,
				Frequency: "monthly",
				Status:    "active",
			},
			formData: url.Values{
				"MemberID":    []string{"member-123"},
				"BodyPart":    []string{"knee"},
				"Description": []string{"Twisted during training"},
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name:        "missing member ID",
			setupMember: nil,
			formData: url.Values{
				"BodyPart":    []string{"knee"},
				"Description": []string{"Twisted during training"},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "missing body part",
			setupMember: &memberDomain.Member{
				ID:        "member-123",
				Name:      "John Doe",
				Email:     "john@example.com",
				Program:   "adults",
				Fee:       150,
				Frequency: "monthly",
				Status:    "active",
			},
			formData: url.Values{
				"MemberID":    []string{"member-123"},
				"Description": []string{"Twisted during training"},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock stores
			mockMember := &mockMemberStore{members: make(map[string]memberDomain.Member)}
			stores = &Stores{
				MemberStore:     mockMember,
				AttendanceStore: &mockAttendanceStore{attendances: make(map[string]attendanceDomain.Attendance)},
				InjuryStore:     &mockInjuryStore{injuries: make(map[string]injuryDomain.Injury)},
				WaiverStore:     &mockWaiverStore{waivers: make(map[string]waiverDomain.Waiver)},
			}

			// Setup member if provided
			if tt.setupMember != nil {
				ctx := context.Background()
				if err := mockMember.Save(ctx, *tt.setupMember); err != nil {
					t.Fatalf("failed to save test member: %v", err)
				}
			}

			// Create request
			req := httptest.NewRequest("POST", "/injuries", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Accept", "text/html")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Call handler
			handlePostInjuriesReportInjury(rec, req)

			// Assert response status
			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d. Body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}

// TestPostSignwaiver tests the POST sign waiver endpoint.
func TestPostSignwaiver(t *testing.T) {
	tests := []struct {
		name       string
		formData   url.Values
		wantStatus int
	}{
		{
			name: "valid waiver signature",
			formData: url.Values{
				"MemberName":    []string{"John Doe"},
				"Email":         []string{"john@example.com"},
				"AcceptedTerms": []string{"true"},
			},
			wantStatus: http.StatusSeeOther,
		},
		{
			name: "missing member name",
			formData: url.Values{
				"Email":         []string{"john@example.com"},
				"AcceptedTerms": []string{"true"},
			},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "terms not accepted",
			formData: url.Values{
				"MemberName":    []string{"John Doe"},
				"Email":         []string{"john@example.com"},
				"AcceptedTerms": []string{"false"},
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock stores
			stores = &Stores{
				MemberStore:     &mockMemberStore{members: make(map[string]memberDomain.Member)},
				AttendanceStore: &mockAttendanceStore{attendances: make(map[string]attendanceDomain.Attendance)},
				InjuryStore:     &mockInjuryStore{injuries: make(map[string]injuryDomain.Injury)},
				WaiverStore:     &mockWaiverStore{waivers: make(map[string]waiverDomain.Waiver)},
			}

			// Create request
			req := httptest.NewRequest("POST", "/waivers", strings.NewReader(tt.formData.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("Accept", "text/html")

			// Create response recorder
			rec := httptest.NewRecorder()

			// Call handler
			handlePostWaiversSignWaiver(rec, req)

			// Assert response status
			if rec.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d. Body: %s", rec.Code, tt.wantStatus, rec.Body.String())
			}
		})
	}
}
