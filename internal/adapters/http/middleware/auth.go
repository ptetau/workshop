package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	domainAccount "workshop/internal/domain/account"
)

// contextKey is an unexported type for context keys in this package.
type contextKey string

const accountContextKey contextKey = "account"

// Session represents an authenticated session.
type Session struct {
	AccountID string
	Email     string
	Role      string
	CreatedAt time.Time

	// DevMode impersonation fields — populated only when an admin is impersonating another role.
	RealAccountID string
	RealEmail     string
	RealRole      string
}

// IsImpersonating returns true if this session is currently impersonating another role.
// INVARIANT: Session fields are not mutated
func (s Session) IsImpersonating() bool {
	return s.RealRole != ""
}

// SessionStore is an in-memory session store.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

// NewSessionStore creates a new in-memory session store.
func NewSessionStore() *SessionStore {
	return &SessionStore{
		sessions: make(map[string]Session),
	}
}

// Create stores a new session and returns the token.
// PRE: accountID, email, role are non-empty
// POST: Session is stored, token is returned
func (ss *SessionStore) Create(accountID, email, role string) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", err
	}
	ss.mu.Lock()
	defer ss.mu.Unlock()
	ss.sessions[token] = Session{
		AccountID: accountID,
		Email:     email,
		Role:      role,
		CreatedAt: time.Now(),
	}
	return token, nil
}

// Get retrieves a session by token.
// PRE: token is non-empty
// POST: Returns session if valid and not expired
func (ss *SessionStore) Get(token string) (Session, bool) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	session, ok := ss.sessions[token]
	if !ok {
		return Session{}, false
	}
	// Sessions expire after 24 hours
	if time.Since(session.CreatedAt) > 24*time.Hour {
		delete(ss.sessions, token)
		return Session{}, false
	}
	return session, true
}

// Delete removes a session by token.
// PRE: token is non-empty
// POST: Session with given token is removed
func (ss *SessionStore) Delete(token string) {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	delete(ss.sessions, token)
}

// Update replaces the session for a given token in-place.
// PRE: token exists in the store
// POST: Session is replaced with the new value
func (ss *SessionStore) Update(token string, session Session) bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	if _, ok := ss.sessions[token]; !ok {
		return false
	}
	ss.sessions[token] = session
	return true
}

const sessionCookieName = "workshop_session"

// Auth returns middleware that extracts the session from the cookie and sets the account in context.
// It does NOT block unauthenticated requests — use RequireAuth or RequireRole for that.
func Auth(sessions *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(sessionCookieName)
			if err == nil && cookie.Value != "" {
				if session, ok := sessions.Get(cookie.Value); ok {
					ctx := context.WithValue(r.Context(), accountContextKey, session)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

// RequireAuth returns middleware that blocks unauthenticated requests.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := GetSessionFromContext(r.Context()); !ok {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireRole returns middleware that blocks requests from users without one of the specified roles.
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r] = true
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			session, ok := GetSessionFromContext(r.Context())
			if !ok {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if !roleSet[session.Role] {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetSessionFromContext extracts the session from the request context.
func GetSessionFromContext(ctx context.Context) (Session, bool) {
	session, ok := ctx.Value(accountContextKey).(Session)
	return session, ok
}

// SetSessionCookie sets the session cookie on the response.
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		HttpOnly: true,
		Secure:   false, // Allow HTTP for local development
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   86400, // 24 hours
	})
}

// ClearSessionCookie removes the session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})
}

// IsRole checks if the current session has one of the given roles.
func IsRole(ctx context.Context, roles ...string) bool {
	session, ok := GetSessionFromContext(ctx)
	if !ok {
		return false
	}
	for _, r := range roles {
		if session.Role == r {
			return true
		}
	}
	return false
}

// IsAdmin checks if the current session is an admin.
func IsAdmin(ctx context.Context) bool {
	return IsRole(ctx, domainAccount.RoleAdmin)
}

// IsRealAdmin checks if the underlying (non-impersonated) identity is an admin.
// Returns true if the session is admin (not impersonating) or if RealRole is admin (impersonating).
func IsRealAdmin(ctx context.Context) bool {
	session, ok := GetSessionFromContext(ctx)
	if !ok {
		return false
	}
	if session.IsImpersonating() {
		return session.RealRole == domainAccount.RoleAdmin
	}
	return session.Role == domainAccount.RoleAdmin
}

// IsCoachOrAdmin checks if the current session is a coach or admin.
func IsCoachOrAdmin(ctx context.Context) bool {
	return IsRole(ctx, domainAccount.RoleAdmin, domainAccount.RoleCoach)
}

// ContextWithSession returns a context with the given session set.
// Intended for use in tests.
func ContextWithSession(ctx context.Context, sess Session) context.Context {
	return context.WithValue(ctx, accountContextKey, sess)
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
