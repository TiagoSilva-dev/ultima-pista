package api

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ultima-pista/ultima-pista/internal/cases"
	"github.com/ultima-pista/ultima-pista/internal/models"
	"github.com/ultima-pista/ultima-pista/internal/store"
)

type Server struct {
	Store      *StoreWrap
	Secure     bool
	StaticDir  string
	createLim  *limiter
	joinLim    *limiter
	msgLim     *limiter
}

type StoreWrap = store.Store

func New(st *store.Store, secure bool, staticDir string) *Server {
	store.CaseLookup = func(id string) (models.Case, bool) {
		c, ok := cases.ByID(id)
		if !ok {
			return models.Case{}, false
		}
		return *c, true
	}
	return &Server{
		Store:     st,
		Secure:    secure,
		StaticDir: staticDir,
		createLim: newLimiter(10, time.Minute),
		joinLim:   newLimiter(30, time.Minute),
		msgLim:    newLimiter(60, time.Minute),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/cases", s.handleCases)
	mux.HandleFunc("/api/rooms", s.handleRooms)
	mux.HandleFunc("/api/rooms/join", s.handleJoin)
	mux.HandleFunc("/api/rooms/", s.handleRoomRoutes)
	mux.HandleFunc("/", s.handleStatic)
	return s.withMiddleware(mux)
}

func (s *Server) withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if strings.HasPrefix(r.URL.Path, "/api/") {
			w.Header().Set("Cache-Control", "no-store")
		}
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleCases(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	writeJSON(w, http.StatusOK, cases.PublicList())
}

func (s *Server) session(w http.ResponseWriter, r *http.Request) (string, error) {
	var sid string
	if c, err := r.Cookie(store.SessionCookie); err == nil {
		sid = c.Value
	}
	id, err := s.Store.EnsureSession(sid)
	if err != nil {
		return "", err
	}
	if sid != id {
		http.SetCookie(w, &http.Cookie{
			Name:     store.SessionCookie,
			Value:    id,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
			Secure:   s.Secure,
			MaxAge:   60 * 60 * 24 * 30,
		})
	}
	return id, nil
}

func (s *Server) checkOrigin(r *http.Request) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true // same-origin navigations / curl
	}
	host := r.Host
	return strings.HasSuffix(origin, "://"+host) || strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1")
}

func (s *Server) handleRooms(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	if !s.checkOrigin(r) {
		writeErr(w, http.StatusForbidden, "origem inválida")
		return
	}
	ip := clientIP(r)
	if !s.createLim.allow(ip) {
		writeErr(w, http.StatusTooManyRequests, "muitas salas criadas; aguarde")
		return
	}
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	var body struct {
		Nickname string `json:"nickname"`
		Mode     string `json:"mode"`
	}
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	room, _, err := s.Store.CreateRoom(sid, body.Nickname, body.Mode)
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	snap, err := s.Store.PublicSnapshot(room.Code, store.SnapshotOpts{SessionID: sid})
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, snap)
}

func (s *Server) handleJoin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}
	if !s.checkOrigin(r) {
		writeErr(w, http.StatusForbidden, "origem inválida")
		return
	}
	ip := clientIP(r)
	if !s.joinLim.allow(ip) {
		writeErr(w, http.StatusTooManyRequests, "muitas tentativas; aguarde")
		return
	}
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	var body struct {
		Code     string `json:"code"`
		Nickname string `json:"nickname"`
	}
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	room, _, err := s.Store.JoinRoom(sid, body.Code, body.Nickname)
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	snap, err := s.Store.PublicSnapshot(room.Code, store.SnapshotOpts{SessionID: sid})
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleRoomRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/rooms/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		writeErr(w, http.StatusNotFound, "não encontrado")
		return
	}
	code := strings.ToUpper(parts[0])
	rest := parts[1:]

	if len(rest) == 0 {
		if r.Method == http.MethodGet {
			s.handleSnapshot(w, r, code, false)
			return
		}
		writeErr(w, http.StatusMethodNotAllowed, "método não permitido")
		return
	}

	switch {
	case len(rest) == 1 && rest[0] == "telao" && r.Method == http.MethodGet:
		s.handleSnapshot(w, r, code, true)
	case len(rest) == 1 && rest[0] == "private" && r.Method == http.MethodGet:
		s.handlePrivate(w, r, code)
	case len(rest) == 1 && rest[0] == "events" && r.Method == http.MethodGet:
		s.handleSSE(w, r, code)
	case len(rest) == 1 && rest[0] == "start" && r.Method == http.MethodPost:
		s.handleStart(w, r, code)
	case len(rest) == 1 && rest[0] == "questions" && r.Method == http.MethodPost:
		s.handleAsk(w, r, code)
	case len(rest) == 1 && rest[0] == "oral" && r.Method == http.MethodPost:
		s.handleOral(w, r, code)
	case len(rest) == 3 && rest[0] == "questions" && rest[2] == "answer" && r.Method == http.MethodPost:
		s.handleAnswer(w, r, code, rest[1])
	case len(rest) == 1 && rest[0] == "theories" && r.Method == http.MethodPost:
		s.handleTheory(w, r, code)
	case len(rest) == 3 && rest[0] == "theories" && rest[2] == "continue" && r.Method == http.MethodPost:
		s.handleContinue(w, r, code, rest[1])
	case len(rest) == 3 && rest[0] == "theories" && rest[2] == "confirm" && r.Method == http.MethodPost:
		s.handleConfirmTheory(w, r, code, rest[1])
	case len(rest) == 1 && rest[0] == "confirm-solution" && r.Method == http.MethodPost:
		s.handleConfirmSpoken(w, r, code)
	case len(rest) == 1 && rest[0] == "lobby" && r.Method == http.MethodPost:
		s.handleLobby(w, r, code)
	default:
		writeErr(w, http.StatusNotFound, "não encontrado")
	}
}

func (s *Server) handleSnapshot(w http.ResponseWriter, r *http.Request, code string, telao bool) {
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	snap, err := s.Store.PublicSnapshot(code, store.SnapshotOpts{SessionID: sid, Telao: telao})
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	if telao {
		snap.You = nil
	}
	writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handlePrivate(w http.ResponseWriter, r *http.Request, code string) {
	w.Header().Set("Cache-Control", "no-store")
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	priv, err := s.Store.PrivateForNarrator(code, sid)
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, priv)
}

func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request, code string) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, http.StatusInternalServerError, "SSE indisponível")
		return
	}
	ver, err := s.Store.CurrentVersion(code)
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := s.Store.Subscribe(code)
	defer s.Store.Unsubscribe(code, ch)

	_, _ = fmt.Fprintf(w, "event: hello\ndata: {\"version\":%d}\n\n", ver)
	flusher.Flush()

	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case v, ok := <-ch:
			if !ok {
				return
			}
			_, _ = fmt.Fprintf(w, "event: invalidate\ndata: {\"version\":%d}\n\n", v)
			flusher.Flush()
		case <-ticker.C:
			_, _ = fmt.Fprintf(w, "event: ping\ndata: {}\n\n")
			flusher.Flush()
		}
	}
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request, code string) {
	if !s.checkOrigin(r) {
		writeErr(w, http.StatusForbidden, "origem inválida")
		return
	}
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	var body struct {
		CaseID string `json:"caseId"`
	}
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if _, ok := cases.ByID(body.CaseID); !ok {
		writeErr(w, http.StatusBadRequest, "caso inválido")
		return
	}
	if err := s.Store.StartRound(code, sid, body.CaseID); err != nil {
		mapStoreErr(w, err)
		return
	}
	snap, err := s.Store.PublicSnapshot(code, store.SnapshotOpts{SessionID: sid})
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleAsk(w http.ResponseWriter, r *http.Request, code string) {
	if !s.checkOrigin(r) {
		writeErr(w, http.StatusForbidden, "origem inválida")
		return
	}
	if !s.msgLim.allow(clientIP(r)) {
		writeErr(w, http.StatusTooManyRequests, "muitas mensagens; aguarde")
		return
	}
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	var body struct {
		Text           string `json:"text"`
		IdempotencyKey string `json:"idempotencyKey"`
	}
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if body.IdempotencyKey == "" {
		body.IdempotencyKey = r.Header.Get("Idempotency-Key")
	}
	q, err := s.Store.AskQuestion(code, sid, body.Text, body.IdempotencyKey)
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, q)
}

func (s *Server) handleOral(w http.ResponseWriter, r *http.Request, code string) {
	if !s.checkOrigin(r) {
		writeErr(w, http.StatusForbidden, "origem inválida")
		return
	}
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	var body struct {
		Text           string `json:"text"`
		Answer         string `json:"answer"`
		IdempotencyKey string `json:"idempotencyKey"`
	}
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if body.IdempotencyKey == "" {
		body.IdempotencyKey = r.Header.Get("Idempotency-Key")
	}
	q, err := s.Store.OralQuestion(code, sid, body.Text, body.Answer, body.IdempotencyKey)
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, q)
}

func (s *Server) handleAnswer(w http.ResponseWriter, r *http.Request, code, qid string) {
	if !s.checkOrigin(r) {
		writeErr(w, http.StatusForbidden, "origem inválida")
		return
	}
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	var body struct {
		Answer string `json:"answer"`
	}
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if err := s.Store.AnswerQuestion(code, sid, qid, body.Answer); err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func (s *Server) handleTheory(w http.ResponseWriter, r *http.Request, code string) {
	if !s.checkOrigin(r) {
		writeErr(w, http.StatusForbidden, "origem inválida")
		return
	}
	if !s.msgLim.allow(clientIP(r)) {
		writeErr(w, http.StatusTooManyRequests, "muitas mensagens; aguarde")
		return
	}
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	var body struct {
		Text           string `json:"text"`
		IdempotencyKey string `json:"idempotencyKey"`
	}
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if body.IdempotencyKey == "" {
		body.IdempotencyKey = r.Header.Get("Idempotency-Key")
	}
	th, err := s.Store.ProposeTheory(code, sid, body.Text, body.IdempotencyKey)
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, th)
}

func (s *Server) handleContinue(w http.ResponseWriter, r *http.Request, code, tid string) {
	if !s.checkOrigin(r) {
		writeErr(w, http.StatusForbidden, "origem inválida")
		return
	}
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	var body struct {
		Guidance string `json:"guidance"`
	}
	if err := readJSON(r, &body); err != nil {
		writeErr(w, http.StatusBadRequest, "json inválido")
		return
	}
	if err := s.Store.ContinueTheory(code, sid, tid, body.Guidance); err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func (s *Server) handleConfirmTheory(w http.ResponseWriter, r *http.Request, code, tid string) {
	if !s.checkOrigin(r) {
		writeErr(w, http.StatusForbidden, "origem inválida")
		return
	}
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	if err := s.Store.ConfirmTheory(code, sid, tid); err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func (s *Server) handleConfirmSpoken(w http.ResponseWriter, r *http.Request, code string) {
	if !s.checkOrigin(r) {
		writeErr(w, http.StatusForbidden, "origem inválida")
		return
	}
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	var body struct {
		Summary string `json:"summary"`
	}
	_ = readJSON(r, &body)
	if err := s.Store.ConfirmSpokenSolution(code, sid, body.Summary); err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func (s *Server) handleLobby(w http.ResponseWriter, r *http.Request, code string) {
	if !s.checkOrigin(r) {
		writeErr(w, http.StatusForbidden, "origem inválida")
		return
	}
	sid, err := s.session(w, r)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "sessão")
		return
	}
	if err := s.Store.ReturnToLobby(code, sid); err != nil {
		mapStoreErr(w, err)
		return
	}
	snap, err := s.Store.PublicSnapshot(code, store.SnapshotOpts{SessionID: sid})
	if err != nil {
		mapStoreErr(w, err)
		return
	}
	writeJSON(w, http.StatusOK, snap)
}

func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if s.StaticDir == "" {
		writeErr(w, http.StatusNotFound, "frontend não compilado")
		return
	}
	path := r.URL.Path
	if path == "/" {
		path = "/index.html"
	}
	full := filepath.Join(s.StaticDir, filepath.Clean(path))
	if !strings.HasPrefix(full, s.StaticDir) {
		writeErr(w, http.StatusForbidden, "caminho inválido")
		return
	}
	if st, err := os.Stat(full); err == nil && !st.IsDir() {
		http.ServeFile(w, r, full)
		return
	}
	// SPA fallback
	http.ServeFile(w, r, filepath.Join(s.StaticDir, "index.html"))
}

func mapStoreErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeErr(w, http.StatusNotFound, "sala não encontrada")
	case errors.Is(err, store.ErrExpired):
		writeErr(w, http.StatusGone, "sala expirada por inatividade (24h)")
	case errors.Is(err, store.ErrForbidden):
		writeErr(w, http.StatusForbidden, "ação não permitida")
	case errors.Is(err, store.ErrConflict), errors.Is(err, store.ErrAlreadyAnswered):
		writeErr(w, http.StatusConflict, err.Error())
	case errors.Is(err, store.ErrValidation), errors.Is(err, store.ErrRoundInactive):
		writeErr(w, http.StatusBadRequest, err.Error())
	default:
		writeErr(w, http.StatusInternalServerError, "erro interno")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func readJSON(r *http.Request, dst any) error {
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	return dec.Decode(dst)
}

func clientIP(r *http.Request) string {
	if x := r.Header.Get("X-Forwarded-For"); x != "" {
		return strings.TrimSpace(strings.Split(x, ",")[0])
	}
	return r.RemoteAddr
}

type limiter struct {
	mu     sync.Mutex
	window time.Duration
	max    int
	hits   map[string][]time.Time
}

func newLimiter(max int, window time.Duration) *limiter {
	return &limiter{window: window, max: max, hits: make(map[string][]time.Time)}
}

func (l *limiter) allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	cut := now.Add(-l.window)
	arr := l.hits[key]
	n := 0
	for _, t := range arr {
		if t.After(cut) {
			arr[n] = t
			n++
		}
	}
	arr = arr[:n]
	if len(arr) >= l.max {
		l.hits[key] = arr
		return false
	}
	l.hits[key] = append(arr, now)
	return true
}

func RandomHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
