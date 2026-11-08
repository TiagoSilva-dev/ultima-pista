package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"

	"github.com/ultima-pista/ultima-pista/internal/models"
)

const (
	RoomTTL          = 24 * time.Hour
	SessionCookie    = "up_session"
	MaxNicknameLen   = 24
	MaxQuestionLen   = 280
	MaxTheoryLen     = 2000
	MaxOralLen       = 280
	MaxGuidanceLen   = 280
)

var (
	ErrNotFound       = errors.New("não encontrado")
	ErrForbidden      = errors.New("não autorizado")
	ErrConflict       = errors.New("conflito")
	ErrExpired        = errors.New("sala expirada")
	ErrValidation     = errors.New("dados inválidos")
	ErrRoundInactive  = errors.New("rodada inativa")
	ErrAlreadyAnswered = errors.New("já respondida")
)

type Store struct {
	db *sql.DB
	mu sync.Mutex

	subsMu sync.RWMutex
	subs   map[string]map[chan int64]struct{}
}

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db, subs: make(map[string]map[chan int64]struct{})}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error { return s.db.Close() }

func (s *Store) migrate() error {
	_, err := s.db.Exec(`
CREATE TABLE IF NOT EXISTS sessions (
  id TEXT PRIMARY KEY,
  created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS rooms (
  id TEXT PRIMARY KEY,
  code TEXT NOT NULL UNIQUE,
  mode TEXT NOT NULL,
  narrator_session TEXT NOT NULL,
  created_at TEXT NOT NULL,
  last_activity TEXT NOT NULL,
  version INTEGER NOT NULL DEFAULT 1,
  active_round_id TEXT
);
CREATE TABLE IF NOT EXISTS participants (
  id TEXT PRIMARY KEY,
  room_id TEXT NOT NULL REFERENCES rooms(id),
  session_id TEXT NOT NULL,
  nickname TEXT NOT NULL,
  role TEXT NOT NULL,
  joined_at TEXT NOT NULL,
  last_seen TEXT NOT NULL,
  UNIQUE(room_id, session_id)
);
CREATE TABLE IF NOT EXISTS rounds (
  id TEXT PRIMARY KEY,
  room_id TEXT NOT NULL REFERENCES rooms(id),
  case_id TEXT NOT NULL,
  status TEXT NOT NULL,
  started_at TEXT NOT NULL,
  resolved_at TEXT,
  accepted_theory_id TEXT,
  accepted_theory_txt TEXT,
  number INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS questions (
  id TEXT PRIMARY KEY,
  round_id TEXT NOT NULL REFERENCES rounds(id),
  asker_id TEXT,
  asker_nickname TEXT NOT NULL,
  text TEXT NOT NULL,
  kind TEXT NOT NULL,
  status TEXT NOT NULL,
  answer TEXT,
  created_at TEXT NOT NULL,
  answered_at TEXT,
  idempotency_key TEXT NOT NULL,
  UNIQUE(round_id, idempotency_key)
);
CREATE TABLE IF NOT EXISTS theories (
  id TEXT PRIMARY KEY,
  round_id TEXT NOT NULL REFERENCES rounds(id),
  author_id TEXT,
  author_nickname TEXT NOT NULL,
  text TEXT NOT NULL,
  status TEXT NOT NULL,
  guidance TEXT,
  created_at TEXT NOT NULL,
  resolved_at TEXT,
  idempotency_key TEXT NOT NULL,
  UNIQUE(round_id, idempotency_key)
);
CREATE INDEX IF NOT EXISTS idx_rooms_code ON rooms(code);
CREATE INDEX IF NOT EXISTS idx_participants_room ON participants(room_id);
CREATE INDEX IF NOT EXISTS idx_questions_round ON questions(round_id, created_at);
CREATE INDEX IF NOT EXISTS idx_theories_round ON theories(round_id, created_at);
`)
	return err
}

func now() time.Time { return time.Now().UTC() }
func fmtTime(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }
func parseTime(v string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, v)
	if err != nil {
		t, _ = time.Parse(time.RFC3339, v)
	}
	return t
}

func newID() string { return uuid.NewString() }

func genCode() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	b := make([]byte, 4)
	u := uuid.New()
	for i := 0; i < 4; i++ {
		b[i] = alphabet[int(u[i])%len(alphabet)]
	}
	return string(b)
}

func (s *Store) EnsureSession(sessionID string) (string, error) {
	if sessionID != "" {
		var id string
		err := s.db.QueryRow(`SELECT id FROM sessions WHERE id = ?`, sessionID).Scan(&id)
		if err == nil {
			return id, nil
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return "", err
		}
	}
	id := newID()
	_, err := s.db.Exec(`INSERT INTO sessions(id, created_at) VALUES(?, ?)`, id, fmtTime(now()))
	return id, err
}

func (s *Store) bump(tx *sql.Tx, roomID string) (int64, error) {
	_, err := tx.Exec(`UPDATE rooms SET version = version + 1, last_activity = ? WHERE id = ?`, fmtTime(now()), roomID)
	if err != nil {
		return 0, err
	}
	var v int64
	err = tx.QueryRow(`SELECT version FROM rooms WHERE id = ?`, roomID).Scan(&v)
	return v, err
}

func (s *Store) notify(roomCode string, version int64) {
	s.subsMu.RLock()
	defer s.subsMu.RUnlock()
	for ch := range s.subs[roomCode] {
		select {
		case ch <- version:
		default:
		}
	}
}

func (s *Store) Subscribe(roomCode string) chan int64 {
	ch := make(chan int64, 8)
	s.subsMu.Lock()
	if s.subs[roomCode] == nil {
		s.subs[roomCode] = make(map[chan int64]struct{})
	}
	s.subs[roomCode][ch] = struct{}{}
	s.subsMu.Unlock()
	return ch
}

func (s *Store) Unsubscribe(roomCode string, ch chan int64) {
	s.subsMu.Lock()
	if m := s.subs[roomCode]; m != nil {
		delete(m, ch)
		if len(m) == 0 {
			delete(s.subs, roomCode)
		}
	}
	s.subsMu.Unlock()
	close(ch)
}

func normalizeNickname(n string) (string, error) {
	n = strings.TrimSpace(n)
	if n == "" || len([]rune(n)) > MaxNicknameLen {
		return "", fmt.Errorf("%w: apelido inválido", ErrValidation)
	}
	return n, nil
}

func (s *Store) CreateRoom(sessionID, nickname, mode string) (*models.Room, *models.Participant, error) {
	nickname, err := normalizeNickname(nickname)
	if err != nil {
		return nil, nil, err
	}
	if mode != models.ModePresencial && mode != models.ModeRemoto {
		return nil, nil, fmt.Errorf("%w: modo inválido", ErrValidation)
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var code string
	for i := 0; i < 12; i++ {
		c := genCode()
		var existing string
		err := tx.QueryRow(`SELECT code FROM rooms WHERE code = ?`, c).Scan(&existing)
		if errors.Is(err, sql.ErrNoRows) {
			code = c
			break
		}
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return nil, nil, err
		}
	}
	if code == "" {
		return nil, nil, errors.New("falha ao gerar código")
	}

	t := now()
	room := &models.Room{
		ID:              newID(),
		Code:            code,
		Mode:            mode,
		NarratorSession: sessionID,
		CreatedAt:       t,
		LastActivity:    t,
		Version:         1,
	}
	_, err = tx.Exec(`INSERT INTO rooms(id, code, mode, narrator_session, created_at, last_activity, version) VALUES(?,?,?,?,?,?,?)`,
		room.ID, room.Code, room.Mode, room.NarratorSession, fmtTime(room.CreatedAt), fmtTime(room.LastActivity), room.Version)
	if err != nil {
		return nil, nil, err
	}
	p := &models.Participant{
		ID:        newID(),
		RoomID:    room.ID,
		SessionID: sessionID,
		Nickname:  nickname,
		Role:      models.RoleNarrador,
		JoinedAt:  t,
		LastSeen:  t,
	}
	_, err = tx.Exec(`INSERT INTO participants(id, room_id, session_id, nickname, role, joined_at, last_seen) VALUES(?,?,?,?,?,?,?)`,
		p.ID, p.RoomID, p.SessionID, p.Nickname, p.Role, fmtTime(p.JoinedAt), fmtTime(p.LastSeen))
	if err != nil {
		return nil, nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	return room, p, nil
}

func (s *Store) getRoomByCodeTx(tx *sql.Tx, code string) (*models.Room, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	r := &models.Room{}
	var active sql.NullString
	var created, last string
	err := tx.QueryRow(`SELECT id, code, mode, narrator_session, created_at, last_activity, version, active_round_id FROM rooms WHERE code = ?`, code).
		Scan(&r.ID, &r.Code, &r.Mode, &r.NarratorSession, &created, &last, &r.Version, &active)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	r.CreatedAt = parseTime(created)
	r.LastActivity = parseTime(last)
	if active.Valid {
		r.ActiveRoundID = &active.String
	}
	if now().Sub(r.LastActivity) > RoomTTL {
		return r, ErrExpired
	}
	return r, nil
}

func (s *Store) JoinRoom(sessionID, code, nickname string) (*models.Room, *models.Participant, error) {
	nickname, err := normalizeNickname(nickname)
	if err != nil {
		return nil, nil, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	tx, err := s.db.Begin()
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = tx.Rollback() }()

	room, err := s.getRoomByCodeTx(tx, code)
	if err != nil {
		return nil, nil, err
	}
	if sessionID == room.NarratorSession {
		// Narrator rejoining — update nickname/seen
		var p models.Participant
		var joined, seen string
		err = tx.QueryRow(`SELECT id, room_id, session_id, nickname, role, joined_at, last_seen FROM participants WHERE room_id = ? AND session_id = ?`,
			room.ID, sessionID).Scan(&p.ID, &p.RoomID, &p.SessionID, &p.Nickname, &p.Role, &joined, &seen)
		if err != nil {
			return nil, nil, err
		}
		p.JoinedAt = parseTime(joined)
		p.LastSeen = now()
		p.Nickname = nickname
		_, _ = tx.Exec(`UPDATE participants SET nickname = ?, last_seen = ? WHERE id = ?`, nickname, fmtTime(p.LastSeen), p.ID)
		v, err := s.bump(tx, room.ID)
		if err != nil {
			return nil, nil, err
		}
		room.Version = v
		if err := tx.Commit(); err != nil {
			return nil, nil, err
		}
		s.notify(room.Code, room.Version)
		return room, &p, nil
	}

	var existing models.Participant
	var joined, seen string
	err = tx.QueryRow(`SELECT id, room_id, session_id, nickname, role, joined_at, last_seen FROM participants WHERE room_id = ? AND session_id = ?`,
		room.ID, sessionID).Scan(&existing.ID, &existing.RoomID, &existing.SessionID, &existing.Nickname, &existing.Role, &joined, &seen)
	if err == nil {
		existing.JoinedAt = parseTime(joined)
		existing.LastSeen = now()
		existing.Nickname = nickname
		_, _ = tx.Exec(`UPDATE participants SET nickname = ?, last_seen = ? WHERE id = ?`, nickname, fmtTime(existing.LastSeen), existing.ID)
		v, err := s.bump(tx, room.ID)
		if err != nil {
			return nil, nil, err
		}
		room.Version = v
		if err := tx.Commit(); err != nil {
			return nil, nil, err
		}
		s.notify(room.Code, room.Version)
		return room, &existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, nil, err
	}

	t := now()
	p := &models.Participant{
		ID: newID(), RoomID: room.ID, SessionID: sessionID, Nickname: nickname,
		Role: models.RoleDetetive, JoinedAt: t, LastSeen: t,
	}
	_, err = tx.Exec(`INSERT INTO participants(id, room_id, session_id, nickname, role, joined_at, last_seen) VALUES(?,?,?,?,?,?,?)`,
		p.ID, p.RoomID, p.SessionID, p.Nickname, p.Role, fmtTime(p.JoinedAt), fmtTime(p.LastSeen))
	if err != nil {
		return nil, nil, err
	}
	v, err := s.bump(tx, room.ID)
	if err != nil {
		return nil, nil, err
	}
	room.Version = v
	if err := tx.Commit(); err != nil {
		return nil, nil, err
	}
	s.notify(room.Code, room.Version)
	return room, p, nil
}

func (s *Store) TouchParticipant(sessionID, roomID string) {
	_, _ = s.db.Exec(`UPDATE participants SET last_seen = ? WHERE session_id = ? AND room_id = ?`, fmtTime(now()), sessionID, roomID)
}

type SnapshotOpts struct {
	SessionID string
	Telao     bool
}

func (s *Store) PublicSnapshot(code string, opts SnapshotOpts) (*models.RoomPublic, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	room, err := s.getRoomByCodeTx(tx, code)
	if err != nil && !errors.Is(err, ErrExpired) {
		return nil, err
	}
	expired := errors.Is(err, ErrExpired)
	if room == nil {
		return nil, ErrNotFound
	}

	parts, err := s.listParticipants(tx, room.ID)
	if err != nil {
		return nil, err
	}
	pub := &models.RoomPublic{
		Code:         room.Code,
		Mode:         room.Mode,
		Version:      room.Version,
		Participants: parts,
		InvitePath:   "/entrar/" + room.Code,
		TelaoPath:    "/telao/" + room.Code,
		Expired:      expired,
	}

	if opts.SessionID != "" && !opts.Telao {
		var sessionID, role, nick, pid string
		err := tx.QueryRow(`SELECT id, session_id, nickname, role FROM participants WHERE room_id = ? AND session_id = ?`, room.ID, opts.SessionID).
			Scan(&pid, &sessionID, &nick, &role)
		if err == nil {
			pub.You = &models.YouInfo{
				ParticipantID: pid,
				Nickname:      nick,
				Role:          role,
				IsNarrator:    role == models.RoleNarrador && sessionID == room.NarratorSession,
			}
		}
	}

	if room.ActiveRoundID != nil {
		rp, err := s.loadRoundPublic(tx, *room.ActiveRoundID, opts)
		if err != nil {
			return nil, err
		}
		pub.Round = rp
	}
	_ = tx.Commit()
	return pub, nil
}

func (s *Store) listParticipants(tx *sql.Tx, roomID string) ([]models.ParticipantPublic, error) {
	rows, err := tx.Query(`SELECT id, nickname, role FROM participants WHERE room_id = ? ORDER BY joined_at`, roomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.ParticipantPublic
	for rows.Next() {
		var p models.ParticipantPublic
		if err := rows.Scan(&p.ID, &p.Nickname, &p.Role); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if out == nil {
		out = []models.ParticipantPublic{}
	}
	return out, rows.Err()
}

func (s *Store) loadRoundPublic(tx *sql.Tx, roundID string, opts SnapshotOpts) (*models.RoundPublic, error) {
	var r models.Round
	var resolved, accID, accTxt sql.NullString
	var started string
	err := tx.QueryRow(`SELECT id, room_id, case_id, status, started_at, resolved_at, accepted_theory_id, accepted_theory_txt, number FROM rounds WHERE id = ?`, roundID).
		Scan(&r.ID, &r.RoomID, &r.CaseID, &r.Status, &started, &resolved, &accID, &accTxt, &r.Number)
	if err != nil {
		return nil, err
	}
	r.StartedAt = parseTime(started)
	rp := &models.RoundPublic{
		ID:        r.ID,
		Number:    r.Number,
		Status:    r.Status,
		StartedAt: r.StartedAt.Format(time.RFC3339),
		Questions: []models.QuestionPublic{},
		Theories:  []models.TheoryPublic{},
	}
	if resolved.Valid {
		rp.ResolvedAt = &resolved.String
	}
	if accTxt.Valid {
		rp.AcceptedTheory = &accTxt.String
	}

	// case public
	cata := mustCase(r.CaseID)
	cp := models.CasePublic{
		ID: cata.ID, Title: cata.Title, Theme: cata.Theme, Difficulty: cata.Difficulty,
		PublicText: cata.PublicText, Illustration: cata.Illustration,
	}
	rp.Case = &cp
	if r.Status == models.RoundResolvida {
		sol := cata.Solution
		rp.RevealedSolution = &sol
	}

	var myParticipant string
	var isNarr bool
	if opts.SessionID != "" && !opts.Telao {
		var role string
		_ = tx.QueryRow(`SELECT id, role FROM participants WHERE room_id = ? AND session_id = ?`, r.RoomID, opts.SessionID).Scan(&myParticipant, &role)
		isNarr = role == models.RoleNarrador
	}

	qrows, err := tx.Query(`SELECT id, asker_id, asker_nickname, text, kind, status, answer, created_at FROM questions WHERE round_id = ? ORDER BY created_at`, roundID)
	if err != nil {
		return nil, err
	}
	for qrows.Next() {
		var q models.QuestionPublic
		var askerID, answer sql.NullString
		var created string
		if err := qrows.Scan(&q.ID, &askerID, &q.AskerNickname, &q.Text, &q.Kind, &q.Status, &answer, &created); err != nil {
			_ = qrows.Close()
			return nil, err
		}
		q.CreatedAt = created
		if answer.Valid {
			q.Answer = &answer.String
		}
		if askerID.Valid && askerID.String == myParticipant {
			q.Mine = true
		}
		if q.Status == models.QuestionPendente {
			if opts.Telao || (!isNarr && !q.Mine) {
				continue
			}
		}
		rp.Questions = append(rp.Questions, q)
	}
	if err := qrows.Close(); err != nil {
		return nil, err
	}

	trows, err := tx.Query(`SELECT id, author_id, author_nickname, text, status, guidance, created_at FROM theories WHERE round_id = ? ORDER BY created_at`, roundID)
	if err != nil {
		return nil, err
	}
	for trows.Next() {
		var t models.TheoryPublic
		var authorID, guidance sql.NullString
		var created string
		if err := trows.Scan(&t.ID, &authorID, &t.AuthorNickname, &t.Text, &t.Status, &guidance, &created); err != nil {
			_ = trows.Close()
			return nil, err
		}
		t.CreatedAt = created
		if guidance.Valid {
			t.Guidance = &guidance.String
		}
		if authorID.Valid && authorID.String == myParticipant {
			t.Mine = true
		}
		if t.Status == models.TheoryPendente {
			if opts.Telao || (!isNarr && !t.Mine) {
				continue
			}
		}
		rp.Theories = append(rp.Theories, t)
	}
	if err := trows.Close(); err != nil {
		return nil, err
	}
	return rp, nil
}

func mustCase(id string) models.Case {
	if c, ok := lookupCase(id); ok {
		return c
	}
	return models.Case{ID: id, Title: id}
}

// CaseLookup is injected from the cases package to avoid an import cycle.
var CaseLookup func(id string) (models.Case, bool)

func lookupCase(id string) (models.Case, bool) {
	if CaseLookup == nil {
		return models.Case{}, false
	}
	return CaseLookup(id)
}

func (s *Store) PrivateForNarrator(code, sessionID string) (*models.PrivatePayload, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	room, err := s.getRoomByCodeTx(tx, code)
	if err != nil {
		return nil, err
	}
	if sessionID != room.NarratorSession {
		return nil, ErrForbidden
	}
	if room.ActiveRoundID == nil {
		return nil, fmt.Errorf("%w: nenhuma rodada ativa", ErrValidation)
	}
	var caseID, status string
	err = tx.QueryRow(`SELECT case_id, status FROM rounds WHERE id = ?`, *room.ActiveRoundID).Scan(&caseID, &status)
	if err != nil {
		return nil, err
	}
	c := mustCase(caseID)
	return &models.PrivatePayload{
		Solution:   c.Solution,
		Essentials: c.Essentials,
		Hints:      c.Hints,
		CaseTitle:  c.Title,
	}, nil
}

func (s *Store) requireNarrator(tx *sql.Tx, code, sessionID string) (*models.Room, error) {
	room, err := s.getRoomByCodeTx(tx, code)
	if err != nil {
		return nil, err
	}
	if sessionID != room.NarratorSession {
		return nil, ErrForbidden
	}
	return room, nil
}

func (s *Store) StartRound(code, sessionID, caseID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	room, err := s.requireNarrator(tx, code, sessionID)
	if err != nil {
		return err
	}
	c, ok := lookupCase(caseID)
	if !ok || c.PublicText == "" {
		return fmt.Errorf("%w: caso inválido", ErrValidation)
	}
	if room.ActiveRoundID != nil {
		var st string
		_ = tx.QueryRow(`SELECT status FROM rounds WHERE id = ?`, *room.ActiveRoundID).Scan(&st)
		if st == models.RoundEmAndamento {
			return fmt.Errorf("%w: já existe rodada em andamento", ErrConflict)
		}
	}
	var n int
	_ = tx.QueryRow(`SELECT COUNT(*) FROM rounds WHERE room_id = ?`, room.ID).Scan(&n)
	rid := newID()
	_, err = tx.Exec(`INSERT INTO rounds(id, room_id, case_id, status, started_at, number) VALUES(?,?,?,?,?,?)`,
		rid, room.ID, caseID, models.RoundEmAndamento, fmtTime(now()), n+1)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`UPDATE rooms SET active_round_id = ? WHERE id = ?`, rid, room.ID)
	if err != nil {
		return err
	}
	v, err := s.bump(tx, room.ID)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.notify(room.Code, v)
	return nil
}

func (s *Store) AskQuestion(code, sessionID, text, idem string) (*models.Question, error) {
	text = strings.TrimSpace(text)
	if text == "" || len([]rune(text)) > MaxQuestionLen {
		return nil, fmt.Errorf("%w: pergunta inválida", ErrValidation)
	}
	if idem == "" {
		return nil, fmt.Errorf("%w: idempotency-key obrigatória", ErrValidation)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	room, err := s.getRoomByCodeTx(tx, code)
	if err != nil {
		return nil, err
	}
	if room.ActiveRoundID == nil {
		return nil, ErrRoundInactive
	}
	var status string
	_ = tx.QueryRow(`SELECT status FROM rounds WHERE id = ?`, *room.ActiveRoundID).Scan(&status)
	if status != models.RoundEmAndamento {
		return nil, ErrRoundInactive
	}
	var pid, nick, role string
	err = tx.QueryRow(`SELECT id, nickname, role FROM participants WHERE room_id = ? AND session_id = ?`, room.ID, sessionID).
		Scan(&pid, &nick, &role)
	if err != nil {
		return nil, ErrForbidden
	}
	if role == models.RoleNarrador {
		return nil, fmt.Errorf("%w: narrador não envia perguntas", ErrForbidden)
	}

	var existing models.Question
	var askerID sql.NullString
	var answered sql.NullString
	var created string
	err = tx.QueryRow(`SELECT id, round_id, asker_id, asker_nickname, text, kind, status, answer, created_at, answered_at, idempotency_key FROM questions WHERE round_id = ? AND idempotency_key = ?`,
		*room.ActiveRoundID, idem).Scan(&existing.ID, &existing.RoundID, &askerID, &existing.AskerNickname, &existing.Text, &existing.Kind, &existing.Status, &existing.Answer, &created, &answered, &existing.IdempotencyKey)
	if err == nil {
		existing.CreatedAt = parseTime(created)
		if askerID.Valid {
			existing.AskerID = &askerID.String
		}
		return &existing, nil
	}

	q := &models.Question{
		ID: newID(), RoundID: *room.ActiveRoundID, AskerID: &pid, AskerNickname: nick,
		Text: text, Kind: models.QuestionKindTyped, Status: models.QuestionPendente,
		CreatedAt: now(), IdempotencyKey: idem,
	}
	_, err = tx.Exec(`INSERT INTO questions(id, round_id, asker_id, asker_nickname, text, kind, status, created_at, idempotency_key) VALUES(?,?,?,?,?,?,?,?,?)`,
		q.ID, q.RoundID, pid, nick, text, q.Kind, q.Status, fmtTime(q.CreatedAt), idem)
	if err != nil {
		return nil, err
	}
	v, err := s.bump(tx, room.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.notify(room.Code, v)
	return q, nil
}

func (s *Store) OralQuestion(code, sessionID, text, answer, idem string) (*models.Question, error) {
	text = strings.TrimSpace(text)
	if text == "" || len([]rune(text)) > MaxOralLen {
		return nil, fmt.Errorf("%w: resumo inválido", ErrValidation)
	}
	if answer != models.AnswerSim && answer != models.AnswerNao && answer != models.AnswerIrrelevante {
		return nil, fmt.Errorf("%w: resposta inválida", ErrValidation)
	}
	if idem == "" {
		return nil, fmt.Errorf("%w: idempotency-key obrigatória", ErrValidation)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	room, err := s.requireNarrator(tx, code, sessionID)
	if err != nil {
		return nil, err
	}
	if room.ActiveRoundID == nil {
		return nil, ErrRoundInactive
	}
	var status string
	_ = tx.QueryRow(`SELECT status FROM rounds WHERE id = ?`, *room.ActiveRoundID).Scan(&status)
	if status != models.RoundEmAndamento {
		return nil, ErrRoundInactive
	}

	var existing models.Question
	var askerID sql.NullString
	var ans sql.NullString
	var created, answered sql.NullString
	err = tx.QueryRow(`SELECT id, round_id, asker_id, asker_nickname, text, kind, status, answer, created_at, answered_at, idempotency_key FROM questions WHERE round_id = ? AND idempotency_key = ?`,
		*room.ActiveRoundID, idem).Scan(&existing.ID, &existing.RoundID, &askerID, &existing.AskerNickname, &existing.Text, &existing.Kind, &existing.Status, &ans, &created, &answered, &existing.IdempotencyKey)
	if err == nil {
		return &existing, nil
	}

	t := now()
	a := answer
	q := &models.Question{
		ID: newID(), RoundID: *room.ActiveRoundID, AskerNickname: "Presencial",
		Text: text, Kind: models.QuestionKindOral, Status: models.QuestionRespondida,
		Answer: &a, CreatedAt: t, AnsweredAt: &t, IdempotencyKey: idem,
	}
	_, err = tx.Exec(`INSERT INTO questions(id, round_id, asker_nickname, text, kind, status, answer, created_at, answered_at, idempotency_key) VALUES(?,?,?,?,?,?,?,?,?,?)`,
		q.ID, q.RoundID, q.AskerNickname, text, q.Kind, q.Status, answer, fmtTime(t), fmtTime(t), idem)
	if err != nil {
		return nil, err
	}
	v, err := s.bump(tx, room.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.notify(room.Code, v)
	return q, nil
}

func (s *Store) AnswerQuestion(code, sessionID, questionID, answer string) error {
	if answer != models.AnswerSim && answer != models.AnswerNao && answer != models.AnswerIrrelevante {
		return fmt.Errorf("%w: resposta inválida", ErrValidation)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	room, err := s.requireNarrator(tx, code, sessionID)
	if err != nil {
		return err
	}
	if room.ActiveRoundID == nil {
		return ErrRoundInactive
	}
	var roundID, status string
	err = tx.QueryRow(`SELECT round_id, status FROM questions WHERE id = ?`, questionID).Scan(&roundID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if roundID != *room.ActiveRoundID {
		return ErrRoundInactive
	}
	var rst string
	_ = tx.QueryRow(`SELECT status FROM rounds WHERE id = ?`, roundID).Scan(&rst)
	if rst != models.RoundEmAndamento {
		return ErrRoundInactive
	}
	if status != models.QuestionPendente {
		return ErrAlreadyAnswered
	}
	res, err := tx.Exec(`UPDATE questions SET status = ?, answer = ?, answered_at = ? WHERE id = ? AND status = ?`,
		models.QuestionRespondida, answer, fmtTime(now()), questionID, models.QuestionPendente)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrAlreadyAnswered
	}
	v, err := s.bump(tx, room.ID)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.notify(room.Code, v)
	return nil
}

func (s *Store) ProposeTheory(code, sessionID, text, idem string) (*models.Theory, error) {
	text = strings.TrimSpace(text)
	if text == "" || len([]rune(text)) > MaxTheoryLen {
		return nil, fmt.Errorf("%w: teoria inválida", ErrValidation)
	}
	if idem == "" {
		return nil, fmt.Errorf("%w: idempotency-key obrigatória", ErrValidation)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	room, err := s.getRoomByCodeTx(tx, code)
	if err != nil {
		return nil, err
	}
	if room.ActiveRoundID == nil {
		return nil, ErrRoundInactive
	}
	var rst string
	_ = tx.QueryRow(`SELECT status FROM rounds WHERE id = ?`, *room.ActiveRoundID).Scan(&rst)
	if rst != models.RoundEmAndamento {
		return nil, ErrRoundInactive
	}
	var pid, nick, role string
	err = tx.QueryRow(`SELECT id, nickname, role FROM participants WHERE room_id = ? AND session_id = ?`, room.ID, sessionID).
		Scan(&pid, &nick, &role)
	if err != nil {
		return nil, ErrForbidden
	}
	if role == models.RoleNarrador {
		return nil, fmt.Errorf("%w: use confirmar solução para teoria falada", ErrForbidden)
	}

	var existing models.Theory
	var authorID, guidance, resolved sql.NullString
	var created string
	err = tx.QueryRow(`SELECT id, round_id, author_id, author_nickname, text, status, guidance, created_at, resolved_at, idempotency_key FROM theories WHERE round_id = ? AND idempotency_key = ?`,
		*room.ActiveRoundID, idem).Scan(&existing.ID, &existing.RoundID, &authorID, &existing.AuthorNickname, &existing.Text, &existing.Status, &guidance, &created, &resolved, &existing.IdempotencyKey)
	if err == nil {
		return &existing, nil
	}

	th := &models.Theory{
		ID: newID(), RoundID: *room.ActiveRoundID, AuthorID: &pid, AuthorNickname: nick,
		Text: text, Status: models.TheoryPendente, CreatedAt: now(), IdempotencyKey: idem,
	}
	_, err = tx.Exec(`INSERT INTO theories(id, round_id, author_id, author_nickname, text, status, created_at, idempotency_key) VALUES(?,?,?,?,?,?,?,?)`,
		th.ID, th.RoundID, pid, nick, text, th.Status, fmtTime(th.CreatedAt), idem)
	if err != nil {
		return nil, err
	}
	v, err := s.bump(tx, room.ID)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	s.notify(room.Code, v)
	return th, nil
}

func (s *Store) ContinueTheory(code, sessionID, theoryID, guidance string) error {
	guidance = strings.TrimSpace(guidance)
	if len([]rune(guidance)) > MaxGuidanceLen {
		return fmt.Errorf("%w: orientação longa demais", ErrValidation)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	room, err := s.requireNarrator(tx, code, sessionID)
	if err != nil {
		return err
	}
	var roundID, status string
	err = tx.QueryRow(`SELECT round_id, status FROM theories WHERE id = ?`, theoryID).Scan(&roundID, &status)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if room.ActiveRoundID == nil || roundID != *room.ActiveRoundID {
		return ErrRoundInactive
	}
	if status != models.TheoryPendente {
		return ErrConflict
	}
	var g any
	if guidance != "" {
		g = guidance
	}
	res, err := tx.Exec(`UPDATE theories SET status = ?, guidance = ?, resolved_at = ? WHERE id = ? AND status = ?`,
		models.TheoryContinuada, g, fmtTime(now()), theoryID, models.TheoryPendente)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrConflict
	}
	v, err := s.bump(tx, room.ID)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.notify(room.Code, v)
	return nil
}

func (s *Store) ConfirmTheory(code, sessionID, theoryID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	room, err := s.requireNarrator(tx, code, sessionID)
	if err != nil {
		return err
	}
	var roundID, status, text string
	err = tx.QueryRow(`SELECT round_id, status, text FROM theories WHERE id = ?`, theoryID).Scan(&roundID, &status, &text)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if room.ActiveRoundID == nil || roundID != *room.ActiveRoundID {
		return ErrRoundInactive
	}
	var rst string
	_ = tx.QueryRow(`SELECT status FROM rounds WHERE id = ?`, roundID).Scan(&rst)
	if rst != models.RoundEmAndamento {
		return ErrRoundInactive
	}
	if status != models.TheoryPendente {
		return ErrConflict
	}
	t := fmtTime(now())
	res, err := tx.Exec(`UPDATE theories SET status = ?, resolved_at = ? WHERE id = ? AND status = ?`,
		models.TheoryAceita, t, theoryID, models.TheoryPendente)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrConflict
	}
	res, err = tx.Exec(`UPDATE rounds SET status = ?, resolved_at = ?, accepted_theory_id = ?, accepted_theory_txt = ? WHERE id = ? AND status = ?`,
		models.RoundResolvida, t, theoryID, text, roundID, models.RoundEmAndamento)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrConflict
	}
	v, err := s.bump(tx, room.ID)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.notify(room.Code, v)
	return nil
}

func (s *Store) ConfirmSpokenSolution(code, sessionID, summary string) error {
	summary = strings.TrimSpace(summary)
	if len([]rune(summary)) > MaxTheoryLen {
		return fmt.Errorf("%w: resumo inválido", ErrValidation)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	room, err := s.requireNarrator(tx, code, sessionID)
	if err != nil {
		return err
	}
	if room.ActiveRoundID == nil {
		return ErrRoundInactive
	}
	roundID := *room.ActiveRoundID
	var rst string
	_ = tx.QueryRow(`SELECT status FROM rounds WHERE id = ?`, roundID).Scan(&rst)
	if rst != models.RoundEmAndamento {
		return ErrRoundInactive
	}
	t := fmtTime(now())
	var theoryID any
	txt := summary
	if summary == "" {
		txt = "Solução confirmada presencialmente"
	}
	tid := newID()
	theoryID = tid
	_, err = tx.Exec(`INSERT INTO theories(id, round_id, author_nickname, text, status, created_at, resolved_at, idempotency_key) VALUES(?,?,?,?,?,?,?,?)`,
		tid, roundID, "Presencial", txt, models.TheoryAceita, t, t, "spoken-"+tid)
	if err != nil {
		return err
	}
	res, err := tx.Exec(`UPDATE rounds SET status = ?, resolved_at = ?, accepted_theory_id = ?, accepted_theory_txt = ? WHERE id = ? AND status = ?`,
		models.RoundResolvida, t, theoryID, txt, roundID, models.RoundEmAndamento)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return ErrConflict
	}
	v, err := s.bump(tx, room.ID)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.notify(room.Code, v)
	return nil
}

func (s *Store) ReturnToLobby(code, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	room, err := s.requireNarrator(tx, code, sessionID)
	if err != nil {
		return err
	}
	if room.ActiveRoundID != nil {
		var st string
		_ = tx.QueryRow(`SELECT status FROM rounds WHERE id = ?`, *room.ActiveRoundID).Scan(&st)
		if st == models.RoundEmAndamento {
			return fmt.Errorf("%w: encerre a rodada antes", ErrConflict)
		}
	}
	_, err = tx.Exec(`UPDATE rooms SET active_round_id = NULL WHERE id = ?`, room.ID)
	if err != nil {
		return err
	}
	v, err := s.bump(tx, room.ID)
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	s.notify(room.Code, v)
	return nil
}

// WaitVersion blocks until version changes or ctx done.
func (s *Store) CurrentVersion(code string) (int64, error) {
	var v int64
	var last string
	err := s.db.QueryRow(`SELECT version, last_activity FROM rooms WHERE code = ?`, strings.ToUpper(code)).Scan(&v, &last)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, err
	}
	if now().Sub(parseTime(last)) > RoomTTL {
		return v, ErrExpired
	}
	return v, nil
}

func (s *Store) WithTx(ctx context.Context, fn func(*sql.Tx) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}
