package domain

import "time"

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

type WordSetWord struct {
	ID        string
	WordSetID string
	Word      string
	SortOrder int
	IsActive  bool
	CreatedAt time.Time
}

type GameRun struct {
	ID               string
	TemplateID       *string
	HostUserID       string
	WordSetID        *string
	Code             string
	Name             string
	Status           string
	ScheduledStartAt *time.Time
	StartedAt        *time.Time
	EndedAt          *time.Time
	WinningPattern   *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
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
