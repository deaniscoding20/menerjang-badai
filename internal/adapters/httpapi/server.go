package httpapi

import (
	"bytes"
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/menerjangbadai/dashboard/internal/application"
	"github.com/menerjangbadai/dashboard/internal/domain"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

const sessionCookie = "__Host-mb_session"
const loginCookie = "__Host-mb_login"
const scope = "phase3-task-preparation"

type Config struct{ PublicURL, CommitSHA, WebDir string }
type Server struct {
	cfg      Config
	brands   application.Brands
	auth     application.AuthStore
	idp      application.IdentityProvider
	dbHealth func(context.Context) error
	limits   limiter
	taskAPI  http.Handler
}

func New(c Config, b application.Brands, a application.AuthStore, p application.IdentityProvider, dbHealth func(context.Context) error, taskAPI ...http.Handler) http.Handler {
	s := &Server{cfg: c, brands: b, auth: a, idp: p, dbHealth: dbHealth}
	if len(taskAPI) > 0 {
		s.taskAPI = taskAPI[0]
	}
	m := http.NewServeMux()
	m.HandleFunc("/api/tasks", s.tasksEndpoint)
	m.HandleFunc("/api/tasks/{id}", s.tasksEndpoint)
	m.HandleFunc("/api/tasks/{id}/events", s.tasksEndpoint)
	m.HandleFunc("/api/tasks/{id}/comments", s.tasksEndpoint)
	m.HandleFunc("/api/tasks/{id}/revision-requests", s.tasksEndpoint)
	m.HandleFunc("/api/tasks/{id}/dependencies", s.tasksEndpoint)
	m.HandleFunc("/api/tasks/{id}/deliverables", s.tasksEndpoint)
	m.HandleFunc("/api/tasks/{id}/deliverables/{deliverableID}", s.tasksEndpoint)
	m.HandleFunc("/api/tasks/{id}/transition", s.tasksEndpoint)
	m.HandleFunc("/api/healthz", method("GET", s.health))
	m.HandleFunc("/api/readyz", method("GET", s.ready))
	m.HandleFunc("/api/session", method("GET", s.session))
	m.HandleFunc("/api/overview", method("GET", s.overview))
	m.HandleFunc("/api/agents", method("GET", s.agents))
	m.HandleFunc("/api/brands", s.brandsCollection)
	m.HandleFunc("/api/brands/{id}", s.brandItem)
	m.HandleFunc("/auth/login", method("GET", s.login))
	m.HandleFunc("/auth/callback", method("GET", s.callback))
	m.HandleFunc("/auth/logout", method("POST", s.logout))
	m.HandleFunc("/", s.static)
	return s.middleware(http.TimeoutHandler(m, 15*time.Second, `{"error":{"code":"timeout","message":"Request timed out"}}`))
}
func method(want string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != want {
			w.Header().Set("Allow", want)
			problem(w, 405, "method_not_allowed", "Method not allowed")
			return
		}
		next(w, r)
	}
}
func send(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}
func problem(w http.ResponseWriter, status int, code, message string) {
	send(w, status, map[string]any{"error": map[string]string{"code": code, "message": message}})
}
func failure(w http.ResponseWriter, e error) {
	switch {
	case errors.Is(e, domain.ErrUnauthorized):
		problem(w, 401, "unauthorized", "Authentication required")
	case errors.Is(e, domain.ErrForbidden):
		problem(w, 403, "forbidden", "Access denied")
	case errors.Is(e, domain.ErrNotFound):
		problem(w, 404, "not_found", "Resource not found")
	case errors.Is(e, domain.ErrInvalid):
		problem(w, 422, "invalid_input", "Invalid request fields")
	default:
		problem(w, 503, "unavailable", "Service temporarily unavailable")
	}
}
func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	return decodeLimit(w, r, v, 16384)
}
func decodeLimit(w http.ResponseWriter, r *http.Request, v any, limit int64) bool {
	if strings.Split(r.Header.Get("Content-Type"), ";")[0] != "application/json" {
		problem(w, 415, "unsupported_media_type", "Use application/json")
		return false
	}
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	body, err := io.ReadAll(r.Body)
	// encoding/json replaces invalid UTF-8, so validate the bounded bytes first.
	if err != nil || !utf8.Valid(body) {
		problem(w, 400, "invalid_json", "Invalid JSON request")
		return false
	}
	d := json.NewDecoder(bytes.NewReader(body))
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		problem(w, 400, "invalid_json", "Invalid JSON request")
		return false
	}
	if e := d.Decode(&struct{}{}); e != io.EOF {
		problem(w, 400, "invalid_json", "Expected one JSON object")
		return false
	}
	return true
}
func (s *Server) current(w http.ResponseWriter, r *http.Request, write bool) (application.Session, string, bool) {
	c, e := r.Cookie(sessionCookie)
	if e != nil || len(c.Value) < 1 || len(c.Value) > 128 {
		failure(w, domain.ErrUnauthorized)
		return application.Session{}, "", false
	}
	v, e := s.auth.Session(r.Context(), application.Hash(c.Value))
	if e != nil {
		failure(w, e)
		return v, "", false
	}
	if write && (r.Header.Get("Origin") != s.cfg.PublicURL || subtle.ConstantTimeCompare([]byte(r.Header.Get("X-CSRF-Token")), []byte(application.CSRF(c.Value))) != 1) {
		problem(w, 403, "csrf_rejected", "Same-origin CSRF token required")
		return v, "", false
	}
	return v, c.Value, true
}
func (s *Server) session(w http.ResponseWriter, r *http.Request) {
	v, raw, ok := s.current(w, r, false)
	if ok {
		send(w, 200, map[string]any{"user": v.User, "csrf_token": application.CSRF(raw)})
	}
}
func (s *Server) agents(w http.ResponseWriter, r *http.Request) {
	if _, _, ok := s.current(w, r, false); ok {
		send(w, 200, map[string]any{"items": domain.Agents()})
	}
}
func (s *Server) overview(w http.ResponseWriter, r *http.Request) {
	v, _, ok := s.current(w, r, false)
	if !ok {
		return
	}
	brands, e := s.brands.List(r.Context(), v.Actor)
	if e != nil {
		failure(w, e)
		return
	}
	type integration struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
		Reason string `json:"reason"`
	}
	items := []integration{}
	for _, v := range []integration{{"hermes", "Hermes Agents", "unverified", "Execution interface and isolation are unverified; dispatch disabled."}, {"temporal", "Temporal", "unavailable", "Workflow execution is not implemented in this release."}, {"postiz", "Postiz", "unavailable", "Publishing integration is not configured or verified."}, {"minio", "MinIO", "unavailable", "Asset storage is not implemented in this release."}, {"superset", "Superset", "unavailable", "Analytics ingestion and embedding are not implemented in this release."}} {
		items = append(items, v)
	}
	count := 0
	for _, b := range brands {
		if b.Status == "active" {
			count++
		}
	}
	send(w, 200, map[string]any{"brands_count": count, "agents_count": 6, "integrations": items, "release": map[string]string{"sha": s.cfg.CommitSHA, "scope": scope}})
}
func (s *Server) brandsCollection(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "POST" {
		problem(w, 405, "method_not_allowed", "Method not allowed")
		return
	}
	v, _, ok := s.current(w, r, r.Method == "POST")
	if !ok {
		return
	}
	if r.Method == "GET" {
		b, e := s.brands.List(r.Context(), v.Actor)
		if e != nil {
			failure(w, e)
			return
		}
		send(w, 200, map[string]any{"items": b})
		return
	}
	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Website     string `json:"website"`
	}
	if !decode(w, r, &input) {
		return
	}
	b, e := s.brands.Create(r.Context(), v.Actor, domain.Brand{Name: input.Name, Description: input.Description, Website: input.Website})
	if e != nil {
		failure(w, e)
		return
	}
	w.Header().Set("Location", "/api/brands/"+b.ID)
	send(w, 201, b)
}
func (s *Server) brandItem(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "PATCH" && r.Method != "DELETE" {
		problem(w, 405, "method_not_allowed", "Method not allowed")
		return
	}
	v, _, ok := s.current(w, r, r.Method != "GET")
	if !ok {
		return
	}
	b, e := s.brands.Get(r.Context(), v.Actor, r.PathValue("id"))
	if e != nil {
		failure(w, e)
		return
	}
	if r.Method == "GET" {
		send(w, 200, b)
		return
	}
	if r.Method == "DELETE" {
		b.Status = "archived"
	} else {
		var in struct {
			Name        *string `json:"name"`
			Description *string `json:"description"`
			Website     *string `json:"website"`
			Status      *string `json:"status"`
		}
		if !decode(w, r, &in) {
			return
		}
		if in.Name != nil {
			b.Name = *in.Name
		}
		if in.Description != nil {
			b.Description = *in.Description
		}
		if in.Website != nil {
			b.Website = *in.Website
		}
		if in.Status != nil {
			b.Status = *in.Status
		}
	}
	b, e = s.brands.Update(r.Context(), v.Actor, b)
	if e != nil {
		failure(w, e)
		return
	}
	if r.Method == "DELETE" {
		send(w, 204, nil)
	} else {
		send(w, 200, b)
	}
}
func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	send(w, 200, map[string]string{"status": "ok"})
}
func (s *Server) ready(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	deps := map[string]string{"database": "ok", "oidc": "ok"}
	status := 200
	if s.dbHealth(ctx) != nil {
		deps["database"] = "unavailable"
		status = 503
	}
	if s.idp.Health(ctx) != nil {
		deps["oidc"] = "unavailable"
		status = 503
	}
	label := "ready"
	if status != 200 {
		label = "not_ready"
	}
	send(w, status, map[string]any{"status": label, "commit": s.cfg.CommitSHA, "sha": s.cfg.CommitSHA, "dependencies": deps, "release": map[string]string{"sha": s.cfg.CommitSHA, "scope": scope}})
}
func cookie(w http.ResponseWriter, name, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{Name: name, Value: value, Path: "/", HttpOnly: true, Secure: true, SameSite: http.SameSiteLaxMode, MaxAge: maxAge})
}
func (s *Server) login(w http.ResponseWriter, r *http.Request) {
	if !s.limits.allow(r.RemoteAddr) {
		problem(w, 429, "rate_limited", "Login rate limit exceeded")
		return
	}
	state, binding, nonce, verifier := application.RandomToken(), application.RandomToken(), application.RandomToken(), application.RandomToken()
	f := application.LoginRequest{StateHash: application.Hash(state), BindingHash: application.Hash(binding), Nonce: nonce, Verifier: verifier, ExpiresAt: time.Now().Add(5 * time.Minute)}
	if e := s.auth.SaveLogin(r.Context(), f); e != nil {
		failure(w, e)
		return
	}
	cookie(w, loginCookie, binding, 300)
	http.Redirect(w, r, s.idp.AuthorizationURL(state, nonce, verifier), http.StatusFound)
}
func (s *Server) callback(w http.ResponseWriter, r *http.Request) {
	cookie(w, loginCookie, "", -1)
	if !s.limits.allow(r.RemoteAddr) {
		problem(w, 429, "rate_limited", "Login rate limit exceeded")
		return
	}
	state, code := r.URL.Query().Get("state"), r.URL.Query().Get("code")
	c, e := r.Cookie(loginCookie)
	if e != nil || state == "" || len(state) > 128 || code == "" || len(code) > 8192 || r.URL.Query().Get("error") != "" {
		failure(w, domain.ErrUnauthorized)
		return
	}
	f, e := s.auth.ConsumeLogin(r.Context(), application.Hash(state), application.Hash(c.Value))
	if e != nil {
		failure(w, domain.ErrUnauthorized)
		return
	}
	i, e := s.idp.Exchange(r.Context(), code, f.Verifier, f.Nonce)
	if e != nil {
		failure(w, domain.ErrUnauthorized)
		return
	}
	raw := application.RandomToken()
	oldHash := ""
	if old, e := r.Cookie(sessionCookie); e == nil {
		oldHash = application.Hash(old.Value)
	}
	if e = s.auth.CreateSession(r.Context(), i, application.Hash(raw), oldHash); e != nil {
		failure(w, e)
		return
	}
	cookie(w, sessionCookie, raw, 43200)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
func (s *Server) logout(w http.ResponseWriter, r *http.Request) {
	_, raw, ok := s.current(w, r, true)
	if !ok {
		return
	}
	if e := s.auth.RevokeSession(r.Context(), application.Hash(raw)); e != nil {
		failure(w, e)
		return
	}
	cookie(w, sessionCookie, "", -1)
	send(w, 204, nil)
}
func (s *Server) static(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/api" || strings.HasPrefix(path, "/api/") || path == "/auth" || strings.HasPrefix(path, "/auth/") {
		problem(w, 404, "not_found", "Route not found")
		return
	}
	if r.Method != "GET" && r.Method != "HEAD" {
		problem(w, 405, "method_not_allowed", "Method not allowed")
		return
	}
	file := "index.html"
	if strings.HasPrefix(path, "/assets/") {
		clean := filepath.Clean(strings.TrimPrefix(path, "/assets/"))
		if clean == "." || strings.HasPrefix(clean, "..") || strings.Contains(clean, "\\") {
			problem(w, 404, "not_found", "Asset not found")
			return
		}
		file = filepath.Join("assets", clean)
	}
	full := filepath.Join(s.cfg.WebDir, file)
	st, e := os.Stat(full)
	if e != nil || !st.Mode().IsRegular() {
		problem(w, 404, "not_found", "UI build or asset not found")
		return
	}
	http.ServeFile(w, r, full)
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}
func (s *Server) middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := uuid.NewString()
		w.Header().Set("X-Request-ID", id)
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'self'")
		sw := &statusWriter{ResponseWriter: w, status: 200}
		defer func() {
			if recover() != nil {
				problem(sw, 500, "internal_error", "Internal server error")
				slog.Error("request panic", "request_id", id)
			}
			slog.Info("http request", "request_id", id, "method", r.Method, "status", sw.status)
		}()
		next.ServeHTTP(sw, r)
	})
}

type rateEntry struct {
	start time.Time
	count int
}
type limiter struct {
	mu     sync.Mutex
	ips    map[string]rateEntry
	global rateEntry
}

func (l *limiter) allow(remote string) bool {
	host, _, e := net.SplitHostPort(remote)
	if e != nil {
		host = remote
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	if l.ips == nil {
		l.ips = map[string]rateEntry{}
	}
	if now.Sub(l.global.start) > time.Minute {
		l.global = rateEntry{start: now}
		for k, v := range l.ips {
			if now.Sub(v.start) > time.Minute {
				delete(l.ips, k)
			}
		}
	}
	if len(l.ips) > 4096 {
		return false
	}
	v := l.ips[host]
	if now.Sub(v.start) > time.Minute {
		v = rateEntry{start: now}
	}
	if v.count >= 30 || l.global.count >= 300 {
		return false
	}
	v.count++
	l.global.count++
	l.ips[host] = v
	return true
}
