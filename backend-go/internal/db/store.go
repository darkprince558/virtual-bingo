package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/audit"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/auth"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, external_subject, display_name, email, role, created_at, updated_at
		FROM users
		WHERE lower(email) = lower($1)
	`, email)

	return scanUser(row)
}

func (s *Store) UpsertUserFromPrincipal(ctx context.Context, principal auth.Principal) (domain.User, error) {
	role := roleFromPrincipal(principal)
	row := s.pool.QueryRow(ctx, `
		INSERT INTO users (external_subject, display_name, email, role)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT ((lower(email))) DO UPDATE
		SET external_subject = EXCLUDED.external_subject,
		    display_name = EXCLUDED.display_name,
		    role = EXCLUDED.role,
		    updated_at = now()
		RETURNING id::text, external_subject, display_name, email, role, created_at, updated_at
	`, principal.ID, principal.DisplayName, strings.ToLower(principal.Email), role)

	return scanUser(row)
}

type CreateGameRunParams struct {
	TemplateID       *string
	HostUserID       string
	WordSetID        *string
	Code             string
	Name             string
	Status           string
	ScheduledStartAt *time.Time
	WinningPattern   *string
}

func (s *Store) CreateGameRun(ctx context.Context, params CreateGameRunParams) (domain.GameRun, error) {
	status := params.Status
	if status == "" {
		status = "draft"
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO game_runs (
			template_id,
			host_user_id,
			word_set_id,
			code,
			name,
			status,
			scheduled_start_at,
			winning_pattern
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::text, template_id::text, host_user_id::text, word_set_id::text, code, name, status, scheduled_start_at, started_at, ended_at, current_called_word_id::text, winning_pattern, created_at, updated_at
	`, params.TemplateID, params.HostUserID, params.WordSetID, params.Code, params.Name, status, params.ScheduledStartAt, params.WinningPattern)

	run, err := scanGameRun(row)
	if err != nil {
		return domain.GameRun{}, mapWriteError(err)
	}

	return run, nil
}

func (s *Store) GetGameRun(ctx context.Context, id string) (domain.GameRun, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, template_id::text, host_user_id::text, word_set_id::text, code, name, status, scheduled_start_at, started_at, ended_at, current_called_word_id::text, winning_pattern, created_at, updated_at
		FROM game_runs
		WHERE id = $1
	`, id)

	return scanGameRun(row)
}

func (s *Store) UpdateGameRunStatus(ctx context.Context, gameRunID, status string, startedAt *time.Time, endedAt *time.Time) (domain.GameRun, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE game_runs
		SET status = $2,
		    started_at = CASE
		      WHEN $3::timestamptz IS NULL THEN started_at
		      WHEN $2 = 'live' THEN COALESCE(started_at, $3)
		      ELSE $3
		    END,
		    ended_at = COALESCE($4, ended_at),
		    updated_at = now()
		WHERE id = $1
		RETURNING id::text, template_id::text, host_user_id::text, word_set_id::text, code, name, status, scheduled_start_at, started_at, ended_at, current_called_word_id::text, winning_pattern, created_at, updated_at
	`, gameRunID, status, startedAt, endedAt)

	return scanGameRun(row)
}

type AddAllowedPlayerParams struct {
	GameRunID   string
	Email       string
	DisplayName string
	Source      string
}

func (s *Store) AddAllowedPlayer(ctx context.Context, params AddAllowedPlayerParams) (domain.AllowedPlayer, error) {
	source := params.Source
	if source == "" {
		source = "manual"
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO allowed_players (game_run_id, email, display_name, source)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, game_run_id::text, email, display_name, source, created_at
	`, params.GameRunID, params.Email, params.DisplayName, source)

	player, err := scanAllowedPlayer(row)
	if err != nil {
		return domain.AllowedPlayer{}, mapWriteError(err)
	}

	return player, nil
}

func (s *Store) GetAllowedPlayerByEmail(ctx context.Context, gameRunID, email string) (domain.AllowedPlayer, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, email, display_name, source, created_at
		FROM allowed_players
		WHERE game_run_id = $1 AND lower(email) = lower($2)
	`, gameRunID, email)

	return scanAllowedPlayer(row)
}

func (s *Store) ListAllowedPlayers(ctx context.Context, gameRunID string) ([]domain.AllowedPlayer, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, game_run_id::text, email, display_name, source, created_at
		FROM allowed_players
		WHERE game_run_id = $1
		ORDER BY created_at, id
	`, gameRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make([]domain.AllowedPlayer, 0)
	for rows.Next() {
		player, err := scanAllowedPlayer(rows)
		if err != nil {
			return nil, err
		}
		players = append(players, player)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return players, nil
}

func (s *Store) CountAllowedPlayers(ctx context.Context, gameRunID string) (int, error) {
	var count int
	if err := s.pool.QueryRow(ctx, `
		SELECT count(*)
		FROM allowed_players
		WHERE game_run_id = $1
	`, gameRunID).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

type CreatePlayerParams struct {
	GameRunID       string
	UserID          *string
	Email           string
	DisplayName     string
	ConnectionState string
	State           string
}

func (s *Store) CreatePlayer(ctx context.Context, params CreatePlayerParams) (domain.Player, error) {
	connectionState := params.ConnectionState
	if connectionState == "" {
		connectionState = "offline"
	}
	state := params.State
	if state == "" {
		state = "joined"
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO players (game_run_id, user_id, email, display_name, connection_state, state)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id::text, game_run_id::text, user_id::text, email, display_name, connection_state, state, joined_at, last_seen_at, created_at, updated_at
	`, params.GameRunID, params.UserID, params.Email, params.DisplayName, connectionState, state)

	player, err := scanPlayer(row)
	if err != nil {
		return domain.Player{}, mapWriteError(err)
	}

	return player, nil
}

func (s *Store) GetPlayerByGameRunAndEmail(ctx context.Context, gameRunID, email string) (domain.Player, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, user_id::text, email, display_name, connection_state, state, joined_at, last_seen_at, created_at, updated_at
		FROM players
		WHERE game_run_id = $1 AND lower(email) = lower($2)
	`, gameRunID, email)

	return scanPlayer(row)
}

func (s *Store) ListWordSetWords(ctx context.Context, wordSetID string) ([]domain.WordSetWord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, word_set_id::text, word, sort_order, is_active, created_at
		FROM word_set_words
		WHERE word_set_id = $1 AND is_active = true
		ORDER BY sort_order
	`, wordSetID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	words := make([]domain.WordSetWord, 0)
	for rows.Next() {
		word, err := scanWordSetWord(rows)
		if err != nil {
			return nil, err
		}
		words = append(words, word)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return words, nil
}

type CreateCalledWordParams struct {
	GameRunID      string
	WordSetWordID  *string
	Word           string
	CalledByUserID *string
}

func (s *Store) CreateCalledWord(ctx context.Context, params CreateCalledWordParams) (domain.CalledWord, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	var lockedGameRunID string
	if err := tx.QueryRow(ctx, `SELECT id::text FROM game_runs WHERE id = $1 FOR UPDATE`, params.GameRunID).Scan(&lockedGameRunID); err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO called_words (game_run_id, word_set_word_id, word, called_by_user_id, sequence)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			(SELECT COALESCE(MAX(sequence), 0) + 1 FROM called_words WHERE game_run_id = $1)
		)
		RETURNING id::text, game_run_id::text, word_set_word_id::text, word, called_by_user_id::text, sequence, called_at, created_at
	`, params.GameRunID, params.WordSetWordID, params.Word, params.CalledByUserID)

	calledWord, err := scanCalledWord(row)
	if err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}

	if _, err := tx.Exec(ctx, `
		UPDATE game_runs
		SET current_called_word_id = $2,
		    updated_at = now()
		WHERE id = $1
	`, params.GameRunID, calledWord.ID); err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}

	return calledWord, nil
}

func (s *Store) ListCalledWords(ctx context.Context, gameRunID string) ([]domain.CalledWord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, game_run_id::text, word_set_word_id::text, word, called_by_user_id::text, sequence, called_at, created_at
		FROM called_words
		WHERE game_run_id = $1
		ORDER BY sequence ASC
	`, gameRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	words := make([]domain.CalledWord, 0)
	for rows.Next() {
		word, err := scanCalledWord(rows)
		if err != nil {
			return nil, err
		}
		words = append(words, word)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return words, nil
}

func (s *Store) GetCalledWordByWord(ctx context.Context, gameRunID, word string) (domain.CalledWord, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, word_set_word_id::text, word, called_by_user_id::text, sequence, called_at, created_at
		FROM called_words
		WHERE game_run_id = $1 AND lower(word) = lower($2)
	`, gameRunID, word)

	return scanCalledWord(row)
}

type CreateCardParams struct {
	GameRunID string
	PlayerID  string
	Seed      string
	Cells     []CreateCardCellParams
}

type CreateCardCellParams struct {
	RowIndex    int
	ColIndex    int
	Word        string
	IsFreeSpace bool
	MarkedAt    *time.Time
}

func (s *Store) CreateCard(ctx context.Context, params CreateCardParams) (domain.BingoCard, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.BingoCard{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	cardRow := tx.QueryRow(ctx, `
		INSERT INTO bingo_cards (game_run_id, player_id, seed)
		VALUES ($1, $2, $3)
		RETURNING id::text, game_run_id::text, player_id::text, seed, created_at
	`, params.GameRunID, params.PlayerID, params.Seed)

	card, err := scanBingoCard(cardRow)
	if err != nil {
		return domain.BingoCard{}, mapWriteError(err)
	}

	card.Cells = make([]domain.BingoCardCell, 0, len(params.Cells))
	for _, cell := range params.Cells {
		cellRow := tx.QueryRow(ctx, `
			INSERT INTO bingo_card_cells (card_id, row_index, col_index, word, is_free_space, marked_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id::text, card_id::text, row_index, col_index, word, is_free_space, marked_at, created_at
		`, card.ID, cell.RowIndex, cell.ColIndex, cell.Word, cell.IsFreeSpace, cell.MarkedAt)

		createdCell, err := scanBingoCardCell(cellRow)
		if err != nil {
			return domain.BingoCard{}, err
		}
		card.Cells = append(card.Cells, createdCell)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.BingoCard{}, err
	}

	return card, nil
}

func (s *Store) RecordAuditEvent(ctx context.Context, event audit.Event) error {
	payload := event.Payload
	if payload == nil {
		payload = map[string]any{}
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal audit payload: %w", err)
	}

	_, err = s.pool.Exec(ctx, `
		INSERT INTO audit_events (game_run_id, actor_user_id, event_type, entity_type, entity_id, payload)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, event.GameRunID, event.ActorUserID, event.EventType, event.EntityType, event.EntityID, payloadJSON)
	if err != nil {
		return mapWriteError(err)
	}

	return nil
}

func (s *Store) GetPlayerCard(ctx context.Context, playerID string) (domain.BingoCard, error) {
	cardRow := s.pool.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, player_id::text, seed, created_at
		FROM bingo_cards
		WHERE player_id = $1
	`, playerID)

	card, err := scanBingoCard(cardRow)
	if err != nil {
		return domain.BingoCard{}, err
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id::text, card_id::text, row_index, col_index, word, is_free_space, marked_at, created_at
		FROM bingo_card_cells
		WHERE card_id = $1
		ORDER BY row_index, col_index
	`, card.ID)
	if err != nil {
		return domain.BingoCard{}, err
	}
	defer rows.Close()

	card.Cells = make([]domain.BingoCardCell, 0)
	for rows.Next() {
		cell, err := scanBingoCardCell(rows)
		if err != nil {
			return domain.BingoCard{}, err
		}
		card.Cells = append(card.Cells, cell)
	}
	if err := rows.Err(); err != nil {
		return domain.BingoCard{}, err
	}

	return card, nil
}

type SetCardCellMarkedParams struct {
	GameRunID string
	PlayerID  string
	CellID    string
	Marked    bool
}

func (s *Store) SetCardCellMarked(ctx context.Context, params SetCardCellMarkedParams) (domain.BingoCardCell, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE bingo_card_cells AS cell
		SET marked_at = CASE WHEN $4 THEN now() ELSE NULL END
		FROM bingo_cards AS card
		WHERE cell.card_id = card.id
		  AND card.game_run_id = $1
		  AND card.player_id = $2
		  AND cell.id = $3
		RETURNING cell.id::text, cell.card_id::text, cell.row_index, cell.col_index, cell.word, cell.is_free_space, cell.marked_at, cell.created_at
	`, params.GameRunID, params.PlayerID, params.CellID, params.Marked)

	return scanBingoCardCell(row)
}

type CreateBingoClaimParams struct {
	GameRunID        string
	PlayerID         string
	Pattern          string
	Status           string
	ValidationResult json.RawMessage
}

func (s *Store) CreateBingoClaim(ctx context.Context, params CreateBingoClaimParams) (domain.BingoClaim, error) {
	status := params.Status
	if status == "" {
		status = "pending"
	}
	validationResult := params.ValidationResult
	if len(validationResult) == 0 {
		validationResult = json.RawMessage(`{}`)
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO bingo_claims (game_run_id, player_id, pattern, status, validation_result)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, game_run_id::text, player_id::text, pattern, status, validation_result, claimed_at, reviewed_by_user_id::text, reviewed_at, created_at, updated_at
	`, params.GameRunID, params.PlayerID, params.Pattern, status, validationResult)

	claim, err := scanBingoClaim(row)
	if err != nil {
		return domain.BingoClaim{}, mapWriteError(err)
	}

	return claim, nil
}

type UpdateBingoClaimValidationParams struct {
	ClaimID          string
	Status           string
	ValidationResult json.RawMessage
	ReviewedByUserID *string
	ReviewedAt       *time.Time
}

func (s *Store) UpdateBingoClaimValidation(ctx context.Context, params UpdateBingoClaimValidationParams) (domain.BingoClaim, error) {
	validationResult := params.ValidationResult
	if len(validationResult) == 0 {
		validationResult = json.RawMessage(`{}`)
	}

	row := s.pool.QueryRow(ctx, `
		UPDATE bingo_claims
		SET status = $2,
		    validation_result = $3,
		    reviewed_by_user_id = $4,
		    reviewed_at = $5,
		    updated_at = now()
		WHERE id = $1
		RETURNING id::text, game_run_id::text, player_id::text, pattern, status, validation_result, claimed_at, reviewed_by_user_id::text, reviewed_at, created_at, updated_at
	`, params.ClaimID, params.Status, validationResult, params.ReviewedByUserID, params.ReviewedAt)

	return scanBingoClaim(row)
}

func (s *Store) GetBingoClaim(ctx context.Context, claimID string) (domain.BingoClaim, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, player_id::text, pattern, status, validation_result, claimed_at, reviewed_by_user_id::text, reviewed_at, created_at, updated_at
		FROM bingo_claims
		WHERE id = $1
	`, claimID)

	return scanBingoClaim(row)
}

func (s *Store) ListBingoClaims(ctx context.Context, gameRunID string) ([]domain.BingoClaim, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, game_run_id::text, player_id::text, pattern, status, validation_result, claimed_at, reviewed_by_user_id::text, reviewed_at, created_at, updated_at
		FROM bingo_claims
		WHERE game_run_id = $1
		ORDER BY claimed_at ASC, id ASC
	`, gameRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	claims := make([]domain.BingoClaim, 0)
	for rows.Next() {
		claim, err := scanBingoClaim(rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return claims, nil
}

type CreateWinnerParams struct {
	GameRunID string
	PlayerID  string
	ClaimID   *string
	Placement int
	Pattern   string
}

func (s *Store) CreateWinner(ctx context.Context, params CreateWinnerParams) (domain.Winner, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO winners (game_run_id, player_id, claim_id, placement, pattern)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, game_run_id::text, player_id::text, claim_id::text, placement, pattern, confirmed_at, created_at
	`, params.GameRunID, params.PlayerID, params.ClaimID, params.Placement, params.Pattern)

	winner, err := scanWinner(row)
	if err != nil {
		return domain.Winner{}, mapWriteError(err)
	}

	return winner, nil
}

func (s *Store) ListWinners(ctx context.Context, gameRunID string) ([]domain.Winner, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, game_run_id::text, player_id::text, claim_id::text, placement, pattern, confirmed_at, created_at
		FROM winners
		WHERE game_run_id = $1
		ORDER BY placement ASC, confirmed_at ASC
	`, gameRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	winners := make([]domain.Winner, 0)
	for rows.Next() {
		winner, err := scanWinner(rows)
		if err != nil {
			return nil, err
		}
		winners = append(winners, winner)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return winners, nil
}

func (s *Store) CountWinners(ctx context.Context, gameRunID string) (int, error) {
	var count int
	if err := s.pool.QueryRow(ctx, `
		SELECT count(*)
		FROM winners
		WHERE game_run_id = $1
	`, gameRunID).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) CountPlayers(ctx context.Context, gameRunID string) (int, error) {
	var count int
	if err := s.pool.QueryRow(ctx, `
		SELECT count(*)
		FROM players
		WHERE game_run_id = $1
	`, gameRunID).Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanGameRun(row scanner) (domain.GameRun, error) {
	var run domain.GameRun
	var templateID, wordSetID, currentCalledWordID, winningPattern sql.NullString
	var scheduledStartAt, startedAt, endedAt sql.NullTime

	err := row.Scan(
		&run.ID,
		&templateID,
		&run.HostUserID,
		&wordSetID,
		&run.Code,
		&run.Name,
		&run.Status,
		&scheduledStartAt,
		&startedAt,
		&endedAt,
		&currentCalledWordID,
		&winningPattern,
		&run.CreatedAt,
		&run.UpdatedAt,
	)
	if err != nil {
		return domain.GameRun{}, mapError(err)
	}

	run.TemplateID = nullableStringPtr(templateID)
	run.WordSetID = nullableStringPtr(wordSetID)
	run.ScheduledStartAt = nullableTimePtr(scheduledStartAt)
	run.StartedAt = nullableTimePtr(startedAt)
	run.EndedAt = nullableTimePtr(endedAt)
	run.CurrentCalledWordID = nullableStringPtr(currentCalledWordID)
	run.WinningPattern = nullableStringPtr(winningPattern)

	return run, nil
}

func scanUser(row scanner) (domain.User, error) {
	var user domain.User
	var externalSubject sql.NullString

	err := row.Scan(&user.ID, &externalSubject, &user.DisplayName, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return domain.User{}, mapError(err)
	}

	user.ExternalSubject = nullableStringPtr(externalSubject)
	return user, nil
}

func scanWordSetWord(row scanner) (domain.WordSetWord, error) {
	var word domain.WordSetWord
	err := row.Scan(&word.ID, &word.WordSetID, &word.Word, &word.SortOrder, &word.IsActive, &word.CreatedAt)
	if err != nil {
		return domain.WordSetWord{}, mapError(err)
	}

	return word, nil
}

func scanAllowedPlayer(row scanner) (domain.AllowedPlayer, error) {
	var player domain.AllowedPlayer
	err := row.Scan(&player.ID, &player.GameRunID, &player.Email, &player.DisplayName, &player.Source, &player.CreatedAt)
	if err != nil {
		return domain.AllowedPlayer{}, mapError(err)
	}

	return player, nil
}

func scanPlayer(row scanner) (domain.Player, error) {
	var player domain.Player
	var userID sql.NullString

	err := row.Scan(
		&player.ID,
		&player.GameRunID,
		&userID,
		&player.Email,
		&player.DisplayName,
		&player.ConnectionState,
		&player.State,
		&player.JoinedAt,
		&player.LastSeenAt,
		&player.CreatedAt,
		&player.UpdatedAt,
	)
	if err != nil {
		return domain.Player{}, mapError(err)
	}

	player.UserID = nullableStringPtr(userID)
	return player, nil
}

func scanBingoCard(row scanner) (domain.BingoCard, error) {
	var card domain.BingoCard
	err := row.Scan(&card.ID, &card.GameRunID, &card.PlayerID, &card.Seed, &card.CreatedAt)
	if err != nil {
		return domain.BingoCard{}, mapError(err)
	}

	return card, nil
}

func scanBingoCardCell(row scanner) (domain.BingoCardCell, error) {
	var cell domain.BingoCardCell
	var markedAt sql.NullTime

	err := row.Scan(
		&cell.ID,
		&cell.CardID,
		&cell.RowIndex,
		&cell.ColIndex,
		&cell.Word,
		&cell.IsFreeSpace,
		&markedAt,
		&cell.CreatedAt,
	)
	if err != nil {
		return domain.BingoCardCell{}, mapError(err)
	}

	cell.MarkedAt = nullableTimePtr(markedAt)
	return cell, nil
}

func scanCalledWord(row scanner) (domain.CalledWord, error) {
	var calledWord domain.CalledWord
	var wordSetWordID, calledByUserID sql.NullString

	err := row.Scan(
		&calledWord.ID,
		&calledWord.GameRunID,
		&wordSetWordID,
		&calledWord.Word,
		&calledByUserID,
		&calledWord.Sequence,
		&calledWord.CalledAt,
		&calledWord.CreatedAt,
	)
	if err != nil {
		return domain.CalledWord{}, mapError(err)
	}

	calledWord.WordSetWordID = nullableStringPtr(wordSetWordID)
	calledWord.CalledByUserID = nullableStringPtr(calledByUserID)
	return calledWord, nil
}

func scanBingoClaim(row scanner) (domain.BingoClaim, error) {
	var claim domain.BingoClaim
	var reviewedByUserID sql.NullString
	var reviewedAt sql.NullTime
	var validationResult []byte

	err := row.Scan(
		&claim.ID,
		&claim.GameRunID,
		&claim.PlayerID,
		&claim.Pattern,
		&claim.Status,
		&validationResult,
		&claim.ClaimedAt,
		&reviewedByUserID,
		&reviewedAt,
		&claim.CreatedAt,
		&claim.UpdatedAt,
	)
	if err != nil {
		return domain.BingoClaim{}, mapError(err)
	}

	claim.ValidationResult = json.RawMessage(validationResult)
	claim.ReviewedByUserID = nullableStringPtr(reviewedByUserID)
	claim.ReviewedAt = nullableTimePtr(reviewedAt)
	return claim, nil
}

func scanWinner(row scanner) (domain.Winner, error) {
	var winner domain.Winner
	var claimID sql.NullString

	err := row.Scan(
		&winner.ID,
		&winner.GameRunID,
		&winner.PlayerID,
		&claimID,
		&winner.Placement,
		&winner.Pattern,
		&winner.ConfirmedAt,
		&winner.CreatedAt,
	)
	if err != nil {
		return domain.Winner{}, mapError(err)
	}

	winner.ClaimID = nullableStringPtr(claimID)
	return winner, nil
}

func nullableStringPtr(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}

	return &value.String
}

func nullableTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	return &value.Time
}

func mapError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	return fmt.Errorf("database query: %w", err)
}

func mapWriteError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}

	return fmt.Errorf("database write: %w", err)
}

func roleFromPrincipal(principal auth.Principal) string {
	if auth.HasRole(principal, "admin") {
		return "admin"
	}
	if auth.HasRole(principal, "host") {
		return "host"
	}
	if auth.HasRole(principal, "viewer") {
		return "viewer"
	}

	return "player"
}
