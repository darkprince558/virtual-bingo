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
	UpdateGameRunStatus(context.Context, string, string, *time.Time, *time.Time) (domain.GameRun, error)
	TransitionGameRunStatus(context.Context, db.GameStatusTransitionParams) (domain.GameRun, error)
	CountAllowedPlayers(context.Context, string) (int, error)
	AddAllowedPlayer(context.Context, db.AddAllowedPlayerParams) (domain.AllowedPlayer, error)
	ListAllowedPlayers(context.Context, string) ([]domain.AllowedPlayer, error)
	GetAllowedPlayerByEmail(context.Context, string, string) (domain.AllowedPlayer, error)
	CreatePlayer(context.Context, db.CreatePlayerParams) (domain.Player, error)
	GetPlayer(context.Context, string, string) (domain.Player, error)
	GetPlayerByGameRunAndEmail(context.Context, string, string) (domain.Player, error)
	UpdatePlayerConnectionState(context.Context, db.UpdatePlayerConnectionStateParams) (domain.Player, error)
	ListWordSetWords(context.Context, string) ([]domain.WordSetWord, error)
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
	CountPlayers(context.Context, string) (int, error)
	ListGameEvents(context.Context, string, int64, int) ([]domain.GameEvent, error)
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
	if !auth.HasRole(principal, "admin", "host") {
		return domain.AllowedPlayer{}, ErrForbidden
	}

	if _, err := s.store.GetGameRun(ctx, input.GameRunID); err != nil {
		return domain.AllowedPlayer{}, mapStoreError(err)
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

func (s *Service) ListAllowedPlayers(ctx context.Context, gameRunID string) ([]domain.AllowedPlayer, error) {
	if _, err := s.store.GetGameRun(ctx, gameRunID); err != nil {
		return nil, mapStoreError(err)
	}

	return s.store.ListAllowedPlayers(ctx, gameRunID)
}

type JoinPlayerInput struct {
	GameRunID   string
	Email       string
	DisplayName string
}

func (s *Service) JoinPlayer(ctx context.Context, input JoinPlayerInput) (domain.Player, error) {
	email := normalizeEmail(input.Email)
	displayName := strings.TrimSpace(input.DisplayName)
	if email == "" || displayName == "" {
		return domain.Player{}, fmt.Errorf("%w: email and displayName are required", ErrValidation)
	}

	if _, err := s.store.GetGameRun(ctx, input.GameRunID); err != nil {
		return domain.Player{}, mapStoreError(err)
	}

	if _, err := s.store.GetAllowedPlayerByEmail(ctx, input.GameRunID, email); err != nil {
		if errors.Is(err, db.ErrNotFound) {
			return domain.Player{}, ErrNotAllowed
		}
		return domain.Player{}, err
	}

	existing, err := s.store.GetPlayerByGameRunAndEmail(ctx, input.GameRunID, email)
	if err == nil {
		updated, err := s.store.UpdatePlayerConnectionState(ctx, db.UpdatePlayerConnectionStateParams{
			GameRunID:       input.GameRunID,
			PlayerID:        existing.ID,
			ConnectionState: "online",
			EventType:       "player.reconnected",
		})
		if err != nil {
			return domain.Player{}, mapStoreError(err)
		}
		s.emit(ctx, events.Event{Type: "player.reconnected", EntityID: updated.ID, Payload: map[string]any{"gameRunId": input.GameRunID, "playerId": updated.ID, "connectionState": updated.ConnectionState}})
		return updated, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return domain.Player{}, err
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
			return s.store.GetPlayerByGameRunAndEmail(ctx, input.GameRunID, email)
		}
		return domain.Player{}, err
	}

	s.emit(ctx, events.Event{Type: "player.joined", EntityID: player.ID, Payload: map[string]any{"gameRunId": input.GameRunID, "email": player.Email}})
	return player, nil
}

func (s *Service) HeartbeatPlayer(ctx context.Context, principal auth.Principal, gameRunID, playerID string) (domain.Player, error) {
	player, err := s.store.GetPlayer(ctx, gameRunID, playerID)
	if err != nil {
		return domain.Player{}, mapStoreError(err)
	}
	if !canActForPlayer(principal, player) {
		return domain.Player{}, ErrForbidden
	}

	eventType := ""
	if player.ConnectionState != "online" {
		eventType = "player.reconnected"
	}
	updated, err := s.store.UpdatePlayerConnectionState(ctx, db.UpdatePlayerConnectionStateParams{
		GameRunID:       gameRunID,
		PlayerID:        playerID,
		ConnectionState: "online",
		EventType:       eventType,
	})
	if err != nil {
		return domain.Player{}, mapStoreError(err)
	}
	if eventType != "" {
		s.emit(ctx, events.Event{Type: eventType, EntityID: updated.ID, Payload: map[string]any{"gameRunId": gameRunID, "playerId": updated.ID, "connectionState": updated.ConnectionState}})
	}

	return updated, nil
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
		s.emit(ctx, events.Event{Type: eventType, EntityID: player.ID, Payload: map[string]any{"gameRunId": gameRunID, "playerId": player.ID, "connectionState": player.ConnectionState}})
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
		GameRun:     run,
		Status:      run.Status,
		CurrentWord: currentWord,
		Pattern:     winningPatternOrDefault(run),
		Player:      player,
		Card:        cardPtr,
		CalledWords: calledWords,
		Claims:      playerClaims,
		Winners:     winners,
	}, nil
}

func (s *Service) ListGameEvents(ctx context.Context, gameRunID string, afterSequence int64, limit int) ([]domain.GameEvent, error) {
	if _, err := s.store.GetGameRun(ctx, gameRunID); err != nil {
		return nil, mapStoreError(err)
	}

	return s.store.ListGameEvents(ctx, gameRunID, afterSequence, limit)
}

func (s *Service) Authenticate(r *http.Request) (auth.Principal, error) {
	return s.authenticator.Authenticate(r)
}

func canActForPlayer(principal auth.Principal, player domain.Player) bool {
	return auth.HasRole(principal, "admin", "host") || strings.EqualFold(principal.Email, player.Email)
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
