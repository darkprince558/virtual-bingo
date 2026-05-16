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

type GameRunSettings struct {
	GameRunID                    string
	MarkingMode                  string
	AllowPlayerMarkingModeChoice bool
	ShowClaimReadiness           bool
	VoiceClaimMode               string
	VoiceClaimAutoplay           bool
	CallerMode                   string
	ThemeMode                    string
	ThemeID                      *string
	CreatedAt                    time.Time
	UpdatedAt                    time.Time
}

type ThemeTokens struct {
	Name          string         `json:"name"`
	Summary       string         `json:"summary"`
	Palette       map[string]any `json:"palette"`
	Icons         []string       `json:"icons"`
	Decorations   []string       `json:"decorations"`
	Motion        string         `json:"motion"`
	CallerTone    string         `json:"callerTone"`
	Accessibility map[string]any `json:"accessibility"`
}

type ContentGenerationJob struct {
	ID           string
	GameRunID    string
	JobType      string
	Status       string
	Provider     string
	ErrorMessage *string
	RetryCount   int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type GeneratedGameContent struct {
	ID                   string
	GameRunID            string
	GenerationJobID      *string
	Status               string
	Topic                string
	Summary              string
	GeneratedWords       []string
	CurrentWords         []string
	CallerStyle          *string
	ThemePrompt          *string
	ReviewWindowOpensAt  *time.Time
	ReviewWindowClosesAt *time.Time
	LockedAt             *time.Time
	LockedWordSetID      *string
	GenerationProvider   string
	GenerationError      *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

type GameRunContentReview struct {
	ID            string
	GameRunID     string
	ContentID     string
	ActorUserID   *string
	EditedTopic   *string
	EditedSummary *string
	EditedWords   []string
	CallerStyle   *string
	CreatedAt     time.Time
}

type PlayerPreferences struct {
	PlayerID    string
	MarkingMode *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
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
	Icon            *string
	AvatarColor     *string
	AvatarLabel     *string
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
	CallerAsset    *CallerAsset
}

type GameCallDeckItem struct {
	ID             string
	GameRunID      string
	WordSetWordID  *string
	Word           string
	Sequence       int
	ShuffleSeed    string
	ShuffleVersion string
	LockedAt       time.Time
	CalledWordID   *string
	CreatedAt      time.Time
}

type CallerAsset struct {
	ID             string
	GameRunID      string
	CallDeckItemID string
	Word           string
	Sequence       int
	Line           string
	AudioURL       *string
	StorageKey     *string
	VoiceName      *string
	Provider       string
	Status         string
	ErrorReason    *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
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
	GameRun            GameRun
	Settings           GameRunSettings
	Status             string
	CurrentWord        *CalledWord
	CurrentCallerAsset *CallerAsset
	AppliedTheme       *Theme
	Pattern            string
	PlayerCount        int
	Players            []Player
	CalledWords        []CalledWord
	Claims             []BingoClaim
	Winners            []Winner
}

type PlayerSnapshot struct {
	GameRun            GameRun
	Settings           GameRunSettings
	MarkingMode        string
	Status             string
	CurrentWord        *CalledWord
	CurrentCallerAsset *CallerAsset
	AppliedTheme       *Theme
	Pattern            string
	Player             Player
	Card               *BingoCard
	CalledWords        []CalledWord
	Claims             []BingoClaim
	Winners            []Winner
	ReconnectNotice    *ReconnectNotice
}

type ReconnectNotice struct {
	LastSeenAt        time.Time
	MissedCalledWords []CalledWord
}

type ClaimReadiness struct {
	Ready             bool
	SupportedPatterns []string
	ReadyPatterns     []string
	BestPattern       string
	MatchedCells      []BingoCardCell
	MissingCells      []BingoCardCell
	Reason            string
}

type DeliveryBatch struct {
	ID        string
	GameRunID string
	Channel   string
	Purpose   string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type DeliveryAttempt struct {
	ID              string
	BatchID         string
	GameRunID       string
	Channel         string
	Purpose         string
	RecipientEmail  string
	RecipientUserID *string
	Subject         string
	TemplateKey     string
	BodyPreview     string
	LinkURL         string
	GameCode        string
	Status          string
	ErrorReason     *string
	SentAt          *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type ThemeGenerationJob struct {
	ID           string
	GameRunID    *string
	Status       string
	Provider     string
	Prompt       string
	ErrorMessage *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Theme struct {
	ID               string
	GameRunID        *string
	GenerationJobID  *string
	Name             string
	Summary          string
	Tokens           ThemeTokens
	Status           string
	Provider         string
	CreatedByUserID  *string
	ApprovedByUserID *string
	ApprovedAt       *time.Time
	RejectedAt       *time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

type ThemeApproval struct {
	ID          string
	ThemeID     string
	GameRunID   *string
	ActorUserID *string
	Status      string
	CreatedAt   time.Time
}
