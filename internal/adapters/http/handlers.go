package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/yuin/goldmark"
	goldmarkHTML "github.com/yuin/goldmark/renderer/html"

	"workshop/internal/adapters/http/middleware"
	accountStore "workshop/internal/adapters/storage/account"
	memberStore "workshop/internal/adapters/storage/member"
	noticeStore "workshop/internal/adapters/storage/notice"
	"workshop/internal/application/orchestrators"
	"workshop/internal/application/projections"
	accountDomain "workshop/internal/domain/account"
	clipDomain "workshop/internal/domain/clip"
	gradingDomain "workshop/internal/domain/grading"
	holidayDomain "workshop/internal/domain/holiday"
	messageDomain "workshop/internal/domain/message"
	milestoneDomain "workshop/internal/domain/milestone"
	noticeDomain "workshop/internal/domain/notice"
	observationDomain "workshop/internal/domain/observation"
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

	funcMap := template.FuncMap{
		"currentRole":     func() string { return role },
		"currentEmail":    func() string { return email },
		"isLoggedIn":      func() bool { return role != "" },
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

// handleMembers handles both GET (list) and POST (register) for /members
func handleMembers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	isHTML := isHTMLRequest(r)

	if r.Method == "GET" {
		// GET: List members
		query := projections.GetMemberListQuery{Program: ""}
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
				"Members": result.Members,
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(result.Members)
		return
	}

	if r.Method == "POST" {
		// POST: Register member
		input := orchestrators.RegisterMemberInput{}

		if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Form error: "+err.Error(), http.StatusBadRequest)
				return
			}
			input.Email = r.FormValue("Email")
			input.Name = r.FormValue("Name")
			input.Program = r.FormValue("Program")
		} else {
			if err := strictDecode(r, &input); err != nil {
				http.Error(w, "JSON error: "+err.Error(), http.StatusBadRequest)
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

	input := orchestrators.CheckInMemberInput{}

	if strings.HasPrefix(r.Header.Get("Content-Type"), "application/x-www-form-urlencoded") {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Form error: "+err.Error(), http.StatusBadRequest)
			return
		}
		input.MemberID = r.FormValue("MemberID")
		input.ScheduleID = r.FormValue("ScheduleID")
		input.ClassDate = r.FormValue("ClassDate")
	} else {
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "JSON error: "+err.Error(), http.StatusBadRequest)
			return
		}
	}

	deps := orchestrators.CheckInMemberDeps{
		MemberStore:     stores.MemberStore,
		AttendanceStore: stores.AttendanceStore,
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

	query := projections.GetAttendanceTodayQuery{Date: ""}
	deps := projections.GetAttendanceTodayDeps{
		AttendanceStore: stores.AttendanceStore,
		MemberStore:     stores.MemberStore,
		InjuryStore:     stores.InjuryStore,
	}

	result, err := projections.QueryGetAttendanceToday(ctx, query, deps)
	if err != nil {
		internalError(w, err)
		return
	}

	if isHTML {
		renderTemplate(w, r, "get_attendance_today.html", map[string]any{
			"Attendees": result.Attendees,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result.Attendees)
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
			http.Error(w, "Form error: "+err.Error(), http.StatusBadRequest)
			return
		}
		input.MemberID = r.FormValue("MemberID")
		input.BodyPart = strings.ToLower(r.FormValue("BodyPart"))
		input.Description = r.FormValue("Description")
	} else {
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "JSON error: "+err.Error(), http.StatusBadRequest)
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
			http.Error(w, "Form error: "+err.Error(), http.StatusBadRequest)
			return
		}
		input.MemberName = r.FormValue("MemberName")
		input.Email = r.FormValue("Email")
		input.AcceptedTerms = r.FormValue("AcceptedTerms") == "true"
	} else {
		if err := strictDecode(r, &input); err != nil {
			http.Error(w, "JSON error: "+err.Error(), http.StatusBadRequest)
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
			http.Error(w, "Form error: "+err.Error(), http.StatusBadRequest)
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
		token, err := sessions.Create(result.AccountID, result.Email, result.Role, result.PasswordChangeRequired)
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
	if _, ok := middleware.GetSessionFromContext(r.Context()); !ok {
		http.Error(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	memberID := r.URL.Query().Get("member_id")
	if memberID == "" {
		http.Error(w, "member_id is required", http.StatusBadRequest)
		return
	}

	query := projections.GetTrainingLogQuery{MemberID: memberID}
	deps := projections.GetTrainingLogDeps{
		AttendanceStore: stores.AttendanceStore,
		MemberStore:     stores.MemberStore,
	}
	result, err := projections.QueryGetTrainingLog(r.Context(), query, deps)
	if err != nil {
		internalError(w, err)
		return
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
		// No type filter — return all notices
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

// handleMessages handles GET/POST for /api/messages
func handleMessages(w http.ResponseWriter, r *http.Request) {
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
		obs := observationDomain.Observation{
			ID:        generateID(),
			MemberID:  input.MemberID,
			AuthorID:  sess.AccountID,
			Content:   input.Content,
			CreatedAt: timeNow(),
		}
		if err := obs.Validate(); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.ObservationStore.Save(ctx, obs); err != nil {
			internalError(w, err)
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
			ID    string `json:"ID"`
			Email string `json:"Email"`
			Role  string `json:"Role"`
		}
		var safe []safeAccount
		for _, a := range accounts {
			safe = append(safe, safeAccount{ID: a.ID, Email: a.Email, Role: a.Role})
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
		if err := acct.SetPassword(input.Password); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err := stores.AccountStore.Save(ctx, acct); err != nil {
			internalError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"ID":    acct.ID,
			"Email": acct.Email,
			"Role":  acct.Role,
		})
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
			Program:         input.Program,
			Belt:            input.Belt,
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

	type readinessEntry struct {
		MemberID     string  `json:"MemberID"`
		MemberName   string  `json:"MemberName"`
		Program      string  `json:"Program"`
		CurrentBelt  string  `json:"CurrentBelt"`
		TargetBelt   string  `json:"TargetBelt"`
		MatHours     float64 `json:"MatHours"`
		RequiredHrs  float64 `json:"RequiredHours"`
		PercentReady float64 `json:"PercentReady"`
	}

	var results []readinessEntry
	for _, m := range members {
		if m.Status != "active" {
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
			results = append(results, readinessEntry{
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

	w.Header().Set("Content-Type", "application/json")
	if results == nil {
		w.Write([]byte("[]"))
		return
	}
	json.NewEncoder(w).Encode(results)
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
	if r.Method != "GET" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}
	types, err := stores.ClassTypeStore.List(r.Context())
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

// --- Phase 3: Dashboard & Kiosk Handlers ---

// handleDashboard handles GET /dashboard — renders role-appropriate dashboard.
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
			AttendanceStore: stores.AttendanceStore,
			MemberStore:     stores.MemberStore,
			InjuryStore:     stores.InjuryStore,
		},
		InactiveDeps: projections.GetInactiveMembersDeps{
			MemberStore:     stores.MemberStore,
			AttendanceStore: stores.AttendanceStore,
		},
		TrainingLogDeps: projections.GetTrainingLogDeps{
			AttendanceStore: stores.AttendanceStore,
			MemberStore:     stores.MemberStore,
		},
		NoticeStore:       stores.NoticeStore,
		ProposalStore:     stores.GradingProposalStore,
		MessageStore:      stores.MessageStore,
		TrainingGoalStore: stores.TrainingGoalStore,
		MemberStore:       stores.MemberStore,
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

// handleKioskPage handles GET /kiosk — renders the standalone kiosk UI.
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
		if _, ok := middleware.GetSessionFromContext(ctx); !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
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
		if _, ok := middleware.GetSessionFromContext(ctx); !ok {
			http.Error(w, "not authenticated", http.StatusUnauthorized)
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

// handleThemesPage handles GET /themes — renders the theme carousel page.
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
	renderTemplate(w, r, "themes.html", map[string]any{
		"Email": sess.Email,
		"Role":  sess.Role,
	})
}

// handleLibraryPage handles GET /library — renders the technical library page.
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
