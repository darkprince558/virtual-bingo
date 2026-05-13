package game

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
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
	CountAllowedPlayers(context.Context, string) (int, error)
	AddAllowedPlayer(context.Context, db.AddAllowedPlayerParams) (domain.AllowedPlayer, error)
	ListAllowedPlayers(context.Context, string) ([]domain.AllowedPlayer, error)
	GetAllowedPlayerByEmail(context.Context, string, string) (domain.AllowedPlayer, error)
	CreatePlayer(context.Context, db.CreatePlayerParams) (domain.Player, error)
	GetPlayerByGameRunAndEmail(context.Context, string, string) (domain.Player, error)
	ListWordSetWords(context.Context, string) ([]domain.WordSetWord, error)
	CreateCard(context.Context, db.CreateCardParams) (domain.BingoCard, error)
	GetPlayerCard(context.Context, string) (domain.BingoCard, error)
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
