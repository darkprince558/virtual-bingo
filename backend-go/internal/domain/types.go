package domain

import (
	"encoding/json"
	"time"
)

type User struct {
	ID              string
	ExternalSubject *string
	DisplayName     string
	Email           string
	Role            string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type WordSet struct {
	ID        string
	Name      string
	Status    string
	Source    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type WordSetWithWords struct {
	WordSet WordSet
	Words   []WordSetWord
}

type WordSetWord struct {
	ID        string
	WordSetID string
	Word      string
	SortOrder int
	IsActive  bool
	CreatedAt time.Time
}

type GameRun struct {
	ID                  string
	TemplateID          *string
	HostUserID          string
	WordSetID           *string
	Code                string
	Name                string
	Status              string
	ScheduledStartAt    *time.Time
	StartedAt           *time.Time
	EndedAt             *time.Time
	CurrentCalledWordID *string
	WinningPattern      *string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

type GameRunSummary struct {
	GameRun            GameRun
	AllowedPlayerCount int
	PlayerCount        int
}

type AllowedPlayer struct {
	ID          string
	GameRunID   string
	Email       string
	DisplayName string
	Source      string
	CreatedAt   time.Time
}

type Player struct {
	ID              string
	GameRunID       string
	UserID          *string
	Email           string
	DisplayName     string
	ConnectionState string
	State           string
	JoinedAt        time.Time
	LastSeenAt      time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type BingoCard struct {
	ID        string
	GameRunID string
	PlayerID  string
	Seed      string
	Cells     []BingoCardCell
	CreatedAt time.Time
}

type BingoCardCell struct {
	ID          string
	CardID      string
	RowIndex    int
	ColIndex    int
	Word        string
	IsFreeSpace bool
	MarkedAt    *time.Time
	CreatedAt   time.Time
}

type CalledWord struct {
	ID             string
	GameRunID      string
	WordSetWordID  *string
	Word           string
	CalledByUserID *string
	Sequence       int
	CalledAt       time.Time
	CreatedAt      time.Time
}

type BingoClaim struct {
	ID               string
	GameRunID        string
	PlayerID         string
	Pattern          string
	Status           string
	ValidationResult json.RawMessage
	ClaimedAt        time.Time
	ReviewedByUserID *string
	ReviewedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type Winner struct {
	ID          string
	GameRunID   string
	PlayerID    string
	ClaimID     *string
	Placement   int
	Pattern     string
	ConfirmedAt time.Time
	CreatedAt   time.Time
}

type GameSummary struct {
	GameRun         GameRun
	PlayerCount     int
	CalledWordCount int
	CurrentWord     *CalledWord
	Claims          []BingoClaim
	Winners         []Winner
	Players         []Player
	CalledWords     []CalledWord
	Status          string
}

type GameEvent struct {
	ID        string
	GameRunID string
	Type      string
	EntityID  *string
	Payload   json.RawMessage
	Sequence  int64
	CreatedAt time.Time
}

type ActivityEvent struct {
	ID          string
	GameRunID   string
	Type        string
	EntityType  *string
	EntityID    *string
	ActorUserID *string
	Payload     json.RawMessage
	Sequence    *int64
	CreatedAt   time.Time
}

type HostSnapshot struct {
	GameRun     GameRun
	Status      string
	CurrentWord *CalledWord
	Pattern     string
	PlayerCount int
	Players     []Player
	CalledWords []CalledWord
	Claims      []BingoClaim
	Winners     []Winner
}

type PlayerSnapshot struct {
	GameRun         GameRun
	Status          string
	CurrentWord     *CalledWord
	Pattern         string
	Player          Player
	Card            *BingoCard
	CalledWords     []CalledWord
	Claims          []BingoClaim
	Winners         []Winner
	ReconnectNotice *ReconnectNotice
}

type ReconnectNotice struct {
	LastSeenAt        time.Time
	MissedCalledWords []CalledWord
}
