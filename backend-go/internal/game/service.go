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
}

type Service struct {
	store         Store
	authenticator auth.Authenticator
	publisher     events.Publisher
	auditLogger   audit.Logger
	clock         clock.Clock
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

	return &Service{
		store:         config.Store,
		authenticator: config.Authenticator,
		publisher:     config.Publisher,
		auditLogger:   config.AuditLogger,
		clock:         config.Clock,
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

	return domain.HostSnapshot{
		GameRun:     summary.GameRun,
		Status:      summary.Status,
		CurrentWord: summary.CurrentWord,
		Pattern:     winningPatternOrDefault(summary.GameRun),
		PlayerCount: summary.PlayerCount,
		Players:     summary.Players,
		CalledWords: summary.CalledWords,
		Claims:      summary.Claims,
		Winners:     summary.Winners,
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

	var currentWord *domain.CalledWord
	if len(calledWords) > 0 {
		word := calledWords[len(calledWords)-1]
		currentWord = &word
	}

	return domain.PlayerSnapshot{
		GameRun:         run,
		Status:          run.Status,
		CurrentWord:     currentWord,
		Pattern:         winningPatternOrDefault(run),
		Player:          player,
		Card:            cardPtr,
		CalledWords:     calledWords,
		Claims:          playerClaims,
		Winners:         winners,
		ReconnectNotice: notice,
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
