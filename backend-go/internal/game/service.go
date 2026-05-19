package game

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/ai"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/audit"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/auth"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/bingo"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/clock"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/db"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/domain"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/events"
)

var (
	ErrForbidden  = errors.New("forbidden")
	ErrNotAllowed = errors.New("player is not allowed")
	ErrValidation = errors.New("validation error")
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")
)

type Store interface {
	UpsertUserFromPrincipal(context.Context, auth.Principal) (domain.User, error)
	CreateGameRun(context.Context, db.CreateGameRunParams) (domain.GameRun, error)
	GetGameRun(context.Context, string) (domain.GameRun, error)
	GetGameRunByCode(context.Context, string) (domain.GameRun, error)
	ListGameRuns(context.Context, db.ListGameRunsParams) ([]domain.GameRunSummary, error)
	UpdateGameRun(context.Context, db.UpdateGameRunParams) (domain.GameRun, error)
	UpdateGameRunStatus(context.Context, string, string, *time.Time, *time.Time) (domain.GameRun, error)
	TransitionGameRunStatus(context.Context, db.GameStatusTransitionParams) (domain.GameRun, error)
	CountAllowedPlayers(context.Context, string) (int, error)
	CountPlayers(context.Context, string) (int, error)
	AddAllowedPlayer(context.Context, db.AddAllowedPlayerParams) (domain.AllowedPlayer, error)
	BulkAddAllowedPlayers(context.Context, db.BulkAddAllowedPlayersParams) ([]domain.AllowedPlayer, error)
	UpdateAllowedPlayer(context.Context, db.UpdateAllowedPlayerParams) (domain.AllowedPlayer, error)
	DeleteAllowedPlayer(context.Context, string, string) (domain.AllowedPlayer, error)
	ListAllowedPlayers(context.Context, string) ([]domain.AllowedPlayer, error)
	GetAllowedPlayerByEmail(context.Context, string, string) (domain.AllowedPlayer, error)
	GetOrCreateGameSettings(context.Context, string) (domain.GameRunSettings, error)
	UpdateGameSettings(context.Context, db.UpdateGameSettingsParams) (domain.GameRunSettings, error)
	GetOrCreatePlayerPreferences(context.Context, string, string) (domain.PlayerPreferences, error)
	UpdatePlayerPreferences(context.Context, db.UpdatePlayerPreferencesParams) (domain.PlayerPreferences, error)
	GetEffectivePlayerMarkingMode(context.Context, string, string) (string, error)
	AutoMarkGame(context.Context, string) (db.AutoMarkRunResult, error)
	CreateContentGenerationJob(context.Context, db.CreateContentGenerationJobParams) (domain.ContentGenerationJob, error)
	UpdateContentGenerationJob(context.Context, db.UpdateContentGenerationJobParams) (domain.ContentGenerationJob, error)
	GetGeneratedGameContent(context.Context, string) (domain.GeneratedGameContent, error)
	UpsertGeneratedGameContent(context.Context, db.UpsertGeneratedGameContentParams) (domain.GeneratedGameContent, error)
	UpdateGeneratedGameContent(context.Context, db.UpdateGeneratedGameContentParams) (domain.GeneratedGameContent, error)
	LockGeneratedGameContent(context.Context, db.LockGeneratedGameContentParams) (db.LockGeneratedGameContentResult, error)
	CreateGameCallDeck(context.Context, db.CreateGameCallDeckParams) ([]domain.GameCallDeckItem, error)
	ListGameCallDeck(context.Context, string) ([]domain.GameCallDeckItem, error)
	CreateCalledWordFromDeck(context.Context, db.CreateCalledWordParams) (domain.CalledWord, error)
	UpsertCallerAsset(context.Context, db.CreateCallerAssetParams) (domain.CallerAsset, error)
	ListCallerAssets(context.Context, string) ([]domain.CallerAsset, error)
	CreateDeliveryBatch(context.Context, db.CreateDeliveryBatchParams) (domain.DeliveryBatch, []domain.DeliveryAttempt, error)
	ListDeliveryAttempts(context.Context, string) ([]domain.DeliveryAttempt, error)
	RetryDeliveryAttempt(context.Context, string) (domain.DeliveryAttempt, error)
	UpdatePlayerProfile(context.Context, db.UpdatePlayerProfileParams) (domain.Player, error)
	CreateThemeGenerationJob(context.Context, *string, string, string) (domain.ThemeGenerationJob, error)
	CreateTheme(context.Context, db.CreateThemeParams) (domain.Theme, error)
	GetTheme(context.Context, string) (domain.Theme, error)
	UpdateTheme(context.Context, string, domain.ThemeTokens, *string, *string) (domain.Theme, error)
	SetThemeApproval(context.Context, string, *string, bool) (domain.Theme, error)
	ApplyThemeToGame(context.Context, string, string, *string) (domain.GameRunSettings, error)
	CreatePlayer(context.Context, db.CreatePlayerParams) (domain.Player, error)
	GetPlayer(context.Context, string, string) (domain.Player, error)
	GetPlayerByGameRunAndEmail(context.Context, string, string) (domain.Player, error)
	UpdatePlayerConnectionState(context.Context, db.UpdatePlayerConnectionStateParams) (domain.Player, error)
	MarkStalePlayersDisconnected(context.Context, db.MarkStalePlayersDisconnectedParams) ([]domain.Player, error)
	ListWordSets(context.Context, bool) ([]domain.WordSet, error)
	GetWordSet(context.Context, string) (domain.WordSet, error)
	ListWordSetWords(context.Context, string) ([]domain.WordSetWord, error)
	ListWordSetWordsForManagement(context.Context, string) ([]domain.WordSetWord, error)
	CreateWordSet(context.Context, db.CreateWordSetParams) (domain.WordSetWithWords, error)
	UpdateWordSet(context.Context, db.UpdateWordSetParams) (domain.WordSet, error)
	CreateWordSetWord(context.Context, db.CreateWordSetWordParams) (domain.WordSetWord, error)
	UpdateWordSetWord(context.Context, db.UpdateWordSetWordParams) (domain.WordSetWord, error)
	DeleteWordSetWord(context.Context, string, string) (domain.WordSetWord, error)
	CreateCalledWord(context.Context, db.CreateCalledWordParams) (domain.CalledWord, error)
	ListCalledWords(context.Context, string) ([]domain.CalledWord, error)
	CreateCard(context.Context, db.CreateCardParams) (domain.BingoCard, error)
	GetPlayerCard(context.Context, string) (domain.BingoCard, error)
	SetCardCellMarked(context.Context, db.SetCardCellMarkedParams) (domain.BingoCardCell, error)
	CreateBingoClaim(context.Context, db.CreateBingoClaimParams) (domain.BingoClaim, error)
	UpdateBingoClaimValidation(context.Context, db.UpdateBingoClaimValidationParams) (domain.BingoClaim, error)
	SubmitBingoClaimTx(context.Context, db.SubmitBingoClaimTxParams) (db.SubmitBingoClaimTxResult, error)
	AcknowledgeBingoClaimTx(context.Context, db.AcknowledgeBingoClaimTxParams) (domain.GameEvent, error)
	GetBingoClaim(context.Context, string) (domain.BingoClaim, error)
	ListBingoClaims(context.Context, string) ([]domain.BingoClaim, error)
	CreateWinner(context.Context, db.CreateWinnerParams) (domain.Winner, error)
	ListWinners(context.Context, string) ([]domain.Winner, error)
	ListPlayers(context.Context, string) ([]domain.Player, error)
	CountWinners(context.Context, string) (int, error)
	ListGameEvents(context.Context, string, int64, int) ([]domain.GameEvent, error)
	ListActivityEvents(context.Context, string, int) ([]domain.ActivityEvent, error)
	RecordAuditEvent(context.Context, audit.Event) error
}

type ServiceConfig struct {
	Store         Store
	Authenticator auth.Authenticator
	Publisher     events.Publisher
	AuditLogger   audit.Logger
	Clock         clock.Clock
	AIClient      ai.Client
}

type Service struct {
	store         Store
	authenticator auth.Authenticator
	publisher     events.Publisher
	auditLogger   audit.Logger
	clock         clock.Clock
	aiClient      ai.Client
}

func NewService(config ServiceConfig) *Service {
	if config.Publisher == nil {
		config.Publisher = events.NoopPublisher{}
	}
	if config.AuditLogger == nil {
		config.AuditLogger = audit.NoopLogger{}
	}
	if config.Clock == nil {
		config.Clock = clock.SystemClock{}
	}
	if config.Authenticator == nil {
		config.Authenticator = auth.DevAuthenticator{Enabled: true}
	}
	if config.AIClient == nil {
		config.AIClient = ai.DisabledClient{}
	}

	return &Service{
		store:         config.Store,
		authenticator: config.Authenticator,
		publisher:     config.Publisher,
		auditLogger:   config.AuditLogger,
		clock:         config.Clock,
		aiClient:      config.AIClient,
	}
}

type GameRunWithCounts struct {
	GameRun            domain.GameRun
	AllowedPlayerCount int
}

type CreateGameRunInput struct {
	Name             string
	Code             string
	WordSetID        *string
	ScheduledStartAt *time.Time
	WinningPattern   *string
}

func (s *Service) CurrentUser(ctx context.Context, principal auth.Principal) (domain.User, error) {
	if strings.TrimSpace(principal.Email) == "" {
		return domain.User{}, auth.ErrUnauthenticated
	}

	return s.store.UpsertUserFromPrincipal(ctx, principal)
}

func (s *Service) CreateGameRun(ctx context.Context, principal auth.Principal, input CreateGameRunInput) (GameRunWithCounts, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return GameRunWithCounts{}, ErrForbidden
	}

	name := strings.TrimSpace(input.Name)
	if name == "" {
		return GameRunWithCounts{}, fmt.Errorf("%w: name is required", ErrValidation)
	}
	if input.WinningPattern != nil {
		pattern := bingo.NormalizePattern(*input.WinningPattern)
		if err := bingo.EnsureSupportedPattern(pattern); err != nil {
			return GameRunWithCounts{}, fmt.Errorf("%w: %v", ErrValidation, err)
		}
		input.WinningPattern = &pattern
	}

	user, err := s.store.UpsertUserFromPrincipal(ctx, principal)
	if err != nil {
		return GameRunWithCounts{}, err
	}

	code := strings.TrimSpace(input.Code)
	if code == "" {
		code = generateGameCode(name, s.clock.Now())
	}

	run, err := s.store.CreateGameRun(ctx, db.CreateGameRunParams{
		HostUserID:       user.ID,
		WordSetID:        input.WordSetID,
		Code:             strings.ToUpper(code),
		Name:             name,
		Status:           "lobby_open",
		ScheduledStartAt: input.ScheduledStartAt,
		WinningPattern:   input.WinningPattern,
	})
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return GameRunWithCounts{}, fmt.Errorf("%w: game code already exists", ErrConflict)
		}
		return GameRunWithCounts{}, err
	}

	s.emit(ctx, events.Event{Type: "game.created", EntityID: run.ID, Payload: map[string]any{"code": run.Code}})
	s.recordAudit(ctx, audit.Event{GameRunID: &run.ID, ActorUserID: &user.ID, EventType: "game.created", EntityType: "game_run", EntityID: &run.ID, Payload: map[string]any{"code": run.Code}})

	return GameRunWithCounts{GameRun: run}, nil
}

func (s *Service) GetGameRun(ctx context.Context, gameRunID string) (GameRunWithCounts, error) {
	run, err := s.store.GetGameRun(ctx, gameRunID)
	if err != nil {
		return GameRunWithCounts{}, mapStoreError(err)
	}

	count, err := s.store.CountAllowedPlayers(ctx, gameRunID)
	if err != nil {
		return GameRunWithCounts{}, err
	}

	return GameRunWithCounts{GameRun: run, AllowedPlayerCount: count}, nil
}

func (s *Service) GetGameRunByCode(ctx context.Context, code string) (GameRunWithCounts, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return GameRunWithCounts{}, fmt.Errorf("%w: code is required", ErrValidation)
	}

	run, err := s.store.GetGameRunByCode(ctx, code)
	if err != nil {
		return GameRunWithCounts{}, mapStoreError(err)
	}

	count, err := s.store.CountAllowedPlayers(ctx, run.ID)
	if err != nil {
		return GameRunWithCounts{}, err
	}

	return GameRunWithCounts{GameRun: run, AllowedPlayerCount: count}, nil
}

func (s *Service) ListGameRuns(ctx context.Context, principal auth.Principal, scope, status string) ([]domain.GameRunSummary, error) {
	scope = strings.ToLower(strings.TrimSpace(scope))
	if scope == "" {
		if auth.HasRole(principal, "admin", "host") {
			scope = "host"
		} else {
			scope = "player"
		}
	}
	switch scope {
	case "admin":
		if !auth.HasRole(principal, "admin") {
			return nil, ErrForbidden
		}
	case "host":
		if !auth.HasRole(principal, "admin", "host") {
			return nil, ErrForbidden
		}
	case "player":
	default:
		return nil, fmt.Errorf("%w: scope must be host, player, or admin", ErrValidation)
	}

	user, err := s.store.UpsertUserFromPrincipal(ctx, principal)
	if err != nil {
		return nil, err
	}

	return s.store.ListGameRuns(ctx, db.ListGameRunsParams{
		Scope:       scope,
		Status:      strings.TrimSpace(status),
		UserID:      user.ID,
		PlayerEmail: normalizeEmail(principal.Email),
	})
}

type UpdateGameRunInput struct {
	GameRunID        string
	Name             *string
	Code             *string
	WordSetID        *string
	ScheduledStartAt *time.Time
	WinningPattern   *string
}

func (s *Service) UpdateGameRun(ctx context.Context, principal auth.Principal, input UpdateGameRunInput) (GameRunWithCounts, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return GameRunWithCounts{}, ErrForbidden
	}

	user, err := s.store.UpsertUserFromPrincipal(ctx, principal)
	if err != nil {
		return GameRunWithCounts{}, err
	}
	run, err := s.store.GetGameRun(ctx, input.GameRunID)
	if err != nil {
		return GameRunWithCounts{}, mapStoreError(err)
	}
	if !auth.HasRole(principal, "admin") && run.HostUserID != user.ID {
		return GameRunWithCounts{}, ErrForbidden
	}

	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return GameRunWithCounts{}, fmt.Errorf("%w: name cannot be blank", ErrValidation)
		}
		input.Name = &name
	}
	if input.Code != nil {
		code := strings.ToUpper(strings.TrimSpace(*input.Code))
		if code == "" {
			return GameRunWithCounts{}, fmt.Errorf("%w: code cannot be blank", ErrValidation)
		}
		input.Code = &code
	}
	if input.WordSetID != nil {
		wordSetID := strings.TrimSpace(*input.WordSetID)
		if wordSetID == "" {
			return GameRunWithCounts{}, fmt.Errorf("%w: wordSetId cannot be blank", ErrValidation)
		}
		if _, err := s.store.GetWordSet(ctx, wordSetID); err != nil {
			return GameRunWithCounts{}, mapStoreError(err)
		}
		input.WordSetID = &wordSetID
	}
	if input.WinningPattern != nil {
		pattern := bingo.NormalizePattern(*input.WinningPattern)
		if err := bingo.EnsureSupportedPattern(pattern); err != nil {
			return GameRunWithCounts{}, fmt.Errorf("%w: %v", ErrValidation, err)
		}
		input.WinningPattern = &pattern
	}

	updated, err := s.store.UpdateGameRun(ctx, db.UpdateGameRunParams{
		GameRunID:        input.GameRunID,
		Name:             input.Name,
		Code:             input.Code,
		WordSetID:        input.WordSetID,
		ScheduledStartAt: input.ScheduledStartAt,
		WinningPattern:   input.WinningPattern,
		ActorUserID:      &user.ID,
	})
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return GameRunWithCounts{}, fmt.Errorf("%w: game can only be updated before it is live, paused, finished, or cancelled", ErrConflict)
		}
		return GameRunWithCounts{}, mapStoreError(err)
	}
	count, err := s.store.CountAllowedPlayers(ctx, input.GameRunID)
	if err != nil {
		return GameRunWithCounts{}, err
	}

	s.emit(ctx, events.Event{Type: "game.updated", EntityID: updated.ID, Payload: map[string]any{"gameRunId": updated.ID, "code": updated.Code}})
	return GameRunWithCounts{GameRun: updated, AllowedPlayerCount: count}, nil
}

func (s *Service) StartGame(ctx context.Context, principal auth.Principal, gameRunID string) (GameRunWithCounts, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return GameRunWithCounts{}, ErrForbidden
	}

	user, err := s.store.UpsertUserFromPrincipal(ctx, principal)
	if err != nil {
		return GameRunWithCounts{}, err
	}

	run, err := s.store.GetGameRun(ctx, gameRunID)
	if err != nil {
		return GameRunWithCounts{}, mapStoreError(err)
	}

	switch run.Status {
	case "live":
		count, err := s.store.CountAllowedPlayers(ctx, gameRunID)
		if err != nil {
			return GameRunWithCounts{}, err
		}
		return GameRunWithCounts{GameRun: run, AllowedPlayerCount: count}, nil
	case "draft", "scheduled", "invites_sent", "lobby_open":
	default:
		return GameRunWithCounts{}, fmt.Errorf("%w: game cannot start from status %s", ErrConflict, run.Status)
	}

	startedAt := s.clock.Now()
	run, err = s.store.TransitionGameRunStatus(ctx, db.GameStatusTransitionParams{
		GameRunID:              gameRunID,
		Status:                 "live",
		StartedAt:              &startedAt,
		ActorUserID:            &user.ID,
		EventType:              "game.started",
		AllowedCurrentStatuses: []string{"draft", "scheduled", "invites_sent", "lobby_open"},
		MarkJoinedPlayersLive:  true,
		Payload:                map[string]any{"status": "live"},
	})
	if err != nil {
		return GameRunWithCounts{}, mapStoreError(err)
	}

	s.emit(ctx, events.Event{Type: "game.started", EntityID: run.ID, Payload: map[string]any{"status": run.Status}})

	count, err := s.store.CountAllowedPlayers(ctx, gameRunID)
	if err != nil {
		return GameRunWithCounts{}, err
	}

	return GameRunWithCounts{GameRun: run, AllowedPlayerCount: count}, nil
}

func (s *Service) PauseGame(ctx context.Context, principal auth.Principal, gameRunID string) (GameRunWithCounts, error) {
	return s.transitionGame(ctx, principal, gameRunID, gameTransition{
		Status:        "paused",
		EventType:     "game.paused",
		AllowedFrom:   []string{"live"},
		ErrorTemplate: "game cannot pause from current status",
	})
}

func (s *Service) ResumeGame(ctx context.Context, principal auth.Principal, gameRunID string) (GameRunWithCounts, error) {
	return s.transitionGame(ctx, principal, gameRunID, gameTransition{
		Status:        "live",
		EventType:     "game.resumed",
		AllowedFrom:   []string{"paused"},
		ErrorTemplate: "game cannot resume from current status",
	})
}

func (s *Service) FinishGame(ctx context.Context, principal auth.Principal, gameRunID string) (GameRunWithCounts, error) {
	endedAt := s.clock.Now()
	return s.transitionGame(ctx, principal, gameRunID, gameTransition{
		Status:        "finished",
		EventType:     "game.finished",
		AllowedFrom:   []string{"live", "paused"},
		EndedAt:       &endedAt,
		ErrorTemplate: "game cannot finish from current status",
	})
}

func (s *Service) CancelGame(ctx context.Context, principal auth.Principal, gameRunID string) (GameRunWithCounts, error) {
	run, err := s.store.GetGameRun(ctx, gameRunID)
	if err != nil {
		return GameRunWithCounts{}, mapStoreError(err)
	}

	var endedAt *time.Time
	if run.StartedAt != nil {
		now := s.clock.Now()
		endedAt = &now
	}

	return s.transitionGame(ctx, principal, gameRunID, gameTransition{
		Status:        "cancelled",
		EventType:     "game.cancelled",
		AllowedFrom:   []string{"draft", "scheduled", "invites_sent", "lobby_open", "live", "paused"},
		EndedAt:       endedAt,
		ErrorTemplate: "game cannot cancel from current status",
	})
}

type gameTransition struct {
	Status        string
	EventType     string
	AllowedFrom   []string
	EndedAt       *time.Time
	ErrorTemplate string
}

func (s *Service) transitionGame(ctx context.Context, principal auth.Principal, gameRunID string, transition gameTransition) (GameRunWithCounts, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return GameRunWithCounts{}, ErrForbidden
	}

	user, err := s.store.UpsertUserFromPrincipal(ctx, principal)
	if err != nil {
		return GameRunWithCounts{}, err
	}

	run, err := s.store.TransitionGameRunStatus(ctx, db.GameStatusTransitionParams{
		GameRunID:              gameRunID,
		Status:                 transition.Status,
		EndedAt:                transition.EndedAt,
		ActorUserID:            &user.ID,
		EventType:              transition.EventType,
		AllowedCurrentStatuses: transition.AllowedFrom,
		Payload:                map[string]any{"status": transition.Status},
	})
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return GameRunWithCounts{}, fmt.Errorf("%w: %s", ErrConflict, transition.ErrorTemplate)
		}
		return GameRunWithCounts{}, mapStoreError(err)
	}

	s.emit(ctx, events.Event{Type: transition.EventType, EntityID: run.ID, Payload: map[string]any{"status": run.Status}})

	count, err := s.store.CountAllowedPlayers(ctx, gameRunID)
	if err != nil {
		return GameRunWithCounts{}, err
	}

	return GameRunWithCounts{GameRun: run, AllowedPlayerCount: count}, nil
}

func (s *Service) CallNextWord(ctx context.Context, principal auth.Principal, gameRunID string) (domain.CalledWord, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return domain.CalledWord{}, ErrForbidden
	}

	user, err := s.store.UpsertUserFromPrincipal(ctx, principal)
	if err != nil {
		return domain.CalledWord{}, err
	}

	run, err := s.store.GetGameRun(ctx, gameRunID)
	if err != nil {
		return domain.CalledWord{}, mapStoreError(err)
	}
	if run.Status != "live" {
		return domain.CalledWord{}, fmt.Errorf("%w: game must be live to call words", ErrConflict)
	}
	if run.WordSetID == nil {
		return domain.CalledWord{}, fmt.Errorf("%w: game has no word set", ErrValidation)
	}
	deck, err := s.store.ListGameCallDeck(ctx, gameRunID)
	if err != nil {
		return domain.CalledWord{}, mapStoreError(err)
	}
	if len(deck) > 0 {
		calledWord, err := s.store.CreateCalledWordFromDeck(ctx, db.CreateCalledWordParams{
			GameRunID:      gameRunID,
			CalledByUserID: &user.ID,
		})
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				return domain.CalledWord{}, fmt.Errorf("%w: locked call deck is exhausted", ErrConflict)
			}
			return domain.CalledWord{}, mapStoreError(err)
		}
		s.emit(ctx, events.Event{Type: "word.called", EntityID: calledWord.ID, Payload: calledWordPayload(gameRunID, calledWord)})
		s.recordAudit(ctx, audit.Event{GameRunID: &gameRunID, ActorUserID: &user.ID, EventType: "word.called", EntityType: "called_word", EntityID: &calledWord.ID, Payload: map[string]any{"word": calledWord.Word, "sequence": calledWord.Sequence, "deck": true}})
		return calledWord, nil
	}

	words, err := s.store.ListWordSetWords(ctx, *run.WordSetID)
	if err != nil {
		return domain.CalledWord{}, err
	}
	calledWords, err := s.store.ListCalledWords(ctx, gameRunID)
	if err != nil {
		return domain.CalledWord{}, err
	}

	called := make(map[string]struct{}, len(calledWords))
	for _, word := range calledWords {
		called[strings.ToLower(word.Word)] = struct{}{}
	}

	for _, word := range words {
		if _, ok := called[strings.ToLower(word.Word)]; ok {
			continue
		}

		calledWord, err := s.store.CreateCalledWord(ctx, db.CreateCalledWordParams{
			GameRunID:      gameRunID,
			WordSetWordID:  &word.ID,
			Word:           word.Word,
			CalledByUserID: &user.ID,
		})
		if err != nil {
			if errors.Is(err, db.ErrConflict) {
				return domain.CalledWord{}, fmt.Errorf("%w: word was already called", ErrConflict)
			}
			return domain.CalledWord{}, err
		}

		s.emit(ctx, events.Event{Type: "word.called", EntityID: calledWord.ID, Payload: map[string]any{"gameRunId": gameRunID, "word": calledWord.Word, "sequence": calledWord.Sequence}})
		s.recordAudit(ctx, audit.Event{GameRunID: &gameRunID, ActorUserID: &user.ID, EventType: "word.called", EntityType: "called_word", EntityID: &calledWord.ID, Payload: map[string]any{"word": calledWord.Word, "sequence": calledWord.Sequence}})
		return calledWord, nil
	}

	return domain.CalledWord{}, fmt.Errorf("%w: no uncalled words remain", ErrConflict)
}

func (s *Service) ListCalledWords(ctx context.Context, gameRunID string) ([]domain.CalledWord, error) {
	if _, err := s.store.GetGameRun(ctx, gameRunID); err != nil {
		return nil, mapStoreError(err)
	}

	return s.store.ListCalledWords(ctx, gameRunID)
}

type AddAllowedPlayerInput struct {
	GameRunID   string
	Email       string
	DisplayName string
	Source      string
}

func (s *Service) AddAllowedPlayer(ctx context.Context, principal auth.Principal, input AddAllowedPlayerInput) (domain.AllowedPlayer, error) {
	if _, err := s.authorizeHostMutation(ctx, principal, input.GameRunID); err != nil {
		return domain.AllowedPlayer{}, err
	}

	email := normalizeEmail(input.Email)
	displayName := strings.TrimSpace(input.DisplayName)
	if email == "" || displayName == "" {
		return domain.AllowedPlayer{}, fmt.Errorf("%w: email and displayName are required", ErrValidation)
	}

	player, err := s.store.AddAllowedPlayer(ctx, db.AddAllowedPlayerParams{
		GameRunID:   input.GameRunID,
		Email:       email,
		DisplayName: displayName,
		Source:      input.Source,
	})
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return domain.AllowedPlayer{}, fmt.Errorf("%w: allowed player already exists", ErrConflict)
		}
		return domain.AllowedPlayer{}, err
	}

	s.emit(ctx, events.Event{Type: "allowed_player.added", EntityID: player.ID, Payload: map[string]any{"gameRunId": input.GameRunID, "email": player.Email}})
	return player, nil
}

func (s *Service) BulkAddAllowedPlayers(ctx context.Context, principal auth.Principal, gameRunID string, rows []AddAllowedPlayerInput) ([]domain.AllowedPlayer, error) {
	if _, err := s.authorizeHostMutation(ctx, principal, gameRunID); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("%w: at least one allowed player is required", ErrValidation)
	}

	seen := make(map[string]struct{}, len(rows))
	params := make([]db.AddAllowedPlayerParams, 0, len(rows))
	for index, row := range rows {
		email := normalizeEmail(row.Email)
		displayName := strings.TrimSpace(row.DisplayName)
		if email == "" || displayName == "" {
			return nil, fmt.Errorf("%w: row %d requires email and displayName", ErrValidation, index+1)
		}
		if _, ok := seen[email]; ok {
			return nil, fmt.Errorf("%w: duplicate email %s in request", ErrConflict, email)
		}
		seen[email] = struct{}{}
		params = append(params, db.AddAllowedPlayerParams{
			GameRunID:   gameRunID,
			Email:       email,
			DisplayName: displayName,
			Source:      row.Source,
		})
	}

	players, err := s.store.BulkAddAllowedPlayers(ctx, db.BulkAddAllowedPlayersParams{
		GameRunID: gameRunID,
		Players:   params,
	})
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return nil, fmt.Errorf("%w: allowed player already exists", ErrConflict)
		}
		return nil, mapStoreError(err)
	}
	for _, player := range players {
		s.emit(ctx, events.Event{Type: "allowed_player.added", EntityID: player.ID, Payload: map[string]any{"gameRunId": gameRunID, "email": player.Email}})
	}

	return players, nil
}

type UpdateAllowedPlayerInput struct {
	GameRunID       string
	AllowedPlayerID string
	Email           *string
	DisplayName     *string
}

func (s *Service) UpdateAllowedPlayer(ctx context.Context, principal auth.Principal, input UpdateAllowedPlayerInput) (domain.AllowedPlayer, error) {
	if _, err := s.authorizeHostMutation(ctx, principal, input.GameRunID); err != nil {
		return domain.AllowedPlayer{}, err
	}
	if input.Email != nil {
		email := normalizeEmail(*input.Email)
		if email == "" {
			return domain.AllowedPlayer{}, fmt.Errorf("%w: email cannot be blank", ErrValidation)
		}
		input.Email = &email
	}
	if input.DisplayName != nil {
		displayName := strings.TrimSpace(*input.DisplayName)
		if displayName == "" {
			return domain.AllowedPlayer{}, fmt.Errorf("%w: displayName cannot be blank", ErrValidation)
		}
		input.DisplayName = &displayName
	}

	player, err := s.store.UpdateAllowedPlayer(ctx, db.UpdateAllowedPlayerParams{
		GameRunID:       input.GameRunID,
		AllowedPlayerID: input.AllowedPlayerID,
		Email:           input.Email,
		DisplayName:     input.DisplayName,
	})
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return domain.AllowedPlayer{}, fmt.Errorf("%w: allowed player already exists", ErrConflict)
		}
		return domain.AllowedPlayer{}, mapStoreError(err)
	}

	s.emit(ctx, events.Event{Type: "allowed_player.updated", EntityID: player.ID, Payload: map[string]any{"gameRunId": input.GameRunID, "email": player.Email}})
	return player, nil
}

func (s *Service) DeleteAllowedPlayer(ctx context.Context, principal auth.Principal, gameRunID, allowedPlayerID string) error {
	if _, err := s.authorizeHostMutation(ctx, principal, gameRunID); err != nil {
		return err
	}

	player, err := s.store.DeleteAllowedPlayer(ctx, gameRunID, allowedPlayerID)
	if err != nil {
		return mapStoreError(err)
	}
	s.emit(ctx, events.Event{Type: "allowed_player.deleted", EntityID: player.ID, Payload: map[string]any{"gameRunId": gameRunID, "email": player.Email}})
	return nil
}

func (s *Service) ListAllowedPlayers(ctx context.Context, gameRunID string) ([]domain.AllowedPlayer, error) {
	if _, err := s.store.GetGameRun(ctx, gameRunID); err != nil {
		return nil, mapStoreError(err)
	}

	return s.store.ListAllowedPlayers(ctx, gameRunID)
}

type UpdateGameSettingsInput struct {
	GameRunID                    string
	MarkingMode                  *string
	AllowPlayerMarkingModeChoice *bool
	ShowClaimReadiness           *bool
	VoiceClaimMode               *string
	VoiceClaimAutoplay           *bool
	CallerMode                   *string
	ThemeMode                    *string
	ThemeID                      *string
}

func (s *Service) GetGameSettings(ctx context.Context, principal auth.Principal, gameRunID string) (domain.GameRunSettings, error) {
	if _, err := s.authorizeHostMutation(ctx, principal, gameRunID); err != nil {
		return domain.GameRunSettings{}, err
	}

	settings, err := s.store.GetOrCreateGameSettings(ctx, gameRunID)
	if err != nil {
		return domain.GameRunSettings{}, mapStoreError(err)
	}

	return settings, nil
}

func (s *Service) UpdateGameSettings(ctx context.Context, principal auth.Principal, input UpdateGameSettingsInput) (domain.GameRunSettings, error) {
	user, err := s.authorizeHostMutation(ctx, principal, input.GameRunID)
	if err != nil {
		return domain.GameRunSettings{}, err
	}
	if input.MarkingMode != nil {
		markingMode, err := normalizeMarkingMode(*input.MarkingMode)
		if err != nil {
			return domain.GameRunSettings{}, err
		}
		input.MarkingMode = &markingMode
	}
	if input.VoiceClaimMode != nil {
		voiceClaimMode, err := normalizeChoice(*input.VoiceClaimMode, "voiceClaimMode", []string{"off", "optional", "required"})
		if err != nil {
			return domain.GameRunSettings{}, err
		}
		input.VoiceClaimMode = &voiceClaimMode
	}
	if input.CallerMode != nil {
		callerMode, err := normalizeChoice(*input.CallerMode, "callerMode", []string{"off", "text_only", "tts"})
		if err != nil {
			return domain.GameRunSettings{}, err
		}
		input.CallerMode = &callerMode
	}
	if input.ThemeMode != nil {
		themeMode, err := normalizeChoice(*input.ThemeMode, "themeMode", []string{"default", "manual", "ai_generated"})
		if err != nil {
			return domain.GameRunSettings{}, err
		}
		input.ThemeMode = &themeMode
	}
	if input.ThemeID != nil {
		themeID := strings.TrimSpace(*input.ThemeID)
		if themeID == "" {
			return domain.GameRunSettings{}, fmt.Errorf("%w: themeId cannot be blank", ErrValidation)
		}
		input.ThemeID = &themeID
	}

	settings, err := s.store.UpdateGameSettings(ctx, db.UpdateGameSettingsParams{
		GameRunID:                    input.GameRunID,
		MarkingMode:                  input.MarkingMode,
		AllowPlayerMarkingModeChoice: input.AllowPlayerMarkingModeChoice,
		ShowClaimReadiness:           input.ShowClaimReadiness,
		VoiceClaimMode:               input.VoiceClaimMode,
		VoiceClaimAutoplay:           input.VoiceClaimAutoplay,
		CallerMode:                   input.CallerMode,
		ThemeMode:                    input.ThemeMode,
		ThemeID:                      input.ThemeID,
		ActorUserID:                  &user.ID,
	})
	if err != nil {
		return domain.GameRunSettings{}, mapStoreError(err)
	}

	s.emit(ctx, events.Event{Type: "game.settings_updated", EntityID: input.GameRunID, Payload: map[string]any{"gameRunId": input.GameRunID, "markingMode": settings.MarkingMode}})
	return settings, nil
}

type UpdatePlayerPreferencesInput struct {
	GameRunID   string
	MarkingMode *string
}

func (s *Service) GetCurrentPlayerPreferences(ctx context.Context, principal auth.Principal, gameRunID string) (domain.PlayerPreferences, error) {
	player, err := s.resolveCurrentPlayer(ctx, principal, gameRunID)
	if err != nil {
		return domain.PlayerPreferences{}, err
	}

	preferences, err := s.store.GetOrCreatePlayerPreferences(ctx, gameRunID, player.ID)
	if err != nil {
		return domain.PlayerPreferences{}, mapStoreError(err)
	}

	return preferences, nil
}

func (s *Service) UpdateCurrentPlayerPreferences(ctx context.Context, principal auth.Principal, input UpdatePlayerPreferencesInput) (domain.PlayerPreferences, error) {
	user, err := s.store.UpsertUserFromPrincipal(ctx, principal)
	if err != nil {
		return domain.PlayerPreferences{}, err
	}
	player, err := s.resolveCurrentPlayer(ctx, principal, input.GameRunID)
	if err != nil {
		return domain.PlayerPreferences{}, err
	}
	if input.MarkingMode != nil {
		markingMode, err := normalizeMarkingMode(*input.MarkingMode)
		if err != nil {
			return domain.PlayerPreferences{}, err
		}
		input.MarkingMode = &markingMode
	}

	preferences, err := s.store.UpdatePlayerPreferences(ctx, db.UpdatePlayerPreferencesParams{
		GameRunID:   input.GameRunID,
		PlayerID:    player.ID,
		MarkingMode: input.MarkingMode,
		ActorUserID: &user.ID,
	})
	if err != nil {
		return domain.PlayerPreferences{}, mapStoreError(err)
	}

	s.emit(ctx, events.Event{Type: "player.preferences_updated", EntityID: player.ID, Payload: map[string]any{"gameRunId": input.GameRunID, "playerId": player.ID, "markingMode": preferences.MarkingMode}})
	return preferences, nil
}

func (s *Service) AutoMarkGame(ctx context.Context, principal auth.Principal, gameRunID string) (db.AutoMarkRunResult, error) {
	if _, err := s.authorizeHostMutation(ctx, principal, gameRunID); err != nil {
		return db.AutoMarkRunResult{}, err
	}

	result, err := s.store.AutoMarkGame(ctx, gameRunID)
	if err != nil {
		return db.AutoMarkRunResult{}, mapStoreError(err)
	}
	if result.CellsMarked > 0 {
		s.emit(ctx, events.Event{Type: "card.auto_marked", EntityID: gameRunID, Payload: map[string]any{"gameRunId": gameRunID, "playersMarked": result.PlayersMarked, "cellsMarked": result.CellsMarked, "mode": result.Mode}})
	}

	return result, nil
}

type UpdateGeneratedContentInput struct {
	GameRunID    string
	Topic        *string
	Summary      *string
	Words        []string
	CallerStyle  *string
	HasWordPatch bool
}

func (s *Service) PrepareGameContent(ctx context.Context, gameRunID string) (domain.GeneratedGameContent, error) {
	return s.prepareGameContent(ctx, gameRunID, nil)
}

func (s *Service) PrepareGameContentForHost(ctx context.Context, principal auth.Principal, gameRunID string) (domain.GeneratedGameContent, error) {
	user, err := s.authorizeHostMutation(ctx, principal, gameRunID)
	if err != nil {
		return domain.GeneratedGameContent{}, err
	}

	return s.prepareGameContent(ctx, gameRunID, &user.ID)
}

func (s *Service) prepareGameContent(ctx context.Context, gameRunID string, actorUserID *string) (domain.GeneratedGameContent, error) {
	run, err := s.store.GetGameRun(ctx, gameRunID)
	if err != nil {
		return domain.GeneratedGameContent{}, mapStoreError(err)
	}
	settings, err := s.store.GetOrCreateGameSettings(ctx, gameRunID)
	if err != nil {
		return domain.GeneratedGameContent{}, mapStoreError(err)
	}

	job, err := s.store.CreateContentGenerationJob(ctx, db.CreateContentGenerationJobParams{
		GameRunID: gameRunID,
		JobType:   "game_prep",
		Status:    "running",
		Provider:  "unknown",
	})
	if err != nil {
		return domain.GeneratedGameContent{}, mapStoreError(err)
	}

	output, err := s.aiClient.GenerateGamePrep(ctx, ai.GamePrepInput{
		GameRunID:   gameRunID,
		TopicPrompt: run.Name,
		WordCount:   75,
		Tone:        "fun",
		Audience:    "internal workplace team",
		Settings: map[string]string{
			"callerStyle": settings.CallerMode,
			"themeMode":   settings.ThemeMode,
		},
	})
	if err != nil {
		message := err.Error()
		_, _ = s.store.UpdateContentGenerationJob(ctx, db.UpdateContentGenerationJobParams{
			JobID:        job.ID,
			Status:       "failed",
			ErrorMessage: &message,
		})
		return domain.GeneratedGameContent{}, fmt.Errorf("%w: AI game prep failed: %v", ErrValidation, err)
	}

	now := s.clock.Now()
	reviewClosesAt := now.Add(30 * time.Minute)
	if run.ScheduledStartAt != nil {
		deadline := run.ScheduledStartAt.Add(-30 * time.Minute)
		if deadline.After(now) {
			reviewClosesAt = deadline
		}
	}
	callerStyle := stringPtrIfNotBlank(output.CallerStyle)
	themePrompt := stringPtrIfNotBlank(output.ThemePrompt)
	content, err := s.store.UpsertGeneratedGameContent(ctx, db.UpsertGeneratedGameContentParams{
		GameRunID:            gameRunID,
		GenerationJobID:      &job.ID,
		Status:               "generated",
		Topic:                output.Topic,
		Summary:              output.Summary,
		GeneratedWords:       output.Words,
		CurrentWords:         output.Words,
		CallerStyle:          callerStyle,
		ThemePrompt:          themePrompt,
		ReviewWindowOpensAt:  &now,
		ReviewWindowClosesAt: &reviewClosesAt,
		GenerationProvider:   output.Provider,
		ActorUserID:          actorUserID,
	})
	if err != nil {
		return domain.GeneratedGameContent{}, mapStoreError(err)
	}

	s.emit(ctx, events.Event{Type: "content.generated", EntityID: content.ID, Payload: map[string]any{"gameRunId": gameRunID, "wordCount": len(content.CurrentWords)}})
	return content, nil
}

func (s *Service) GetGeneratedGameContent(ctx context.Context, principal auth.Principal, gameRunID string) (domain.GeneratedGameContent, error) {
	if _, err := s.authorizeHostMutation(ctx, principal, gameRunID); err != nil {
		return domain.GeneratedGameContent{}, err
	}
	content, err := s.store.GetGeneratedGameContent(ctx, gameRunID)
	if err != nil {
		return domain.GeneratedGameContent{}, mapStoreError(err)
	}

	return content, nil
}

func (s *Service) UpdateGeneratedGameContent(ctx context.Context, principal auth.Principal, input UpdateGeneratedContentInput) (domain.GeneratedGameContent, error) {
	user, err := s.authorizeHostMutation(ctx, principal, input.GameRunID)
	if err != nil {
		return domain.GeneratedGameContent{}, err
	}
	if input.Topic != nil {
		topic := strings.TrimSpace(*input.Topic)
		if topic == "" {
			return domain.GeneratedGameContent{}, fmt.Errorf("%w: topic cannot be blank", ErrValidation)
		}
		input.Topic = &topic
	}
	if input.Summary != nil {
		summary := strings.TrimSpace(*input.Summary)
		if summary == "" {
			return domain.GeneratedGameContent{}, fmt.Errorf("%w: summary cannot be blank", ErrValidation)
		}
		input.Summary = &summary
	}
	if input.CallerStyle != nil {
		callerStyle := strings.TrimSpace(*input.CallerStyle)
		if callerStyle == "" {
			return domain.GeneratedGameContent{}, fmt.Errorf("%w: callerStyle cannot be blank", ErrValidation)
		}
		input.CallerStyle = &callerStyle
	}
	if input.HasWordPatch {
		words, err := normalizeContentWords(input.Words)
		if err != nil {
			return domain.GeneratedGameContent{}, err
		}
		input.Words = words
	}

	content, err := s.store.UpdateGeneratedGameContent(ctx, db.UpdateGeneratedGameContentParams{
		GameRunID:    input.GameRunID,
		Topic:        input.Topic,
		Summary:      input.Summary,
		Words:        input.Words,
		CallerStyle:  input.CallerStyle,
		ActorUserID:  &user.ID,
		HasWordPatch: input.HasWordPatch,
	})
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return domain.GeneratedGameContent{}, fmt.Errorf("%w: generated content is locked and cannot be edited", ErrConflict)
		}
		return domain.GeneratedGameContent{}, mapStoreError(err)
	}

	s.emit(ctx, events.Event{Type: "content.edited", EntityID: content.ID, Payload: map[string]any{"gameRunId": input.GameRunID, "wordCount": len(content.CurrentWords)}})
	return content, nil
}

func (s *Service) LockGameContent(ctx context.Context, gameRunID string) (domain.GeneratedGameContent, error) {
	content, _, err := s.lockGameContent(ctx, gameRunID, nil)
	return content, err
}

func (s *Service) LockGameContentForHost(ctx context.Context, principal auth.Principal, gameRunID string) (domain.GeneratedGameContent, error) {
	user, err := s.authorizeHostMutation(ctx, principal, gameRunID)
	if err != nil {
		return domain.GeneratedGameContent{}, err
	}
	content, _, err := s.lockGameContent(ctx, gameRunID, &user.ID)
	return content, err
}

func (s *Service) lockGameContent(ctx context.Context, gameRunID string, actorUserID *string) (domain.GeneratedGameContent, *domain.WordSetWithWords, error) {
	content, err := s.store.GetGeneratedGameContent(ctx, gameRunID)
	if err != nil {
		return domain.GeneratedGameContent{}, nil, mapStoreError(err)
	}
	if _, err := ai.NormalizeGamePrepOutput(ai.GamePrepOutput{
		Topic:   content.Topic,
		Summary: content.Summary,
		Words:   content.CurrentWords,
	}); err != nil {
		return domain.GeneratedGameContent{}, nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}

	result, err := s.store.LockGeneratedGameContent(ctx, db.LockGeneratedGameContentParams{
		GameRunID:   gameRunID,
		ActorUserID: actorUserID,
	})
	if err != nil {
		return domain.GeneratedGameContent{}, nil, mapStoreError(err)
	}
	if len(result.WordSet.Words) > 0 {
		seed := "game:" + gameRunID + ":content:" + result.Content.ID
		deckWords := BuildCallDeck(result.WordSet.Words, seed, CallDeckShuffleVersion)
		if _, err := s.store.CreateGameCallDeck(ctx, db.CreateGameCallDeckParams{
			GameRunID:      gameRunID,
			ShuffleSeed:    seed,
			ShuffleVersion: CallDeckShuffleVersion,
			Words:          deckWords,
		}); err != nil {
			return domain.GeneratedGameContent{}, nil, mapStoreError(err)
		}
	}

	s.emit(ctx, events.Event{Type: "content.locked", EntityID: result.Content.ID, Payload: map[string]any{"gameRunId": gameRunID, "wordSetId": result.Content.LockedWordSetID, "wordCount": len(result.Content.CurrentWords)}})
	return result.Content, &result.WordSet, nil
}

func (s *Service) GenerateCallerAssets(ctx context.Context, principal auth.Principal, gameRunID string) ([]domain.CallerAsset, error) {
	if _, err := s.authorizeHostMutation(ctx, principal, gameRunID); err != nil {
		return nil, err
	}
	deck, err := s.store.ListGameCallDeck(ctx, gameRunID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	if len(deck) == 0 {
		return nil, fmt.Errorf("%w: locked call deck is required before caller assets can be generated", ErrConflict)
	}
	s.emit(ctx, events.Event{Type: "caller.assets_generation_started", EntityID: gameRunID, Payload: map[string]any{"gameRunId": gameRunID, "items": len(deck)}})
	inputDeck := make([]ai.CallDeckItemInput, 0, len(deck))
	for _, item := range deck {
		inputDeck = append(inputDeck, ai.CallDeckItemInput{CallDeckItemID: item.ID, Word: item.Word, Sequence: item.Sequence})
	}
	output, err := s.aiClient.GenerateCallerAssetsBulk(ctx, ai.CallerAssetsBulkInput{
		GameRunID: gameRunID,
		VoiceName: "local-default",
		Tone:      "fun",
		Deck:      inputDeck,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: caller asset generation failed: %v", ErrValidation, err)
	}
	byDeckID := make(map[string]ai.CallerAssetOutput, len(output.Assets))
	for _, asset := range output.Assets {
		byDeckID[asset.CallDeckItemID] = asset
	}
	assets := make([]domain.CallerAsset, 0, len(deck))
	for _, item := range deck {
		assetOut, ok := byDeckID[item.ID]
		if !ok {
			assetOut = ai.CallerAssetOutput{CallDeckItemID: item.ID, Word: item.Word, Sequence: item.Sequence, Line: "Next word is " + item.Word + ".", Status: "fallback", Provider: output.Provider, ErrorReason: "missing_from_ai_response"}
		}
		status := normalizeCallerAssetStatus(assetOut.Status)
		errorReason := stringPtrIfNotBlank(assetOut.ErrorReason)
		if status == "failed" && assetOut.Line == "" {
			assetOut.Line = "Next word is " + item.Word + "."
		}
		audioURL := stringPtrIfNotBlank(assetOut.AudioURL)
		storageKey := stringPtrIfNotBlank(assetOut.StorageKey)
		voiceName := stringPtrIfNotBlank("local-default")
		asset, err := s.store.UpsertCallerAsset(ctx, db.CreateCallerAssetParams{
			GameRunID:      gameRunID,
			CallDeckItemID: item.ID,
			Word:           item.Word,
			Sequence:       item.Sequence,
			Line:           firstNonBlank(assetOut.Line, "Next word is "+item.Word+"."),
			AudioURL:       audioURL,
			StorageKey:     storageKey,
			VoiceName:      voiceName,
			Provider:       firstNonBlank(assetOut.Provider, output.Provider),
			Status:         status,
			ErrorReason:    errorReason,
		})
		if err != nil {
			return nil, mapStoreError(err)
		}
		eventType := "caller.audio_ready"
		if asset.Status == "failed" || asset.Status == "fallback" {
			eventType = "caller.failed"
		}
		s.emit(ctx, events.Event{Type: eventType, EntityID: asset.ID, Payload: map[string]any{"gameRunId": gameRunID, "word": asset.Word, "sequence": asset.Sequence, "status": asset.Status}})
		assets = append(assets, asset)
	}
	return assets, nil
}

func (s *Service) SendMockPlayerInvites(ctx context.Context, principal auth.Principal, gameRunID string) ([]domain.DeliveryAttempt, error) {
	if _, err := s.authorizeHostMutation(ctx, principal, gameRunID); err != nil {
		return nil, err
	}
	run, err := s.store.GetGameRun(ctx, gameRunID)
	if err != nil {
		return nil, mapStoreError(err)
	}
	allowed, err := s.store.ListAllowedPlayers(ctx, gameRunID)
	if err != nil {
		return nil, err
	}
	attempts := make([]db.CreateDeliveryAttemptParams, 0, len(allowed))
	for _, player := range allowed {
		link := "/join?code=" + run.Code
		attempts = append(attempts, db.CreateDeliveryAttemptParams{
			RecipientEmail: player.Email,
			Subject:        "You're invited to " + run.Name,
			TemplateKey:    "local.player_invite",
			BodyPreview:    "Join " + run.Name + " with code " + run.Code + ".",
			LinkURL:        link,
			GameCode:       run.Code,
			Status:         "sent",
		})
	}
	_, created, err := s.store.CreateDeliveryBatch(ctx, db.CreateDeliveryBatchParams{GameRunID: gameRunID, Channel: "email", Purpose: "player_invite", Attempts: attempts})
	if err != nil {
		return nil, mapStoreError(err)
	}
	return created, nil
}

func (s *Service) ListDeliveries(ctx context.Context, principal auth.Principal, gameRunID string) ([]domain.DeliveryAttempt, error) {
	if _, err := s.authorizeHostMutation(ctx, principal, gameRunID); err != nil {
		return nil, err
	}
	return s.store.ListDeliveryAttempts(ctx, gameRunID)
}

func (s *Service) RetryDelivery(ctx context.Context, principal auth.Principal, deliveryID string) (domain.DeliveryAttempt, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return domain.DeliveryAttempt{}, ErrForbidden
	}
	attempt, err := s.store.RetryDeliveryAttempt(ctx, deliveryID)
	if err != nil {
		return domain.DeliveryAttempt{}, mapStoreError(err)
	}
	return attempt, nil
}

func (s *Service) OpenLobby(ctx context.Context, principal auth.Principal, gameRunID string) (GameRunWithCounts, error) {
	return s.transitionGame(ctx, principal, gameRunID, gameTransition{
		Status:        "lobby_open",
		EventType:     "lobby.opened",
		AllowedFrom:   []string{"draft", "scheduled", "invites_sent"},
		ErrorTemplate: "lobby cannot open from current status",
	})
}

type UpdatePlayerProfileInput struct {
	GameRunID   string
	Icon        string
	AvatarColor string
	AvatarLabel string
}

func (s *Service) UpdateCurrentPlayerProfile(ctx context.Context, principal auth.Principal, input UpdatePlayerProfileInput) (domain.Player, error) {
	player, err := s.resolveCurrentPlayer(ctx, principal, input.GameRunID)
	if err != nil {
		return domain.Player{}, err
	}
	icon, color, label, err := normalizePlayerProfile(input.Icon, input.AvatarColor, input.AvatarLabel)
	if err != nil {
		return domain.Player{}, err
	}
	updated, err := s.store.UpdatePlayerProfile(ctx, db.UpdatePlayerProfileParams{
		GameRunID: input.GameRunID, PlayerID: player.ID, Icon: icon, AvatarColor: color, AvatarLabel: label,
	})
	if err != nil {
		return domain.Player{}, mapStoreError(err)
	}
	s.emit(ctx, events.Event{Type: "player.profile_updated", EntityID: updated.ID, Payload: map[string]any{"gameRunId": input.GameRunID, "playerId": updated.ID}})
	return updated, nil
}

type GenerateThemeInput struct {
	GameRunID *string
	Prompt    string
	Tone      string
}

func (s *Service) GenerateTheme(ctx context.Context, principal auth.Principal, input GenerateThemeInput) (domain.Theme, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return domain.Theme{}, ErrForbidden
	}
	user, err := s.store.UpsertUserFromPrincipal(ctx, principal)
	if err != nil {
		return domain.Theme{}, err
	}
	prompt := strings.TrimSpace(input.Prompt)
	if prompt == "" {
		return domain.Theme{}, fmt.Errorf("%w: prompt is required", ErrValidation)
	}
	if input.GameRunID != nil {
		if _, err := s.authorizeHostMutation(ctx, principal, *input.GameRunID); err != nil {
			return domain.Theme{}, err
		}
	}
	job, err := s.store.CreateThemeGenerationJob(ctx, input.GameRunID, prompt, "unknown")
	if err != nil {
		return domain.Theme{}, mapStoreError(err)
	}
	output, err := s.aiClient.GenerateTheme(ctx, ai.ThemeInput{GameRunID: stringValue(input.GameRunID), Prompt: prompt, Tone: input.Tone, AllowedAssets: ThemeAssetIDs()})
	if err != nil {
		return domain.Theme{}, fmt.Errorf("%w: theme generation failed: %v", ErrValidation, err)
	}
	tokens := domain.ThemeTokens{Name: output.Name, Summary: output.Summary, Palette: output.Palette, Icons: output.Icons, Decorations: output.Decorations, Motion: output.Motion, CallerTone: output.CallerTone, Accessibility: output.Accessibility}
	if err := validateThemeTokens(tokens); err != nil {
		return domain.Theme{}, err
	}
	theme, err := s.store.CreateTheme(ctx, db.CreateThemeParams{GameRunID: input.GameRunID, GenerationJobID: &job.ID, Name: output.Name, Summary: output.Summary, Tokens: tokens, Provider: output.Provider, CreatedByUserID: &user.ID})
	if err != nil {
		return domain.Theme{}, mapStoreError(err)
	}
	s.emit(ctx, events.Event{Type: "theme.generated", EntityID: theme.ID, Payload: map[string]any{"themeId": theme.ID, "name": theme.Name}})
	return theme, nil
}

func (s *Service) GetTheme(ctx context.Context, principal auth.Principal, themeID string) (domain.Theme, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return domain.Theme{}, ErrForbidden
	}
	return s.store.GetTheme(ctx, themeID)
}

func (s *Service) UpdateTheme(ctx context.Context, principal auth.Principal, themeID string, tokens domain.ThemeTokens, name, summary *string) (domain.Theme, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return domain.Theme{}, ErrForbidden
	}
	if name != nil {
		trimmed := strings.TrimSpace(*name)
		if trimmed == "" {
			return domain.Theme{}, fmt.Errorf("%w: name cannot be blank", ErrValidation)
		}
		name = &trimmed
		tokens.Name = trimmed
	}
	if summary != nil {
		trimmed := strings.TrimSpace(*summary)
		if trimmed == "" {
			return domain.Theme{}, fmt.Errorf("%w: summary cannot be blank", ErrValidation)
		}
		summary = &trimmed
		tokens.Summary = trimmed
	}
	if err := validateThemeTokens(tokens); err != nil {
		return domain.Theme{}, err
	}
	theme, err := s.store.UpdateTheme(ctx, themeID, tokens, name, summary)
	if err != nil {
		return domain.Theme{}, mapStoreError(err)
	}
	return theme, nil
}

func (s *Service) SetThemeApproval(ctx context.Context, principal auth.Principal, themeID string, approved bool) (domain.Theme, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return domain.Theme{}, ErrForbidden
	}
	user, err := s.store.UpsertUserFromPrincipal(ctx, principal)
	if err != nil {
		return domain.Theme{}, err
	}
	theme, err := s.store.SetThemeApproval(ctx, themeID, &user.ID, approved)
	if err != nil {
		return domain.Theme{}, mapStoreError(err)
	}
	eventType := "theme.approved"
	if !approved {
		eventType = "theme.rejected"
	}
	s.emit(ctx, events.Event{Type: eventType, EntityID: theme.ID, Payload: map[string]any{"themeId": theme.ID, "status": theme.Status}})
	return theme, nil
}

func (s *Service) ApplyThemeToGame(ctx context.Context, principal auth.Principal, gameRunID, themeID string) (domain.GameRunSettings, error) {
	user, err := s.authorizeHostMutation(ctx, principal, gameRunID)
	if err != nil {
		return domain.GameRunSettings{}, err
	}
	settings, err := s.store.ApplyThemeToGame(ctx, gameRunID, themeID, &user.ID)
	if err != nil {
		return domain.GameRunSettings{}, mapStoreError(err)
	}
	s.emit(ctx, events.Event{Type: "theme.applied", EntityID: themeID, Payload: map[string]any{"gameRunId": gameRunID, "themeId": themeID}})
	return settings, nil
}

func (s *Service) GetCurrentPlayerClaimReadiness(ctx context.Context, principal auth.Principal, gameRunID string) (domain.ClaimReadiness, error) {
	run, err := s.store.GetGameRun(ctx, gameRunID)
	if err != nil {
		return domain.ClaimReadiness{}, mapStoreError(err)
	}
	player, err := s.resolveCurrentPlayer(ctx, principal, gameRunID)
	if err != nil {
		return domain.ClaimReadiness{}, err
	}
	card, err := s.store.GetPlayerCard(ctx, player.ID)
	if err != nil {
		return domain.ClaimReadiness{}, mapStoreError(err)
	}
	if card.GameRunID != gameRunID {
		return domain.ClaimReadiness{}, ErrNotFound
	}
	calledWords, err := s.store.ListCalledWords(ctx, gameRunID)
	if err != nil {
		return domain.ClaimReadiness{}, err
	}

	supportedPatterns := []string{winningPatternOrDefault(run)}
	cellByID := make(map[string]domain.BingoCardCell, len(card.Cells))
	for _, cell := range card.Cells {
		cellByID[cell.ID] = cell
	}

	readiness := domain.ClaimReadiness{
		SupportedPatterns: supportedPatterns,
		ReadyPatterns:     []string{},
		MatchedCells:      []domain.BingoCardCell{},
		MissingCells:      []domain.BingoCardCell{},
		Reason:            "not_ready",
	}
	bestMissingCount := 26
	for _, pattern := range supportedPatterns {
		validation := bingo.Validate(bingo.ValidationInput{
			GameRunID:       gameRunID,
			ClaimGameRunID:  run.ID,
			PlayerGameRunID: player.GameRunID,
			CardGameRunID:   card.GameRunID,
			Pattern:         pattern,
			Cells:           bingoCellsFromDomain(card.Cells),
			CalledWords:     calledWordStrings(calledWords),
		})
		if validation.Valid {
			readiness.Ready = true
			readiness.ReadyPatterns = append(readiness.ReadyPatterns, pattern)
		}
		if validation.Valid || len(validation.MissingCells) < bestMissingCount {
			bestMissingCount = len(validation.MissingCells)
			readiness.BestPattern = pattern
			readiness.MatchedCells = domainCellsFromBingo(validation.MatchedCells, cellByID)
			readiness.MissingCells = domainCellsFromBingo(validation.MissingCells, cellByID)
			readiness.Reason = validation.Reason
		}
	}
	if readiness.Ready && readiness.Reason == "" {
		readiness.Reason = "ready_to_claim"
	}

	return readiness, nil
}

func (s *Service) ListWordSets(ctx context.Context, principal auth.Principal) ([]domain.WordSetWithWords, error) {
	approvedOnly := !auth.HasRole(principal, "admin", "host")
	wordSets, err := s.store.ListWordSets(ctx, approvedOnly)
	if err != nil {
		return nil, err
	}

	result := make([]domain.WordSetWithWords, 0, len(wordSets))
	for _, wordSet := range wordSets {
		words, err := s.store.ListWordSetWordsForManagement(ctx, wordSet.ID)
		if err != nil {
			return nil, err
		}
		if approvedOnly && wordSet.Status != "approved" {
			continue
		}
		result = append(result, domain.WordSetWithWords{WordSet: wordSet, Words: words})
	}

	return result, nil
}

func (s *Service) GetWordSet(ctx context.Context, principal auth.Principal, wordSetID string) (domain.WordSetWithWords, error) {
	wordSet, err := s.store.GetWordSet(ctx, wordSetID)
	if err != nil {
		return domain.WordSetWithWords{}, mapStoreError(err)
	}
	if !auth.HasRole(principal, "admin", "host") && wordSet.Status != "approved" {
		return domain.WordSetWithWords{}, ErrForbidden
	}
	words, err := s.store.ListWordSetWordsForManagement(ctx, wordSet.ID)
	if err != nil {
		return domain.WordSetWithWords{}, err
	}

	return domain.WordSetWithWords{WordSet: wordSet, Words: words}, nil
}

type WordSetWordInput struct {
	Word      string
	SortOrder *int
	IsActive  *bool
}

type CreateWordSetInput struct {
	Name   string
	Status string
	Source string
	Words  []WordSetWordInput
}

func (s *Service) CreateWordSet(ctx context.Context, principal auth.Principal, input CreateWordSetInput) (domain.WordSetWithWords, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return domain.WordSetWithWords{}, ErrForbidden
	}
	user, err := s.store.UpsertUserFromPrincipal(ctx, principal)
	if err != nil {
		return domain.WordSetWithWords{}, err
	}

	name, status, source, err := normalizeWordSetFields(input.Name, input.Status, input.Source)
	if err != nil {
		return domain.WordSetWithWords{}, err
	}
	words, err := normalizeWordInputs("", input.Words)
	if err != nil {
		return domain.WordSetWithWords{}, err
	}

	wordSet, err := s.store.CreateWordSet(ctx, db.CreateWordSetParams{
		Name:            name,
		Status:          status,
		Source:          source,
		CreatedByUserID: &user.ID,
		Words:           words,
	})
	if err != nil {
		return domain.WordSetWithWords{}, mapStoreError(err)
	}

	return wordSet, nil
}

type UpdateWordSetInput struct {
	WordSetID string
	Name      *string
	Status    *string
	Source    *string
}

func (s *Service) UpdateWordSet(ctx context.Context, principal auth.Principal, input UpdateWordSetInput) (domain.WordSetWithWords, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return domain.WordSetWithWords{}, ErrForbidden
	}
	if input.Name != nil {
		name := strings.TrimSpace(*input.Name)
		if name == "" {
			return domain.WordSetWithWords{}, fmt.Errorf("%w: name cannot be blank", ErrValidation)
		}
		input.Name = &name
	}
	if input.Status != nil {
		status := strings.ToLower(strings.TrimSpace(*input.Status))
		if err := ensureWordSetStatus(status); err != nil {
			return domain.WordSetWithWords{}, err
		}
		input.Status = &status
	}
	if input.Source != nil {
		source := strings.ToLower(strings.TrimSpace(*input.Source))
		if err := ensureManualWordSetSource(source); err != nil {
			return domain.WordSetWithWords{}, err
		}
		input.Source = &source
	}

	wordSet, err := s.store.UpdateWordSet(ctx, db.UpdateWordSetParams{
		WordSetID: input.WordSetID,
		Name:      input.Name,
		Status:    input.Status,
		Source:    input.Source,
	})
	if err != nil {
		return domain.WordSetWithWords{}, mapStoreError(err)
	}
	words, err := s.store.ListWordSetWordsForManagement(ctx, wordSet.ID)
	if err != nil {
		return domain.WordSetWithWords{}, err
	}

	return domain.WordSetWithWords{WordSet: wordSet, Words: words}, nil
}

func (s *Service) CreateWordSetWord(ctx context.Context, principal auth.Principal, wordSetID string, input WordSetWordInput) (domain.WordSetWord, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return domain.WordSetWord{}, ErrForbidden
	}
	words, err := normalizeWordInputs(wordSetID, []WordSetWordInput{input})
	if err != nil {
		return domain.WordSetWord{}, err
	}

	word, err := s.store.CreateWordSetWord(ctx, words[0])
	if err != nil {
		return domain.WordSetWord{}, mapStoreError(err)
	}

	return word, nil
}

type UpdateWordSetWordInput struct {
	WordSetID string
	WordID    string
	Word      *string
	SortOrder *int
	IsActive  *bool
}

func (s *Service) UpdateWordSetWord(ctx context.Context, principal auth.Principal, input UpdateWordSetWordInput) (domain.WordSetWord, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return domain.WordSetWord{}, ErrForbidden
	}
	if input.Word != nil {
		word := strings.TrimSpace(*input.Word)
		if word == "" {
			return domain.WordSetWord{}, fmt.Errorf("%w: word cannot be blank", ErrValidation)
		}
		input.Word = &word
	}
	if input.SortOrder != nil && *input.SortOrder <= 0 {
		return domain.WordSetWord{}, fmt.Errorf("%w: sortOrder must be positive", ErrValidation)
	}

	word, err := s.store.UpdateWordSetWord(ctx, db.UpdateWordSetWordParams{
		WordSetID: input.WordSetID,
		WordID:    input.WordID,
		Word:      input.Word,
		SortOrder: input.SortOrder,
		IsActive:  input.IsActive,
	})
	if err != nil {
		return domain.WordSetWord{}, mapStoreError(err)
	}

	return word, nil
}

func (s *Service) DeleteWordSetWord(ctx context.Context, principal auth.Principal, wordSetID, wordID string) error {
	if !auth.HasRole(principal, "admin", "host") {
		return ErrForbidden
	}
	if _, err := s.store.DeleteWordSetWord(ctx, wordSetID, wordID); err != nil {
		return mapStoreError(err)
	}

	return nil
}

type JoinPlayerInput struct {
	GameRunID   string
	Email       string
	DisplayName string
}

type PlayerConnectionUpdate struct {
	Player          domain.Player
	ReconnectNotice *domain.ReconnectNotice
}

func (s *Service) JoinPlayer(ctx context.Context, input JoinPlayerInput) (PlayerConnectionUpdate, error) {
	email := normalizeEmail(input.Email)
	displayName := strings.TrimSpace(input.DisplayName)
	if email == "" || displayName == "" {
		return PlayerConnectionUpdate{}, fmt.Errorf("%w: email and displayName are required", ErrValidation)
	}

	if _, err := s.store.GetGameRun(ctx, input.GameRunID); err != nil {
		return PlayerConnectionUpdate{}, mapStoreError(err)
	}

	if _, err := s.store.GetAllowedPlayerByEmail(ctx, input.GameRunID, email); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return PlayerConnectionUpdate{}, ErrNotAllowed
		}
		return PlayerConnectionUpdate{}, err
	}

	existing, err := s.store.GetPlayerByGameRunAndEmail(ctx, input.GameRunID, email)
	if err == nil {
		notice, err := s.reconnectNotice(ctx, input.GameRunID, existing, existing.ConnectionState != "online")
		if err != nil {
			return PlayerConnectionUpdate{}, err
		}
		updated, err := s.store.UpdatePlayerConnectionState(ctx, db.UpdatePlayerConnectionStateParams{
			GameRunID:       input.GameRunID,
			PlayerID:        existing.ID,
			ConnectionState: "online",
			EventType:       "player.reconnected",
		})
		if err != nil {
			return PlayerConnectionUpdate{}, mapStoreError(err)
		}
		s.emit(ctx, events.Event{Type: "player.reconnected", EntityID: updated.ID, Payload: reconnectPayload(input.GameRunID, updated.ID, updated.ConnectionState, notice)})
		return PlayerConnectionUpdate{Player: updated, ReconnectNotice: notice}, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return PlayerConnectionUpdate{}, err
	}

	player, err := s.store.CreatePlayer(ctx, db.CreatePlayerParams{
		GameRunID:       input.GameRunID,
		Email:           email,
		DisplayName:     displayName,
		ConnectionState: "online",
		State:           "joined",
	})
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			player, err := s.store.GetPlayerByGameRunAndEmail(ctx, input.GameRunID, email)
			return PlayerConnectionUpdate{Player: player}, err
		}
		return PlayerConnectionUpdate{}, err
	}

	s.emit(ctx, events.Event{Type: "player.joined", EntityID: player.ID, Payload: map[string]any{"gameRunId": input.GameRunID, "email": player.Email}})
	return PlayerConnectionUpdate{Player: player}, nil
}

func (s *Service) HeartbeatPlayer(ctx context.Context, principal auth.Principal, gameRunID, playerID string) (PlayerConnectionUpdate, error) {
	player, err := s.store.GetPlayer(ctx, gameRunID, playerID)
	if err != nil {
		return PlayerConnectionUpdate{}, mapStoreError(err)
	}
	if !canActForPlayer(principal, player) {
		return PlayerConnectionUpdate{}, ErrForbidden
	}

	eventType := ""
	if player.ConnectionState != "online" {
		eventType = "player.reconnected"
	}
	notice, err := s.reconnectNotice(ctx, gameRunID, player, eventType != "")
	if err != nil {
		return PlayerConnectionUpdate{}, err
	}
	updated, err := s.store.UpdatePlayerConnectionState(ctx, db.UpdatePlayerConnectionStateParams{
		GameRunID:       gameRunID,
		PlayerID:        playerID,
		ConnectionState: "online",
		EventType:       eventType,
	})
	if err != nil {
		return PlayerConnectionUpdate{}, mapStoreError(err)
	}
	if eventType != "" {
		s.emit(ctx, events.Event{Type: eventType, EntityID: updated.ID, Payload: reconnectPayload(gameRunID, updated.ID, updated.ConnectionState, notice)})
	}

	return PlayerConnectionUpdate{Player: updated, ReconnectNotice: notice}, nil
}

func (s *Service) AssignPlayerCard(ctx context.Context, gameRunID, playerID string) (domain.BingoCard, error) {
	if _, err := s.store.GetGameRun(ctx, gameRunID); err != nil {
		return domain.BingoCard{}, mapStoreError(err)
	}

	existing, err := s.store.GetPlayerCard(ctx, playerID)
	if err == nil {
		if existing.GameRunID != gameRunID {
			return domain.BingoCard{}, ErrNotFound
		}
		return existing, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return domain.BingoCard{}, err
	}

	run, err := s.store.GetGameRun(ctx, gameRunID)
	if err != nil {
		return domain.BingoCard{}, mapStoreError(err)
	}
	if run.WordSetID == nil {
		return domain.BingoCard{}, fmt.Errorf("%w: game has no word set", ErrValidation)
	}

	words, err := s.store.ListWordSetWords(ctx, *run.WordSetID)
	if err != nil {
		return domain.BingoCard{}, err
	}
	if len(words) < 24 {
		return domain.BingoCard{}, fmt.Errorf("%w: at least 24 active word set words are required", ErrValidation)
	}

	seed := gameRunID + ":" + playerID
	card, err := s.store.CreateCard(ctx, db.CreateCardParams{
		GameRunID: gameRunID,
		PlayerID:  playerID,
		Seed:      seed,
		Cells:     makeCardCells(seed, words),
	})
	if err != nil {
		if errors.Is(err, db.ErrConflict) {
			return s.store.GetPlayerCard(ctx, playerID)
		}
		return domain.BingoCard{}, err
	}

	s.emit(ctx, events.Event{Type: "card.assigned", EntityID: card.ID, Payload: map[string]any{"gameRunId": gameRunID, "playerId": playerID}})
	return card, nil
}

func (s *Service) AssignCurrentPlayerCard(ctx context.Context, principal auth.Principal, gameRunID string) (domain.BingoCard, error) {
	player, err := s.resolveCurrentPlayer(ctx, principal, gameRunID)
	if err != nil {
		return domain.BingoCard{}, err
	}

	return s.AssignPlayerCard(ctx, gameRunID, player.ID)
}

func (s *Service) GetPlayerCard(ctx context.Context, gameRunID, playerID string) (domain.BingoCard, error) {
	card, err := s.store.GetPlayerCard(ctx, playerID)
	if err != nil {
		return domain.BingoCard{}, mapStoreError(err)
	}
	if card.GameRunID != gameRunID {
		return domain.BingoCard{}, ErrNotFound
	}

	return card, nil
}

func (s *Service) GetCurrentPlayerCard(ctx context.Context, principal auth.Principal, gameRunID string) (domain.BingoCard, error) {
	player, err := s.resolveCurrentPlayer(ctx, principal, gameRunID)
	if err != nil {
		return domain.BingoCard{}, err
	}

	return s.GetPlayerCard(ctx, gameRunID, player.ID)
}

type MarkCardCellInput struct {
	GameRunID string
	PlayerID  string
	CellID    string
	Marked    bool
}

func (s *Service) MarkCardCell(ctx context.Context, input MarkCardCellInput) (domain.BingoCardCell, error) {
	if _, err := s.store.GetGameRun(ctx, input.GameRunID); err != nil {
		return domain.BingoCardCell{}, mapStoreError(err)
	}

	cell, err := s.store.SetCardCellMarked(ctx, db.SetCardCellMarkedParams{
		GameRunID: input.GameRunID,
		PlayerID:  input.PlayerID,
		CellID:    input.CellID,
		Marked:    input.Marked,
	})
	if err != nil {
		return domain.BingoCardCell{}, mapStoreError(err)
	}

	s.emit(ctx, events.Event{Type: "card.cell_marked", EntityID: cell.ID, Payload: map[string]any{"gameRunId": input.GameRunID, "playerId": input.PlayerID, "marked": input.Marked}})
	return cell, nil
}

func (s *Service) MarkCurrentPlayerCardCell(ctx context.Context, principal auth.Principal, input MarkCardCellInput) (domain.BingoCardCell, error) {
	player, err := s.resolveCurrentPlayer(ctx, principal, input.GameRunID)
	if err != nil {
		return domain.BingoCardCell{}, err
	}
	input.PlayerID = player.ID

	return s.MarkCardCell(ctx, input)
}

type SubmitBingoClaimInput struct {
	GameRunID string
	PlayerID  string
	Pattern   string
}

type BingoClaimResult struct {
	Claim  domain.BingoClaim
	Winner *domain.Winner
}

func (s *Service) SubmitBingoClaim(ctx context.Context, input SubmitBingoClaimInput) (BingoClaimResult, error) {
	pattern := bingo.NormalizePattern(input.Pattern)
	if err := bingo.EnsureSupportedPattern(pattern); err != nil {
		return BingoClaimResult{}, fmt.Errorf("%w: %v", ErrValidation, err)
	}

	run, err := s.store.GetGameRun(ctx, input.GameRunID)
	if err != nil {
		return BingoClaimResult{}, mapStoreError(err)
	}
	if err := enforceAllowedPattern(run, pattern); err != nil {
		return BingoClaimResult{}, err
	}

	txResult, err := s.store.SubmitBingoClaimTx(ctx, db.SubmitBingoClaimTxParams{
		GameRunID: input.GameRunID,
		PlayerID:  input.PlayerID,
		Pattern:   pattern,
		Validate: func(data db.ClaimValidationData) (db.ClaimValidationDecision, error) {
			if data.GameRun.Status != "live" && data.GameRun.Status != "paused" {
				return db.ClaimValidationDecision{}, fmt.Errorf("%w: game must be live or paused to submit a claim", ErrConflict)
			}
			if err := enforceAllowedPattern(data.GameRun, pattern); err != nil {
				return db.ClaimValidationDecision{}, err
			}

			validation := bingo.Validate(bingo.ValidationInput{
				GameRunID:       input.GameRunID,
				ClaimGameRunID:  data.GameRun.ID,
				PlayerGameRunID: data.Player.GameRunID,
				CardGameRunID:   data.Card.GameRunID,
				Pattern:         pattern,
				Cells:           bingoCellsFromDomain(data.Card.Cells),
				CalledWords:     calledWordStrings(data.CalledWords),
			})
			validationJSON, err := json.Marshal(validation)
			if err != nil {
				return db.ClaimValidationDecision{}, fmt.Errorf("marshal claim validation: %w", err)
			}

			status := "invalid"
			if validation.Valid {
				status = "confirmed"
			}
			return db.ClaimValidationDecision{Status: status, ValidationResult: validationJSON, Valid: validation.Valid}, nil
		},
	})
	if err != nil {
		return BingoClaimResult{}, mapStoreError(err)
	}
	for _, event := range txResult.Events {
		s.emit(ctx, event)
	}

	return BingoClaimResult{Claim: txResult.Claim, Winner: txResult.Winner}, nil
}

func (s *Service) ListBingoClaims(ctx context.Context, gameRunID string) ([]domain.BingoClaim, error) {
	if _, err := s.store.GetGameRun(ctx, gameRunID); err != nil {
		return nil, mapStoreError(err)
	}

	return s.store.ListBingoClaims(ctx, gameRunID)
}

func (s *Service) GetBingoClaim(ctx context.Context, principal auth.Principal, gameRunID, claimID string) (domain.BingoClaim, error) {
	claim, err := s.store.GetBingoClaim(ctx, claimID)
	if err != nil {
		return domain.BingoClaim{}, mapStoreError(err)
	}
	if claim.GameRunID != gameRunID {
		return domain.BingoClaim{}, ErrNotFound
	}
	if auth.HasRole(principal, "admin", "host") {
		return claim, nil
	}
	player, err := s.store.GetPlayer(ctx, gameRunID, claim.PlayerID)
	if err != nil {
		return domain.BingoClaim{}, mapStoreError(err)
	}
	if !strings.EqualFold(player.Email, principal.Email) {
		return domain.BingoClaim{}, ErrForbidden
	}

	return claim, nil
}

type AcknowledgeBingoClaimInput struct {
	GameRunID string
	ClaimID   string
	Decision  string
	Note      string
}

func (s *Service) AcknowledgeBingoClaim(ctx context.Context, principal auth.Principal, input AcknowledgeBingoClaimInput) (domain.ActivityEvent, error) {
	user, err := s.authorizeHostMutation(ctx, principal, input.GameRunID)
	if err != nil {
		return domain.ActivityEvent{}, err
	}

	decision := strings.ToLower(strings.TrimSpace(input.Decision))
	if decision != "approve" && decision != "reject" {
		return domain.ActivityEvent{}, fmt.Errorf("%w: decision must be approve or reject", ErrValidation)
	}

	note := strings.TrimSpace(input.Note)
	if len(note) > 240 {
		return domain.ActivityEvent{}, fmt.Errorf("%w: note must be 240 characters or fewer", ErrValidation)
	}

	claim, err := s.store.GetBingoClaim(ctx, input.ClaimID)
	if err != nil {
		return domain.ActivityEvent{}, mapStoreError(err)
	}
	if claim.GameRunID != input.GameRunID {
		return domain.ActivityEvent{}, ErrNotFound
	}
	if decision == "approve" && claim.Status != "confirmed" {
		return domain.ActivityEvent{}, fmt.Errorf("%w: only confirmed claims can be acknowledged as valid", ErrConflict)
	}
	if decision == "reject" && claim.Status != "invalid" {
		return domain.ActivityEvent{}, fmt.Errorf("%w: only invalid claims can be acknowledged as rejected", ErrConflict)
	}

	event, err := s.store.AcknowledgeBingoClaimTx(ctx, db.AcknowledgeBingoClaimTxParams{
		GameRunID:   input.GameRunID,
		ClaimID:     input.ClaimID,
		Status:      claim.Status,
		Decision:    decision,
		Note:        note,
		ActorUserID: &user.ID,
	})
	if err != nil {
		return domain.ActivityEvent{}, mapStoreError(err)
	}
	s.emit(ctx, events.Event{
		Type:     event.Type,
		EntityID: input.ClaimID,
		Payload:  map[string]any{"gameRunId": event.GameRunID, "claimId": input.ClaimID, "decision": decision},
	})

	sequence := event.Sequence
	return domain.ActivityEvent{
		ID:          event.ID,
		GameRunID:   event.GameRunID,
		Type:        event.Type,
		EntityType:  stringPtr("bingo_claim"),
		EntityID:    event.EntityID,
		ActorUserID: &user.ID,
		Payload:     event.Payload,
		Sequence:    &sequence,
		CreatedAt:   event.CreatedAt,
	}, nil
}

func (s *Service) GetGameSummary(ctx context.Context, gameRunID string) (domain.GameSummary, error) {
	run, err := s.store.GetGameRun(ctx, gameRunID)
	if err != nil {
		return domain.GameSummary{}, mapStoreError(err)
	}

	playerCount, err := s.store.CountPlayers(ctx, gameRunID)
	if err != nil {
		return domain.GameSummary{}, err
	}
	calledWords, err := s.store.ListCalledWords(ctx, gameRunID)
	if err != nil {
		return domain.GameSummary{}, err
	}
	claims, err := s.store.ListBingoClaims(ctx, gameRunID)
	if err != nil {
		return domain.GameSummary{}, err
	}
	winners, err := s.store.ListWinners(ctx, gameRunID)
	if err != nil {
		return domain.GameSummary{}, err
	}
	players, err := s.store.ListPlayers(ctx, gameRunID)
	if err != nil {
		return domain.GameSummary{}, err
	}

	var currentWord *domain.CalledWord
	if len(calledWords) > 0 {
		word := calledWords[len(calledWords)-1]
		currentWord = &word
	}

	return domain.GameSummary{
		GameRun:         run,
		PlayerCount:     playerCount,
		CalledWordCount: len(calledWords),
		CurrentWord:     currentWord,
		Claims:          claims,
		Winners:         winners,
		Players:         players,
		CalledWords:     calledWords,
		Status:          run.Status,
	}, nil
}

func (s *Service) GetHostSnapshot(ctx context.Context, principal auth.Principal, gameRunID string) (domain.HostSnapshot, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return domain.HostSnapshot{}, ErrForbidden
	}

	summary, err := s.GetGameSummary(ctx, gameRunID)
	if err != nil {
		return domain.HostSnapshot{}, err
	}
	settings, err := s.store.GetOrCreateGameSettings(ctx, gameRunID)
	if err != nil {
		return domain.HostSnapshot{}, mapStoreError(err)
	}
	currentAsset, theme := s.snapshotExtras(ctx, gameRunID, settings, summary.CurrentWord)

	return domain.HostSnapshot{
		GameRun:            summary.GameRun,
		Settings:           settings,
		Status:             summary.Status,
		CurrentWord:        summary.CurrentWord,
		CurrentCallerAsset: currentAsset,
		AppliedTheme:       theme,
		Pattern:            winningPatternOrDefault(summary.GameRun),
		PlayerCount:        summary.PlayerCount,
		Players:            summary.Players,
		CalledWords:        summary.CalledWords,
		Claims:             summary.Claims,
		Winners:            summary.Winners,
	}, nil
}

func (s *Service) GetPlayerSnapshot(ctx context.Context, principal auth.Principal, gameRunID, playerID string) (domain.PlayerSnapshot, error) {
	run, err := s.store.GetGameRun(ctx, gameRunID)
	if err != nil {
		return domain.PlayerSnapshot{}, mapStoreError(err)
	}

	player, err := s.store.GetPlayer(ctx, gameRunID, playerID)
	if err != nil {
		return domain.PlayerSnapshot{}, mapStoreError(err)
	}
	if !canActForPlayer(principal, player) {
		return domain.PlayerSnapshot{}, ErrForbidden
	}
	eventType := ""
	if player.ConnectionState != "online" {
		eventType = "player.reconnected"
	}
	notice, err := s.reconnectNotice(ctx, gameRunID, player, eventType != "")
	if err != nil {
		return domain.PlayerSnapshot{}, err
	}
	player, err = s.store.UpdatePlayerConnectionState(ctx, db.UpdatePlayerConnectionStateParams{
		GameRunID:       gameRunID,
		PlayerID:        playerID,
		ConnectionState: "online",
		EventType:       eventType,
	})
	if err != nil {
		return domain.PlayerSnapshot{}, mapStoreError(err)
	}
	if eventType != "" {
		s.emit(ctx, events.Event{Type: eventType, EntityID: player.ID, Payload: reconnectPayload(gameRunID, player.ID, player.ConnectionState, notice)})
	}

	card, err := s.store.GetPlayerCard(ctx, playerID)
	if err != nil && !errors.Is(err, db.ErrNotFound) {
		return domain.PlayerSnapshot{}, mapStoreError(err)
	}
	var cardPtr *domain.BingoCard
	if err == nil {
		if card.GameRunID != gameRunID {
			return domain.PlayerSnapshot{}, ErrNotFound
		}
		cardPtr = &card
	}

	calledWords, err := s.store.ListCalledWords(ctx, gameRunID)
	if err != nil {
		return domain.PlayerSnapshot{}, err
	}
	claims, err := s.store.ListBingoClaims(ctx, gameRunID)
	if err != nil {
		return domain.PlayerSnapshot{}, err
	}
	playerClaims := make([]domain.BingoClaim, 0)
	for _, claim := range claims {
		if claim.PlayerID == playerID {
			playerClaims = append(playerClaims, claim)
		}
	}
	winners, err := s.store.ListWinners(ctx, gameRunID)
	if err != nil {
		return domain.PlayerSnapshot{}, err
	}
	settings, err := s.store.GetOrCreateGameSettings(ctx, gameRunID)
	if err != nil {
		return domain.PlayerSnapshot{}, mapStoreError(err)
	}
	markingMode, err := s.store.GetEffectivePlayerMarkingMode(ctx, gameRunID, playerID)
	if err != nil {
		return domain.PlayerSnapshot{}, mapStoreError(err)
	}

	var currentWord *domain.CalledWord
	if len(calledWords) > 0 {
		word := calledWords[len(calledWords)-1]
		currentWord = &word
	}
	currentAsset, theme := s.snapshotExtras(ctx, gameRunID, settings, currentWord)

	return domain.PlayerSnapshot{
		GameRun:            run,
		Settings:           settings,
		MarkingMode:        markingMode,
		Status:             run.Status,
		CurrentWord:        currentWord,
		CurrentCallerAsset: currentAsset,
		AppliedTheme:       theme,
		Pattern:            winningPatternOrDefault(run),
		Player:             player,
		Card:               cardPtr,
		CalledWords:        calledWords,
		Claims:             playerClaims,
		Winners:            winners,
		ReconnectNotice:    notice,
	}, nil
}

func (s *Service) GetCurrentPlayerSnapshot(ctx context.Context, principal auth.Principal, gameRunID string) (domain.PlayerSnapshot, error) {
	player, err := s.resolveCurrentPlayer(ctx, principal, gameRunID)
	if err != nil {
		return domain.PlayerSnapshot{}, err
	}

	return s.GetPlayerSnapshot(ctx, principal, gameRunID, player.ID)
}

func (s *Service) HeartbeatCurrentPlayer(ctx context.Context, principal auth.Principal, gameRunID string) (PlayerConnectionUpdate, error) {
	player, err := s.resolveCurrentPlayer(ctx, principal, gameRunID)
	if err != nil {
		return PlayerConnectionUpdate{}, err
	}

	return s.HeartbeatPlayer(ctx, principal, gameRunID, player.ID)
}

func (s *Service) SweepStalePlayerConnections(ctx context.Context, timeout time.Duration, limit int) ([]domain.Player, error) {
	if timeout <= 0 {
		return nil, nil
	}
	if limit <= 0 {
		limit = 100
	}

	cutoff := s.clock.Now().Add(-timeout)
	players, err := s.store.MarkStalePlayersDisconnected(ctx, db.MarkStalePlayersDisconnectedParams{
		Cutoff: cutoff,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}
	for _, player := range players {
		s.emit(ctx, events.Event{Type: "player.disconnected", EntityID: player.ID, Payload: map[string]any{
			"gameRunId":       player.GameRunID,
			"playerId":        player.ID,
			"connectionState": player.ConnectionState,
			"lastSeenAt":      player.LastSeenAt,
		}})
	}

	return players, nil
}

func (s *Service) ListGameEvents(ctx context.Context, gameRunID string, afterSequence int64, limit int) ([]domain.GameEvent, error) {
	if _, err := s.store.GetGameRun(ctx, gameRunID); err != nil {
		return nil, mapStoreError(err)
	}

	return s.store.ListGameEvents(ctx, gameRunID, afterSequence, limit)
}

func (s *Service) ListActivityEvents(ctx context.Context, principal auth.Principal, gameRunID string, limit int) ([]domain.ActivityEvent, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return nil, ErrForbidden
	}
	if _, err := s.store.GetGameRun(ctx, gameRunID); err != nil {
		return nil, mapStoreError(err)
	}

	return s.store.ListActivityEvents(ctx, gameRunID, limit)
}

func (s *Service) Authenticate(r *http.Request) (auth.Principal, error) {
	return s.authenticator.Authenticate(r)
}

func canActForPlayer(principal auth.Principal, player domain.Player) bool {
	return auth.HasRole(principal, "admin", "host") || strings.EqualFold(principal.Email, player.Email)
}

func (s *Service) resolveCurrentPlayer(ctx context.Context, principal auth.Principal, gameRunID string) (domain.Player, error) {
	email := normalizeEmail(principal.Email)
	if email == "" {
		return domain.Player{}, auth.ErrUnauthenticated
	}
	player, err := s.store.GetPlayerByGameRunAndEmail(ctx, gameRunID, email)
	if err != nil {
		return domain.Player{}, mapStoreError(err)
	}

	return player, nil
}

func (s *Service) authorizeHostMutation(ctx context.Context, principal auth.Principal, gameRunID string) (domain.User, error) {
	if !auth.HasRole(principal, "admin", "host") {
		return domain.User{}, ErrForbidden
	}
	user, err := s.store.UpsertUserFromPrincipal(ctx, principal)
	if err != nil {
		return domain.User{}, err
	}
	run, err := s.store.GetGameRun(ctx, gameRunID)
	if err != nil {
		return domain.User{}, mapStoreError(err)
	}
	if !auth.HasRole(principal, "admin") && run.HostUserID != user.ID {
		return domain.User{}, ErrForbidden
	}

	return user, nil
}

func normalizeWordSetFields(name, status, source string) (string, string, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", "", "", fmt.Errorf("%w: name is required", ErrValidation)
	}
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		status = "draft"
	}
	if err := ensureWordSetStatus(status); err != nil {
		return "", "", "", err
	}
	source = strings.ToLower(strings.TrimSpace(source))
	if source == "" {
		source = "manual"
	}
	if err := ensureManualWordSetSource(source); err != nil {
		return "", "", "", err
	}

	return name, status, source, nil
}

func ensureWordSetStatus(status string) error {
	switch status {
	case "draft", "approved", "archived":
		return nil
	default:
		return fmt.Errorf("%w: status must be draft, approved, or archived", ErrValidation)
	}
}

func ensureManualWordSetSource(source string) error {
	switch source {
	case "manual", "seed":
		return nil
	default:
		return fmt.Errorf("%w: source must be manual or seed", ErrValidation)
	}
}

func normalizeWordInputs(wordSetID string, inputs []WordSetWordInput) ([]db.CreateWordSetWordParams, error) {
	words := make([]db.CreateWordSetWordParams, 0, len(inputs))
	seenSortOrders := make(map[int]struct{}, len(inputs))
	for index, input := range inputs {
		word := strings.TrimSpace(input.Word)
		if word == "" {
			return nil, fmt.Errorf("%w: word %d cannot be blank", ErrValidation, index+1)
		}
		sortOrder := index + 1
		if input.SortOrder != nil {
			if *input.SortOrder <= 0 {
				return nil, fmt.Errorf("%w: sortOrder must be positive", ErrValidation)
			}
			sortOrder = *input.SortOrder
		}
		if _, ok := seenSortOrders[sortOrder]; ok {
			return nil, fmt.Errorf("%w: duplicate sortOrder %d in request", ErrConflict, sortOrder)
		}
		seenSortOrders[sortOrder] = struct{}{}
		isActive := true
		if input.IsActive != nil {
			isActive = *input.IsActive
		}
		words = append(words, db.CreateWordSetWordParams{
			WordSetID: wordSetID,
			Word:      word,
			SortOrder: sortOrder,
			IsActive:  isActive,
		})
	}

	return words, nil
}

func normalizeContentWords(inputs []string) ([]string, error) {
	output, err := ai.NormalizeGamePrepOutput(ai.GamePrepOutput{
		Topic:   "content validation",
		Summary: "content validation",
		Words:   inputs,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrValidation, err)
	}

	return output.Words, nil
}

func normalizeCallerAssetStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "ready", "failed", "fallback", "pending":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return "fallback"
	}
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func stringPtrIfNotBlank(value string) *string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	return &value
}

func stringPtr(value string) *string {
	return &value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func normalizePlayerProfile(icon, color, label string) (string, string, string, error) {
	icon = strings.ToLower(strings.TrimSpace(icon))
	color = strings.ToUpper(strings.TrimSpace(color))
	label = strings.TrimSpace(label)
	allowedIcons := map[string]struct{}{"star": {}, "sparkles": {}, "briefcase": {}, "rocket": {}, "trophy": {}, "coffee": {}, "smile": {}, "bolt": {}}
	if _, ok := allowedIcons[icon]; !ok {
		return "", "", "", fmt.Errorf("%w: icon must be one of star, sparkles, briefcase, rocket, trophy, coffee, smile, bolt", ErrValidation)
	}
	if len(color) != 7 || color[0] != '#' {
		return "", "", "", fmt.Errorf("%w: avatarColor must be a hex color like #1F7A8C", ErrValidation)
	}
	for _, r := range color[1:] {
		if (r < '0' || r > '9') && (r < 'A' || r > 'F') {
			return "", "", "", fmt.Errorf("%w: avatarColor must be a hex color like #1F7A8C", ErrValidation)
		}
	}
	if label == "" || len(label) > 4 {
		return "", "", "", fmt.Errorf("%w: avatarLabel must be 1 to 4 characters", ErrValidation)
	}
	return icon, color, label, nil
}

func ThemeAssetIDs() []string {
	return []string{"sparkles", "briefcase", "rocket", "trophy", "coffee", "star", "confetti", "subtle", "none"}
}

func validateThemeTokens(tokens domain.ThemeTokens) error {
	if strings.TrimSpace(tokens.Name) == "" || strings.TrimSpace(tokens.Summary) == "" {
		return fmt.Errorf("%w: theme name and summary are required", ErrValidation)
	}
	if len(tokens.Palette) == 0 {
		return fmt.Errorf("%w: theme palette is required", ErrValidation)
	}
	allowed := make(map[string]struct{})
	for _, id := range ThemeAssetIDs() {
		allowed[id] = struct{}{}
	}
	for _, icon := range tokens.Icons {
		if _, ok := allowed[strings.ToLower(strings.TrimSpace(icon))]; !ok {
			return fmt.Errorf("%w: theme icon %q is not approved", ErrValidation, icon)
		}
	}
	for _, decoration := range tokens.Decorations {
		if _, ok := allowed[strings.ToLower(strings.TrimSpace(decoration))]; !ok {
			return fmt.Errorf("%w: theme decoration %q is not approved", ErrValidation, decoration)
		}
	}
	payload, _ := json.Marshal(tokens)
	value := strings.ToLower(string(payload))
	if strings.Contains(value, "<script") || strings.Contains(value, "javascript:") || strings.Contains(value, "http://") || strings.Contains(value, "https://") || strings.Contains(value, "url(") {
		return fmt.Errorf("%w: theme tokens cannot include scripts or external asset URLs", ErrValidation)
	}
	return nil
}

func (s *Service) reconnectNotice(ctx context.Context, gameRunID string, player domain.Player, reconnected bool) (*domain.ReconnectNotice, error) {
	if !reconnected {
		return nil, nil
	}

	calledWords, err := s.store.ListCalledWords(ctx, gameRunID)
	if err != nil {
		return nil, err
	}
	missed := make([]domain.CalledWord, 0)
	for _, calledWord := range calledWords {
		if calledWord.CalledAt.After(player.LastSeenAt) {
			missed = append(missed, calledWord)
		}
	}

	return &domain.ReconnectNotice{
		LastSeenAt:        player.LastSeenAt,
		MissedCalledWords: missed,
	}, nil
}

func reconnectPayload(gameRunID, playerID, connectionState string, notice *domain.ReconnectNotice) map[string]any {
	payload := map[string]any{
		"gameRunId":       gameRunID,
		"playerId":        playerID,
		"connectionState": connectionState,
	}
	if notice != nil {
		payload["lastSeenAt"] = notice.LastSeenAt
		payload["missedCalledWordCount"] = len(notice.MissedCalledWords)
	}

	return payload
}

func calledWordPayload(gameRunID string, calledWord domain.CalledWord) map[string]any {
	payload := map[string]any{"gameRunId": gameRunID, "word": calledWord.Word, "sequence": calledWord.Sequence}
	if calledWord.CallerAsset != nil {
		payload["callerAssetStatus"] = calledWord.CallerAsset.Status
		payload["callerLine"] = calledWord.CallerAsset.Line
		payload["callerAudioUrl"] = calledWord.CallerAsset.AudioURL
	}
	return payload
}

func (s *Service) snapshotExtras(ctx context.Context, gameRunID string, settings domain.GameRunSettings, currentWord *domain.CalledWord) (*domain.CallerAsset, *domain.Theme) {
	var currentAsset *domain.CallerAsset
	if currentWord != nil {
		assets, err := s.store.ListCallerAssets(ctx, gameRunID)
		if err == nil {
			for _, asset := range assets {
				if asset.Sequence == currentWord.Sequence {
					copy := asset
					currentAsset = &copy
					break
				}
			}
		}
	}
	var theme *domain.Theme
	if settings.ThemeID != nil {
		found, err := s.store.GetTheme(ctx, *settings.ThemeID)
		if err == nil {
			theme = &found
		}
	}
	return currentAsset, theme
}

func makeCardCells(seed string, words []domain.WordSetWord) []db.CreateCardCellParams {
	shuffled := append([]domain.WordSetWord(nil), words...)
	random := rand.New(rand.NewSource(seedInt64(seed)))
	random.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	cells := make([]db.CreateCardCellParams, 0, 25)
	wordIndex := 0
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			if row == 2 && col == 2 {
				cells = append(cells, db.CreateCardCellParams{
					RowIndex:    row,
					ColIndex:    col,
					Word:        "FREE",
					IsFreeSpace: true,
				})
				continue
			}

			cells = append(cells, db.CreateCardCellParams{
				RowIndex: row,
				ColIndex: col,
				Word:     shuffled[wordIndex].Word,
			})
			wordIndex++
		}
	}

	return cells
}

func seedInt64(seed string) int64 {
	hash := sha256.Sum256([]byte(seed))
	return int64(binary.BigEndian.Uint64(hash[:8]))
}

func enforceAllowedPattern(run domain.GameRun, pattern string) error {
	if run.WinningPattern == nil || strings.TrimSpace(*run.WinningPattern) == "" {
		if pattern != bingo.PatternSingleLine {
			return fmt.Errorf("%w: game default winning pattern is %s", ErrValidation, bingo.PatternSingleLine)
		}
		return nil
	}

	allowed := bingo.NormalizePattern(*run.WinningPattern)
	if pattern != allowed {
		return fmt.Errorf("%w: game winning pattern is %s", ErrValidation, allowed)
	}

	return nil
}

func winningPatternOrDefault(run domain.GameRun) string {
	if run.WinningPattern == nil || strings.TrimSpace(*run.WinningPattern) == "" {
		return bingo.PatternSingleLine
	}

	return bingo.NormalizePattern(*run.WinningPattern)
}

func normalizeMarkingMode(value string) (string, error) {
	return normalizeChoice(value, "markingMode", []string{"manual", "assist", "auto_mark"})
}

func normalizeChoice(value, field string, allowed []string) (string, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	for _, candidate := range allowed {
		if normalized == candidate {
			return normalized, nil
		}
	}

	return "", fmt.Errorf("%w: %s must be one of %s", ErrValidation, field, strings.Join(allowed, ", "))
}

func bingoCellsFromDomain(cells []domain.BingoCardCell) []bingo.Cell {
	result := make([]bingo.Cell, 0, len(cells))
	for _, cell := range cells {
		result = append(result, bingo.Cell{
			ID:          cell.ID,
			RowIndex:    cell.RowIndex,
			ColIndex:    cell.ColIndex,
			Word:        cell.Word,
			IsFreeSpace: cell.IsFreeSpace,
		})
	}

	return result
}

func domainCellsFromBingo(cells []bingo.Cell, byID map[string]domain.BingoCardCell) []domain.BingoCardCell {
	result := make([]domain.BingoCardCell, 0, len(cells))
	for _, cell := range cells {
		if domainCell, ok := byID[cell.ID]; ok {
			result = append(result, domainCell)
			continue
		}
		result = append(result, domain.BingoCardCell{
			ID:          cell.ID,
			RowIndex:    cell.RowIndex,
			ColIndex:    cell.ColIndex,
			Word:        cell.Word,
			IsFreeSpace: cell.IsFreeSpace,
		})
	}

	return result
}

func calledWordStrings(words []domain.CalledWord) []string {
	result := make([]string, 0, len(words))
	for _, word := range words {
		result = append(result, word.Word)
	}

	return result
}

func generateGameCode(name string, now time.Time) string {
	prefix := "GAME"
	words := strings.FieldsFunc(strings.ToUpper(name), func(r rune) bool {
		return r < 'A' || r > 'Z'
	})
	if len(words) > 0 && words[0] != "" {
		if len(words[0]) > 4 {
			prefix = words[0][:4]
		} else {
			prefix = words[0]
		}
	}

	return fmt.Sprintf("%s-%d", prefix, now.Unix()%100000)
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func mapStoreError(err error) error {
	if errors.Is(err, db.ErrNotFound) {
		return ErrNotFound
	}
	if errors.Is(err, db.ErrConflict) {
		return ErrConflict
	}

	return err
}

func (s *Service) emit(ctx context.Context, event events.Event) {
	_ = s.publisher.Publish(ctx, event)
}

func (s *Service) recordAudit(ctx context.Context, event audit.Event) {
	_ = s.auditLogger.Record(ctx, event)
}
