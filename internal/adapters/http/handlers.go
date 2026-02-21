package web

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/yuin/goldmark"
	goldmarkHTML "github.com/yuin/goldmark/renderer/html"

	"workshop/internal/adapters/http/middleware"
	accountStore "workshop/internal/adapters/storage/account"
	emailStoreImport "workshop/internal/adapters/storage/email"
	memberStore "workshop/internal/adapters/storage/member"
	noticeStore "workshop/internal/adapters/storage/notice"
	"workshop/internal/application/listutil"
	"workshop/internal/application/orchestrators"
	"workshop/internal/application/projections"
	accountDomain "workshop/internal/domain/account"
	"workshop/internal/domain/attendance"
	calendarDomain "workshop/internal/domain/calendar"
	classTypeDomain "workshop/internal/domain/classtype"
	clipDomain "workshop/internal/domain/clip"
	emailDomain "workshop/internal/domain/email"
	estimatedHoursDomain "workshop/internal/domain/estimatedhours"
	featureflagDomain "workshop/internal/domain/featureflag"
	gradingDomain "workshop/internal/domain/grading"
	holidayDomain "workshop/internal/domain/holiday"
	memberDomain "workshop/internal/domain/member"
	messageDomain "workshop/internal/domain/message"
	milestoneDomain "workshop/internal/domain/milestone"
	noticeDomain "workshop/internal/domain/notice"
	rotorDomain "workshop/internal/domain/rotor"
	scheduleDomain "workshop/internal/domain/schedule"
	termDomain "workshop/internal/domain/term"
	themeDomain "workshop/internal/domain/theme"
	trainingGoalDomain "workshop/internal/domain/traininggoal"
)

// timeNow is a variable for testability.
var timeNow = time.Now

// mdRenderer is a goldmark instance configured for safe HTML output.
// Raw HTML in markdown input is escaped (WithUnsafe is NOT set), preventing XSS.
var mdRenderer = goldmark.New(
	goldmark.WithRendererOptions(
		goldmarkHTML.WithHardWraps(),
	),
)

// generateID creates a new UUID string.
func generateID() string {
	return uuid.New().String()
}

func csvSafeCell(s string) string {
	// Mitigate CSV formula injection when opening exports in Excel/Sheets.
	// Any cell starting with these characters may be treated as a formula.
	// Prefix with a single quote to force literal interpretation.
	if s == "" {
		return s
	}
	switch s[0] {
	case '=', '+', '-', '@':
		return "'" + s
	default:
		return s
	}
}

// handleMembersExportCSV handles GET /api/members/export
// Exports the members list as CSV, respecting the same search/filter/sort params as /members.
func handleMembersExportCSV(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !middleware.IsCoachOrAdmin(r.Context()) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !requireFeatureAPI(w, r, sess, "member_mgmt") {
		return
	}

	// Same query params as the /members list.
	lp := listutil.ParseListParams(r.URL.Query(),
		[]string{"name", "email", "program", "status"},
		[]string{"program", "status"},
	)

	filter := memberStore.ListFilter{
		Program: lp.Filters["program"],
		Status:  lp.Filters["status"],
		Search:  lp.Search,
		Sort:    lp.Sort,
		Dir:     lp.Dir,
		Limit:   100000,
		Offset:  0,
	}

	members, err := stores.MemberStore.List(r.Context(), filter)
	if err != nil {
		internalError(w, err)
		return
	}

	filename := fmt.Sprintf("members-%s.csv", timeNow().Format("2006-01-02"))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Cache-Control", "no-store")

	cw := csv.NewWriter(w)
	cw.UseCRLF = true

	// Note: we call Flush explicitly and check cw.Error() so write failures don't get swallowed.

	if err := cw.Write([]string{"ID", "AccountID", "Name", "Email", "Program", "Status", "Fee", "Frequency", "GradingMetric"}); err != nil {
		internalError(w, err)
		return
	}
	for _, m := range members {
		rec := []string{
			csvSafeCell(m.ID),
			csvSafeCell(m.AccountID),
			csvSafeCell(m.Name),
			csvSafeCell(m.Email),
			csvSafeCell(m.Program),
			csvSafeCell(m.Status),
			csvSafeCell(strconv.Itoa(m.Fee)),
			csvSafeCell(m.Frequency),
			csvSafeCell(m.GradingMetric),
		}
		if err := cw.Write(rec); err != nil {
			internalError(w, err)
			return
		}
	}

	cw.Flush()
	if err := cw.Error(); err != nil {
		// Response may be partially written; log only.
		slog.Error("csv_export_flush_error", "error", err.Error())
	}
}

// internalError logs the real error and returns a generic message to the client.
// This prevents leaking internal details per OWASP A05.
func internalError(w http.ResponseWriter, err error) {
	slog.Error("internal_error", "error", err.Error())
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

// strictDecode decodes JSON from the request body, rejecting unknown fields.
func strictDecode(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

const templatesDir = "internal/adapters/http/templates"

func isHTMLRequest(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "text/html") || strings.Contains(accept, "application/xhtml+xml")
}

func renderTemplate(w http.ResponseWriter, r *http.Request, templateName string, data any) {
	sess, ok := middleware.GetSessionFromContext(r.Context())
	role := ""
	email := ""
	if ok {
		role = sess.Role
		email = sess.Email
	}

	impersonating := false
	realRole := ""
	isRealAdmin := false
	if ok && sess.IsImpersonating() {
		impersonating = true
		realRole = sess.RealRole
		isRealAdmin = sess.RealRole == "admin"
	} else if ok {
		isRealAdmin = sess.Role == "admin"
	}

	flagsByKey := mergedFeatureFlagsByKey(r.Context())

	funcMap := template.FuncMap{
		"currentRole":  func() string { return role },
		"currentEmail": func() string { return email },
		"isLoggedIn":   func() bool { return role != "" },
		"featureEnabled": func(key string) bool {
			if !ok {
				return false
			}
			ff, exists := flagsByKey[key]
			if !exists {
				return true
			}
			return ff.EnabledForRole(sess.Role, sess.BetaTester)
		},
		"csrfToken":       func() string { return csrf.Token(r) },
		"isImpersonating": func() bool { return impersonating },
		"realRole":        func() string { return realRole },
		"isRealAdmin":     func() bool { return isRealAdmin },
		"list":            func(items ...string) []string { return items },
		"renderMarkdown": func(md string) template.HTML {
			var buf bytes.Buffer
			if err := mdRenderer.Convert([]byte(md), &buf); err != nil {
				return template.HTML(template.HTMLEscapeString(md))
			}
			return template.HTML(buf.String())
		},
		"noticeColorHex": func(color string) string {
			if hex, ok := noticeDomain.ColorHex[color]; ok {
				return hex
			}
			return noticeDomain.ColorHex[noticeDomain.ColorOrange]
		},
		"stripeRange": func(n int) []int {
			s := make([]int, n)
			for i := range s {
				s[i] = i
			}
			return s
		},
		"add": func(a, b int) int { return a + b },
		"sub": func(a, b int) int { return a - b },
		"sortHeaderArgs": func(col, label, activeSort, activeDir, search, program, status string, perPage int) map[string]string {
			nextDir := "asc"
			if col == activeSort && activeDir == "asc" {
				nextDir = "desc"
			}
			return map[string]string{
				"Col": col, "Label": label,
				"ActiveSort": activeSort, "ActiveDir": activeDir, "NextDir": nextDir,
				"Search": search, "Program": program, "Status": status,
				"PerPage": fmt.Sprintf("%d", perPage),
			}
		},
		"paginationQuery": func(page int, sort, dir, search, program, status string, perPage int) template.URL {
			q := fmt.Sprintf("page=%d", page)
			if sort != "" {
				q += "&sort=" + sort
			}
			if dir != "" {
				q += "&dir=" + dir
			}
			if search != "" {
				q += "&q=" + search
			}
			if program != "" {
				q += "&program=" + program
			}
			if status != "" {
				q += "&status=" + status
			}
			if perPage != 0 {
				q += fmt.Sprintf("&per_page=%d", perPage)
			}
			return template.URL(q)
		},
		"exportMembersQuery": func(sort, dir, search, program, status string) template.URL {
			v := url.Values{}
			if sort != "" {
				v.Set("sort", sort)
			}
			if dir != "" {
				v.Set("dir", dir)
			}
			if search != "" {
				v.Set("q", search)
			}
			if program != "" {
				v.Set("program", program)
			}
			if status != "" {
				v.Set("status", status)
			}
			return template.URL(v.Encode())
		},
	}

	layoutPath := filepath.Join(templatesDir, "layout.html")
	pagePath := filepath.Join(templatesDir, templateName)
	tpl, err := template.New("layout.html").Funcs(funcMap).ParseFiles(layoutPath, pagePath)
	if err != nil {
		http.Error(w, "Template error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.Execute(w, data); err != nil {
		http.Error(w, "Render error: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func mergedFeatureFlagsByKey(ctx context.Context) map[string]featureflagDomain.FeatureFlag {
	m := make(map[string]featureflagDomain.FeatureFlag)
	for _, d := range featureflagDomain.DefaultFlags() {
		m[d.Key] = d
	}
	if stores == nil || stores.FeatureFlagStore == nil {
		return m
	}
	persisted, err := stores.FeatureFlagStore.List(ctx)
	if err != nil {
		return m
	}
	for _, p := range persisted {
		if d, ok := m[p.Key]; ok {
			if p.Description != "" {
				d.Description = p.Description
			}
			d.EnabledAdmin = p.EnabledAdmin
			d.EnabledCoach = p.EnabledCoach
			d.EnabledMember = p.EnabledMember
			d.EnabledTrial = p.EnabledTrial
			d.BetaOverride = p.BetaOverride
			m[p.Key] = d
		} else {
			m[p.Key] = p
		}
	}
	return m
}

func featureEnabledForSession(ctx context.Context, sess middleware.Session, featureKey string) bool {
	ff, ok := mergedFeatureFlagsByKey(ctx)[featureKey]
	if !ok {
		return true
	}
	return ff.EnabledForRole(sess.Role, sess.BetaTester)
}

func requireFeaturePage(w http.ResponseWriter, r *http.Request, sess middleware.Session, featureKey string) bool {
	if featureEnabledForSession(r.Context(), sess, featureKey) {
		return true
	}
	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	return false
}

func requireFeatureAPI(w http.ResponseWriter, r *http.Request, sess middleware.Session, featureKey string) bool {
	if featureEnabledForSession(r.Context(), sess, featureKey) {
		return true
	}
	http.Error(w, "Forbidden", http.StatusForbidden)
	return false
}

// handleMembers handles both GET (list) and POST (register) for /members
func handleMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	isHTML := isHTMLRequest(r)
	if sess, ok := middleware.GetSessionFromContext(ctx); ok {
		if isHTML {
			if !requireFeaturePage(w, r, sess, "member_mgmt") {
				return
			}
		} else {
			if !requireFeatureAPI(w, r, sess, "member_mgmt") {
				return
			}
		}
	}

	if r.Method == "GET" {
		// GET: List members with pagination, sorting, search, and filtering
		lp := listutil.ParseListParams(r.URL.Query(),
			[]string{"name", "email", "program", "status"},
			[]string{"program", "status"},
		)

		query := projections.GetMemberListQuery{
			Program: lp.Filters["program"],
			Status:  lp.Filters["status"],
			Search:  lp.Search,
			Sort:    lp.Sort,
			Dir:     lp.Dir,
			Page:    lp.Page,
			PerPage: lp.PerPage,
		}
		deps := projections.GetMemberListDeps{
			MemberStore: stores.MemberStore,
			InjuryStore: stores.InjuryStore,
		}

		result, err := projections.QueryGetMemberList(ctx, query, deps)
		if err != nil {
			internalError(w, err)
			return
		}

		if isHTML {
			renderTemplate(w, r, "get_member_list.html", map[string]any{
				"Members":        result.Members,
				"PageInfo":       result.PageInfo,
				"Sort":           lp.Sort,
				"Dir":            lp.Dir,
				"Search":         lp.Search,
				"Program":        lp.Filters["program"],
				"Status":         lp.Filters["status"],
				"PerPageOptions": listutil.PerPageOptions,
				"HasFilters":     lp.Search != "" || lp.Filters["program"] != "" || lp.Filters["status"] != "",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result)
		return
	}

	if r.Method == "POST" {
		// POST: Register member
		input := orchestrators.RegisterMemberInput{}

		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form submission", http.StatusBadRequest)
				return
			}
			input.Email = r.FormValue("Email")
			input.Name = r.FormValue("Name")
			input.Program = r.FormValue("Program")
		} else {
			if err := strictDecode(r, &input); err != nil {
				http.Error(w, "Invalid request", http.StatusBadRequest)
				return
			}
		}

		deps := orchestrators.RegisterMemberDeps{
			MemberStore: stores.MemberStore,
		}
		_, err := orchestrators.ExecuteRegisterMember(ctx, input, deps)
		if err != nil {
			internalError(w, err)
			return
		}

		if isHTML {
			http.Redirect(w, r, "/", http.StatusSeeOther)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handlePostCheckinCheckInMember handles POST /checkin
func handlePostCheckinCheckInMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	isHTML := isHTMLRequest(r)

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if sess, ok := middleware.GetSessionFromContext(ctx); ok {
		if isHTML {
			if !requireFeaturePage(w, r, sess, "attendance") {
				return
			}
		} else {
			if !requireFeatureAPI(w, r, sess, "attendance") {
				return
			}
		}
	}

	input := orchestrators.CheckInMemberInput{}

	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form submission", http.StatusBadRequest)
			return
		}
		input.MemberID = r.FormValue("MemberID")
		input.ScheduleID = r.FormValue("ScheduleID")
		input.ClassDate = r.FormValue("ClassDate")
	} else {
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
	}

	deps := orchestrators.CheckInMemberDeps{
		MemberStore:     stores.MemberStore,
		AttendanceStore: stores.AttendanceStore,
		ScheduleStore:   stores.ScheduleStore,
	}
	if stores.GradingRecordStore != nil && stores.GradingConfigStore != nil {
		deps.InferStripeDeps = &orchestrators.InferStripeDeps{
			MemberStore:         stores.MemberStore,
			AttendanceStore:     stores.AttendanceStore,
			EstimatedHoursStore: stores.EstimatedHoursStore,
			GradingRecordStore:  stores.GradingRecordStore,
			GradingConfigStore:  stores.GradingConfigStore,
		}
	}
	err := orchestrators.ExecuteCheckInMember(ctx, input, deps)
	if err != nil {
		internalError(w, err)
		return
	}

	if isHTML {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

// handleGetAttendanceGetAttendanceToday handles GET /attendance
func handleGetAttendanceGetAttendanceToday(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	isHTML := isHTMLRequest(r)

	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if sess, ok := middleware.GetSessionFromContext(ctx); ok {
		if isHTML {
			if !requireFeaturePage(w, r, sess, "attendance") {
				return
			}
		} else {
			if !requireFeatureAPI(w, r, sess, "attendance") {
				return
			}
		}
	}

	dateParam := r.URL.Query().Get("date")
	query := projections.GetAttendanceTodayQuery{Date: dateParam}
	deps := projections.GetAttendanceTodayDeps{
		AttendanceStore:    stores.AttendanceStore,
		MemberStore:        stores.MemberStore,
		InjuryStore:        stores.InjuryStore,
		GradingRecordStore: stores.GradingRecordStore,
		ScheduleStore:      stores.ScheduleStore,
		ClassTypeStore:     stores.ClassTypeStore,
	}

	result, err := projections.QueryGetAttendanceToday(ctx, query, deps)
	if err != nil {
		internalError(w, err)
		return
	}

	// Determine display date and whether it's today
	today := time.Now().Format("2006-01-02")
	displayDate := today
	if dateParam != "" {
		displayDate = dateParam
	}
	isToday := displayDate == today

	// Compute prev/next dates for navigation
	parsed, _ := time.Parse("2006-01-02", displayDate)
	if parsed.IsZero() {
		parsed = time.Now()
	}
	prevDate := parsed.AddDate(0, 0, -1).Format("2006-01-02")
	nextDate := parsed.AddDate(0, 0, 1).Format("2006-01-02")

	if isHTML {
		renderTemplate(w, r, "get_attendance_today.html", map[string]any{
			"Attendees":   result.Attendees,
			"DisplayDate": displayDate,
			"IsToday":     isToday,
			"PrevDate":    prevDate,
			"NextDate":    nextDate,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Attendees)
}

// handleMemberAttendanceToday handles GET /api/attendance/member?member_id=X
// Returns today's check-ins for a specific member (used by kiosk for un-check-in).
func handleMemberAttendanceToday(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if sess, ok := middleware.GetSessionFromContext(r.Context()); ok {
		if !requireFeatureAPI(w, r, sess, "attendance") {
			return
		}
	}

	memberID := r.URL.Query().Get("member_id")
	if memberID == "" {
		http.Error(w, "member_id is required", http.StatusBadRequest)
		return
	}

	today := time.Now().Format("2006-01-02")
	records, err := stores.AttendanceStore.ListByMemberIDAndDate(r.Context(), memberID, today)
	if err != nil {
		internalError(w, err)
		return
	}
	if records == nil {
		records = []attendance.Attendance{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

// handleUndoCheckIn handles DELETE /api/attendance/undo
// Removes an attendance record (only today's check-ins).
func handleUndoCheckIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if sess, ok := middleware.GetSessionFromContext(r.Context()); ok {
		if !requireFeatureAPI(w, r, sess, "attendance") {
			return
		}
	}

	var input struct {
		AttendanceID string `json:"AttendanceID"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	deps := orchestrators.UndoCheckInDeps{
		AttendanceStore: stores.AttendanceStore,
	}
	err := orchestrators.ExecuteUndoCheckIn(r.Context(), orchestrators.UndoCheckInInput{
		AttendanceID: input.AttendanceID,
	}, deps)
	if err != nil {
		if err.Error() == "can only undo today's check-ins" || err.Error() == "attendance record not found" {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleCheckOut handles POST /api/attendance/checkout
// Sets CheckOutTime and calculates MatHours for an active check-in.
func handleCheckOut(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	if sess, ok := middleware.GetSessionFromContext(r.Context()); ok {
		if !requireFeatureAPI(w, r, sess, "attendance") {
			return
		}
	}

	var input struct {
		AttendanceID string `json:"AttendanceID"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if input.AttendanceID == "" {
		http.Error(w, "AttendanceID is required", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	record, err := stores.AttendanceStore.GetByID(ctx, input.AttendanceID)
	if err != nil {
		http.Error(w, "attendance record not found", http.StatusNotFound)
		return
	}

	if record.IsCheckedOut() {
		http.Error(w, "already checked out", http.StatusConflict)
		return
	}

	now := time.Now()
	record.CheckOutTime = now
	record.MatHours = now.Sub(record.CheckInTime).Hours()

	if err := record.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := stores.AttendanceStore.Save(ctx, record); err != nil {
		internalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(record)
}

// handleEstimatedHours handles GET/POST for /api/estimated-hours
func handleEstimatedHours(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if sess.Role != "admin" && sess.Role != "coach" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		memberID := r.URL.Query().Get("member_id")
		if memberID == "" {
			http.Error(w, "member_id is required", http.StatusBadRequest)
			return
		}
		entries, err := stores.EstimatedHoursStore.ListByMemberID(ctx, memberID)
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if entries == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(entries)
		return
	}

	if r.Method == "POST" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if sess.Role != "admin" && sess.Role != "coach" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		var input struct {
			MemberID    string  `json:"MemberID"`
			StartDate   string  `json:"StartDate"`
			EndDate     string  `json:"EndDate"`
			WeeklyHours float64 `json:"WeeklyHours"`
			Note        string  `json:"Note"`
			OverlapMode string  `json:"OverlapMode"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		orchInput := orchestrators.BulkAddEstimatedHoursInput{
			MemberID:    input.MemberID,
			StartDate:   input.StartDate,
			EndDate:     input.EndDate,
			WeeklyHours: input.WeeklyHours,
			Note:        input.Note,
			CreatedBy:   sess.AccountID,
			OverlapMode: input.OverlapMode,
		}
		orchDeps := orchestrators.BulkAddEstimatedHoursDeps{
			EstimatedHoursStore: stores.EstimatedHoursStore,
			AttendanceStore:     stores.AttendanceStore,
			GenerateID:          generateID,
			Now:                 timeNow,
		}
		entry, err := orchestrators.ExecuteBulkAddEstimatedHours(ctx, orchInput, orchDeps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(entry)
		return
	}

	if r.Method == "DELETE" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if sess.Role != "admin" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := stores.EstimatedHoursStore.Delete(ctx, id); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleEstimatedHoursCheckOverlap handles GET /api/estimated-hours/check-overlap
func handleEstimatedHoursCheckOverlap(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	memberID := r.URL.Query().Get("member_id")
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")
	if memberID == "" || startDate == "" || endDate == "" {
		http.Error(w, "member_id, start_date, and end_date are required", http.StatusBadRequest)
		return
	}
	result, err := orchestrators.CheckEstimatedHoursOverlap(r.Context(), memberID, startDate, endDate, stores.AttendanceStore)
	if err != nil {
		internalError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleSelfEstimates handles POST /api/self-estimates ΓÇö member submits a self-estimate.
func handleSelfEstimates(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	// Look up the member record for the logged-in user
	m, err := stores.MemberStore.GetByEmail(r.Context(), sess.Email)
	if err != nil {
		http.Error(w, "no member record found for this account", http.StatusForbidden)
		return
	}
	var input struct {
		StartDate   string  `json:"StartDate"`
		EndDate     string  `json:"EndDate"`
		WeeklyHours float64 `json:"WeeklyHours"`
		Note        string  `json:"Note"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	orchInput := orchestrators.SubmitSelfEstimateInput{
		MemberID:    m.ID,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		WeeklyHours: input.WeeklyHours,
		Note:        input.Note,
	}
	orchDeps := orchestrators.SubmitSelfEstimateDeps{
		EstimatedHoursStore: stores.EstimatedHoursStore,
		GenerateID:          generateID,
		Now:                 timeNow,
	}
	entry, err := orchestrators.ExecuteSubmitSelfEstimate(r.Context(), orchInput, orchDeps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

// handleSelfEstimatesPending handles GET /api/self-estimates/pending ΓÇö admin/coach review queue.
func handleSelfEstimatesPending(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	entries, err := stores.EstimatedHoursStore.ListPending(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}
	// Enrich with member names
	type pendingEntry struct {
		ID          string  `json:"ID"`
		MemberID    string  `json:"MemberID"`
		MemberName  string  `json:"MemberName"`
		StartDate   string  `json:"StartDate"`
		EndDate     string  `json:"EndDate"`
		WeeklyHours float64 `json:"WeeklyHours"`
		TotalHours  float64 `json:"TotalHours"`
		Note        string  `json:"Note"`
		CreatedAt   string  `json:"CreatedAt"`
	}
	result := make([]pendingEntry, 0, len(entries))
	for _, e := range entries {
		name := ""
		if m, err := stores.MemberStore.GetByID(r.Context(), e.MemberID); err == nil {
			name = m.Name
		}
		result = append(result, pendingEntry{
			ID:          e.ID,
			MemberID:    e.MemberID,
			MemberName:  name,
			StartDate:   e.StartDate,
			EndDate:     e.EndDate,
			WeeklyHours: e.WeeklyHours,
			TotalHours:  e.TotalHours,
			Note:        e.Note,
			CreatedAt:   e.CreatedAt.Format(time.RFC3339),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	if len(result) == 0 {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(result)
}

// handleSelfEstimatesReview handles POST /api/self-estimates/review ΓÇö admin/coach approves or rejects.
func handleSelfEstimatesReview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var input struct {
		ID            string  `json:"ID"`
		Action        string  `json:"Action"`
		AdjustedHours float64 `json:"AdjustedHours"`
		ReviewNote    string  `json:"ReviewNote"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.ID == "" || (input.Action != "approve" && input.Action != "reject") {
		http.Error(w, "ID and Action (approve/reject) are required", http.StatusBadRequest)
		return
	}
	orchInput := orchestrators.ReviewSelfEstimateInput{
		ID:            input.ID,
		Action:        input.Action,
		AdjustedHours: input.AdjustedHours,
		ReviewNote:    input.ReviewNote,
		ReviewerID:    sess.AccountID,
	}
	orchDeps := orchestrators.ReviewSelfEstimateDeps{
		EstimatedHoursStore: stores.EstimatedHoursStore,
		Now:                 timeNow,
	}
	entry, err := orchestrators.ExecuteReviewSelfEstimate(r.Context(), orchInput, orchDeps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(entry)
}

// handleSelfEstimatesPage handles GET /admin/self-estimates ΓÇö admin review queue page.
func handleSelfEstimatesPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	renderTemplate(w, r, "admin_self_estimates.html", nil)
}

// handlePostInjuriesReportInjury handles POST /injuries
func handlePostInjuriesReportInjury(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	isHTML := isHTMLRequest(r)

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	input := orchestrators.ReportInjuryInput{}

	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form submission", http.StatusBadRequest)
			return
		}
		input.MemberID = r.FormValue("MemberID")
		input.BodyPart = strings.ToLower(r.FormValue("BodyPart"))
		input.Description = r.FormValue("Description")
	} else {
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
	}

	deps := orchestrators.ReportInjuryDeps{
		InjuryStore: stores.InjuryStore,
		MemberStore: stores.MemberStore,
	}
	err := orchestrators.ExecuteReportInjury(ctx, input, deps)
	if err != nil {
		internalError(w, err)
		return
	}

	if isHTML {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

// handlePostWaiversSignWaiver handles POST /waivers
func handlePostWaiversSignWaiver(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	isHTML := isHTMLRequest(r)

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	input := orchestrators.SignWaiverInput{}

	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form submission", http.StatusBadRequest)
			return
		}
		input.MemberName = r.FormValue("MemberName")
		input.Email = r.FormValue("Email")
		input.AcceptedTerms = r.FormValue("AcceptedTerms") == "true"
	} else {
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
	}

	deps := orchestrators.SignWaiverDeps{
		WaiverStore: stores.WaiverStore,
		MemberStore: stores.MemberStore,
	}
	err := orchestrators.ExecuteSignWaiver(ctx, input, deps)
	if err != nil {
		internalError(w, err)
		return
	}

	if isHTML {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleGetMembersRegisterForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	renderTemplate(w, r, "form_register_member.html", map[string]any{
		"CSRFToken": csrf.Token(r),
	})
}

func handleGetCheckInForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	renderTemplate(w, r, "form_check_in_member.html", map[string]any{
		"CSRFToken": csrf.Token(r),
	})
}

func handleGetInjuryForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	renderTemplate(w, r, "form_report_injury.html", map[string]any{
		"CSRFToken": csrf.Token(r),
	})
}

func handleGetWaiverForm(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	renderTemplate(w, r, "form_sign_waiver.html", map[string]any{
		"CSRFToken": csrf.Token(r),
	})
}

func handleGetMemberProfile(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		if isHTMLRequest(r) {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		if isHTMLRequest(r) {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if isHTMLRequest(r) {
		if !requireFeaturePage(w, r, sess, "member_mgmt") {
			return
		}
	} else {
		if !requireFeatureAPI(w, r, sess, "member_mgmt") {
			return
		}
	}

	memberID := r.URL.Query().Get("id")
	if memberID == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	query := projections.GetMemberProfileQuery{MemberID: memberID}
	deps := projections.GetMemberProfileDeps{
		MemberStore:     stores.MemberStore,
		WaiverStore:     stores.WaiverStore,
		InjuryStore:     stores.InjuryStore,
		AttendanceStore: stores.AttendanceStore,
	}

	result, err := projections.QueryGetMemberProfile(r.Context(), query, deps)
	if err != nil {
		internalError(w, err)
		return
	}

	if isHTMLRequest(r) {
		renderTemplate(w, r, "get_member_profile.html", result)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleLogin handles GET (form) and POST (authenticate) for /login
func handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		// If already logged in, redirect to dashboard
		if _, ok := middleware.GetSessionFromContext(r.Context()); ok {
			http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
			return
		}
		renderTemplate(w, r, "login.html", map[string]any{
			"CSRFToken": csrf.Token(r),
		})
		return
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form submission", http.StatusBadRequest)
			return
		}

		input := orchestrators.LoginInput{
			Email:    r.FormValue("Email"),
			Password: r.FormValue("Password"),
		}

		deps := orchestrators.LoginDeps{
			AccountStore: stores.AccountStore,
		}

		result, err := orchestrators.ExecuteLogin(r.Context(), input, deps)
		if err != nil {
			renderTemplate(w, r, "login.html", map[string]any{
				"CSRFToken": csrf.Token(r),
				"Error":     err.Error(),
			})
			return
		}

		// Create session
		betaTester := false
		if acct, err := stores.AccountStore.GetByID(r.Context(), result.AccountID); err == nil {
			betaTester = acct.BetaTester
		}
		token, err := sessions.Create(result.AccountID, result.Email, result.Role, result.PasswordChangeRequired, betaTester)
		if err != nil {
			http.Error(w, "Session error", http.StatusInternalServerError)
			return
		}

		middleware.SetSessionCookie(w, token)
		if result.PasswordChangeRequired {
			http.Redirect(w, r, "/change-password", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleLogout handles POST /logout
func handleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Delete session
	cookie, err := r.Cookie("workshop_session")
	if err == nil {
		sessions.Delete(cookie.Value)
	}

	middleware.ClearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// handleChangePassword handles GET (form) and POST (update) for /change-password
func handleChangePassword(w http.ResponseWriter, r *http.Request) {
	session, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if r.Method == "GET" {
		renderTemplate(w, r, "change_password.html", map[string]any{
			"CSRFToken": csrf.Token(r),
			"Forced":    session.PasswordChangeRequired,
		})
		return
	}

	if r.Method == "POST" {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Form error", http.StatusBadRequest)
			return
		}

		input := orchestrators.ChangePasswordInput{
			AccountID:       session.AccountID,
			CurrentPassword: r.FormValue("CurrentPassword"),
			NewPassword:     r.FormValue("NewPassword"),
		}

		// Validate confirm matches
		if r.FormValue("NewPassword") != r.FormValue("ConfirmPassword") {
			renderTemplate(w, r, "change_password.html", map[string]any{
				"CSRFToken": csrf.Token(r),
				"Forced":    session.PasswordChangeRequired,
				"Error":     "New passwords do not match",
			})
			return
		}

		deps := orchestrators.ChangePasswordDeps{
			AccountStore: stores.AccountStore,
		}

		if err := orchestrators.ExecuteChangePassword(r.Context(), input, deps); err != nil {
			renderTemplate(w, r, "change_password.html", map[string]any{
				"CSRFToken": csrf.Token(r),
				"Forced":    session.PasswordChangeRequired,
				"Error":     err.Error(),
			})
			return
		}

		// Update session to clear the flag
		cookie, err := r.Cookie("workshop_session")
		if err == nil {
			session.PasswordChangeRequired = false
			sessions.Update(cookie.Value, session)
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleMemberSearch handles GET /api/members/search?q=<name>
func handleMemberSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if sess, ok := middleware.GetSessionFromContext(r.Context()); ok {
		if !requireFeatureAPI(w, r, sess, "member_mgmt") {
			return
		}
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	results, err := stores.MemberStore.SearchByName(r.Context(), query, 10)
	if err != nil {
		internalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// handleArchiveMember handles POST /api/members/archive
func handleArchiveMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if sess, ok := middleware.GetSessionFromContext(r.Context()); ok {
		if !requireFeatureAPI(w, r, sess, "member_mgmt") {
			return
		}
	}

	var input orchestrators.ArchiveMemberInput
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		r.ParseForm()
		input.MemberID = r.FormValue("MemberID")
	} else {
		strictDecode(r, &input)
	}

	deps := orchestrators.ArchiveMemberDeps{MemberStore: stores.MemberStore}
	if err := orchestrators.ExecuteArchiveMember(r.Context(), input, deps); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleRestoreMember handles POST /api/members/restore
func handleRestoreMember(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	if sess, ok := middleware.GetSessionFromContext(r.Context()); ok {
		if !requireFeatureAPI(w, r, sess, "member_mgmt") {
			return
		}
	}

	var input orchestrators.RestoreMemberInput
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		r.ParseForm()
		input.MemberID = r.FormValue("MemberID")
	} else {
		strictDecode(r, &input)
	}

	deps := orchestrators.RestoreMemberDeps{MemberStore: stores.MemberStore}
	if err := orchestrators.ExecuteRestoreMember(r.Context(), input, deps); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleGuestCheckIn handles POST /api/guest/checkin
func handleGuestCheckIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var input orchestrators.GuestCheckInInput
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		r.ParseForm()
		input.Name = r.FormValue("Name")
		input.Email = r.FormValue("Email")
		input.AcceptedTerms = r.FormValue("AcceptedTerms") == "true"
		input.ScheduleID = r.FormValue("ScheduleID")
		input.ClassDate = r.FormValue("ClassDate")
	} else {
		strictDecode(r, &input)
	}
	input.IPAddress = r.RemoteAddr

	deps := orchestrators.GuestCheckInDeps{
		MemberStore:     stores.MemberStore,
		WaiverStore:     stores.WaiverStore,
		AttendanceStore: stores.AttendanceStore,
	}

	result, err := orchestrators.ExecuteGuestCheckIn(r.Context(), input, deps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleTodaysClasses handles GET /api/classes/today
func handleTodaysClasses(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	deps := projections.GetTodaysClassesDeps{
		ScheduleStore:  stores.ScheduleStore,
		TermStore:      stores.TermStore,
		HolidayStore:   stores.HolidayStore,
		ClassTypeStore: stores.ClassTypeStore,
		ProgramStore:   stores.ProgramStore,
	}

	results, err := projections.QueryGetTodaysClasses(r.Context(), timeNow(), deps)
	if err != nil {
		internalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if results == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(results)
}

// handleKioskLaunch handles POST /api/kiosk/launch
func handleKioskLaunch(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	deps := orchestrators.LaunchKioskDeps{AccountStore: stores.AccountStore}
	input := orchestrators.LaunchKioskInput{AccountID: sess.AccountID}

	session, err := orchestrators.ExecuteLaunchKiosk(r.Context(), input, deps)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}

// handleKioskExit handles POST /api/kiosk/exit
func handleKioskExit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var input orchestrators.ExitKioskInput
	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		r.ParseForm()
		input.AccountID = r.FormValue("AccountID")
		input.Password = r.FormValue("Password")
	} else {
		strictDecode(r, &input)
	}

	deps := orchestrators.ExitKioskDeps{AccountStore: stores.AccountStore}
	if err := orchestrators.ExecuteExitKiosk(r.Context(), input, deps); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Layer 1b: Engagement API Handlers ---

// handleGetTrainingLog handles GET /api/training-log?member_id=<id>
func handleGetTrainingLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !requireFeatureAPI(w, r, sess, "training_log") {
		return
	}
	memberID := r.URL.Query().Get("member_id")
	if memberID == "" {
		http.Error(w, "member_id is required", http.StatusBadRequest)
		return
	}

	query := projections.GetTrainingLogQuery{MemberID: memberID}
	deps := projections.GetTrainingLogDeps{
		AttendanceStore:     stores.AttendanceStore,
		MemberStore:         stores.MemberStore,
		GradingRecordStore:  stores.GradingRecordStore,
		GradingConfigStore:  stores.GradingConfigStore,
		EstimatedHoursStore: stores.EstimatedHoursStore,
	}
	result, err := projections.QueryGetTrainingLog(r.Context(), query, deps)
	if err != nil {
		internalError(w, err)
		return
	}

	// For kids in sessions mode, enrich with term attendance data
	if result.Program == "kids" && (result.GradingMetric == "" || result.GradingMetric == memberDomain.MetricSessions) {
		if result.GradingMetric == "" {
			result.GradingMetric = memberDomain.MetricSessions
		}
		kidsQuery := projections.GetKidsTermReadinessQuery{Now: time.Now()}
		kidsDeps := projections.GetKidsTermReadinessDeps{
			TermStore:          stores.TermStore,
			ProgramStore:       stores.ProgramStore,
			ClassTypeStore:     stores.ClassTypeStore,
			ScheduleStore:      stores.ScheduleStore,
			HolidayStore:       stores.HolidayStore,
			MemberStore:        stores.MemberStore,
			AttendanceStore:    stores.AttendanceStore,
			GradingRecordStore: stores.GradingRecordStore,
			GradingConfigStore: stores.GradingConfigStore,
		}
		kidsResult, err := projections.QueryGetKidsTermReadiness(r.Context(), kidsQuery, kidsDeps)
		if err == nil {
			result.TermName = kidsResult.TermName
			for _, e := range kidsResult.Entries {
				if e.MemberID == memberID {
					result.TermAttended = e.Attended
					result.TermTotal = e.TotalSessions
					result.TermAttendancePct = e.AttendancePct
					result.TermThresholdPct = e.ThresholdPct
					result.TermEligible = e.Eligible
					break
				}
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleGetInactiveMembers handles GET /api/members/inactive?days=<n>
func handleGetInactiveMembers(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	days := 30
	if d := r.URL.Query().Get("days"); d != "" {
		fmt.Sscanf(d, "%d", &days)
	}

	query := projections.GetInactiveMembersQuery{DaysSinceLastCheckIn: days}
	deps := projections.GetInactiveMembersDeps{
		MemberStore:     stores.MemberStore,
		AttendanceStore: stores.AttendanceStore,
	}
	results, err := projections.QueryGetInactiveMembers(r.Context(), query, deps)
	if err != nil {
		internalError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if results == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(results)
}

// handleNotices handles GET/POST for /api/notices
func handleNotices(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		if _, ok := middleware.GetSessionFromContext(ctx); !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		noticeType := r.URL.Query().Get("type")
		if noticeType != "" {
			results, err := stores.NoticeStore.ListPublished(ctx, noticeType, timeNow())
			if err != nil {
				internalError(w, err)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			if results == nil {
				w.Write([]byte("[]"))
				return
			}
			json.NewEncoder(w).Encode(results)
			return
		}
		// No type filter ΓÇö return all notices
		results, err := stores.NoticeStore.List(ctx, noticeStore.ListFilter{Limit: 100})
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if results == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(results)
		return
	}

	if r.Method == "POST" {
		sess, ok := requireAdmin(w, r)
		if !ok {
			return
		}
		var input struct {
			Type         string `json:"Type"`
			Title        string `json:"Title"`
			Content      string `json:"Content"`
			TargetID     string `json:"TargetID"`
			AuthorName   string `json:"AuthorName"`
			ShowAuthor   bool   `json:"ShowAuthor"`
			Color        string `json:"Color"`
			VisibleFrom  string `json:"VisibleFrom"`
			VisibleUntil string `json:"VisibleUntil"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		orchInput := orchestrators.CreateNoticeInput{
			Type:       input.Type,
			Title:      input.Title,
			Content:    input.Content,
			TargetID:   input.TargetID,
			AuthorName: input.AuthorName,
			ShowAuthor: input.ShowAuthor,
			Color:      input.Color,
			CreatedBy:  sess.AccountID,
		}
		if input.VisibleFrom != "" {
			if t, err := time.Parse(time.RFC3339, input.VisibleFrom); err == nil {
				orchInput.VisibleFrom = t
			}
		}
		if input.VisibleUntil != "" {
			if t, err := time.Parse(time.RFC3339, input.VisibleUntil); err == nil {
				orchInput.VisibleUntil = t
			}
		}
		n, err := orchestrators.ExecuteCreateNotice(ctx, orchInput, orchestrators.CreateNoticeDeps{
			NoticeStore: stores.NoticeStore,
			GenerateID:  generateID,
			Now:         timeNow,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(n)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleGradingProposals handles GET/POST for /api/grading/proposals
func handleGradingProposals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		if _, ok := middleware.GetSessionFromContext(ctx); !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		proposals, err := stores.GradingProposalStore.ListPending(ctx)
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if proposals == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(proposals)
		return
	}

	if r.Method == "POST" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		var input struct {
			MemberID   string `json:"MemberID"`
			TargetBelt string `json:"TargetBelt"`
			Notes      string `json:"Notes"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		proposal := gradingDomain.Proposal{
			ID:         generateID(),
			MemberID:   input.MemberID,
			TargetBelt: input.TargetBelt,
			Notes:      input.Notes,
			ProposedBy: sess.AccountID,
			Status:     gradingDomain.ProposalPending,
			CreatedAt:  timeNow(),
		}
		if err := proposal.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.GradingProposalStore.Save(ctx, proposal); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(proposal)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleGradingNotes handles GET/POST for /api/grading/notes
func handleGradingNotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		if _, ok := middleware.GetSessionFromContext(ctx); !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		memberID := r.URL.Query().Get("member_id")
		if memberID == "" {
			http.Error(w, "member_id is required", http.StatusBadRequest)
			return
		}
		notes, err := stores.GradingNoteStore.ListByMemberID(ctx, memberID)
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if notes == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(notes)
		return
	}

	if r.Method == "POST" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		var input struct {
			MemberID string `json:"MemberID"`
			Content  string `json:"Content"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		note := gradingDomain.Note{
			ID:        generateID(),
			MemberID:  input.MemberID,
			Content:   input.Content,
			CreatedBy: sess.AccountID,
			CreatedAt: timeNow(),
		}
		if err := note.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.GradingNoteStore.Save(ctx, note); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(note)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleMessages handles GET/POST for /api/messages
func handleMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if !requireFeatureAPI(w, r, sess, "messages") {
			return
		}
		memberID := r.URL.Query().Get("member_id")
		if memberID == "" {
			http.Error(w, "member_id is required", http.StatusBadRequest)
			return
		}
		messages, err := stores.MessageStore.ListByReceiverID(ctx, memberID)
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if messages == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(messages)
		return
	}

	if r.Method == "POST" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		var input struct {
			ReceiverID string `json:"ReceiverID"`
			Subject    string `json:"Subject"`
			Content    string `json:"Content"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		msg := messageDomain.Message{
			ID:         generateID(),
			SenderID:   sess.AccountID,
			ReceiverID: input.ReceiverID,
			Subject:    input.Subject,
			Content:    input.Content,
			CreatedAt:  timeNow(),
		}
		if err := msg.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.MessageStore.Save(ctx, msg); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(msg)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleObservations handles GET/POST for /api/observations
func handleObservations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		if _, ok := middleware.GetSessionFromContext(ctx); !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		memberID := r.URL.Query().Get("member_id")
		if memberID == "" {
			http.Error(w, "member_id is required", http.StatusBadRequest)
			return
		}
		obs, err := stores.ObservationStore.ListByMemberID(ctx, memberID)
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if obs == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(obs)
		return
	}

	if r.Method == "POST" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		var input struct {
			MemberID string `json:"MemberID"`
			Content  string `json:"Content"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		obs, err := orchestrators.ExecuteCreateObservation(ctx, orchestrators.CreateObservationInput{
			MemberID: input.MemberID,
			Content:  input.Content,
			AuthorID: sess.AccountID,
		}, orchestrators.CreateObservationDeps{
			ObservationStore: stores.ObservationStore,
			GenerateID:       generateID,
			Now:              timeNow,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(obs)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// --- Phase 1: Admin CRUD API Handlers ---

// requireAdmin checks the session for admin role and returns the session.
// Returns false if the request should not proceed.
func requireAdmin(w http.ResponseWriter, r *http.Request) (middleware.Session, bool) {
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		slog.Warn("auth_denied", "path", r.URL.Path, "reason", "no session")
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return middleware.Session{}, false
	}
	if sess.Role != "admin" {
		slog.Warn("auth_denied", "path", r.URL.Path, "account_id", sess.AccountID, "role", sess.Role, "required", "admin")
		http.Error(w, "Forbidden", http.StatusForbidden)
		return middleware.Session{}, false
	}
	return sess, true
}

// handleSchedules handles GET/POST/DELETE for /api/schedules
func handleSchedules(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		day := r.URL.Query().Get("day")
		var schedules []scheduleDomain.Schedule
		var err error
		if day != "" {
			schedules, err = stores.ScheduleStore.ListByDay(ctx, day)
		} else {
			schedules, err = stores.ScheduleStore.List(ctx)
		}
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if schedules == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(schedules)
		return
	}

	if r.Method == "POST" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		var input struct {
			ClassTypeID string `json:"ClassTypeID"`
			Day         string `json:"Day"`
			StartTime   string `json:"StartTime"`
			EndTime     string `json:"EndTime"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		sched := scheduleDomain.Schedule{
			ID:          generateID(),
			ClassTypeID: input.ClassTypeID,
			Day:         strings.ToLower(input.Day),
			StartTime:   input.StartTime,
			EndTime:     input.EndTime,
		}
		if err := sched.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.ScheduleStore.Save(ctx, sched); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(sched)
		return
	}

	if r.Method == "DELETE" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := stores.ScheduleStore.Delete(ctx, id); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// handleHolidays handles GET/POST/DELETE for /api/holidays
func handleHolidays(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		holidays, err := stores.HolidayStore.List(ctx)
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if holidays == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(holidays)
		return
	}

	if r.Method == "POST" {
		sess, ok := requireAdmin(w, r)
		if !ok {
			return
		}
		var input struct {
			Name      string `json:"Name"`
			StartDate string `json:"StartDate"`
			EndDate   string `json:"EndDate"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		startDate, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			http.Error(w, "StartDate must be YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		endDate, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			http.Error(w, "EndDate must be YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		h := holidayDomain.Holiday{
			ID:        generateID(),
			Name:      input.Name,
			StartDate: startDate,
			EndDate:   endDate,
		}
		if err := h.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.HolidayStore.Save(ctx, h); err != nil {
			internalError(w, err)
			return
		}
		// Auto-generate a holiday notice
		notice := noticeDomain.Notice{
			ID:        generateID(),
			Type:      noticeDomain.TypeHoliday,
			Status:    noticeDomain.StatusPublished,
			Title:     "Holiday: " + h.Name,
			Content:   fmt.Sprintf("Gym closed %s to %s: %s", input.StartDate, input.EndDate, h.Name),
			CreatedBy: sess.AccountID,
			CreatedAt: timeNow(),
		}
		_ = stores.NoticeStore.Save(ctx, notice)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(h)
		return
	}

	if r.Method == "DELETE" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := stores.HolidayStore.Delete(ctx, id); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// handleTerms handles GET/POST/DELETE for /api/terms
func handleTerms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		terms, err := stores.TermStore.List(ctx)
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if terms == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(terms)
		return
	}

	if r.Method == "POST" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		var input struct {
			Name      string `json:"Name"`
			StartDate string `json:"StartDate"`
			EndDate   string `json:"EndDate"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		startDate, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			http.Error(w, "StartDate must be YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		endDate, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			http.Error(w, "EndDate must be YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		t := termDomain.Term{
			ID:        generateID(),
			Name:      input.Name,
			StartDate: startDate,
			EndDate:   endDate,
		}
		if err := t.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.TermStore.Save(ctx, t); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)
		return
	}

	if r.Method == "DELETE" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := stores.TermStore.Delete(ctx, id); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// handleAccounts handles GET/POST for /api/accounts
func handleAccounts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		role := r.URL.Query().Get("role")
		filter := accountStore.ListFilter{Limit: 1000}
		if role != "" {
			filter.Role = role
		}
		accounts, err := stores.AccountStore.List(ctx, filter)
		if err != nil {
			internalError(w, err)
			return
		}
		// Strip password hashes from response
		type safeAccount struct {
			ID     string `json:"ID"`
			Email  string `json:"Email"`
			Role   string `json:"Role"`
			Status string `json:"Status"`
		}
		var safe []safeAccount
		for _, a := range accounts {
			safe = append(safe, safeAccount{ID: a.ID, Email: a.Email, Role: a.Role, Status: a.Status})
		}
		w.Header().Set("Content-Type", "application/json")
		if safe == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(safe)
		return
	}

	if r.Method == "POST" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		var input struct {
			Email    string `json:"Email"`
			Password string `json:"Password"`
			Role     string `json:"Role"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		acct := accountDomain.Account{
			ID:        generateID(),
			Email:     input.Email,
			Role:      input.Role,
			CreatedAt: timeNow(),
		}
		if err := acct.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response := map[string]string{
			"ID":    acct.ID,
			"Email": acct.Email,
			"Role":  acct.Role,
		}

		// Non-admin accounts use activation flow (no password required)
		if input.Role != accountDomain.RoleAdmin {
			acct.Status = accountDomain.StatusPendingActivation
			// Set a random placeholder password (will be replaced on activation)
			acct.PasswordHash = "pending_activation"
			if err := stores.AccountStore.Save(ctx, acct); err != nil {
				internalError(w, err)
				return
			}
			// Generate activation token
			tokenStr := generateID()
			tok := accountDomain.ActivationToken{
				ID:        generateID(),
				AccountID: acct.ID,
				Token:     tokenStr,
				ExpiresAt: timeNow().Add(72 * time.Hour),
				CreatedAt: timeNow(),
			}
			if err := stores.AccountStore.SaveActivationToken(ctx, tok); err != nil {
				internalError(w, err)
				return
			}
			response["Status"] = accountDomain.StatusPendingActivation
			response["ActivationToken"] = tokenStr
			slog.Info("auth_event", "event", "account_created_pending", "email", acct.Email, "role", acct.Role)
		} else {
			// Admin accounts require a password and are active immediately
			if err := acct.SetPassword(input.Password); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			acct.Status = accountDomain.StatusActive
			if err := stores.AccountStore.Save(ctx, acct); err != nil {
				internalError(w, err)
				return
			}
			response["Status"] = accountDomain.StatusActive
			slog.Info("auth_event", "event", "account_created", "email", acct.Email, "role", acct.Role)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// handleChangeRole handles POST /api/accounts/role
func handleChangeRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var input struct {
		AccountID string `json:"AccountID"`
		NewRole   string `json:"NewRole"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.AccountID == "" || input.NewRole == "" {
		http.Error(w, "AccountID and NewRole are required", http.StatusBadRequest)
		return
	}
	acct, err := stores.AccountStore.GetByID(r.Context(), input.AccountID)
	if err != nil {
		http.Error(w, "account not found", http.StatusNotFound)
		return
	}
	acct.Role = input.NewRole
	if err := acct.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := stores.AccountStore.Save(r.Context(), acct); err != nil {
		internalError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"ID":    acct.ID,
		"Email": acct.Email,
		"Role":  acct.Role,
	})
}

// handleAdminFeatureFlags handles GET/POST /api/admin/feature-flags
func handleAdminFeatureFlags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	type flagDTO struct {
		Key           string `json:"Key"`
		Description   string `json:"Description"`
		EnabledAdmin  bool   `json:"EnabledAdmin"`
		EnabledCoach  bool   `json:"EnabledCoach"`
		EnabledMember bool   `json:"EnabledMember"`
		EnabledTrial  bool   `json:"EnabledTrial"`
		BetaOverride  bool   `json:"BetaOverride"`
	}

	if r.Method == "GET" {
		persisted, err := stores.FeatureFlagStore.List(ctx)
		if err != nil {
			internalError(w, err)
			return
		}
		defaults := featureflagDomain.DefaultFlags()

		persistedByKey := make(map[string]featureflagDomain.FeatureFlag, len(persisted))
		for _, p := range persisted {
			persistedByKey[p.Key] = p
		}

		seen := make(map[string]bool, len(defaults))
		merged := make([]featureflagDomain.FeatureFlag, 0, len(defaults)+len(persisted))
		for _, d := range defaults {
			if p, ok := persistedByKey[d.Key]; ok {
				// Persisted values override defaults.
				if p.Description != "" {
					d.Description = p.Description
				}
				d.EnabledAdmin = p.EnabledAdmin
				d.EnabledCoach = p.EnabledCoach
				d.EnabledMember = p.EnabledMember
				d.EnabledTrial = p.EnabledTrial
				d.BetaOverride = p.BetaOverride
			}
			seen[d.Key] = true
			merged = append(merged, d)
		}
		for _, p := range persisted {
			if !seen[p.Key] {
				merged = append(merged, p)
			}
		}
		sort.Slice(merged, func(i, j int) bool { return merged[i].Key < merged[j].Key })

		out := make([]flagDTO, 0, len(merged))
		for _, ff := range merged {
			out = append(out, flagDTO{
				Key:           ff.Key,
				Description:   ff.Description,
				EnabledAdmin:  ff.EnabledAdmin,
				EnabledCoach:  ff.EnabledCoach,
				EnabledMember: ff.EnabledMember,
				EnabledTrial:  ff.EnabledTrial,
				BetaOverride:  ff.BetaOverride,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(out)
		return
	}

	if r.Method == "POST" {
		var input struct {
			Flags []flagDTO `json:"Flags"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if input.Flags == nil {
			input.Flags = []flagDTO{}
		}
		for _, dto := range input.Flags {
			ff := featureflagDomain.FeatureFlag{
				Key:           strings.TrimSpace(dto.Key),
				Description:   strings.TrimSpace(dto.Description),
				EnabledAdmin:  dto.EnabledAdmin,
				EnabledCoach:  dto.EnabledCoach,
				EnabledMember: dto.EnabledMember,
				EnabledTrial:  dto.EnabledTrial,
				BetaOverride:  dto.BetaOverride,
			}
			if err := ff.Validate(); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if err := stores.FeatureFlagStore.Save(ctx, ff); err != nil {
				internalError(w, err)
				return
			}
		}

		sess, _ := middleware.GetSessionFromContext(ctx)
		slog.Info("audit_event",
			"actor_id", sess.AccountID,
			"actor_role", sess.Role,
			"action", "admin.feature_flags.save",
			"count", len(input.Flags),
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// handleAdminBetaTesters handles GET/POST /api/admin/beta-testers
func handleAdminBetaTesters(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	type safeAccount struct {
		ID    string `json:"ID"`
		Email string `json:"Email"`
		Role  string `json:"Role"`
	}

	if r.Method == "GET" {
		accounts, err := stores.AccountStore.List(ctx, accountStore.ListFilter{Limit: 1000})
		if err != nil {
			internalError(w, err)
			return
		}
		var out []safeAccount
		for _, a := range accounts {
			if a.BetaTester {
				out = append(out, safeAccount{ID: a.ID, Email: a.Email, Role: a.Role})
			}
		}
		sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Email) < strings.ToLower(out[j].Email) })
		w.Header().Set("Content-Type", "application/json")
		if out == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(out)
		return
	}

	if r.Method == "POST" {
		var input struct {
			Email     string `json:"Email"`
			AccountID string `json:"AccountID"`
			Beta      bool   `json:"Beta"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if input.Email == "" && input.AccountID == "" {
			http.Error(w, "Email or AccountID is required", http.StatusBadRequest)
			return
		}
		var acct accountDomain.Account
		var err error
		if input.AccountID != "" {
			acct, err = stores.AccountStore.GetByID(ctx, input.AccountID)
		} else {
			acct, err = stores.AccountStore.GetByEmail(ctx, input.Email)
		}
		if err != nil {
			http.Error(w, "account not found", http.StatusNotFound)
			return
		}
		acct.BetaTester = input.Beta
		if err := stores.AccountStore.Save(ctx, acct); err != nil {
			internalError(w, err)
			return
		}

		sess, _ := middleware.GetSessionFromContext(ctx)
		slog.Info("audit_event",
			"actor_id", sess.AccountID,
			"actor_role", sess.Role,
			"action", "admin.beta_tester.set",
			"target_account_id", acct.ID,
			"beta", input.Beta,
		)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(safeAccount{ID: acct.ID, Email: acct.Email, Role: acct.Role})
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// --- Phase 2: Engagement Workflow Handlers ---

// handleNoticePublish handles POST /api/notices/publish
func handleNoticePublish(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var input struct {
		NoticeID string `json:"NoticeID"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	n, err := orchestrators.ExecutePublishNotice(r.Context(), orchestrators.PublishNoticeInput{
		NoticeID:    input.NoticeID,
		PublisherID: sess.AccountID,
	}, orchestrators.PublishNoticeDeps{
		NoticeStore: stores.NoticeStore,
		Now:         timeNow,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(n)
}

// handleNoticeEdit handles POST /api/notices/edit
func handleNoticeEdit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var input struct {
		NoticeID     string `json:"NoticeID"`
		Title        string `json:"Title"`
		Content      string `json:"Content"`
		Type         string `json:"Type"`
		AuthorName   string `json:"AuthorName"`
		ShowAuthor   bool   `json:"ShowAuthor"`
		Color        string `json:"Color"`
		VisibleFrom  string `json:"VisibleFrom"`
		VisibleUntil string `json:"VisibleUntil"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	orchInput := orchestrators.EditNoticeInput{
		NoticeID:   input.NoticeID,
		Title:      input.Title,
		Content:    input.Content,
		Type:       input.Type,
		AuthorName: input.AuthorName,
		ShowAuthor: input.ShowAuthor,
		Color:      input.Color,
	}
	if input.VisibleFrom != "" {
		if t, err := time.Parse(time.RFC3339, input.VisibleFrom); err == nil {
			orchInput.VisibleFrom = t
		}
	}
	if input.VisibleUntil != "" {
		if t, err := time.Parse(time.RFC3339, input.VisibleUntil); err == nil {
			orchInput.VisibleUntil = t
		}
	}
	n, err := orchestrators.ExecuteEditNotice(r.Context(), orchInput, orchestrators.EditNoticeDeps{
		NoticeStore: stores.NoticeStore,
		Now:         timeNow,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(n)
}

// handleNoticePin handles POST /api/notices/pin (toggle pin/unpin)
func handleNoticePin(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	var input struct {
		NoticeID string `json:"NoticeID"`
		Pinned   bool   `json:"Pinned"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	n, err := orchestrators.ExecutePinNotice(r.Context(), orchestrators.PinNoticeInput{
		NoticeID: input.NoticeID,
		Pinned:   input.Pinned,
	}, orchestrators.PinNoticeDeps{
		NoticeStore: stores.NoticeStore,
		Now:         timeNow,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(n)
}

// handleGradingDecide handles POST /api/grading/proposals/decide
func handleGradingDecide(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var input struct {
		ProposalID string `json:"ProposalID"`
		Decision   string `json:"Decision"` // "approve" or "reject"
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.ProposalID == "" {
		http.Error(w, "ProposalID is required", http.StatusBadRequest)
		return
	}
	proposal, err := stores.GradingProposalStore.GetByID(r.Context(), input.ProposalID)
	if err != nil {
		http.Error(w, "proposal not found", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	switch input.Decision {
	case "approve":
		if err := proposal.Approve(sess.AccountID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.GradingProposalStore.Save(ctx, proposal); err != nil {
			internalError(w, err)
			return
		}
		// Create the official grading record
		record := gradingDomain.Record{
			ID:         generateID(),
			MemberID:   proposal.MemberID,
			Belt:       proposal.TargetBelt,
			Stripe:     0,
			PromotedAt: timeNow(),
			ProposedBy: proposal.ProposedBy,
			ApprovedBy: sess.AccountID,
			Method:     gradingDomain.MethodStandard,
		}
		if err := stores.GradingRecordStore.Save(ctx, record); err != nil {
			internalError(w, err)
			return
		}
	case "reject":
		if err := proposal.Reject(sess.AccountID); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.GradingProposalStore.Save(ctx, proposal); err != nil {
			internalError(w, err)
			return
		}
	default:
		http.Error(w, "Decision must be 'approve' or 'reject'", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(proposal)
}

// handleGradingConfig handles GET/POST for /api/grading/config
func handleGradingConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		configs, err := stores.GradingConfigStore.List(ctx)
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if configs == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(configs)
		return
	}

	if r.Method == "POST" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		var input struct {
			Program         string  `json:"Program"`
			Belt            string  `json:"Belt"`
			FlightTimeHours float64 `json:"FlightTimeHours"`
			AttendancePct   float64 `json:"AttendancePct"`
			StripeCount     int     `json:"StripeCount"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		config := gradingDomain.Config{
			ID:              generateID(),
			Program:         strings.ToLower(strings.TrimSpace(input.Program)),
			Belt:            strings.ToLower(strings.TrimSpace(input.Belt)),
			FlightTimeHours: input.FlightTimeHours,
			AttendancePct:   input.AttendancePct,
			StripeCount:     input.StripeCount,
		}
		if err := config.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.GradingConfigStore.Save(ctx, config); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(config)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// handleGradingReadiness handles GET /api/grading/readiness
func handleGradingReadiness(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	ctx := r.Context()

	// Get all grading configs
	configs, err := stores.GradingConfigStore.List(ctx)
	if err != nil {
		internalError(w, err)
		return
	}

	// Get all active members
	members, err := stores.MemberStore.List(ctx, memberStore.ListFilter{Limit: 10000})
	if err != nil {
		internalError(w, err)
		return
	}

	type adultEntry struct {
		MemberID     string  `json:"MemberID"`
		MemberName   string  `json:"MemberName"`
		Program      string  `json:"Program"`
		CurrentBelt  string  `json:"CurrentBelt"`
		TargetBelt   string  `json:"TargetBelt"`
		MatHours     float64 `json:"MatHours"`
		RequiredHrs  float64 `json:"RequiredHours"`
		PercentReady float64 `json:"PercentReady"`
	}

	type kidsEntry struct {
		MemberID      string  `json:"MemberID"`
		MemberName    string  `json:"MemberName"`
		CurrentBelt   string  `json:"CurrentBelt"`
		TargetBelt    string  `json:"TargetBelt"`
		Attended      int     `json:"Attended"`
		TotalSessions int     `json:"TotalSessions"`
		AttendancePct float64 `json:"AttendancePct"`
		ThresholdPct  float64 `json:"ThresholdPct"`
		Eligible      bool    `json:"Eligible"`
	}

	type readinessResponse struct {
		Adults   []adultEntry `json:"Adults"`
		Kids     []kidsEntry  `json:"Kids"`
		TermName string       `json:"TermName"`
	}

	var adults []adultEntry
	for _, m := range members {
		if m.Status != "active" {
			continue
		}
		// Skip kids in sessions mode ΓÇö they appear in kids term attendance section
		if m.Program == "kids" && m.GradingMetric != memberDomain.MetricHours {
			continue
		}
		// Get member's latest grading record to find current belt
		records, err := stores.GradingRecordStore.ListByMemberID(ctx, m.ID)
		if err != nil {
			continue
		}
		currentBelt := "white"
		if len(records) > 0 {
			currentBelt = records[len(records)-1].Belt
		}

		// Find the config for their program + next belt
		nextBelt := nextBeltFor(currentBelt, m.Program)
		if nextBelt == "" {
			continue // already at highest belt
		}

		var requiredHours float64
		for _, c := range configs {
			if c.Program == m.Program && c.Belt == nextBelt {
				requiredHours = c.FlightTimeHours
				break
			}
		}
		if requiredHours <= 0 {
			continue // no config for this belt
		}

		// Get training log for mat hours
		query := projections.GetTrainingLogQuery{MemberID: m.ID}
		deps := projections.GetTrainingLogDeps{
			AttendanceStore: stores.AttendanceStore,
			MemberStore:     stores.MemberStore,
		}
		log, err := projections.QueryGetTrainingLog(ctx, query, deps)
		if err != nil {
			continue
		}

		pct := (log.TotalMatHours / requiredHours) * 100
		if pct > 100 {
			pct = 100
		}
		if pct >= 50 { // only show members at 50%+ readiness
			adults = append(adults, adultEntry{
				MemberID:     m.ID,
				MemberName:   m.Name,
				Program:      m.Program,
				CurrentBelt:  currentBelt,
				TargetBelt:   nextBelt,
				MatHours:     log.TotalMatHours,
				RequiredHrs:  requiredHours,
				PercentReady: pct,
			})
		}
	}

	// Kids term attendance readiness
	var kids []kidsEntry
	termName := ""
	kidsQuery := projections.GetKidsTermReadinessQuery{Now: time.Now()}
	kidsDeps := projections.GetKidsTermReadinessDeps{
		TermStore:          stores.TermStore,
		ProgramStore:       stores.ProgramStore,
		ClassTypeStore:     stores.ClassTypeStore,
		ScheduleStore:      stores.ScheduleStore,
		HolidayStore:       stores.HolidayStore,
		MemberStore:        stores.MemberStore,
		AttendanceStore:    stores.AttendanceStore,
		GradingRecordStore: stores.GradingRecordStore,
		GradingConfigStore: stores.GradingConfigStore,
	}
	kidsResult, err := projections.QueryGetKidsTermReadiness(ctx, kidsQuery, kidsDeps)
	if err == nil {
		termName = kidsResult.TermName
		for _, e := range kidsResult.Entries {
			kids = append(kids, kidsEntry{
				MemberID:      e.MemberID,
				MemberName:    e.MemberName,
				CurrentBelt:   e.CurrentBelt,
				TargetBelt:    e.TargetBelt,
				Attended:      e.Attended,
				TotalSessions: e.TotalSessions,
				AttendancePct: e.AttendancePct,
				ThresholdPct:  e.ThresholdPct,
				Eligible:      e.Eligible,
			})
		}
	}

	// Sort by proximity to eligibility (highest % first) ΓÇö #63
	sort.Slice(adults, func(i, j int) bool {
		return adults[i].PercentReady > adults[j].PercentReady
	})
	sort.Slice(kids, func(i, j int) bool {
		return kids[i].AttendancePct > kids[j].AttendancePct
	})

	resp := readinessResponse{Adults: adults, Kids: kids, TermName: termName}
	if resp.Adults == nil {
		resp.Adults = []adultEntry{}
	}
	if resp.Kids == nil {
		resp.Kids = []kidsEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleGradingForcePromote handles POST /api/grading/force-promote
// Allows admin to immediately promote a member, bypassing the proposal flow.
func handleGradingForcePromote(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var input struct {
		MemberID   string `json:"MemberID"`
		TargetBelt string `json:"TargetBelt"`
		Reason     string `json:"Reason"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	record := gradingDomain.Record{
		ID:         generateID(),
		MemberID:   input.MemberID,
		Belt:       input.TargetBelt,
		Stripe:     0,
		PromotedAt: timeNow(),
		ProposedBy: sess.AccountID,
		ApprovedBy: sess.AccountID,
		Method:     gradingDomain.MethodOverride,
	}
	if err := record.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := stores.GradingRecordStore.Save(r.Context(), record); err != nil {
		internalError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(record)
}

// handleGradingMemberConfig handles GET/POST for /api/grading/member-config
// Allows admin to set per-member threshold overrides for grading eligibility.
func handleGradingMemberConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if sess.Role != "admin" && sess.Role != "coach" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		memberID := r.URL.Query().Get("member_id")
		if memberID == "" {
			http.Error(w, "member_id is required", http.StatusBadRequest)
			return
		}
		configs, err := stores.GradingMemberConfigStore.ListByMemberID(ctx, memberID)
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if configs == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(configs)
		return
	}

	if r.Method == "POST" {
		sess, ok := requireAdmin(w, r)
		if !ok {
			return
		}
		_ = sess
		var input struct {
			MemberID        string  `json:"MemberID"`
			Belt            string  `json:"Belt"`
			FlightTimeHours float64 `json:"FlightTimeHours"`
			AttendancePct   float64 `json:"AttendancePct"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		mc := gradingDomain.MemberConfig{
			ID:              generateID(),
			MemberID:        input.MemberID,
			Belt:            input.Belt,
			FlightTimeHours: input.FlightTimeHours,
			AttendancePct:   input.AttendancePct,
		}
		if err := mc.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.GradingMemberConfigStore.Save(ctx, mc); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(mc)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleGradingCredit handles POST /api/grading/credit
// Allows admin to add a direct mat hours credit to a member's record.
func handleGradingCredit(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := requireAdmin(w, r)
	if !ok {
		return
	}
	var input struct {
		MemberID string  `json:"MemberID"`
		Hours    float64 `json:"Hours"`
		Reason   string  `json:"Reason"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.MemberID == "" {
		http.Error(w, "MemberID is required", http.StatusBadRequest)
		return
	}
	if input.Hours <= 0 || input.Hours > 1000 {
		http.Error(w, "Hours must be between 0 and 1000", http.StatusBadRequest)
		return
	}
	today := timeNow().Format("2006-01-02")
	entry := estimatedHoursDomain.EstimatedHours{
		ID:          generateID(),
		MemberID:    input.MemberID,
		StartDate:   today,
		EndDate:     today,
		WeeklyHours: input.Hours,
		TotalHours:  input.Hours,
		Source:      estimatedHoursDomain.SourceCredit,
		Status:      estimatedHoursDomain.StatusApproved,
		Note:        input.Reason,
		CreatedBy:   sess.AccountID,
		CreatedAt:   timeNow(),
	}
	if err := entry.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := stores.EstimatedHoursStore.Save(r.Context(), entry); err != nil {
		internalError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(entry)
}

// handleGradingMetricToggle handles POST /api/grading/metric
// Toggles a kid's grading metric between "sessions" and "hours".
func handleGradingMetricToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var body struct {
		MemberID string `json:"MemberID"`
		Metric   string `json:"Metric"`
	}
	if err := strictDecode(r, &body); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	if body.Metric != memberDomain.MetricSessions && body.Metric != memberDomain.MetricHours {
		http.Error(w, "Metric must be 'sessions' or 'hours'", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	m, err := stores.MemberStore.GetByID(ctx, body.MemberID)
	if err != nil {
		http.Error(w, "Member not found", http.StatusNotFound)
		return
	}
	if m.Program != memberDomain.ProgramKids {
		http.Error(w, "Metric toggle is only for kids", http.StatusBadRequest)
		return
	}

	m.GradingMetric = body.Metric
	if err := stores.MemberStore.Save(ctx, m); err != nil {
		internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// nextBeltFor returns the next belt in progression, or "" if at highest.
func nextBeltFor(current, program string) string {
	var progression []string
	if program == "kids" {
		progression = gradingDomain.KidsBelts
	} else {
		progression = gradingDomain.AdultBelts
	}
	for i, b := range progression {
		if b == current && i+1 < len(progression) {
			return progression[i+1]
		}
	}
	return ""
}

// handleTrainingGoals handles GET/POST/DELETE for /api/training-goals
func handleTrainingGoals(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	if r.Method == "GET" {
		memberID := r.URL.Query().Get("member_id")
		if memberID == "" {
			http.Error(w, "member_id is required", http.StatusBadRequest)
			return
		}
		goals, err := stores.TrainingGoalStore.ListByMemberID(ctx, memberID)
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if goals == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(goals)
		return
	}

	if r.Method == "POST" {
		var input struct {
			MemberID string `json:"MemberID"`
			Target   int    `json:"Target"`
			Period   string `json:"Period"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		goal := trainingGoalDomain.TrainingGoal{
			ID:        generateID(),
			MemberID:  input.MemberID,
			Target:    input.Target,
			Period:    input.Period,
			CreatedAt: timeNow(),
			Active:    true,
		}
		if err := goal.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.TrainingGoalStore.Save(ctx, goal); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(goal)
		return
	}

	if r.Method == "DELETE" {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := stores.TrainingGoalStore.Delete(ctx, id); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	_ = sess // used for auth check
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// handleMilestones handles GET/POST/DELETE for /api/milestones
func handleMilestones(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		if _, ok := middleware.GetSessionFromContext(ctx); !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		milestones, err := stores.MilestoneStore.List(ctx)
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if milestones == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(milestones)
		return
	}

	if r.Method == "POST" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		var input struct {
			Name      string  `json:"Name"`
			Metric    string  `json:"Metric"`
			Threshold float64 `json:"Threshold"`
			BadgeIcon string  `json:"BadgeIcon"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		ms := milestoneDomain.Milestone{
			ID:        generateID(),
			Name:      input.Name,
			Metric:    input.Metric,
			Threshold: input.Threshold,
			BadgeIcon: input.BadgeIcon,
		}
		if err := ms.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.MilestoneStore.Save(ctx, ms); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ms)
		return
	}

	if r.Method == "DELETE" {
		if _, ok := requireAdmin(w, r); !ok {
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := stores.MilestoneStore.Delete(ctx, id); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// handleMemberMilestones handles GET /api/member-milestones?member_id=<id>
// Returns earned milestones for the member, evaluating current training stats.
func handleMemberMilestones(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := middleware.GetSessionFromContext(r.Context()); !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	memberID := r.URL.Query().Get("member_id")
	if memberID == "" {
		http.Error(w, "member_id is required", http.StatusBadRequest)
		return
	}

	// Get training stats
	logQuery := projections.GetTrainingLogQuery{MemberID: memberID}
	logDeps := projections.GetTrainingLogDeps{
		AttendanceStore: stores.AttendanceStore,
		MemberStore:     stores.MemberStore,
	}
	logResult, err := projections.QueryGetTrainingLog(r.Context(), logQuery, logDeps)
	if err != nil {
		internalError(w, err)
		return
	}

	// Check milestones
	checkDeps := projections.CheckMilestonesDeps{
		MilestoneStore:       stores.MilestoneStore,
		MemberMilestoneStore: stores.MemberMilestoneStore,
	}
	earned, err := projections.QueryCheckMilestones(r.Context(), projections.CheckMilestonesInput{
		MemberID:      memberID,
		TotalClasses:  logResult.TotalClasses,
		TotalMatHours: logResult.TotalMatHours,
		CurrentStreak: logResult.CurrentStreak,
	}, checkDeps)
	if err != nil {
		internalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if earned == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(earned)
}

// handleMemberMilestoneDismiss handles POST /api/member-milestones/dismiss
// Marks a milestone notification as seen.
func handleMemberMilestoneDismiss(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := middleware.GetSessionFromContext(r.Context()); !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	var input struct {
		ID string `json:"ID"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.ID == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}
	if err := stores.MemberMilestoneStore.MarkNotified(r.Context(), input.ID); err != nil {
		internalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleMessageRead handles POST /api/messages/read
func handleMessageRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := middleware.GetSessionFromContext(r.Context()); !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	var input struct {
		MessageID string `json:"MessageID"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.MessageID == "" {
		http.Error(w, "MessageID is required", http.StatusBadRequest)
		return
	}
	msg, err := stores.MessageStore.GetByID(r.Context(), input.MessageID)
	if err != nil {
		http.Error(w, "message not found", http.StatusNotFound)
		return
	}
	msg.MarkRead()
	if err := stores.MessageStore.Save(r.Context(), msg); err != nil {
		internalError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(msg)
}

// --- Admin Page Handlers ---

// handleClassTypes handles GET /api/class-types
func handleClassTypes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		programID := r.URL.Query().Get("program_id")
		var (
			types []classTypeDomain.ClassType
			err   error
		)
		if programID != "" {
			types, err = stores.ClassTypeStore.ListByProgramID(ctx, programID)
		} else {
			types, err = stores.ClassTypeStore.List(ctx)
		}
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if types == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(types)
		return
	}

	// Writes are admin-only.
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	if r.Method == "POST" {
		var input struct {
			ProgramID   string `json:"ProgramID"`
			Name        string `json:"Name"`
			Description string `json:"Description"`
			Attire      string `json:"Attire"`
			Level       string `json:"Level"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		ct := classTypeDomain.ClassType{
			ID:          generateID(),
			ProgramID:   input.ProgramID,
			Name:        input.Name,
			Description: input.Description,
			Attire:      input.Attire,
			Level:       input.Level,
		}
		if err := ct.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.ClassTypeStore.Save(ctx, ct); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(ct)
		return
	}

	if r.Method == "PUT" {
		var input struct {
			ID          string `json:"ID"`
			ProgramID   string `json:"ProgramID"`
			Name        string `json:"Name"`
			Description string `json:"Description"`
			Attire      string `json:"Attire"`
			Level       string `json:"Level"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if input.ID == "" {
			http.Error(w, "ID is required", http.StatusBadRequest)
			return
		}
		ct := classTypeDomain.ClassType{
			ID:          input.ID,
			ProgramID:   input.ProgramID,
			Name:        input.Name,
			Description: input.Description,
			Attire:      input.Attire,
			Level:       input.Level,
		}
		if err := ct.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.ClassTypeStore.Save(ctx, ct); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ct)
		return
	}

	if r.Method == "DELETE" {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := stores.ClassTypeStore.Delete(ctx, id); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// handlePrograms handles GET /api/programs (admin-only).
func handlePrograms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	items, err := stores.ProgramStore.List(ctx)
	if err != nil {
		internalError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if items == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(items)
}

// handleAdminSchedulesPage handles GET /admin/schedules
func handleAdminSchedulesPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	renderTemplate(w, r, "admin_schedules.html", nil)
}

// handleAdminClassTypesPage handles GET /admin/class-types
func handleAdminClassTypesPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	renderTemplate(w, r, "admin_class_types.html", nil)
}

// handleAdminHolidaysPage handles GET /admin/holidays
func handleAdminHolidaysPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	renderTemplate(w, r, "admin_holidays.html", nil)
}

// handleAdminTermsPage handles GET /admin/terms
func handleAdminTermsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	renderTemplate(w, r, "admin_terms.html", nil)
}

// handleAdminAccountsPage handles GET /admin/accounts
func handleAdminAccountsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	renderTemplate(w, r, "admin_accounts.html", nil)
}

// handleAdminFeaturesPage handles GET /admin/features
func handleAdminFeaturesPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	renderTemplate(w, r, "admin_features.html", nil)
}

// handleAdminNoticesPage handles GET /admin/notices
func handleAdminNoticesPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	renderTemplate(w, r, "admin_notices.html", nil)
}

// handleAdminGradingPage handles GET /admin/grading
func handleAdminGradingPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	renderTemplate(w, r, "admin_grading.html", nil)
}

// handleAdminInactivePage handles GET /admin/inactive
func handleAdminInactivePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	renderTemplate(w, r, "admin_inactive.html", nil)
}

// handleAdminMilestonesPage handles GET /admin/milestones
func handleAdminMilestonesPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	renderTemplate(w, r, "admin_milestones.html", nil)
}

// handleAdminPerfPage handles GET /admin/perf
func handleAdminPerfPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	if perfCollector == nil {
		renderTemplate(w, r, "admin_perf.html", nil)
		return
	}
	snap := perfCollector.Snapshot(time.Now().Add(-1*time.Hour), 10)
	renderTemplate(w, r, "admin_perf.html", snap)
}

// handleTrainingLogPage handles GET /training-log
func handleTrainingLogPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	m, err := stores.MemberStore.GetByEmail(r.Context(), sess.Email)
	memberID := ""
	if err == nil {
		memberID = m.ID
	}
	renderTemplate(w, r, "member_training_log.html", map[string]any{
		"Email":    sess.Email,
		"MemberID": memberID,
	})
}

// handleMessagesPage handles GET /messages
func handleMessagesPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	m, err := stores.MemberStore.GetByEmail(r.Context(), sess.Email)
	memberID := ""
	if err == nil {
		memberID = m.ID
	}
	renderTemplate(w, r, "member_messages.html", map[string]any{
		"Email":    sess.Email,
		"MemberID": memberID,
	})
}

// handleMemberInboxPage handles GET /inbox ΓÇö shows emails sent to the current member.
func handleMemberInboxPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	m, err := stores.MemberStore.GetByEmail(r.Context(), sess.Email)
	memberID := ""
	if err == nil {
		memberID = m.ID
	}
	renderTemplate(w, r, "member_inbox.html", map[string]any{
		"Email":    sess.Email,
		"MemberID": memberID,
	})
}

// handleMemberInboxAPI handles GET /api/inbox?member_id=...
func handleMemberInboxAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	// Members can only view their own inbox; admins can view any
	memberID := r.URL.Query().Get("member_id")
	if memberID == "" {
		// Look up member by session email
		m, err := stores.MemberStore.GetByEmail(r.Context(), sess.Email)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("[]"))
			return
		}
		memberID = m.ID
	} else if sess.Role != "admin" {
		// Non-admin trying to view another member's inbox
		m, err := stores.MemberStore.GetByEmail(r.Context(), sess.Email)
		if err != nil || m.ID != memberID {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	emails, err := stores.EmailStore.ListByRecipientMemberID(r.Context(), memberID)
	if err != nil {
		internalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if emails == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(emails)
}

// --- Phase 3: Dashboard & Kiosk Handlers ---

// handleDashboard handles GET /dashboard ΓÇö renders role-appropriate dashboard.
func handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	ctx := r.Context()
	query := projections.GetDashboardQuery{
		Role:         sess.Role,
		AccountEmail: sess.Email,
	}
	deps := projections.GetDashboardDeps{
		TodaysClassesDeps: projections.GetTodaysClassesDeps{
			ScheduleStore:  stores.ScheduleStore,
			TermStore:      stores.TermStore,
			HolidayStore:   stores.HolidayStore,
			ClassTypeStore: stores.ClassTypeStore,
			ProgramStore:   stores.ProgramStore,
		},
		AttendanceDeps: projections.GetAttendanceTodayDeps{
			AttendanceStore:    stores.AttendanceStore,
			MemberStore:        stores.MemberStore,
			InjuryStore:        stores.InjuryStore,
			GradingRecordStore: stores.GradingRecordStore,
			ScheduleStore:      stores.ScheduleStore,
			ClassTypeStore:     stores.ClassTypeStore,
		},
		InactiveDeps: projections.GetInactiveMembersDeps{
			MemberStore:     stores.MemberStore,
			AttendanceStore: stores.AttendanceStore,
		},
		TrainingLogDeps: projections.GetTrainingLogDeps{
			AttendanceStore: stores.AttendanceStore,
			MemberStore:     stores.MemberStore,
		},
		NoticeStore:        stores.NoticeStore,
		ProposalStore:      stores.GradingProposalStore,
		MessageStore:       stores.MessageStore,
		TrainingGoalStore:  stores.TrainingGoalStore,
		MemberStore:        stores.MemberStore,
		GradingRecordStore: stores.GradingRecordStore,
		WaiverStore:        stores.WaiverStore,
	}

	result, err := projections.QueryGetDashboard(ctx, query, deps, timeNow())
	if err != nil {
		internalError(w, err)
		return
	}

	var templateName string
	switch sess.Role {
	case "admin":
		templateName = "dashboard_admin.html"
	case "coach":
		templateName = "dashboard_coach.html"
	default:
		templateName = "dashboard_member.html"
	}

	renderTemplate(w, r, templateName, result)
}

// handleKioskPage handles GET /kiosk ΓÇö renders the standalone kiosk UI.
func handleKioskPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Kiosk is a standalone template (no layout.html)
	kioskPath := filepath.Join(templatesDir, "kiosk.html")
	tpl, err := template.ParseFiles(kioskPath)
	if err != nil {
		internalError(w, err)
		return
	}
	tpl.Execute(w, nil)
}

// --- Layer 2: Spine Handlers ---

// handleThemes handles GET/POST for /api/themes
func handleThemes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if !requireFeatureAPI(w, r, sess, "library") {
			return
		}
		program := r.URL.Query().Get("program")
		var themes []themeDomain.Theme
		var err error
		if program != "" {
			themes, err = stores.ThemeStore.ListByProgram(ctx, program)
		} else {
			themes, err = stores.ThemeStore.List(ctx)
		}
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if themes == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(themes)
		return
	}

	if r.Method == "POST" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if sess.Role != "admin" && sess.Role != "coach" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if !requireFeatureAPI(w, r, sess, "library") {
			return
		}
		var input struct {
			Name        string `json:"Name"`
			Description string `json:"Description"`
			Program     string `json:"Program"`
			StartDate   string `json:"StartDate"`
			EndDate     string `json:"EndDate"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		startDate, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			http.Error(w, "invalid start date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		endDate, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			http.Error(w, "invalid end date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		theme := themeDomain.Theme{
			ID:          generateID(),
			Name:        input.Name,
			Description: input.Description,
			Program:     input.Program,
			StartDate:   startDate,
			EndDate:     endDate,
			CreatedBy:   sess.AccountID,
			CreatedAt:   timeNow(),
		}
		if err := theme.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.ThemeStore.Save(ctx, theme); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(theme)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleClips handles GET/POST for /api/clips
func handleClips(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if r.Method == "GET" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if !requireFeatureAPI(w, r, sess, "library") {
			return
		}
		themeID := r.URL.Query().Get("theme_id")
		promoted := r.URL.Query().Get("promoted")
		var clips []clipDomain.Clip
		var err error
		if promoted == "true" {
			clips, err = stores.ClipStore.ListPromoted(ctx)
		} else if themeID != "" {
			clips, err = stores.ClipStore.ListByThemeID(ctx, themeID)
		} else {
			http.Error(w, "theme_id or promoted=true is required", http.StatusBadRequest)
			return
		}
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if clips == nil {
			w.Write([]byte("[]"))
			return
		}
		json.NewEncoder(w).Encode(clips)
		return
	}

	if r.Method == "POST" {
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		if !requireFeatureAPI(w, r, sess, "library") {
			return
		}
		var input struct {
			ThemeID      string `json:"ThemeID"`
			Title        string `json:"Title"`
			YouTubeURL   string `json:"YouTubeURL"`
			StartSeconds int    `json:"StartSeconds"`
			EndSeconds   int    `json:"EndSeconds"`
			Notes        string `json:"Notes"`
		}
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		clip := clipDomain.Clip{
			ID:           generateID(),
			ThemeID:      input.ThemeID,
			Title:        input.Title,
			YouTubeURL:   input.YouTubeURL,
			StartSeconds: input.StartSeconds,
			EndSeconds:   input.EndSeconds,
			Notes:        input.Notes,
			CreatedBy:    sess.AccountID,
			CreatedAt:    timeNow(),
		}
		if err := clip.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := clip.ExtractYouTubeID(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.ClipStore.Save(ctx, clip); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(clip)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleClipPromote handles POST /api/clips/promote
func handleClipPromote(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !requireFeatureAPI(w, r, sess, "library") {
		return
	}
	var input struct {
		ClipID string `json:"ClipID"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.ClipID == "" {
		http.Error(w, "ClipID is required", http.StatusBadRequest)
		return
	}
	clip, err := stores.ClipStore.GetByID(r.Context(), input.ClipID)
	if err != nil {
		http.Error(w, "clip not found", http.StatusNotFound)
		return
	}
	if err := clip.Promote(sess.AccountID); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if err := stores.ClipStore.Save(r.Context(), clip); err != nil {
		internalError(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(clip)
}

// handleThemesPage handles GET /themes ΓÇö renders the theme carousel page.
func handleThemesPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	if !requireFeaturePage(w, r, sess, "library") {
		return
	}
	renderTemplate(w, r, "themes.html", map[string]any{
		"Email": sess.Email,
		"Role":  sess.Role,
	})
}

// handleLibraryPage handles GET /library ΓÇö renders the technical library page.
func handleLibraryPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	if !requireFeaturePage(w, r, sess, "library") {
		return
	}
	renderTemplate(w, r, "library.html", map[string]any{
		"Email": sess.Email,
		"Role":  sess.Role,
	})
}

// --- DevMode: Admin Impersonation ---

// handleDevModeImpersonate handles POST /api/devmode/impersonate
func handleDevModeImpersonate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	if !middleware.IsRealAdmin(r.Context()) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Form error", http.StatusBadRequest)
		return
	}

	targetRole := r.FormValue("role")
	input := orchestrators.DevModeImpersonateInput{
		TargetRole:    targetRole,
		CurrentRole:   sess.Role,
		AccountID:     sess.AccountID,
		Email:         sess.Email,
		RealAccountID: sess.RealAccountID,
		RealRole:      sess.RealRole,
		RealEmail:     sess.RealEmail,
	}

	result, err := orchestrators.ExecuteDevModeImpersonate(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update session in-place
	cookie, err := r.Cookie("workshop_session")
	if err != nil {
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	sess.Role = result.Role
	sess.RealAccountID = result.RealAccountID
	sess.RealEmail = result.RealEmail
	sess.RealRole = result.RealRole
	// Restore AccountID/Email when switching back to admin
	if result.RealRole == "" && result.Role == "admin" {
		if sess.RealAccountID != "" {
			sess.AccountID = sess.RealAccountID
			sess.Email = sess.RealEmail
		}
		sess.RealAccountID = ""
		sess.RealEmail = ""
		sess.RealRole = ""
	}

	sessions.Update(cookie.Value, sess)

	slog.Info("devmode_event",
		"event", "impersonate",
		"admin_account_id", func() string {
			if result.RealAccountID != "" {
				return result.RealAccountID
			}
			return sess.AccountID
		}(),
		"target_role", result.Role,
	)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// handleDevModeRestore handles POST /api/devmode/restore
func handleDevModeRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	input := orchestrators.DevModeRestoreInput{
		CurrentRole:   sess.Role,
		RealAccountID: sess.RealAccountID,
		RealEmail:     sess.RealEmail,
		RealRole:      sess.RealRole,
	}

	result, err := orchestrators.ExecuteDevModeRestore(input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update session in-place
	cookie, err := r.Cookie("workshop_session")
	if err != nil {
		http.Error(w, "session error", http.StatusInternalServerError)
		return
	}

	sess.AccountID = result.AccountID
	sess.Email = result.Email
	sess.Role = result.Role
	sess.RealAccountID = ""
	sess.RealEmail = ""
	sess.RealRole = ""

	sessions.Update(cookie.Value, sess)

	slog.Info("devmode_event",
		"event", "restore",
		"admin_account_id", result.AccountID,
	)

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

// --- Phase 3: Email System Handlers ---

// memberLookupAdapter bridges the member store to the orchestrator's MemberLookup interface.
type memberLookupAdapter struct{}

// GetEmailByMemberID resolves a member's name and email from the member store.
// PRE: memberID is non-empty
// POST: Returns member name and email, or error if not found
func (a *memberLookupAdapter) GetEmailByMemberID(ctx context.Context, memberID string) (string, string, error) {
	m, err := stores.MemberStore.GetByID(ctx, memberID)
	if err != nil {
		return "", "", err
	}
	return m.Name, m.Email, nil
}

// handleAdminEmailsPage handles GET /admin/emails
func handleAdminEmailsPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	ctx := r.Context()
	emails, err := stores.EmailStore.List(ctx, emailStoreImport.ListFilter{})
	if err != nil {
		internalError(w, err)
		return
	}

	renderTemplate(w, r, "admin_emails.html", map[string]any{
		"Emails": emails,
	})
}

// handleAdminComposeEmailPage handles GET /admin/emails/compose
func handleAdminComposeEmailPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	emailID := r.URL.Query().Get("id")
	var draft emailDomain.Email
	var recipients []emailDomain.Recipient
	if emailID != "" {
		var err error
		draft, err = stores.EmailStore.GetByID(r.Context(), emailID)
		if err != nil {
			http.Error(w, "email not found", http.StatusNotFound)
			return
		}
		recipients, _ = stores.EmailStore.GetRecipients(r.Context(), emailID)
	}

	renderTemplate(w, r, "admin_compose_email.html", map[string]any{
		"Draft":      draft,
		"Recipients": recipients,
	})
}

// handleEmailCompose handles POST /api/emails/compose (save draft)
func handleEmailCompose(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := requireAdmin(w, r)
	if !ok {
		return
	}

	var input struct {
		EmailID   string   `json:"EmailID"`
		Subject   string   `json:"Subject"`
		Body      string   `json:"Body"`
		MemberIDs []string `json:"MemberIDs"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	em, err := orchestrators.ExecuteComposeEmail(r.Context(), orchestrators.ComposeEmailInput{
		EmailID:   input.EmailID,
		Subject:   input.Subject,
		Body:      input.Body,
		SenderID:  sess.AccountID,
		MemberIDs: input.MemberIDs,
	}, orchestrators.ComposeEmailDeps{
		EmailStore:   stores.EmailStore,
		MemberLookup: &memberLookupAdapter{},
		GenerateID:   generateID,
		Now:          timeNow,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(em)
}

// handleEmailSend handles POST /api/emails/send
func handleEmailSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := requireAdmin(w, r)
	if !ok {
		return
	}

	var input struct {
		EmailID string `json:"EmailID"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if emailSender == nil {
		http.Error(w, "email sending is not configured", http.StatusServiceUnavailable)
		return
	}

	em, err := orchestrators.ExecuteSendEmail(r.Context(), orchestrators.SendEmailInput{
		EmailID:  input.EmailID,
		SenderID: sess.AccountID,
	}, orchestrators.SendEmailDeps{
		EmailStore:  stores.EmailStore,
		EmailSender: emailSender,
		Now:         timeNow,
		FromAddress: emailFromAddress,
		ReplyTo:     emailReplyTo,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(em)
}

// handleEmailTestSend handles POST /api/emails/test-send
func handleEmailTestSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	var input struct {
		EmailID     string `json:"EmailID"`
		TestAddress string `json:"TestAddress"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if emailSender == nil {
		http.Error(w, "email sending is not configured", http.StatusServiceUnavailable)
		return
	}

	err := orchestrators.ExecuteTestSendEmail(r.Context(), orchestrators.TestSendEmailInput{
		EmailID:     input.EmailID,
		TestAddress: input.TestAddress,
	}, orchestrators.TestSendEmailDeps{
		EmailStore:  stores.EmailStore,
		EmailSender: emailSender,
		FromAddress: emailFromAddress,
		ReplyTo:     emailReplyTo,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "sent", "address": input.TestAddress})
}

// handleEmailList handles GET /api/emails
func handleEmailList(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("q")
	emails, err := stores.EmailStore.List(r.Context(), emailStoreImport.ListFilter{Status: status, Search: search})
	if err != nil {
		internalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if emails == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(emails)
}

// handleEmailDetail handles GET /api/emails/detail?id=...
func handleEmailDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	em, err := stores.EmailStore.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "email not found", http.StatusNotFound)
		return
	}

	recipients, _ := stores.EmailStore.GetRecipients(r.Context(), id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"Email":      em,
		"Recipients": recipients,
	})
}

// handleEmailDelete handles DELETE /api/emails?id=...
func handleEmailDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	em, err := stores.EmailStore.GetByID(r.Context(), id)
	if err != nil {
		http.Error(w, "email not found", http.StatusNotFound)
		return
	}
	if !em.IsDraft() {
		http.Error(w, "only draft emails can be deleted", http.StatusBadRequest)
		return
	}

	if err := stores.EmailStore.Delete(r.Context(), id); err != nil {
		internalError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleEmailSchedule handles POST /api/emails/schedule
func handleEmailSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := requireAdmin(w, r)
	if !ok {
		return
	}

	var input struct {
		EmailID     string `json:"EmailID"`
		ScheduledAt string `json:"ScheduledAt"` // RFC3339
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.EmailID == "" || input.ScheduledAt == "" {
		http.Error(w, "EmailID and ScheduledAt are required", http.StatusBadRequest)
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, input.ScheduledAt)
	if err != nil {
		http.Error(w, "ScheduledAt must be in RFC3339 format", http.StatusBadRequest)
		return
	}

	_ = sess // sender verified via requireAdmin
	em, err := orchestrators.ExecuteScheduleEmail(r.Context(), orchestrators.ScheduleEmailInput{
		EmailID:     input.EmailID,
		ScheduledAt: scheduledAt,
	}, orchestrators.ScheduleEmailDeps{
		EmailStore: stores.EmailStore,
		Now:        timeNow,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(em)
}

// handleEmailCancel handles POST /api/emails/cancel
func handleEmailCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	var input struct {
		EmailID string `json:"EmailID"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.EmailID == "" {
		http.Error(w, "EmailID is required", http.StatusBadRequest)
		return
	}

	em, err := orchestrators.ExecuteCancelEmail(r.Context(), orchestrators.CancelEmailInput{
		EmailID: input.EmailID,
	}, orchestrators.CancelEmailDeps{
		EmailStore: stores.EmailStore,
		Now:        timeNow,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(em)
}

// handleEmailReschedule handles POST /api/emails/reschedule
func handleEmailReschedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	var input struct {
		EmailID     string `json:"EmailID"`
		ScheduledAt string `json:"ScheduledAt"` // RFC3339
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.EmailID == "" || input.ScheduledAt == "" {
		http.Error(w, "EmailID and ScheduledAt are required", http.StatusBadRequest)
		return
	}

	scheduledAt, err := time.Parse(time.RFC3339, input.ScheduledAt)
	if err != nil {
		http.Error(w, "ScheduledAt must be in RFC3339 format", http.StatusBadRequest)
		return
	}

	em, err := orchestrators.ExecuteRescheduleEmail(r.Context(), orchestrators.RescheduleEmailInput{
		EmailID:     input.EmailID,
		ScheduledAt: scheduledAt,
	}, orchestrators.RescheduleEmailDeps{
		EmailStore: stores.EmailStore,
		Now:        timeNow,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(em)
}

// handleEmailTemplatePage handles GET /admin/emails/template
func handleEmailTemplatePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}
	renderTemplate(w, r, "admin_email_template.html", nil)
}

// handleEmailTemplateGet handles GET /api/emails/template
func handleEmailTemplateGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	t, err := stores.EmailStore.GetActiveTemplate(r.Context())
	if err != nil {
		// No template yet ΓÇö return empty defaults
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ID": "", "Header": "", "Footer": ""})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// handleEmailTemplateSave handles POST /api/emails/template
func handleEmailTemplateSave(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	var input struct {
		Header string `json:"Header"`
		Footer string `json:"Footer"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	t := emailDomain.EmailTemplate{
		ID:        generateID(),
		Header:    input.Header,
		Footer:    input.Footer,
		CreatedAt: timeNow(),
		Active:    true,
	}

	if err := stores.EmailStore.SaveTemplate(r.Context(), t); err != nil {
		internalError(w, err)
		return
	}

	slog.Info("email_event", "event", "template_saved", "template_id", t.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(t)
}

// handleEmailPreview handles POST /api/emails/preview ΓÇö wraps body with active template
func handleEmailPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	var input struct {
		Body string `json:"Body"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	t, err := stores.EmailStore.GetActiveTemplate(r.Context())
	if err != nil {
		// No template ΓÇö return body as-is
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"HTML": input.Body})
		return
	}

	wrapped := t.WrapBody(input.Body)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"HTML": wrapped})
}

// handleMemberFilterForEmail handles GET /api/emails/recipients/filter?program=...
func handleMemberFilterForEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	program := r.URL.Query().Get("program")

	filter := memberStore.ListFilter{
		Status: "active",
	}
	if program != "" {
		filter.Program = program
	}

	members, err := stores.MemberStore.List(r.Context(), filter)
	if err != nil {
		internalError(w, err)
		return
	}

	type memberResult struct {
		ID    string `json:"ID"`
		Name  string `json:"Name"`
		Email string `json:"Email"`
	}
	var results []memberResult
	for _, m := range members {
		results = append(results, memberResult{ID: m.ID, Name: m.Name, Email: m.Email})
	}

	w.Header().Set("Content-Type", "application/json")
	if results == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(results)
}

// handleMemberSearchForEmail handles GET /api/emails/recipients/search?q=...
func handleMemberSearchForEmail(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	members, err := stores.MemberStore.SearchByName(r.Context(), query, 20)
	if err != nil {
		internalError(w, err)
		return
	}

	type memberResult struct {
		ID    string `json:"ID"`
		Name  string `json:"Name"`
		Email string `json:"Email"`
	}
	var results []memberResult
	for _, m := range members {
		results = append(results, memberResult{ID: m.ID, Name: m.Name, Email: m.Email})
	}

	w.Header().Set("Content-Type", "application/json")
	if results == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(results)
}

// handleRecipientsFilterBySession handles GET /api/emails/recipients/by-session?scheduleID=...&date=...
func handleRecipientsFilterBySession(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	scheduleID := r.URL.Query().Get("scheduleID")
	classDate := r.URL.Query().Get("date")
	if scheduleID == "" || classDate == "" {
		http.Error(w, "scheduleID and date are required", http.StatusBadRequest)
		return
	}

	memberIDs, err := stores.AttendanceStore.ListDistinctMemberIDsByScheduleAndDate(r.Context(), scheduleID, classDate)
	if err != nil {
		internalError(w, err)
		return
	}

	type memberResult struct {
		ID    string `json:"ID"`
		Name  string `json:"Name"`
		Email string `json:"Email"`
	}
	var results []memberResult
	for _, id := range memberIDs {
		m, err := stores.MemberStore.GetByID(r.Context(), id)
		if err != nil {
			continue
		}
		results = append(results, memberResult{ID: m.ID, Name: m.Name, Email: m.Email})
	}

	w.Header().Set("Content-Type", "application/json")
	if results == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(results)
}

// handleRecipientsFilterByClassType handles GET /api/emails/recipients/by-class-type?classTypeID=...&days=30
func handleRecipientsFilterByClassType(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	classTypeID := r.URL.Query().Get("classTypeID")
	daysStr := r.URL.Query().Get("days")
	if classTypeID == "" {
		http.Error(w, "classTypeID is required", http.StatusBadRequest)
		return
	}
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	schedules, err := stores.ScheduleStore.ListByClassTypeID(r.Context(), classTypeID)
	if err != nil {
		internalError(w, err)
		return
	}
	var scheduleIDs []string
	for _, s := range schedules {
		scheduleIDs = append(scheduleIDs, s.ID)
	}
	if len(scheduleIDs) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("[]"))
		return
	}

	since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	memberIDs, err := stores.AttendanceStore.ListDistinctMemberIDsByScheduleIDsSince(r.Context(), scheduleIDs, since)
	if err != nil {
		internalError(w, err)
		return
	}

	type memberResult struct {
		ID    string `json:"ID"`
		Name  string `json:"Name"`
		Email string `json:"Email"`
	}
	var results []memberResult
	for _, id := range memberIDs {
		m, err := stores.MemberStore.GetByID(r.Context(), id)
		if err != nil {
			continue
		}
		results = append(results, memberResult{ID: m.ID, Name: m.Name, Email: m.Email})
	}

	w.Header().Set("Content-Type", "application/json")
	if results == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(results)
}

// handleRecentSessions handles GET /api/schedules/recent-sessions ΓÇö lists recent class sessions for the filter dropdown.
func handleRecentSessions(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	schedules, err := stores.ScheduleStore.List(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}

	classTypes, err := stores.ClassTypeStore.List(r.Context())
	if err != nil {
		internalError(w, err)
		return
	}
	ctMap := map[string]string{}
	for _, ct := range classTypes {
		ctMap[ct.ID] = ct.Name
	}

	type sessionInfo struct {
		ScheduleID string `json:"ScheduleID"`
		ClassDate  string `json:"ClassDate"`
		Label      string `json:"Label"`
	}

	// Generate sessions for the last 14 days
	var sessions []sessionInfo
	now := time.Now()
	for daysAgo := 0; daysAgo < 14; daysAgo++ {
		date := now.AddDate(0, 0, -daysAgo)
		dayName := strings.ToLower(date.Weekday().String())
		for _, s := range schedules {
			if s.Day == dayName {
				ctName := ctMap[s.ClassTypeID]
				if ctName == "" {
					ctName = "Unknown"
				}
				label := date.Format("Mon 2 Jan") + " " + s.StartTime + " " + ctName
				sessions = append(sessions, sessionInfo{
					ScheduleID: s.ID,
					ClassDate:  date.Format("2006-01-02"),
					Label:      label,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	if sessions == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(sessions)
}

// handleActivatePage handles GET /activate?token=... ΓÇö shows the password-setting form.
func handleActivatePage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "Missing activation token", http.StatusBadRequest)
		return
	}

	tok, err := stores.AccountStore.GetActivationTokenByToken(r.Context(), token)
	if err != nil {
		renderTemplate(w, r, "activate.html", map[string]any{"Error": "Invalid activation link."})
		return
	}
	if tok.Used {
		renderTemplate(w, r, "activate.html", map[string]any{"Error": "This activation link has already been used."})
		return
	}
	if tok.IsExpired(timeNow()) {
		renderTemplate(w, r, "activate.html", map[string]any{"Error": "This activation link has expired. Please contact the gym to resend."})
		return
	}

	renderTemplate(w, r, "activate.html", map[string]any{"Token": token})
}

// handleActivateAccount handles POST /api/activate ΓÇö sets password and activates account.
func handleActivateAccount(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var input struct {
		Token    string `json:"Token"`
		Password string `json:"Password"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.Token == "" || input.Password == "" {
		http.Error(w, "Token and Password are required", http.StatusBadRequest)
		return
	}

	tok, err := stores.AccountStore.GetActivationTokenByToken(r.Context(), input.Token)
	if err != nil {
		http.Error(w, "Invalid activation token", http.StatusBadRequest)
		return
	}
	if tok.Used {
		http.Error(w, "This activation link has already been used", http.StatusBadRequest)
		return
	}
	if tok.IsExpired(timeNow()) {
		http.Error(w, "Link expired - contact your gym to resend", http.StatusBadRequest)
		return
	}

	acct, err := stores.AccountStore.GetByID(r.Context(), tok.AccountID)
	if err != nil {
		internalError(w, err)
		return
	}

	if err := acct.Activate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := acct.SetPassword(input.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	acct.PasswordChangeRequired = false

	if err := stores.AccountStore.Save(r.Context(), acct); err != nil {
		internalError(w, err)
		return
	}

	tok.Invalidate()
	stores.AccountStore.SaveActivationToken(r.Context(), tok)
	stores.AccountStore.InvalidateTokensForAccount(r.Context(), tok.AccountID)

	slog.Info("auth_event", "event", "account_activated", "account_id", acct.ID, "email", acct.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "activated"})
}

// handleResendActivation handles POST /api/admin/resend-activation ΓÇö admin resends activation email.
func handleResendActivation(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	if _, ok := requireAdmin(w, r); !ok {
		return
	}

	var input struct {
		AccountID string `json:"AccountID"`
	}
	if err := strictDecode(r, &input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.AccountID == "" {
		http.Error(w, "AccountID is required", http.StatusBadRequest)
		return
	}

	acct, err := stores.AccountStore.GetByID(r.Context(), input.AccountID)
	if err != nil {
		http.Error(w, "Account not found", http.StatusNotFound)
		return
	}
	if acct.Status != accountDomain.StatusPendingActivation {
		http.Error(w, "Account is already activated", http.StatusBadRequest)
		return
	}

	stores.AccountStore.InvalidateTokensForAccount(r.Context(), acct.ID)

	tokenStr := generateID()
	tok := accountDomain.ActivationToken{
		ID:        generateID(),
		AccountID: acct.ID,
		Token:     tokenStr,
		ExpiresAt: timeNow().Add(72 * time.Hour),
		CreatedAt: timeNow(),
	}
	if err := stores.AccountStore.SaveActivationToken(r.Context(), tok); err != nil {
		internalError(w, err)
		return
	}

	slog.Info("auth_event", "event", "activation_resent", "account_id", acct.ID, "email", acct.Email)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "sent", "token": tokenStr})
}

// --- Phase 6: Curriculum Rotor Handlers ---

// handleCurriculumPage handles GET /curriculum ΓÇö renders the curriculum management page.
func handleCurriculumPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	if !requireFeaturePage(w, r, sess, "curriculum") {
		return
	}
	renderTemplate(w, r, "curriculum.html", map[string]interface{}{
		"Title": "Curriculum",
	})
}

// handleRotors handles GET/POST for /api/rotors
func handleRotors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !requireFeatureAPI(w, r, sess, "curriculum") {
		return
	}

	if r.Method == "GET" {
		if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		classTypeID := r.URL.Query().Get("class_type_id")
		if classTypeID == "" {
			http.Error(w, "class_type_id is required", http.StatusBadRequest)
			return
		}
		rotors, err := stores.RotorStore.ListRotorsByClassType(ctx, classTypeID)
		if err != nil {
			internalError(w, err)
			return
		}
		if rotors == nil {
			rotors = []rotorDomain.Rotor{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rotors)
		return
	}

	if r.Method == "POST" {
		if sess.Role != "admin" && sess.Role != "coach" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		var input struct {
			ClassTypeID string `json:"class_type_id"`
			Name        string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Determine next version number
		existing, _ := stores.RotorStore.ListRotorsByClassType(ctx, input.ClassTypeID)
		nextVersion := 1
		for _, r := range existing {
			if r.Version >= nextVersion {
				nextVersion = r.Version + 1
			}
		}

		rotor := rotorDomain.Rotor{
			ID:          generateID(),
			ClassTypeID: input.ClassTypeID,
			Name:        input.Name,
			Version:     nextVersion,
			Status:      rotorDomain.StatusDraft,
			CreatedBy:   sess.AccountID,
			CreatedAt:   timeNow(),
		}
		if err := rotor.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.RotorStore.SaveRotor(ctx, rotor); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(rotor)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleRotorByID handles GET/DELETE for /api/rotors/by-id?id=<id>
func handleRotorByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !requireFeatureAPI(w, r, sess, "curriculum") {
		return
	}
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		rotor, err := stores.RotorStore.GetRotor(ctx, id)
		if err != nil {
			http.Error(w, "Rotor not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rotor)
		return
	}

	if r.Method == "PUT" {
		if sess.Role != "admin" && sess.Role != "coach" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		var input struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		rotor, err := stores.RotorStore.GetRotor(ctx, id)
		if err != nil {
			http.Error(w, "Rotor not found", http.StatusNotFound)
			return
		}
		if err := rotor.Rename(input.Name); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.RotorStore.SaveRotor(ctx, rotor); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rotor)
		return
	}

	if r.Method == "DELETE" {
		if sess.Role != "admin" && sess.Role != "coach" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		rotor, err := stores.RotorStore.GetRotor(ctx, id)
		if err != nil {
			http.Error(w, "Rotor not found", http.StatusNotFound)
			return
		}
		if rotor.IsActive() {
			http.Error(w, "cannot delete an active rotor", http.StatusBadRequest)
			return
		}
		if err := stores.RotorStore.DeleteRotor(ctx, id); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleRotorActivate handles POST /api/rotors/activate
func handleRotorActivate(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !requireFeatureAPI(w, r, sess, "curriculum") {
		return
	}

	var input struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	rotor, err := stores.RotorStore.GetRotor(ctx, input.ID)
	if err != nil {
		http.Error(w, "Rotor not found", http.StatusNotFound)
		return
	}

	// Archive any currently active rotor for this class
	activeRotor, err := stores.RotorStore.GetActiveRotor(ctx, rotor.ClassTypeID)
	if err == nil && activeRotor.ID != rotor.ID {
		activeRotor.Archive()
		stores.RotorStore.SaveRotor(ctx, activeRotor)
	}

	if err := rotor.Activate(timeNow()); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := stores.RotorStore.SaveRotor(ctx, rotor); err != nil {
		internalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rotor)
}

// handleRotorPreview handles POST /api/rotors/preview (toggle preview on/off)
func handleRotorPreview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !requireFeatureAPI(w, r, sess, "curriculum") {
		return
	}
	var input struct {
		ID        string `json:"id"`
		PreviewOn bool   `json:"preview_on"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	rotor, err := stores.RotorStore.GetRotor(ctx, input.ID)
	if err != nil {
		http.Error(w, "Rotor not found", http.StatusNotFound)
		return
	}

	rotor.PreviewOn = input.PreviewOn
	if err := stores.RotorStore.SaveRotor(ctx, rotor); err != nil {
		internalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rotor)
}

// handleRotorThemes handles GET/POST/DELETE for /api/rotors/themes
func handleRotorThemes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !requireFeatureAPI(w, r, sess, "curriculum") {
		return
	}

	if r.Method == "GET" {
		if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		rotorID := r.URL.Query().Get("rotor_id")
		if rotorID == "" {
			http.Error(w, "rotor_id is required", http.StatusBadRequest)
			return
		}
		themes, err := stores.RotorStore.ListThemesByRotor(ctx, rotorID)
		if err != nil {
			internalError(w, err)
			return
		}
		if themes == nil {
			themes = []rotorDomain.RotorTheme{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(themes)
		return
	}

	if r.Method == "POST" {
		var input struct {
			RotorID  string `json:"rotor_id"`
			Name     string `json:"name"`
			Position int    `json:"position"`
			Hidden   bool   `json:"hidden"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		// Verify rotor exists and is in draft
		rotor, err := stores.RotorStore.GetRotor(ctx, input.RotorID)
		if err != nil {
			http.Error(w, "Rotor not found", http.StatusNotFound)
			return
		}
		if !rotor.IsDraft() {
			http.Error(w, "can only add themes to draft rotors", http.StatusBadRequest)
			return
		}

		theme := rotorDomain.RotorTheme{
			ID:       generateID(),
			RotorID:  input.RotorID,
			Name:     input.Name,
			Position: input.Position,
			Hidden:   input.Hidden,
		}
		if err := theme.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.RotorStore.SaveRotorTheme(ctx, theme); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(theme)
		return
	}

	if r.Method == "DELETE" {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := stores.RotorStore.DeleteRotorTheme(ctx, id); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleTopics handles GET/POST/DELETE for /api/rotors/topics
func handleTopics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !requireFeatureAPI(w, r, sess, "curriculum") {
		return
	}

	if r.Method == "GET" {
		if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		themeID := r.URL.Query().Get("theme_id")
		if themeID == "" {
			http.Error(w, "theme_id is required", http.StatusBadRequest)
			return
		}
		topics, err := stores.RotorStore.ListTopicsByTheme(ctx, themeID)
		if err != nil {
			internalError(w, err)
			return
		}
		if topics == nil {
			topics = []rotorDomain.Topic{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(topics)
		return
	}

	if r.Method == "POST" {
		var input struct {
			RotorThemeID  string `json:"rotor_theme_id"`
			Name          string `json:"name"`
			Description   string `json:"description"`
			DurationWeeks int    `json:"duration_weeks"`
			Position      int    `json:"position"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if input.DurationWeeks == 0 {
			input.DurationWeeks = 1
		}

		topic := rotorDomain.Topic{
			ID:            generateID(),
			RotorThemeID:  input.RotorThemeID,
			Name:          input.Name,
			Description:   input.Description,
			DurationWeeks: input.DurationWeeks,
			Position:      input.Position,
		}
		if err := topic.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.RotorStore.SaveTopic(ctx, topic); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(topic)
		return
	}

	if r.Method == "PUT" {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		var input struct {
			Name          *string `json:"name"`
			Description   *string `json:"description"`
			DurationWeeks *int    `json:"duration_weeks"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		topic, err := stores.RotorStore.GetTopic(ctx, id)
		if err != nil {
			http.Error(w, "topic not found", http.StatusNotFound)
			return
		}
		if input.Name != nil {
			topic.Name = *input.Name
		}
		if input.Description != nil {
			topic.Description = *input.Description
		}
		if input.DurationWeeks != nil {
			topic.DurationWeeks = *input.DurationWeeks
		}
		if err := topic.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.RotorStore.SaveTopic(ctx, topic); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(topic)
		return
	}

	if r.Method == "DELETE" {
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := stores.RotorStore.DeleteTopic(ctx, id); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleTopicReorder handles POST /api/rotors/topics/reorder
func handleTopicReorder(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !requireFeatureAPI(w, r, sess, "curriculum") {
		return
	}

	var input struct {
		RotorThemeID string   `json:"rotor_theme_id"`
		TopicIDs     []string `json:"topic_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if input.RotorThemeID == "" || len(input.TopicIDs) == 0 {
		http.Error(w, "rotor_theme_id and topic_ids are required", http.StatusBadRequest)
		return
	}

	if err := stores.RotorStore.ReorderTopics(r.Context(), input.RotorThemeID, input.TopicIDs); err != nil {
		internalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleTopicScheduleAction handles POST /api/rotors/schedule/action
// Actions: "activate" (start a topic), "complete", "skip", "extend"
func handleTopicScheduleAction(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !requireFeatureAPI(w, r, sess, "curriculum") {
		return
	}
	var input struct {
		Action       string `json:"action"` // activate, complete, skip, extend
		TopicID      string `json:"topic_id"`
		RotorThemeID string `json:"rotor_theme_id"`
		ExtendWeeks  int    `json:"extend_weeks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	now := timeNow()

	switch input.Action {
	case "activate":
		// Complete any currently active schedule for this theme
		activeSched, err := stores.RotorStore.GetActiveScheduleForTheme(ctx, input.RotorThemeID)
		if err == nil {
			activeSched.Status = rotorDomain.ScheduleStatusCompleted
			activeSched.EndDate = now
			stores.RotorStore.SaveTopicSchedule(ctx, activeSched)

			// Update last_covered on the completed topic
			completedTopic, topicErr := stores.RotorStore.GetTopic(ctx, activeSched.TopicID)
			if topicErr == nil {
				completedTopic.LastCovered = now
				stores.RotorStore.SaveTopic(ctx, completedTopic)
			}
		}

		topic, err := stores.RotorStore.GetTopic(ctx, input.TopicID)
		if err != nil {
			http.Error(w, "Topic not found", http.StatusNotFound)
			return
		}

		sched := rotorDomain.TopicSchedule{
			ID:           generateID(),
			TopicID:      topic.ID,
			RotorThemeID: topic.RotorThemeID,
			StartDate:    now,
			EndDate:      now.AddDate(0, 0, topic.DurationWeeks*7),
			Status:       rotorDomain.ScheduleStatusActive,
		}
		if err := stores.RotorStore.SaveTopicSchedule(ctx, sched); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sched)

	case "complete":
		sched, err := stores.RotorStore.GetActiveScheduleForTheme(ctx, input.RotorThemeID)
		if err != nil {
			http.Error(w, "No active schedule for theme", http.StatusNotFound)
			return
		}
		sched.Status = rotorDomain.ScheduleStatusCompleted
		sched.EndDate = now
		if err := stores.RotorStore.SaveTopicSchedule(ctx, sched); err != nil {
			internalError(w, err)
			return
		}
		// Update last_covered
		topic, topicErr := stores.RotorStore.GetTopic(ctx, sched.TopicID)
		if topicErr == nil {
			topic.LastCovered = now
			stores.RotorStore.SaveTopic(ctx, topic)
		}
		// Clear votes for the completed topic
		stores.RotorStore.DeleteVotesForTopic(ctx, sched.TopicID)

		// Auto-advance: activate the next topic in queue order (wraps around)
		var nextSched *rotorDomain.TopicSchedule
		allTopics, _ := stores.RotorStore.ListTopicsByTheme(ctx, input.RotorThemeID)
		nextTopic := rotorDomain.NextTopicInQueue(allTopics, sched.TopicID)
		if nextTopic != nil {
			ns := rotorDomain.TopicSchedule{
				ID:           generateID(),
				TopicID:      nextTopic.ID,
				RotorThemeID: input.RotorThemeID,
				StartDate:    now,
				EndDate:      now.AddDate(0, 0, nextTopic.DurationWeeks*7),
				Status:       rotorDomain.ScheduleStatusActive,
			}
			if err := stores.RotorStore.SaveTopicSchedule(ctx, ns); err == nil {
				nextSched = &ns
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"completed":    sched,
			"next_started": nextSched,
		})

	case "skip":
		sched, err := stores.RotorStore.GetActiveScheduleForTheme(ctx, input.RotorThemeID)
		if err != nil {
			http.Error(w, "No active schedule for theme", http.StatusNotFound)
			return
		}
		sched.Status = rotorDomain.ScheduleStatusSkipped
		sched.EndDate = now
		if err := stores.RotorStore.SaveTopicSchedule(ctx, sched); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sched)

	case "extend":
		sched, err := stores.RotorStore.GetActiveScheduleForTheme(ctx, input.RotorThemeID)
		if err != nil {
			http.Error(w, "No active schedule for theme", http.StatusNotFound)
			return
		}
		weeks := input.ExtendWeeks
		if weeks < 1 {
			weeks = 1
		}
		sched.EndDate = sched.EndDate.AddDate(0, 0, weeks*7)
		if err := stores.RotorStore.SaveTopicSchedule(ctx, sched); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sched)

	default:
		http.Error(w, "invalid action: must be activate, complete, skip, or extend", http.StatusBadRequest)
	}
}

// handleVotes handles POST /api/votes (cast a vote) and GET /api/votes?topic_id=<id> (get vote count)
func handleVotes(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !requireFeatureAPI(w, r, sess, "curriculum") {
		return
	}

	if r.Method == "GET" {
		topicID := r.URL.Query().Get("topic_id")
		if topicID == "" {
			http.Error(w, "topic_id is required", http.StatusBadRequest)
			return
		}
		count, err := stores.RotorStore.CountVotesForTopic(ctx, topicID)
		if err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"votes": count})
		return
	}

	if r.Method == "POST" {
		session, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var input struct {
			TopicID string `json:"topic_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if input.TopicID == "" {
			http.Error(w, "topic_id is required", http.StatusBadRequest)
			return
		}

		vote := rotorDomain.Vote{
			ID:        generateID(),
			TopicID:   input.TopicID,
			AccountID: session.AccountID,
			CreatedAt: timeNow(),
		}
		if err := stores.RotorStore.SaveVote(ctx, vote); err != nil {
			if err == rotorDomain.ErrAlreadyVoted {
				http.Error(w, err.Error(), http.StatusConflict)
				return
			}
			internalError(w, err)
			return
		}

		count, _ := stores.RotorStore.CountVotesForTopic(ctx, input.TopicID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "voted", "votes": count})
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}

// handleTopicBump handles POST /api/rotors/topics/bump ΓÇö bumps a voted topic to current position
func handleTopicBump(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()

	var input struct {
		TopicID      string `json:"topic_id"`
		RotorThemeID string `json:"rotor_theme_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	topic, err := stores.RotorStore.GetTopic(ctx, input.TopicID)
	if err != nil {
		http.Error(w, "Topic not found", http.StatusNotFound)
		return
	}

	// Complete current active schedule if any
	activeSched, err := stores.RotorStore.GetActiveScheduleForTheme(ctx, input.RotorThemeID)
	if err == nil {
		activeSched.Status = rotorDomain.ScheduleStatusCompleted
		activeSched.EndDate = timeNow()
		stores.RotorStore.SaveTopicSchedule(ctx, activeSched)
	}

	// Activate the bumped topic
	now := timeNow()
	sched := rotorDomain.TopicSchedule{
		ID:           generateID(),
		TopicID:      topic.ID,
		RotorThemeID: input.RotorThemeID,
		StartDate:    now,
		EndDate:      now.AddDate(0, 0, topic.DurationWeeks*7),
		Status:       rotorDomain.ScheduleStatusActive,
	}
	if err := stores.RotorStore.SaveTopicSchedule(ctx, sched); err != nil {
		internalError(w, err)
		return
	}

	// Clear votes for the bumped topic
	stores.RotorStore.DeleteVotesForTopic(ctx, input.TopicID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sched)
}

// handleCurriculumOverview handles GET /api/curriculum/overview
// Returns the active curriculum across all class types for the Theme Carousel page.
func handleCurriculumOverview(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !requireFeatureAPI(w, r, sess, "curriculum") {
		return
	}
	query := projections.GetCurriculumOverviewQuery{
		Role: sess.Role,
	}
	deps := projections.GetCurriculumOverviewDeps{
		ClassTypeStore: stores.ClassTypeStore,
		RotorStore:     stores.RotorStore,
	}
	result, err := projections.QueryGetCurriculumOverview(r.Context(), query, deps)
	if err != nil {
		internalError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleCurriculumView handles GET /api/curriculum/view?class_type_id=<id>
// Returns the full curriculum state for a class: active rotor, themes, topics, schedules, votes.
func handleCurriculumView(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	if !requireFeatureAPI(w, r, sess, "curriculum") {
		return
	}
	classTypeID := r.URL.Query().Get("class_type_id")
	if classTypeID == "" {
		http.Error(w, "class_type_id is required", http.StatusBadRequest)
		return
	}

	rotor, err := stores.RotorStore.GetActiveRotor(ctx, classTypeID)
	if err != nil {
		// No active rotor ΓÇö return empty state
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"rotor":  nil,
			"themes": []interface{}{},
		})
		return
	}

	themes, _ := stores.RotorStore.ListThemesByRotor(ctx, rotor.ID)

	type topicWithVotes struct {
		rotorDomain.Topic
		Votes    int  `json:"votes"`
		IsActive bool `json:"is_active"`
	}
	type themeView struct {
		rotorDomain.RotorTheme
		Topics         []topicWithVotes           `json:"topics"`
		ActiveSchedule *rotorDomain.TopicSchedule `json:"active_schedule"`
	}

	var themeViews []themeView
	for _, th := range themes {
		tv := themeView{RotorTheme: th}
		topics, _ := stores.RotorStore.ListTopicsByTheme(ctx, th.ID)
		activeSched, schedErr := stores.RotorStore.GetActiveScheduleForTheme(ctx, th.ID)
		if schedErr == nil {
			tv.ActiveSchedule = &activeSched
		}
		for _, tp := range topics {
			votes, _ := stores.RotorStore.CountVotesForTopic(ctx, tp.ID)
			isActive := tv.ActiveSchedule != nil && tv.ActiveSchedule.TopicID == tp.ID
			tv.Topics = append(tv.Topics, topicWithVotes{Topic: tp, Votes: votes, IsActive: isActive})
		}
		if tv.Topics == nil {
			tv.Topics = []topicWithVotes{}
		}
		themeViews = append(themeViews, tv)
	}
	if themeViews == nil {
		themeViews = []themeView{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"rotor":  rotor,
		"themes": themeViews,
	})
}

// --- Phase 9: Calendar Handlers ---

// handleCalendarPage handles GET /calendar ΓÇö renders the club calendar page.
func handleCalendarPage(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	sess, ok := middleware.GetSessionFromContext(r.Context())
	if !ok {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}
	if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" && sess.Role != "trial" {
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
		return
	}
	if !requireFeaturePage(w, r, sess, "calendar") {
		return
	}
	renderTemplate(w, r, "calendar.html", map[string]interface{}{
		"Title": "Club Calendar",
	})
}

// handleCalendarEvents handles GET/POST/PUT/DELETE for /api/calendar/events
func handleCalendarEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sess, ok := middleware.GetSessionFromContext(ctx)
	if !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	if !requireFeatureAPI(w, r, sess, "calendar") {
		return
	}

	if r.Method == "GET" {
		if sess.Role != "admin" && sess.Role != "coach" && sess.Role != "member" && sess.Role != "trial" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		from := r.URL.Query().Get("from")
		to := r.URL.Query().Get("to")
		if from == "" || to == "" {
			http.Error(w, "from and to date params required (YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		events, err := stores.CalendarEventStore.ListByDateRange(ctx, from, to)
		if err != nil {
			internalError(w, err)
			return
		}
		if events == nil {
			events = []calendarDomain.Event{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(events)
		return
	}

	if r.Method == "POST" {
		if !middleware.IsRole(ctx, "admin") {
			http.Error(w, "admin only", http.StatusForbidden)
			return
		}
		var input struct {
			Title           string `json:"title"`
			Type            string `json:"type"`
			Description     string `json:"description"`
			Location        string `json:"location"`
			StartDate       string `json:"start_date"`
			EndDate         string `json:"end_date"`
			RegistrationURL string `json:"registration_url"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		startDate, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			http.Error(w, "invalid start_date format (use YYYY-MM-DD)", http.StatusBadRequest)
			return
		}
		var endDate time.Time
		if input.EndDate != "" {
			endDate, err = time.Parse("2006-01-02", input.EndDate)
			if err != nil {
				http.Error(w, "invalid end_date format (use YYYY-MM-DD)", http.StatusBadRequest)
				return
			}
		}
		if input.Type == "" {
			input.Type = "event"
		}
		sess, ok := middleware.GetSessionFromContext(ctx)
		if !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
			return
		}
		event := calendarDomain.Event{
			ID:              generateID(),
			Title:           input.Title,
			Type:            input.Type,
			Description:     input.Description,
			Location:        input.Location,
			StartDate:       startDate,
			EndDate:         endDate,
			RegistrationURL: input.RegistrationURL,
			CreatedBy:       sess.AccountID,
			CreatedAt:       time.Now(),
		}
		if err := event.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.CalendarEventStore.Save(ctx, event); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(event)
		return
	}

	if r.Method == "DELETE" {
		if !middleware.IsRole(ctx, "admin") {
			http.Error(w, "admin only", http.StatusForbidden)
			return
		}
		id := r.URL.Query().Get("id")
		if id == "" {
			http.Error(w, "id is required", http.StatusBadRequest)
			return
		}
		if err := stores.CalendarEventStore.Delete(ctx, id); err != nil {
			internalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.WriteHeader(http.StatusMethodNotAllowed)
}
