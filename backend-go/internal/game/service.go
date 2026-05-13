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
	CountAllowedPlayers(context.Context, string) (int, error)
	AddAllowedPlayer(context.Context, db.AddAllowedPlayerParams) (domain.AllowedPlayer, error)
	ListAllowedPlayers(context.Context, string) ([]domain.AllowedPlayer, error)
	GetAllowedPlayerByEmail(context.Context, string, string) (domain.AllowedPlayer, error)
	CreatePlayer(context.Context, db.CreatePlayerParams) (domain.Player, error)
	GetPlayerByGameRunAndEmail(context.Context, string, string) (domain.Player, error)
	ListWordSetWords(context.Context, string) ([]domain.WordSetWord, error)
	CreateCalledWord(context.Context, db.CreateCalledWordParams) (domain.CalledWord, error)
	ListCalledWords(context.Context, string) ([]domain.CalledWord, error)
	CreateCard(context.Context, db.CreateCardParams) (domain.BingoCard, error)
	GetPlayerCard(context.Context, string) (domain.BingoCard, error)
	SetCardCellMarked(context.Context, db.SetCardCellMarkedParams) (domain.BingoCardCell, error)
	CreateBingoClaim(context.Context, db.CreateBingoClaimParams) (domain.BingoClaim, error)
	UpdateBingoClaimValidation(context.Context, db.UpdateBingoClaimValidationParams) (domain.BingoClaim, error)
	GetBingoClaim(context.Context, string) (domain.BingoClaim, error)
	ListBingoClaims(context.Context, string) ([]domain.BingoClaim, error)
	CreateWinner(context.Context, db.CreateWinnerParams) (domain.Winner, error)
	ListWinners(context.Context, string) ([]domain.Winner, error)
	CountWinners(context.Context, string) (int, error)
	CountPlayers(context.Context, string) (int, error)
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
	run, err = s.store.UpdateGameRunStatus(ctx, gameRunID, "live", &startedAt, nil)
	if err != nil {
		return GameRunWithCounts{}, mapStoreError(err)
	}

	s.emit(ctx, events.Event{Type: "game.started", EntityID: run.ID, Payload: map[string]any{"status": run.Status}})
	s.recordAudit(ctx, audit.Event{GameRunID: &run.ID, ActorUserID: &user.ID, EventType: "game.started", EntityType: "game_run", EntityID: &run.ID, Payload: map[string]any{"status": run.Status}})

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
		return existing, nil
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

type claimValidationResult struct {
	Valid        bool                  `json:"valid"`
	MatchedCells []claimValidationCell `json:"matchedCells"`
	MissingCells []claimValidationCell `json:"missingCells"`
	Reason       string                `json:"reason"`
	Pattern      string                `json:"pattern"`
}

type claimValidationCell struct {
	ID          string `json:"id"`
	RowIndex    int    `json:"rowIndex"`
	ColIndex    int    `json:"colIndex"`
	Word        string `json:"word"`
	IsFreeSpace bool   `json:"isFreeSpace"`
}

func (s *Service) SubmitBingoClaim(ctx context.Context, input SubmitBingoClaimInput) (BingoClaimResult, error) {
	pattern := strings.TrimSpace(input.Pattern)
	if pattern == "" {
		pattern = "single_line"
	}
	if pattern != "single_line" {
		return BingoClaimResult{}, fmt.Errorf("%w: unsupported claim pattern %s", ErrValidation, pattern)
	}

	run, err := s.store.GetGameRun(ctx, input.GameRunID)
	if err != nil {
		return BingoClaimResult{}, mapStoreError(err)
	}
	if run.Status != "live" {
		return BingoClaimResult{}, fmt.Errorf("%w: game must be live to submit a claim", ErrConflict)
	}

	card, err := s.store.GetPlayerCard(ctx, input.PlayerID)
	if err != nil {
		return BingoClaimResult{}, mapStoreError(err)
	}
	if card.GameRunID != input.GameRunID {
		return BingoClaimResult{}, ErrNotFound
	}

	claim, err := s.store.CreateBingoClaim(ctx, db.CreateBingoClaimParams{
		GameRunID: input.GameRunID,
		PlayerID:  input.PlayerID,
		Pattern:   pattern,
		Status:    "pending",
	})
	if err != nil {
		return BingoClaimResult{}, err
	}

	s.emit(ctx, events.Event{Type: "claim.submitted", EntityID: claim.ID, Payload: map[string]any{"gameRunId": input.GameRunID, "playerId": input.PlayerID, "pattern": pattern}})
	s.recordAudit(ctx, audit.Event{GameRunID: &input.GameRunID, EventType: "claim.submitted", EntityType: "bingo_claim", EntityID: &claim.ID, Payload: map[string]any{"playerId": input.PlayerID, "pattern": pattern}})

	calledWords, err := s.store.ListCalledWords(ctx, input.GameRunID)
	if err != nil {
		return BingoClaimResult{}, err
	}

	validation := validateSingleLineClaim(card, calledWords, pattern)
	validationJSON, err := json.Marshal(validation)
	if err != nil {
		return BingoClaimResult{}, fmt.Errorf("marshal claim validation: %w", err)
	}

	status := "invalid"
	if validation.Valid {
		status = "confirmed"
	}
	now := s.clock.Now()
	claim, err = s.store.UpdateBingoClaimValidation(ctx, db.UpdateBingoClaimValidationParams{
		ClaimID:          claim.ID,
		Status:           status,
		ValidationResult: validationJSON,
		ReviewedAt:       &now,
	})
	if err != nil {
		return BingoClaimResult{}, err
	}

	result := BingoClaimResult{Claim: claim}
	if validation.Valid {
		winners, err := s.store.ListWinners(ctx, input.GameRunID)
		if err != nil {
			return BingoClaimResult{}, err
		}
		for _, winner := range winners {
			if winner.ClaimID != nil && *winner.ClaimID == claim.ID {
				result.Winner = &winner
				break
			}
		}

		if result.Winner == nil && len(winners) < 3 {
			placement := len(winners) + 1
			winner, err := s.store.CreateWinner(ctx, db.CreateWinnerParams{
				GameRunID: input.GameRunID,
				PlayerID:  input.PlayerID,
				ClaimID:   &claim.ID,
				Placement: placement,
				Pattern:   pattern,
			})
			if err != nil {
				if !errors.Is(err, db.ErrConflict) {
					return BingoClaimResult{}, err
				}
			} else {
				result.Winner = &winner
				s.emit(ctx, events.Event{Type: "winner.created", EntityID: winner.ID, Payload: map[string]any{"gameRunId": input.GameRunID, "playerId": input.PlayerID, "placement": winner.Placement}})
				s.recordAudit(ctx, audit.Event{GameRunID: &input.GameRunID, EventType: "winner.created", EntityType: "winner", EntityID: &winner.ID, Payload: map[string]any{"playerId": input.PlayerID, "placement": winner.Placement}})
				if placement == 3 {
					endedAt := s.clock.Now()
					if _, err := s.store.UpdateGameRunStatus(ctx, input.GameRunID, "finished", nil, &endedAt); err != nil {
						return BingoClaimResult{}, err
					}
				}
			}
		}
	}

	s.emit(ctx, events.Event{Type: "claim.validated", EntityID: claim.ID, Payload: map[string]any{"gameRunId": input.GameRunID, "status": claim.Status, "valid": validation.Valid}})
	s.recordAudit(ctx, audit.Event{GameRunID: &input.GameRunID, EventType: "claim.validated", EntityType: "bingo_claim", EntityID: &claim.ID, Payload: map[string]any{"status": claim.Status, "valid": validation.Valid}})

	return result, nil
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
		Status:          run.Status,
	}, nil
}

func (s *Service) Authenticate(r *http.Request) (auth.Principal, error) {
	return s.authenticator.Authenticate(r)
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

func validateSingleLineClaim(card domain.BingoCard, calledWords []domain.CalledWord, pattern string) claimValidationResult {
	called := make(map[string]struct{}, len(calledWords))
	for _, word := range calledWords {
		called[strings.ToLower(word.Word)] = struct{}{}
	}

	cellsByPosition := make(map[[2]int]domain.BingoCardCell, len(card.Cells))
	for _, cell := range card.Cells {
		cellsByPosition[[2]int{cell.RowIndex, cell.ColIndex}] = cell
	}

	lines := singleLinePositions()
	best := claimValidationResult{
		Valid:        false,
		MatchedCells: []claimValidationCell{},
		MissingCells: []claimValidationCell{},
		Reason:       "no_complete_line",
		Pattern:      pattern,
	}
	bestMissingCount := 26

	for _, line := range lines {
		matched := make([]claimValidationCell, 0, len(line))
		missing := make([]claimValidationCell, 0)

		for _, position := range line {
			cell, ok := cellsByPosition[position]
			if !ok {
				continue
			}
			cellRef := claimCellRef(cell)
			if cell.IsFreeSpace {
				matched = append(matched, cellRef)
				continue
			}
			if _, ok := called[strings.ToLower(cell.Word)]; ok {
				matched = append(matched, cellRef)
				continue
			}
			missing = append(missing, cellRef)
		}

		if len(missing) == 0 && len(matched) == len(line) {
			return claimValidationResult{
				Valid:        true,
				MatchedCells: matched,
				MissingCells: []claimValidationCell{},
				Reason:       "single_line_complete",
				Pattern:      pattern,
			}
		}
		if len(missing) < bestMissingCount {
			bestMissingCount = len(missing)
			best.MatchedCells = matched
			best.MissingCells = missing
		}
	}

	return best
}

func singleLinePositions() [][][2]int {
	lines := make([][][2]int, 0, 12)

	for row := 0; row < 5; row++ {
		line := make([][2]int, 0, 5)
		for col := 0; col < 5; col++ {
			line = append(line, [2]int{row, col})
		}
		lines = append(lines, line)
	}

	for col := 0; col < 5; col++ {
		line := make([][2]int, 0, 5)
		for row := 0; row < 5; row++ {
			line = append(line, [2]int{row, col})
		}
		lines = append(lines, line)
	}

	firstDiagonal := make([][2]int, 0, 5)
	secondDiagonal := make([][2]int, 0, 5)
	for index := 0; index < 5; index++ {
		firstDiagonal = append(firstDiagonal, [2]int{index, index})
		secondDiagonal = append(secondDiagonal, [2]int{index, 4 - index})
	}
	lines = append(lines, firstDiagonal, secondDiagonal)

	return lines
}

func claimCellRef(cell domain.BingoCardCell) claimValidationCell {
	return claimValidationCell{
		ID:          cell.ID,
		RowIndex:    cell.RowIndex,
		ColIndex:    cell.ColIndex,
		Word:        cell.Word,
		IsFreeSpace: cell.IsFreeSpace,
	}
}

func seedInt64(seed string) int64 {
	hash := sha256.Sum256([]byte(seed))
	return int64(binary.BigEndian.Uint64(hash[:8]))
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
