package models

import "time"

const (
	ModePresencial = "presencial"
	ModeRemoto     = "remoto"

	RoleNarrador  = "narrador"
	RoleDetetive  = "detetive"
	RoleTelao     = "telao"

	RoundEmAndamento = "em_andamento"
	RoundResolvida   = "resolvida"

	QuestionPendente   = "pendente"
	QuestionRespondida = "respondida"

	AnswerSim         = "sim"
	AnswerNao         = "nao"
	AnswerIrrelevante = "irrelevante"

	TheoryPendente    = "pendente"
	TheoryContinuada  = "continuada"
	TheoryAceita      = "aceita"

	QuestionKindTyped = "digitada"
	QuestionKindOral  = "oral"
)

type Case struct {
	ID           string   `json:"id"`
	Title        string   `json:"title"`
	Theme        string   `json:"theme"`
	Difficulty   string   `json:"difficulty"`
	PublicText   string   `json:"publicText"`
	Solution     string   `json:"-"`
	Essentials   []string `json:"-"`
	Hints        []string `json:"-"`
	Illustration string   `json:"illustration,omitempty"`
}

type CasePublic struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Theme        string `json:"theme"`
	Difficulty   string `json:"difficulty"`
	PublicText   string `json:"publicText,omitempty"`
	Illustration string `json:"illustration,omitempty"`
}

type Room struct {
	ID               string
	Code             string
	Mode             string
	NarratorSession  string
	CreatedAt        time.Time
	LastActivity     time.Time
	Version          int64
	ActiveRoundID    *string
}

type Participant struct {
	ID        string
	RoomID    string
	SessionID string
	Nickname  string
	Role      string
	JoinedAt  time.Time
	LastSeen  time.Time
}

type Round struct {
	ID                string
	RoomID            string
	CaseID            string
	Status            string
	StartedAt         time.Time
	ResolvedAt        *time.Time
	AcceptedTheoryID  *string
	AcceptedTheoryTxt *string
	Number            int
}

type Question struct {
	ID             string     `json:"id"`
	RoundID        string     `json:"roundId"`
	AskerID        *string    `json:"askerId,omitempty"`
	AskerNickname  string     `json:"askerNickname"`
	Text           string     `json:"text"`
	Kind           string     `json:"kind"`
	Status         string     `json:"status"`
	Answer         *string    `json:"answer,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	AnsweredAt     *time.Time `json:"answeredAt,omitempty"`
	IdempotencyKey string     `json:"idempotencyKey"`
}

type Theory struct {
	ID             string     `json:"id"`
	RoundID        string     `json:"roundId"`
	AuthorID       *string    `json:"authorId,omitempty"`
	AuthorNickname string     `json:"authorNickname"`
	Text           string     `json:"text"`
	Status         string     `json:"status"`
	Guidance       *string    `json:"guidance,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	IdempotencyKey string     `json:"idempotencyKey"`
}

type ParticipantPublic struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

type QuestionPublic struct {
	ID            string  `json:"id"`
	AskerNickname string  `json:"askerNickname"`
	Text          string  `json:"text"`
	Kind          string  `json:"kind"`
	Status        string  `json:"status"`
	Answer        *string `json:"answer,omitempty"`
	CreatedAt     string  `json:"createdAt"`
	Mine          bool    `json:"mine,omitempty"`
}

type TheoryPublic struct {
	ID             string  `json:"id"`
	AuthorNickname string  `json:"authorNickname"`
	Text           string  `json:"text"`
	Status         string  `json:"status"`
	Guidance       *string `json:"guidance,omitempty"`
	CreatedAt      string  `json:"createdAt"`
	Mine           bool    `json:"mine,omitempty"`
}

type RoundPublic struct {
	ID               string          `json:"id"`
	Number           int             `json:"number"`
	Status           string          `json:"status"`
	Case             *CasePublic     `json:"case,omitempty"`
	StartedAt        string          `json:"startedAt"`
	ResolvedAt       *string         `json:"resolvedAt,omitempty"`
	Questions        []QuestionPublic `json:"questions"`
	Theories         []TheoryPublic  `json:"theories"`
	AcceptedTheory   *string         `json:"acceptedTheory,omitempty"`
	RevealedSolution *string         `json:"revealedSolution,omitempty"`
}

type RoomPublic struct {
	Code         string              `json:"code"`
	Mode         string              `json:"mode"`
	Version      int64               `json:"version"`
	Participants []ParticipantPublic `json:"participants"`
	Round        *RoundPublic        `json:"round,omitempty"`
	You          *YouInfo            `json:"you,omitempty"`
	InvitePath   string              `json:"invitePath"`
	TelaoPath    string              `json:"telaoPath"`
	Expired      bool                `json:"expired,omitempty"`
}

type YouInfo struct {
	ParticipantID string `json:"participantId"`
	Nickname      string `json:"nickname"`
	Role          string `json:"role"`
	IsNarrator    bool   `json:"isNarrator"`
}

type PrivatePayload struct {
	Solution   string   `json:"solution"`
	Essentials []string `json:"essentials"`
	Hints      []string `json:"hints"`
	CaseTitle  string   `json:"caseTitle"`
}
