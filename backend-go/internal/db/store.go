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
	"github.com/darkprince558/virtual-bingo/backend-go/internal/events"
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

type UpdateGameSettingsParams struct {
	GameRunID                    string
	MarkingMode                  *string
	AllowPlayerMarkingModeChoice *bool
	ShowClaimReadiness           *bool
	VoiceClaimMode               *string
	VoiceClaimAutoplay           *bool
	CallerMode                   *string
	ThemeMode                    *string
	ThemeID                      *string
	ActorUserID                  *string
}

type UpdatePlayerPreferencesParams struct {
	GameRunID   string
	PlayerID    string
	MarkingMode *string
	ActorUserID *string
}

type AutoMarkRunResult struct {
	PlayersScanned     int
	PlayersMarked      int
	CalledWordsScanned int
	CellsMarked        int
	Mode               string
	SkippedReason      string
}

type CreateContentGenerationJobParams struct {
	GameRunID string
	JobType   string
	Status    string
	Provider  string
}

type UpdateContentGenerationJobParams struct {
	JobID        string
	Status       string
	Provider     *string
	ErrorMessage *string
}

type UpsertGeneratedGameContentParams struct {
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
	GenerationProvider   string
	GenerationError      *string
	ActorUserID          *string
}

type UpdateGeneratedGameContentParams struct {
	GameRunID    string
	Topic        *string
	Summary      *string
	Words        []string
	CallerStyle  *string
	ActorUserID  *string
	HasWordPatch bool
}

type LockGeneratedGameContentParams struct {
	GameRunID   string
	ActorUserID *string
}

type LockGeneratedGameContentResult struct {
	Content domain.GeneratedGameContent
	WordSet domain.WordSetWithWords
	GameRun domain.GameRun
}

type CreateGameCallDeckParams struct {
	GameRunID      string
	ShuffleSeed    string
	ShuffleVersion string
	Words          []domain.WordSetWord
}

type CreateCallerAssetParams struct {
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
}

type CreateDeliveryBatchParams struct {
	GameRunID string
	Channel   string
	Purpose   string
	Attempts  []CreateDeliveryAttemptParams
}

type CreateDeliveryAttemptParams struct {
	RecipientEmail  string
	RecipientUserID *string
	Subject         string
	TemplateKey     string
	BodyPreview     string
	LinkURL         string
	GameCode        string
	Status          string
	ErrorReason     *string
}

type UpdatePlayerProfileParams struct {
	GameRunID   string
	PlayerID    string
	Icon        string
	AvatarColor string
	AvatarLabel string
}

type CreateThemeParams struct {
	GameRunID       *string
	GenerationJobID *string
	Name            string
	Summary         string
	Tokens          domain.ThemeTokens
	Provider        string
	CreatedByUserID *string
}

func (s *Store) CreateGameRun(ctx context.Context, params CreateGameRunParams) (domain.GameRun, error) {
	status := params.Status
	if status == "" {
		status = "draft"
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.GameRun{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
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

	if _, err := insertOutboxEventInTx(ctx, tx, run.ID, "game.created", &run.ID, map[string]any{"code": run.Code, "status": run.Status}); err != nil {
		return domain.GameRun{}, err
	}
	if err := tx.Commit(ctx); err != nil {
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

func (s *Store) GetGameRunByCode(ctx context.Context, code string) (domain.GameRun, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, template_id::text, host_user_id::text, word_set_id::text, code, name, status, scheduled_start_at, started_at, ended_at, current_called_word_id::text, winning_pattern, created_at, updated_at
		FROM game_runs
		WHERE lower(code) = lower($1)
	`, code)

	return scanGameRun(row)
}

type ListGameRunsParams struct {
	Scope       string
	Status      string
	UserID      string
	PlayerEmail string
}

func (s *Store) ListGameRuns(ctx context.Context, params ListGameRunsParams) ([]domain.GameRunSummary, error) {
	scope := strings.TrimSpace(params.Scope)
	if scope == "" {
		scope = "host"
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
		  g.id::text,
		  g.template_id::text,
		  g.host_user_id::text,
		  g.word_set_id::text,
		  g.code,
		  g.name,
		  g.status,
		  g.scheduled_start_at,
		  g.started_at,
		  g.ended_at,
		  g.current_called_word_id::text,
		  g.winning_pattern,
		  g.created_at,
		  g.updated_at,
		  count(DISTINCT ap.id)::int AS allowed_player_count,
		  count(DISTINCT p.id)::int AS player_count
		FROM game_runs AS g
		LEFT JOIN allowed_players AS ap ON ap.game_run_id = g.id
		LEFT JOIN players AS p ON p.game_run_id = g.id
		WHERE ($1 = '' OR g.status = $1)
		  AND (
		    $2 = 'admin'
		    OR ($2 = 'host' AND g.host_user_id = $3)
		    OR (
		      $2 = 'player'
		      AND (
		        EXISTS (
		          SELECT 1
		          FROM players AS player_scope
		          WHERE player_scope.game_run_id = g.id
		            AND lower(player_scope.email) = lower($4)
		        )
		        OR EXISTS (
		          SELECT 1
		          FROM allowed_players AS allowed_scope
		          WHERE allowed_scope.game_run_id = g.id
		            AND lower(allowed_scope.email) = lower($4)
		        )
		      )
		    )
		  )
		GROUP BY g.id
		ORDER BY g.scheduled_start_at NULLS LAST, g.created_at DESC, g.id
	`, strings.TrimSpace(params.Status), scope, params.UserID, params.PlayerEmail)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]domain.GameRunSummary, 0)
	for rows.Next() {
		var summary domain.GameRunSummary
		run, err := scanGameRunWithCounts(rows, &summary.AllowedPlayerCount, &summary.PlayerCount)
		if err != nil {
			return nil, err
		}
		summary.GameRun = run
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return summaries, nil
}

type UpdateGameRunParams struct {
	GameRunID        string
	Name             *string
	Code             *string
	WordSetID        *string
	ScheduledStartAt *time.Time
	WinningPattern   *string
	ActorUserID      *string
}

func (s *Store) UpdateGameRun(ctx context.Context, params UpdateGameRunParams) (domain.GameRun, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.GameRun{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	run, err := getGameRunForUpdate(ctx, tx, params.GameRunID)
	if err != nil {
		return domain.GameRun{}, mapWriteError(err)
	}
	if !statusAllowed(run.Status, []string{"draft", "scheduled", "invites_sent", "lobby_open"}) {
		return domain.GameRun{}, ErrConflict
	}

	row := tx.QueryRow(ctx, `
		UPDATE game_runs
		SET name = COALESCE($2, name),
		    code = COALESCE($3, code),
		    word_set_id = COALESCE($4, word_set_id),
		    scheduled_start_at = COALESCE($5, scheduled_start_at),
		    winning_pattern = COALESCE($6, winning_pattern),
		    updated_at = now()
		WHERE id = $1
		RETURNING id::text, template_id::text, host_user_id::text, word_set_id::text, code, name, status, scheduled_start_at, started_at, ended_at, current_called_word_id::text, winning_pattern, created_at, updated_at
	`, params.GameRunID, params.Name, params.Code, params.WordSetID, params.ScheduledStartAt, params.WinningPattern)
	updated, err := scanGameRun(row)
	if err != nil {
		return domain.GameRun{}, mapWriteError(err)
	}

	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "game.updated", &updated.ID, map[string]any{
		"name":           updated.Name,
		"code":           updated.Code,
		"wordSetId":      updated.WordSetID,
		"winningPattern": updated.WinningPattern,
	}); err != nil {
		return domain.GameRun{}, err
	}
	if err := recordAuditEventInTx(ctx, tx, audit.Event{
		GameRunID:   &params.GameRunID,
		ActorUserID: params.ActorUserID,
		EventType:   "game.updated",
		EntityType:  "game_run",
		EntityID:    &updated.ID,
		Payload:     map[string]any{"code": updated.Code, "name": updated.Name},
	}); err != nil {
		return domain.GameRun{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.GameRun{}, mapWriteError(err)
	}

	return updated, nil
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

type GameStatusTransitionParams struct {
	GameRunID              string
	Status                 string
	StartedAt              *time.Time
	EndedAt                *time.Time
	ActorUserID            *string
	EventType              string
	AllowedCurrentStatuses []string
	MarkJoinedPlayersLive  bool
	Payload                map[string]any
}

func (s *Store) TransitionGameRunStatus(ctx context.Context, params GameStatusTransitionParams) (domain.GameRun, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.GameRun{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	run, err := getGameRunForUpdate(ctx, tx, params.GameRunID)
	if err != nil {
		return domain.GameRun{}, mapWriteError(err)
	}
	if len(params.AllowedCurrentStatuses) > 0 && !statusAllowed(run.Status, params.AllowedCurrentStatuses) {
		return domain.GameRun{}, ErrConflict
	}

	row := tx.QueryRow(ctx, `
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
	`, params.GameRunID, params.Status, params.StartedAt, params.EndedAt)
	run, err = scanGameRun(row)
	if err != nil {
		return domain.GameRun{}, mapWriteError(err)
	}

	if params.MarkJoinedPlayersLive {
		if _, err := tx.Exec(ctx, `
			UPDATE players
			SET state = 'playing',
			    updated_at = now()
			WHERE game_run_id = $1
			  AND state IN ('joined', 'waiting')
		`, params.GameRunID); err != nil {
			return domain.GameRun{}, mapWriteError(err)
		}
	}

	if params.EventType != "" {
		payload := params.Payload
		if payload == nil {
			payload = map[string]any{"status": run.Status}
		}
		if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, params.EventType, &run.ID, payload); err != nil {
			return domain.GameRun{}, err
		}
		if err := recordAuditEventInTx(ctx, tx, audit.Event{
			GameRunID:   &params.GameRunID,
			ActorUserID: params.ActorUserID,
			EventType:   params.EventType,
			EntityType:  "game_run",
			EntityID:    &run.ID,
			Payload:     payload,
		}); err != nil {
			return domain.GameRun{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.GameRun{}, mapWriteError(err)
	}

	return run, nil
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

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.AllowedPlayer{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return domain.AllowedPlayer{}, mapWriteError(err)
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO allowed_players (game_run_id, email, display_name, source)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, game_run_id::text, email, display_name, source, created_at
	`, params.GameRunID, params.Email, params.DisplayName, source)

	player, err := scanAllowedPlayer(row)
	if err != nil {
		return domain.AllowedPlayer{}, mapWriteError(err)
	}

	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "allowed_player.added", &player.ID, map[string]any{"email": player.Email, "displayName": player.DisplayName}); err != nil {
		return domain.AllowedPlayer{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.AllowedPlayer{}, mapWriteError(err)
	}

	return player, nil
}

type BulkAddAllowedPlayersParams struct {
	GameRunID string
	Players   []AddAllowedPlayerParams
}

func (s *Store) BulkAddAllowedPlayers(ctx context.Context, params BulkAddAllowedPlayersParams) ([]domain.AllowedPlayer, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return nil, mapWriteError(err)
	}

	players := make([]domain.AllowedPlayer, 0, len(params.Players))
	for _, input := range params.Players {
		source := input.Source
		if source == "" {
			source = "manual"
		}
		row := tx.QueryRow(ctx, `
			INSERT INTO allowed_players (game_run_id, email, display_name, source)
			VALUES ($1, $2, $3, $4)
			RETURNING id::text, game_run_id::text, email, display_name, source, created_at
		`, params.GameRunID, input.Email, input.DisplayName, source)
		player, err := scanAllowedPlayer(row)
		if err != nil {
			return nil, mapWriteError(err)
		}
		if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "allowed_player.added", &player.ID, map[string]any{"email": player.Email, "displayName": player.DisplayName}); err != nil {
			return nil, err
		}
		players = append(players, player)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, mapWriteError(err)
	}

	return players, nil
}

type UpdateAllowedPlayerParams struct {
	GameRunID       string
	AllowedPlayerID string
	Email           *string
	DisplayName     *string
}

func (s *Store) UpdateAllowedPlayer(ctx context.Context, params UpdateAllowedPlayerParams) (domain.AllowedPlayer, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.AllowedPlayer{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return domain.AllowedPlayer{}, mapWriteError(err)
	}

	row := tx.QueryRow(ctx, `
		UPDATE allowed_players
		SET email = COALESCE($3, email),
		    display_name = COALESCE($4, display_name)
		WHERE game_run_id = $1
		  AND id = $2
		RETURNING id::text, game_run_id::text, email, display_name, source, created_at
	`, params.GameRunID, params.AllowedPlayerID, params.Email, params.DisplayName)
	player, err := scanAllowedPlayer(row)
	if err != nil {
		return domain.AllowedPlayer{}, mapWriteError(err)
	}

	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "allowed_player.updated", &player.ID, map[string]any{"email": player.Email, "displayName": player.DisplayName}); err != nil {
		return domain.AllowedPlayer{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.AllowedPlayer{}, mapWriteError(err)
	}

	return player, nil
}

func (s *Store) DeleteAllowedPlayer(ctx context.Context, gameRunID, allowedPlayerID string) (domain.AllowedPlayer, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.AllowedPlayer{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, gameRunID); err != nil {
		return domain.AllowedPlayer{}, mapWriteError(err)
	}

	row := tx.QueryRow(ctx, `
		DELETE FROM allowed_players
		WHERE game_run_id = $1
		  AND id = $2
		RETURNING id::text, game_run_id::text, email, display_name, source, created_at
	`, gameRunID, allowedPlayerID)
	player, err := scanAllowedPlayer(row)
	if err != nil {
		return domain.AllowedPlayer{}, mapWriteError(err)
	}

	if _, err := insertOutboxEventInTx(ctx, tx, gameRunID, "allowed_player.deleted", &player.ID, map[string]any{"email": player.Email}); err != nil {
		return domain.AllowedPlayer{}, err
	}
	if err := tx.Commit(ctx); err != nil {
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

func (s *Store) GetOrCreateGameSettings(ctx context.Context, gameRunID string) (domain.GameRunSettings, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, gameRunID); err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}
	settings, err := ensureGameSettingsInTx(ctx, tx, gameRunID)
	if err != nil {
		return domain.GameRunSettings{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}

	return settings, nil
}

func (s *Store) UpdateGameSettings(ctx context.Context, params UpdateGameSettingsParams) (domain.GameRunSettings, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}
	if _, err := ensureGameSettingsInTx(ctx, tx, params.GameRunID); err != nil {
		return domain.GameRunSettings{}, err
	}

	row := tx.QueryRow(ctx, `
		UPDATE game_run_settings
		SET marking_mode = COALESCE($2, marking_mode),
		    allow_player_marking_mode_choice = COALESCE($3, allow_player_marking_mode_choice),
		    show_claim_readiness = COALESCE($4, show_claim_readiness),
		    voice_claim_mode = COALESCE($5, voice_claim_mode),
		    voice_claim_autoplay = COALESCE($6, voice_claim_autoplay),
		    caller_mode = COALESCE($7, caller_mode),
		    theme_mode = COALESCE($8, theme_mode),
		    theme_id = COALESCE($9, theme_id),
		    updated_at = now()
		WHERE game_run_id = $1
		RETURNING game_run_id::text, marking_mode, allow_player_marking_mode_choice, show_claim_readiness, voice_claim_mode, voice_claim_autoplay, caller_mode, theme_mode, theme_id::text, created_at, updated_at
	`, params.GameRunID, params.MarkingMode, params.AllowPlayerMarkingModeChoice, params.ShowClaimReadiness, params.VoiceClaimMode, params.VoiceClaimAutoplay, params.CallerMode, params.ThemeMode, params.ThemeID)
	settings, err := scanGameRunSettings(row)
	if err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}

	payload := gameSettingsPayload(settings)
	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "game.settings_updated", &params.GameRunID, payload); err != nil {
		return domain.GameRunSettings{}, err
	}
	if err := recordAuditEventInTx(ctx, tx, audit.Event{
		GameRunID:   &params.GameRunID,
		ActorUserID: params.ActorUserID,
		EventType:   "game.settings_updated",
		EntityType:  "game_run_settings",
		EntityID:    &params.GameRunID,
		Payload:     payload,
	}); err != nil {
		return domain.GameRunSettings{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}

	return settings, nil
}

func (s *Store) GetOrCreatePlayerPreferences(ctx context.Context, gameRunID, playerID string) (domain.PlayerPreferences, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.PlayerPreferences{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, gameRunID); err != nil {
		return domain.PlayerPreferences{}, mapWriteError(err)
	}
	player, err := getPlayerForUpdate(ctx, tx, playerID)
	if err != nil {
		return domain.PlayerPreferences{}, mapWriteError(err)
	}
	if player.GameRunID != gameRunID {
		return domain.PlayerPreferences{}, ErrNotFound
	}
	preferences, err := ensurePlayerPreferencesInTx(ctx, tx, playerID)
	if err != nil {
		return domain.PlayerPreferences{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.PlayerPreferences{}, mapWriteError(err)
	}

	return preferences, nil
}

func (s *Store) UpdatePlayerPreferences(ctx context.Context, params UpdatePlayerPreferencesParams) (domain.PlayerPreferences, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.PlayerPreferences{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return domain.PlayerPreferences{}, mapWriteError(err)
	}
	player, err := getPlayerForUpdate(ctx, tx, params.PlayerID)
	if err != nil {
		return domain.PlayerPreferences{}, mapWriteError(err)
	}
	if player.GameRunID != params.GameRunID {
		return domain.PlayerPreferences{}, ErrNotFound
	}
	if _, err := ensurePlayerPreferencesInTx(ctx, tx, params.PlayerID); err != nil {
		return domain.PlayerPreferences{}, err
	}

	row := tx.QueryRow(ctx, `
		UPDATE player_preferences
		SET marking_mode = $2,
		    updated_at = now()
		WHERE player_id = $1
		RETURNING player_id::text, marking_mode, created_at, updated_at
	`, params.PlayerID, params.MarkingMode)
	preferences, err := scanPlayerPreferences(row)
	if err != nil {
		return domain.PlayerPreferences{}, mapWriteError(err)
	}

	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "player.preferences_updated", &params.PlayerID, map[string]any{
		"playerId":    params.PlayerID,
		"markingMode": preferences.MarkingMode,
	}); err != nil {
		return domain.PlayerPreferences{}, err
	}
	if err := recordAuditEventInTx(ctx, tx, audit.Event{
		GameRunID:   &params.GameRunID,
		ActorUserID: params.ActorUserID,
		EventType:   "player.preferences_updated",
		EntityType:  "player_preferences",
		EntityID:    &params.PlayerID,
		Payload:     map[string]any{"playerId": params.PlayerID, "markingMode": preferences.MarkingMode},
	}); err != nil {
		return domain.PlayerPreferences{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.PlayerPreferences{}, mapWriteError(err)
	}

	return preferences, nil
}

func (s *Store) GetEffectivePlayerMarkingMode(ctx context.Context, gameRunID, playerID string) (string, error) {
	settings, err := s.GetOrCreateGameSettings(ctx, gameRunID)
	if err != nil {
		return "", err
	}
	preferences, err := s.GetOrCreatePlayerPreferences(ctx, gameRunID, playerID)
	if err != nil {
		return "", err
	}
	if settings.AllowPlayerMarkingModeChoice && preferences.MarkingMode != nil {
		return *preferences.MarkingMode, nil
	}

	return settings.MarkingMode, nil
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

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Player{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return domain.Player{}, mapWriteError(err)
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO players (game_run_id, user_id, email, display_name, player_icon, player_avatar_color, player_avatar_label, connection_state, state)
		VALUES ($1, $2, $3, $4, NULL, NULL, NULL, $5, $6)
		RETURNING id::text, game_run_id::text, user_id::text, email, display_name, player_icon, player_avatar_color, player_avatar_label, connection_state, state, joined_at, last_seen_at, created_at, updated_at
	`, params.GameRunID, params.UserID, params.Email, params.DisplayName, connectionState, state)

	player, err := scanPlayer(row)
	if err != nil {
		return domain.Player{}, mapWriteError(err)
	}

	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "player.joined", &player.ID, map[string]any{"playerId": player.ID, "email": player.Email, "displayName": player.DisplayName}); err != nil {
		return domain.Player{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Player{}, mapWriteError(err)
	}

	return player, nil
}

func (s *Store) GetPlayerByGameRunAndEmail(ctx context.Context, gameRunID, email string) (domain.Player, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, user_id::text, email, display_name, player_icon, player_avatar_color, player_avatar_label, connection_state, state, joined_at, last_seen_at, created_at, updated_at
		FROM players
		WHERE game_run_id = $1 AND lower(email) = lower($2)
	`, gameRunID, email)

	return scanPlayer(row)
}

func (s *Store) GetPlayer(ctx context.Context, gameRunID, playerID string) (domain.Player, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, user_id::text, email, display_name, player_icon, player_avatar_color, player_avatar_label, connection_state, state, joined_at, last_seen_at, created_at, updated_at
		FROM players
		WHERE game_run_id = $1 AND id = $2
	`, gameRunID, playerID)

	return scanPlayer(row)
}

type UpdatePlayerConnectionStateParams struct {
	GameRunID       string
	PlayerID        string
	ConnectionState string
	EventType       string
}

func (s *Store) UpdatePlayerConnectionState(ctx context.Context, params UpdatePlayerConnectionStateParams) (domain.Player, error) {
	connectionState := strings.TrimSpace(params.ConnectionState)
	if connectionState == "" {
		connectionState = "online"
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Player{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return domain.Player{}, mapWriteError(err)
	}

	row := tx.QueryRow(ctx, `
		UPDATE players
		SET connection_state = $3,
		    state = CASE
		      WHEN $3 = 'disconnected' AND state <> 'confirmed_winner' THEN 'disconnected'
		      WHEN $3 = 'online' AND state = 'disconnected' AND game_runs.status IN ('live', 'paused') THEN 'playing'
		      WHEN $3 = 'online' AND state = 'disconnected' THEN 'joined'
		      ELSE state
		    END,
		    last_seen_at = now(),
		    updated_at = now()
		FROM game_runs
		WHERE players.game_run_id = $1
		  AND players.id = $2
		  AND game_runs.id = players.game_run_id
		RETURNING players.id::text, players.game_run_id::text, players.user_id::text, players.email, players.display_name, players.player_icon, players.player_avatar_color, players.player_avatar_label, players.connection_state, players.state, players.joined_at, players.last_seen_at, players.created_at, players.updated_at
	`, params.GameRunID, params.PlayerID, connectionState)
	player, err := scanPlayer(row)
	if err != nil {
		return domain.Player{}, mapWriteError(err)
	}

	if params.EventType != "" {
		if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, params.EventType, &player.ID, map[string]any{
			"playerId":        player.ID,
			"connectionState": player.ConnectionState,
		}); err != nil {
			return domain.Player{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.Player{}, mapWriteError(err)
	}

	return player, nil
}

type MarkStalePlayersDisconnectedParams struct {
	Cutoff time.Time
	Limit  int
}

func (s *Store) MarkStalePlayersDisconnected(ctx context.Context, params MarkStalePlayersDisconnectedParams) ([]domain.Player, error) {
	limit := params.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
		SELECT p.id::text, p.game_run_id::text, p.user_id::text, p.email, p.display_name, p.player_icon, p.player_avatar_color, p.player_avatar_label, p.connection_state, p.state, p.joined_at, p.last_seen_at, p.created_at, p.updated_at
		FROM players AS p
		JOIN game_runs AS g ON g.id = p.game_run_id
		WHERE p.connection_state = 'online'
		  AND p.last_seen_at < $1
		  AND g.status IN ('lobby_open', 'live', 'paused')
		ORDER BY p.last_seen_at ASC, p.id ASC
		LIMIT $2
		FOR UPDATE OF p
	`, params.Cutoff, limit)
	if err != nil {
		return nil, mapWriteError(err)
	}

	candidates := make([]domain.Player, 0)
	for rows.Next() {
		player, err := scanPlayer(rows)
		if err != nil {
			rows.Close()
			return nil, err
		}
		candidates = append(candidates, player)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	updatedPlayers := make([]domain.Player, 0, len(candidates))
	for _, candidate := range candidates {
		if _, err := getGameRunForUpdate(ctx, tx, candidate.GameRunID); err != nil {
			return nil, mapWriteError(err)
		}

		row := tx.QueryRow(ctx, `
			UPDATE players
			SET connection_state = 'disconnected',
			    state = CASE WHEN state <> 'confirmed_winner' THEN 'disconnected' ELSE state END,
			    updated_at = now()
			WHERE id = $1
			RETURNING id::text, game_run_id::text, user_id::text, email, display_name, player_icon, player_avatar_color, player_avatar_label, connection_state, state, joined_at, last_seen_at, created_at, updated_at
		`, candidate.ID)
		player, err := scanPlayer(row)
		if err != nil {
			return nil, mapWriteError(err)
		}

		if _, err := insertOutboxEventInTx(ctx, tx, player.GameRunID, "player.disconnected", &player.ID, map[string]any{
			"playerId":        player.ID,
			"connectionState": player.ConnectionState,
			"lastSeenAt":      player.LastSeenAt,
		}); err != nil {
			return nil, err
		}
		updatedPlayers = append(updatedPlayers, player)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, mapWriteError(err)
	}

	return updatedPlayers, nil
}

func (s *Store) CreateContentGenerationJob(ctx context.Context, params CreateContentGenerationJobParams) (domain.ContentGenerationJob, error) {
	status := strings.TrimSpace(params.Status)
	if status == "" {
		status = "running"
	}
	provider := strings.TrimSpace(params.Provider)
	if provider == "" {
		provider = "unknown"
	}
	jobType := strings.TrimSpace(params.JobType)
	if jobType == "" {
		jobType = "game_prep"
	}

	row := s.pool.QueryRow(ctx, `
		INSERT INTO content_generation_jobs (game_run_id, job_type, status, provider)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, game_run_id::text, job_type, status, provider, error_message, retry_count, created_at, updated_at
	`, params.GameRunID, jobType, status, provider)

	return scanContentGenerationJob(row)
}

func (s *Store) UpdateContentGenerationJob(ctx context.Context, params UpdateContentGenerationJobParams) (domain.ContentGenerationJob, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE content_generation_jobs
		SET status = $2,
		    provider = COALESCE($3, provider),
		    error_message = $4,
		    retry_count = CASE WHEN $2 = 'failed' THEN retry_count + 1 ELSE retry_count END,
		    updated_at = now()
		WHERE id = $1
		RETURNING id::text, game_run_id::text, job_type, status, provider, error_message, retry_count, created_at, updated_at
	`, params.JobID, params.Status, params.Provider, params.ErrorMessage)

	return scanContentGenerationJob(row)
}

func (s *Store) GetGeneratedGameContent(ctx context.Context, gameRunID string) (domain.GeneratedGameContent, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, generation_job_id::text, status, topic, summary, generated_words, current_words, caller_style, theme_prompt, review_window_opens_at, review_window_closes_at, locked_at, locked_word_set_id::text, generation_provider, generation_error, created_at, updated_at
		FROM generated_game_content
		WHERE game_run_id = $1
	`, gameRunID)

	return scanGeneratedGameContent(row)
}

func (s *Store) UpsertGeneratedGameContent(ctx context.Context, params UpsertGeneratedGameContentParams) (domain.GeneratedGameContent, error) {
	generatedWordsJSON, err := json.Marshal(params.GeneratedWords)
	if err != nil {
		return domain.GeneratedGameContent{}, fmt.Errorf("marshal generated words: %w", err)
	}
	currentWordsJSON, err := json.Marshal(params.CurrentWords)
	if err != nil {
		return domain.GeneratedGameContent{}, fmt.Errorf("marshal current words: %w", err)
	}
	status := strings.TrimSpace(params.Status)
	if status == "" {
		status = "generated"
	}
	provider := strings.TrimSpace(params.GenerationProvider)
	if provider == "" {
		provider = "unknown"
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.GeneratedGameContent{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return domain.GeneratedGameContent{}, mapWriteError(err)
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO generated_game_content (
			game_run_id,
			generation_job_id,
			status,
			topic,
			summary,
			generated_words,
			current_words,
			caller_style,
			theme_prompt,
			review_window_opens_at,
			review_window_closes_at,
			generation_provider,
			generation_error
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (game_run_id) DO UPDATE
		SET generation_job_id = EXCLUDED.generation_job_id,
		    status = EXCLUDED.status,
		    topic = EXCLUDED.topic,
		    summary = EXCLUDED.summary,
		    generated_words = EXCLUDED.generated_words,
		    current_words = EXCLUDED.current_words,
		    caller_style = EXCLUDED.caller_style,
		    theme_prompt = EXCLUDED.theme_prompt,
		    review_window_opens_at = EXCLUDED.review_window_opens_at,
		    review_window_closes_at = EXCLUDED.review_window_closes_at,
		    generation_provider = EXCLUDED.generation_provider,
		    generation_error = EXCLUDED.generation_error,
		    updated_at = now()
		WHERE generated_game_content.locked_at IS NULL
		RETURNING id::text, game_run_id::text, generation_job_id::text, status, topic, summary, generated_words, current_words, caller_style, theme_prompt, review_window_opens_at, review_window_closes_at, locked_at, locked_word_set_id::text, generation_provider, generation_error, created_at, updated_at
	`, params.GameRunID, params.GenerationJobID, status, params.Topic, params.Summary, generatedWordsJSON, currentWordsJSON, params.CallerStyle, params.ThemePrompt, params.ReviewWindowOpensAt, params.ReviewWindowClosesAt, provider, params.GenerationError)
	content, err := scanGeneratedGameContent(row)
	if err != nil {
		return domain.GeneratedGameContent{}, mapWriteError(err)
	}

	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "content.generated", &content.ID, map[string]any{
		"gameRunId": params.GameRunID,
		"status":    content.Status,
		"provider":  content.GenerationProvider,
	}); err != nil {
		return domain.GeneratedGameContent{}, err
	}
	if err := recordAuditEventInTx(ctx, tx, audit.Event{
		GameRunID:   &params.GameRunID,
		ActorUserID: params.ActorUserID,
		EventType:   "content.generated",
		EntityType:  "generated_game_content",
		EntityID:    &content.ID,
		Payload:     map[string]any{"provider": content.GenerationProvider, "wordCount": len(content.CurrentWords)},
	}); err != nil {
		return domain.GeneratedGameContent{}, err
	}

	if params.GenerationJobID != nil {
		if _, err := tx.Exec(ctx, `
			UPDATE content_generation_jobs
			SET status = 'succeeded',
			    provider = $2,
			    error_message = NULL,
			    updated_at = now()
			WHERE id = $1
		`, *params.GenerationJobID, provider); err != nil {
			return domain.GeneratedGameContent{}, mapWriteError(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.GeneratedGameContent{}, mapWriteError(err)
	}

	return content, nil
}

func (s *Store) UpdateGeneratedGameContent(ctx context.Context, params UpdateGeneratedGameContentParams) (domain.GeneratedGameContent, error) {
	wordsJSON, err := json.Marshal(params.Words)
	if err != nil {
		return domain.GeneratedGameContent{}, fmt.Errorf("marshal edited words: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.GeneratedGameContent{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	content, err := getGeneratedGameContentForUpdate(ctx, tx, params.GameRunID)
	if err != nil {
		return domain.GeneratedGameContent{}, mapWriteError(err)
	}
	if content.LockedAt != nil || content.Status == "locked" {
		return domain.GeneratedGameContent{}, ErrConflict
	}

	row := tx.QueryRow(ctx, `
		UPDATE generated_game_content
		SET topic = COALESCE($2, topic),
		    summary = COALESCE($3, summary),
		    current_words = CASE WHEN $4 THEN $5 ELSE current_words END,
		    caller_style = COALESCE($6, caller_style),
		    status = 'edited',
		    updated_at = now()
		WHERE game_run_id = $1
		  AND locked_at IS NULL
		RETURNING id::text, game_run_id::text, generation_job_id::text, status, topic, summary, generated_words, current_words, caller_style, theme_prompt, review_window_opens_at, review_window_closes_at, locked_at, locked_word_set_id::text, generation_provider, generation_error, created_at, updated_at
	`, params.GameRunID, params.Topic, params.Summary, params.HasWordPatch, wordsJSON, params.CallerStyle)
	content, err = scanGeneratedGameContent(row)
	if err != nil {
		return domain.GeneratedGameContent{}, mapWriteError(err)
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO game_run_content_reviews (game_run_id, content_id, actor_user_id, edited_topic, edited_summary, edited_words, caller_style)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, params.GameRunID, content.ID, params.ActorUserID, params.Topic, params.Summary, wordsJSON, params.CallerStyle); err != nil {
		return domain.GeneratedGameContent{}, mapWriteError(err)
	}
	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "content.edited", &content.ID, map[string]any{
		"gameRunId": params.GameRunID,
		"wordCount": len(content.CurrentWords),
	}); err != nil {
		return domain.GeneratedGameContent{}, err
	}
	if err := recordAuditEventInTx(ctx, tx, audit.Event{
		GameRunID:   &params.GameRunID,
		ActorUserID: params.ActorUserID,
		EventType:   "content.edited",
		EntityType:  "generated_game_content",
		EntityID:    &content.ID,
		Payload:     map[string]any{"wordCount": len(content.CurrentWords)},
	}); err != nil {
		return domain.GeneratedGameContent{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.GeneratedGameContent{}, mapWriteError(err)
	}

	return content, nil
}

func (s *Store) LockGeneratedGameContent(ctx context.Context, params LockGeneratedGameContentParams) (LockGeneratedGameContentResult, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return LockGeneratedGameContentResult{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return LockGeneratedGameContentResult{}, mapWriteError(err)
	}
	content, err := getGeneratedGameContentForUpdate(ctx, tx, params.GameRunID)
	if err != nil {
		return LockGeneratedGameContentResult{}, mapWriteError(err)
	}
	if content.LockedAt != nil || content.Status == "locked" {
		wordSet := domain.WordSetWithWords{}
		if content.LockedWordSetID != nil {
			wordSet.WordSet, _ = getWordSetForUpdate(ctx, tx, *content.LockedWordSetID)
		}
		return LockGeneratedGameContentResult{Content: content, WordSet: wordSet}, nil
	}

	wordSetRow := tx.QueryRow(ctx, `
		INSERT INTO word_sets (name, status, source, created_by_user_id, approved_by_user_id, approved_at)
		VALUES ($1, 'approved', 'ai_generated', $2, $2, now())
		RETURNING id::text, name, status, source, created_at, updated_at
	`, content.Topic+" Word Set", params.ActorUserID)
	wordSet, err := scanWordSet(wordSetRow)
	if err != nil {
		return LockGeneratedGameContentResult{}, mapWriteError(err)
	}

	words := make([]domain.WordSetWord, 0, len(content.CurrentWords))
	for index, word := range content.CurrentWords {
		wordRow := tx.QueryRow(ctx, `
			INSERT INTO word_set_words (word_set_id, word, sort_order, is_active)
			VALUES ($1, $2, $3, true)
			RETURNING id::text, word_set_id::text, word, sort_order, is_active, created_at
		`, wordSet.ID, word, index+1)
		createdWord, err := scanWordSetWord(wordRow)
		if err != nil {
			return LockGeneratedGameContentResult{}, mapWriteError(err)
		}
		words = append(words, createdWord)
	}

	contentRow := tx.QueryRow(ctx, `
		UPDATE generated_game_content
		SET status = 'locked',
		    locked_at = now(),
		    locked_word_set_id = $2,
		    updated_at = now()
		WHERE game_run_id = $1
		RETURNING id::text, game_run_id::text, generation_job_id::text, status, topic, summary, generated_words, current_words, caller_style, theme_prompt, review_window_opens_at, review_window_closes_at, locked_at, locked_word_set_id::text, generation_provider, generation_error, created_at, updated_at
	`, params.GameRunID, wordSet.ID)
	content, err = scanGeneratedGameContent(contentRow)
	if err != nil {
		return LockGeneratedGameContentResult{}, mapWriteError(err)
	}

	runRow := tx.QueryRow(ctx, `
		UPDATE game_runs
		SET word_set_id = $2,
		    status = CASE WHEN status IN ('draft', 'content_generating', 'content_review') THEN 'scheduled' ELSE status END,
		    updated_at = now()
		WHERE id = $1
		RETURNING id::text, template_id::text, host_user_id::text, word_set_id::text, code, name, status, scheduled_start_at, started_at, ended_at, current_called_word_id::text, winning_pattern, created_at, updated_at
	`, params.GameRunID, wordSet.ID)
	run, err := scanGameRun(runRow)
	if err != nil {
		return LockGeneratedGameContentResult{}, mapWriteError(err)
	}

	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "content.locked", &content.ID, map[string]any{
		"gameRunId": params.GameRunID,
		"wordSetId": wordSet.ID,
		"wordCount": len(content.CurrentWords),
	}); err != nil {
		return LockGeneratedGameContentResult{}, err
	}
	if err := recordAuditEventInTx(ctx, tx, audit.Event{
		GameRunID:   &params.GameRunID,
		ActorUserID: params.ActorUserID,
		EventType:   "content.locked",
		EntityType:  "generated_game_content",
		EntityID:    &content.ID,
		Payload:     map[string]any{"wordSetId": wordSet.ID, "wordCount": len(content.CurrentWords)},
	}); err != nil {
		return LockGeneratedGameContentResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return LockGeneratedGameContentResult{}, mapWriteError(err)
	}

	return LockGeneratedGameContentResult{
		Content: content,
		WordSet: domain.WordSetWithWords{
			WordSet: wordSet,
			Words:   words,
		},
		GameRun: run,
	}, nil
}

func (s *Store) CreateGameCallDeck(ctx context.Context, params CreateGameCallDeckParams) ([]domain.GameCallDeckItem, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, mapWriteError(err)
	}
	defer tx.Rollback(ctx)
	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return nil, mapWriteError(err)
	}
	var existing int
	if err := tx.QueryRow(ctx, `SELECT count(*)::int FROM game_call_deck WHERE game_run_id = $1`, params.GameRunID).Scan(&existing); err != nil {
		return nil, mapWriteError(err)
	}
	if existing > 0 {
		items, err := listGameCallDeckInTx(ctx, tx, params.GameRunID)
		if err != nil {
			return nil, err
		}
		return items, tx.Commit(ctx)
	}
	items := make([]domain.GameCallDeckItem, 0, len(params.Words))
	for index, word := range params.Words {
		row := tx.QueryRow(ctx, `
			INSERT INTO game_call_deck (game_run_id, word_set_word_id, word, sequence, shuffle_seed, shuffle_version)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id::text, game_run_id::text, word_set_word_id::text, word, sequence, shuffle_seed, shuffle_version, locked_at, called_word_id::text, created_at
		`, params.GameRunID, word.ID, word.Word, index+1, params.ShuffleSeed, params.ShuffleVersion)
		item, err := scanGameCallDeckItem(row)
		if err != nil {
			return nil, mapWriteError(err)
		}
		items = append(items, item)
	}
	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "call_deck.locked", nil, map[string]any{
		"gameRunId":      params.GameRunID,
		"wordCount":      len(items),
		"shuffleSeed":    params.ShuffleSeed,
		"shuffleVersion": params.ShuffleVersion,
	}); err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, mapWriteError(err)
	}
	return items, nil
}

func (s *Store) ListGameCallDeck(ctx context.Context, gameRunID string) ([]domain.GameCallDeckItem, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, game_run_id::text, word_set_word_id::text, word, sequence, shuffle_seed, shuffle_version, locked_at, called_word_id::text, created_at
		FROM game_call_deck
		WHERE game_run_id = $1
		ORDER BY sequence ASC
	`, gameRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGameCallDeckRows(rows)
}

func (s *Store) CreateCalledWordFromDeck(ctx context.Context, params CreateCalledWordParams) (domain.CalledWord, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)
	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}
	deckRow := tx.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, word_set_word_id::text, word, sequence, shuffle_seed, shuffle_version, locked_at, called_word_id::text, created_at
		FROM game_call_deck
		WHERE game_run_id = $1 AND called_word_id IS NULL
		ORDER BY sequence ASC
		LIMIT 1
		FOR UPDATE
	`, params.GameRunID)
	item, err := scanGameCallDeckItem(deckRow)
	if err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}
	row := tx.QueryRow(ctx, `
		INSERT INTO called_words (game_run_id, word_set_word_id, word, called_by_user_id, sequence)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, game_run_id::text, word_set_word_id::text, word, called_by_user_id::text, sequence, called_at, created_at
	`, params.GameRunID, item.WordSetWordID, item.Word, params.CalledByUserID, item.Sequence)
	calledWord, err := scanCalledWord(row)
	if err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE game_call_deck SET called_word_id = $2 WHERE id = $1
	`, item.ID, calledWord.ID); err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}
	if _, err := tx.Exec(ctx, `
		UPDATE game_runs SET current_called_word_id = $2, updated_at = now() WHERE id = $1
	`, params.GameRunID, calledWord.ID); err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}
	asset, assetErr := getCallerAssetByDeckItemInTx(ctx, tx, item.ID)
	payload := map[string]any{"word": calledWord.Word, "sequence": calledWord.Sequence}
	if assetErr == nil {
		calledWord.CallerAsset = &asset
		payload["callerAssetStatus"] = asset.Status
		payload["callerLine"] = asset.Line
		payload["callerAudioUrl"] = asset.AudioURL
	}
	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "word.called", &calledWord.ID, payload); err != nil {
		return domain.CalledWord{}, err
	}
	if _, err := autoMarkInTx(ctx, tx, params.GameRunID, &calledWord.Word, &calledWord.ID); err != nil {
		return domain.CalledWord{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}
	return calledWord, nil
}

func (s *Store) UpsertCallerAsset(ctx context.Context, params CreateCallerAssetParams) (domain.CallerAsset, error) {
	provider := strings.TrimSpace(params.Provider)
	if provider == "" {
		provider = "unknown"
	}
	status := strings.TrimSpace(params.Status)
	if status == "" {
		status = "pending"
	}
	line := strings.TrimSpace(params.Line)
	if line == "" {
		line = "Next word is " + strings.TrimSpace(params.Word) + "."
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO caller_assets (game_run_id, call_deck_item_id, word, sequence, line, audio_url, storage_key, voice_name, provider, status, error_reason)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (call_deck_item_id) DO UPDATE
		SET word = EXCLUDED.word,
		    sequence = EXCLUDED.sequence,
		    line = EXCLUDED.line,
		    audio_url = EXCLUDED.audio_url,
		    storage_key = EXCLUDED.storage_key,
		    voice_name = EXCLUDED.voice_name,
		    provider = EXCLUDED.provider,
		    status = EXCLUDED.status,
		    error_reason = EXCLUDED.error_reason,
		    updated_at = now()
		RETURNING id::text, game_run_id::text, call_deck_item_id::text, word, sequence, line, audio_url, storage_key, voice_name, provider, status, error_reason, created_at, updated_at
	`, params.GameRunID, params.CallDeckItemID, params.Word, params.Sequence, line, params.AudioURL, params.StorageKey, params.VoiceName, provider, status, params.ErrorReason)
	return scanCallerAsset(row)
}

func (s *Store) ListCallerAssets(ctx context.Context, gameRunID string) ([]domain.CallerAsset, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, game_run_id::text, call_deck_item_id::text, word, sequence, line, audio_url, storage_key, voice_name, provider, status, error_reason, created_at, updated_at
		FROM caller_assets
		WHERE game_run_id = $1
		ORDER BY sequence ASC
	`, gameRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	assets := make([]domain.CallerAsset, 0)
	for rows.Next() {
		asset, err := scanCallerAsset(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	return assets, rows.Err()
}

func (s *Store) CreateDeliveryBatch(ctx context.Context, params CreateDeliveryBatchParams) (domain.DeliveryBatch, []domain.DeliveryAttempt, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.DeliveryBatch{}, nil, mapWriteError(err)
	}
	defer tx.Rollback(ctx)
	batchRow := tx.QueryRow(ctx, `
		INSERT INTO delivery_batches (game_run_id, channel, purpose, status)
		VALUES ($1, $2, $3, 'sent')
		RETURNING id::text, game_run_id::text, channel, purpose, status, created_at, updated_at
	`, params.GameRunID, params.Channel, params.Purpose)
	batch, err := scanDeliveryBatch(batchRow)
	if err != nil {
		return domain.DeliveryBatch{}, nil, mapWriteError(err)
	}
	attempts := make([]domain.DeliveryAttempt, 0, len(params.Attempts))
	for _, attempt := range params.Attempts {
		status := attempt.Status
		if status == "" {
			status = "sent"
		}
		row := tx.QueryRow(ctx, `
			INSERT INTO delivery_attempts (batch_id, game_run_id, channel, purpose, recipient_email, recipient_user_id, subject, template_key, body_preview, link_url, game_code, status, error_reason, sent_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, CASE WHEN $12 = 'sent' THEN now() ELSE NULL END)
			RETURNING id::text, batch_id::text, game_run_id::text, channel, purpose, recipient_email, recipient_user_id::text, subject, template_key, body_preview, link_url, game_code, status, error_reason, sent_at, created_at, updated_at
		`, batch.ID, params.GameRunID, params.Channel, params.Purpose, attempt.RecipientEmail, attempt.RecipientUserID, attempt.Subject, attempt.TemplateKey, attempt.BodyPreview, attempt.LinkURL, attempt.GameCode, status, attempt.ErrorReason)
		created, err := scanDeliveryAttempt(row)
		if err != nil {
			return domain.DeliveryBatch{}, nil, mapWriteError(err)
		}
		attempts = append(attempts, created)
	}
	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "delivery.batch_created", &batch.ID, map[string]any{"channel": params.Channel, "purpose": params.Purpose, "attempts": len(attempts)}); err != nil {
		return domain.DeliveryBatch{}, nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.DeliveryBatch{}, nil, mapWriteError(err)
	}
	return batch, attempts, nil
}

func (s *Store) ListDeliveryAttempts(ctx context.Context, gameRunID string) ([]domain.DeliveryAttempt, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, batch_id::text, game_run_id::text, channel, purpose, recipient_email, recipient_user_id::text, subject, template_key, body_preview, link_url, game_code, status, error_reason, sent_at, created_at, updated_at
		FROM delivery_attempts
		WHERE game_run_id = $1
		ORDER BY created_at DESC, id
	`, gameRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	attempts := make([]domain.DeliveryAttempt, 0)
	for rows.Next() {
		attempt, err := scanDeliveryAttempt(rows)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, attempt)
	}
	return attempts, rows.Err()
}

func (s *Store) RetryDeliveryAttempt(ctx context.Context, deliveryID string) (domain.DeliveryAttempt, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE delivery_attempts
		SET status = 'sent', error_reason = NULL, sent_at = now(), updated_at = now()
		WHERE id = $1
		RETURNING id::text, batch_id::text, game_run_id::text, channel, purpose, recipient_email, recipient_user_id::text, subject, template_key, body_preview, link_url, game_code, status, error_reason, sent_at, created_at, updated_at
	`, deliveryID)
	return scanDeliveryAttempt(row)
}

func (s *Store) UpdatePlayerProfile(ctx context.Context, params UpdatePlayerProfileParams) (domain.Player, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.Player{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)
	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return domain.Player{}, mapWriteError(err)
	}
	row := tx.QueryRow(ctx, `
		UPDATE players
		SET player_icon = $3,
		    player_avatar_color = $4,
		    player_avatar_label = $5,
		    updated_at = now()
		WHERE game_run_id = $1 AND id = $2
		RETURNING id::text, game_run_id::text, user_id::text, email, display_name, player_icon, player_avatar_color, player_avatar_label, connection_state, state, joined_at, last_seen_at, created_at, updated_at
	`, params.GameRunID, params.PlayerID, params.Icon, params.AvatarColor, params.AvatarLabel)
	player, err := scanPlayer(row)
	if err != nil {
		return domain.Player{}, mapWriteError(err)
	}
	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "player.profile_updated", &player.ID, map[string]any{"playerId": player.ID, "icon": player.Icon, "avatarColor": player.AvatarColor, "avatarLabel": player.AvatarLabel}); err != nil {
		return domain.Player{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Player{}, mapWriteError(err)
	}
	return player, nil
}

func (s *Store) CreateThemeGenerationJob(ctx context.Context, gameRunID *string, prompt, provider string) (domain.ThemeGenerationJob, error) {
	row := s.pool.QueryRow(ctx, `
		INSERT INTO theme_generation_jobs (game_run_id, status, provider, prompt)
		VALUES ($1, 'running', $2, $3)
		RETURNING id::text, game_run_id::text, status, provider, prompt, error_message, created_at, updated_at
	`, gameRunID, provider, prompt)
	return scanThemeGenerationJob(row)
}

func (s *Store) CreateTheme(ctx context.Context, params CreateThemeParams) (domain.Theme, error) {
	tokensJSON, err := json.Marshal(params.Tokens)
	if err != nil {
		return domain.Theme{}, err
	}
	provider := params.Provider
	if provider == "" {
		provider = "unknown"
	}
	row := s.pool.QueryRow(ctx, `
		INSERT INTO themes (game_run_id, generation_job_id, name, summary, tokens, provider, created_by_user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id::text, game_run_id::text, generation_job_id::text, name, summary, tokens, status, provider, created_by_user_id::text, approved_by_user_id::text, approved_at, rejected_at, created_at, updated_at
	`, params.GameRunID, params.GenerationJobID, params.Name, params.Summary, tokensJSON, provider, params.CreatedByUserID)
	theme, err := scanTheme(row)
	if err != nil {
		return domain.Theme{}, mapWriteError(err)
	}
	if params.GenerationJobID != nil {
		_, _ = s.pool.Exec(ctx, `UPDATE theme_generation_jobs SET status = 'succeeded', provider = $2, updated_at = now() WHERE id = $1`, *params.GenerationJobID, provider)
	}
	return theme, nil
}

func (s *Store) GetTheme(ctx context.Context, themeID string) (domain.Theme, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, generation_job_id::text, name, summary, tokens, status, provider, created_by_user_id::text, approved_by_user_id::text, approved_at, rejected_at, created_at, updated_at
		FROM themes WHERE id = $1
	`, themeID)
	return scanTheme(row)
}

func (s *Store) UpdateTheme(ctx context.Context, themeID string, tokens domain.ThemeTokens, name, summary *string) (domain.Theme, error) {
	tokensJSON, err := json.Marshal(tokens)
	if err != nil {
		return domain.Theme{}, err
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE themes
		SET name = COALESCE($2, name), summary = COALESCE($3, summary), tokens = $4, updated_at = now()
		WHERE id = $1 AND status = 'draft'
		RETURNING id::text, game_run_id::text, generation_job_id::text, name, summary, tokens, status, provider, created_by_user_id::text, approved_by_user_id::text, approved_at, rejected_at, created_at, updated_at
	`, themeID, name, summary, tokensJSON)
	return scanTheme(row)
}

func (s *Store) SetThemeApproval(ctx context.Context, themeID string, actorUserID *string, approved bool) (domain.Theme, error) {
	status := "approved"
	if !approved {
		status = "rejected"
	}
	row := s.pool.QueryRow(ctx, `
		UPDATE themes
		SET status = $2,
		    approved_by_user_id = CASE WHEN $2 = 'approved' THEN $3 ELSE approved_by_user_id END,
		    approved_at = CASE WHEN $2 = 'approved' THEN now() ELSE approved_at END,
		    rejected_at = CASE WHEN $2 = 'rejected' THEN now() ELSE rejected_at END,
		    updated_at = now()
		WHERE id = $1
		RETURNING id::text, game_run_id::text, generation_job_id::text, name, summary, tokens, status, provider, created_by_user_id::text, approved_by_user_id::text, approved_at, rejected_at, created_at, updated_at
	`, themeID, status, actorUserID)
	theme, err := scanTheme(row)
	if err != nil {
		return domain.Theme{}, mapWriteError(err)
	}
	_, _ = s.pool.Exec(ctx, `INSERT INTO theme_approvals (theme_id, game_run_id, actor_user_id, status) VALUES ($1, $2, $3, $4)`, theme.ID, theme.GameRunID, actorUserID, status)
	return theme, nil
}

func (s *Store) ApplyThemeToGame(ctx context.Context, gameRunID, themeID string, actorUserID *string) (domain.GameRunSettings, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)
	if _, err := getGameRunForUpdate(ctx, tx, gameRunID); err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}
	theme, err := getThemeForUpdate(ctx, tx, themeID)
	if err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}
	if theme.Status != "approved" {
		return domain.GameRunSettings{}, ErrConflict
	}
	if _, err := ensureGameSettingsInTx(ctx, tx, gameRunID); err != nil {
		return domain.GameRunSettings{}, err
	}
	row := tx.QueryRow(ctx, `
		UPDATE game_run_settings
		SET theme_mode = 'ai_generated', theme_id = $2, updated_at = now()
		WHERE game_run_id = $1
		RETURNING game_run_id::text, marking_mode, allow_player_marking_mode_choice, show_claim_readiness, voice_claim_mode, voice_claim_autoplay, caller_mode, theme_mode, theme_id::text, created_at, updated_at
	`, gameRunID, themeID)
	settings, err := scanGameRunSettings(row)
	if err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}
	if _, err := insertOutboxEventInTx(ctx, tx, gameRunID, "theme.applied", &theme.ID, map[string]any{"themeId": theme.ID, "name": theme.Name}); err != nil {
		return domain.GameRunSettings{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}
	return settings, nil
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

func (s *Store) ListWordSets(ctx context.Context, approvedOnly bool) ([]domain.WordSet, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, name, status, source, created_at, updated_at
		FROM word_sets
		WHERE ($1 = false OR status = 'approved')
		ORDER BY name ASC, created_at DESC
	`, approvedOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	wordSets := make([]domain.WordSet, 0)
	for rows.Next() {
		wordSet, err := scanWordSet(rows)
		if err != nil {
			return nil, err
		}
		wordSets = append(wordSets, wordSet)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return wordSets, nil
}

func (s *Store) GetWordSet(ctx context.Context, wordSetID string) (domain.WordSet, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id::text, name, status, source, created_at, updated_at
		FROM word_sets
		WHERE id = $1
	`, wordSetID)

	return scanWordSet(row)
}

func (s *Store) ListWordSetWordsForManagement(ctx context.Context, wordSetID string) ([]domain.WordSetWord, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, word_set_id::text, word, sort_order, is_active, created_at
		FROM word_set_words
		WHERE word_set_id = $1
		ORDER BY sort_order ASC, created_at ASC, id ASC
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

type CreateWordSetParams struct {
	Name            string
	Status          string
	Source          string
	CreatedByUserID *string
	Words           []CreateWordSetWordParams
}

func (s *Store) CreateWordSet(ctx context.Context, params CreateWordSetParams) (domain.WordSetWithWords, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.WordSetWithWords{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `
		INSERT INTO word_sets (name, status, source, created_by_user_id, approved_by_user_id, approved_at)
		VALUES (
		  $1,
		  $2,
		  $3,
		  $4,
		  CASE WHEN $2 = 'approved' THEN $4::uuid ELSE NULL END,
		  CASE WHEN $2 = 'approved' THEN now() ELSE NULL END
		)
		RETURNING id::text, name, status, source, created_at, updated_at
	`, params.Name, params.Status, params.Source, params.CreatedByUserID)
	wordSet, err := scanWordSet(row)
	if err != nil {
		return domain.WordSetWithWords{}, mapWriteError(err)
	}

	words := make([]domain.WordSetWord, 0, len(params.Words))
	for index, input := range params.Words {
		sortOrder := input.SortOrder
		if sortOrder <= 0 {
			sortOrder = index + 1
		}
		wordRow := tx.QueryRow(ctx, `
			INSERT INTO word_set_words (word_set_id, word, sort_order, is_active)
			VALUES ($1, $2, $3, $4)
			RETURNING id::text, word_set_id::text, word, sort_order, is_active, created_at
		`, wordSet.ID, input.Word, sortOrder, input.IsActive)
		word, err := scanWordSetWord(wordRow)
		if err != nil {
			return domain.WordSetWithWords{}, mapWriteError(err)
		}
		words = append(words, word)
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.WordSetWithWords{}, mapWriteError(err)
	}

	return domain.WordSetWithWords{WordSet: wordSet, Words: words}, nil
}

type UpdateWordSetParams struct {
	WordSetID string
	Name      *string
	Status    *string
	Source    *string
}

func (s *Store) UpdateWordSet(ctx context.Context, params UpdateWordSetParams) (domain.WordSet, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE word_sets
		SET name = COALESCE($2, name),
		    status = COALESCE($3, status),
		    source = COALESCE($4, source),
		    approved_at = CASE WHEN COALESCE($3, status) = 'approved' AND approved_at IS NULL THEN now() ELSE approved_at END,
		    updated_at = now()
		WHERE id = $1
		RETURNING id::text, name, status, source, created_at, updated_at
	`, params.WordSetID, params.Name, params.Status, params.Source)

	return scanWordSet(row)
}

type CreateWordSetWordParams struct {
	WordSetID string
	Word      string
	SortOrder int
	IsActive  bool
}

func (s *Store) CreateWordSetWord(ctx context.Context, params CreateWordSetWordParams) (domain.WordSetWord, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.WordSetWord{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getWordSetForUpdate(ctx, tx, params.WordSetID); err != nil {
		return domain.WordSetWord{}, mapWriteError(err)
	}

	sortOrder := params.SortOrder
	if sortOrder <= 0 {
		if err := tx.QueryRow(ctx, `SELECT COALESCE(MAX(sort_order), 0) + 1 FROM word_set_words WHERE word_set_id = $1`, params.WordSetID).Scan(&sortOrder); err != nil {
			return domain.WordSetWord{}, mapWriteError(err)
		}
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO word_set_words (word_set_id, word, sort_order, is_active)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, word_set_id::text, word, sort_order, is_active, created_at
	`, params.WordSetID, params.Word, sortOrder, params.IsActive)
	word, err := scanWordSetWord(row)
	if err != nil {
		return domain.WordSetWord{}, mapWriteError(err)
	}

	if _, err := tx.Exec(ctx, `UPDATE word_sets SET updated_at = now() WHERE id = $1`, params.WordSetID); err != nil {
		return domain.WordSetWord{}, mapWriteError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.WordSetWord{}, mapWriteError(err)
	}

	return word, nil
}

type UpdateWordSetWordParams struct {
	WordSetID string
	WordID    string
	Word      *string
	SortOrder *int
	IsActive  *bool
}

func (s *Store) UpdateWordSetWord(ctx context.Context, params UpdateWordSetWordParams) (domain.WordSetWord, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.WordSetWord{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getWordSetForUpdate(ctx, tx, params.WordSetID); err != nil {
		return domain.WordSetWord{}, mapWriteError(err)
	}

	row := tx.QueryRow(ctx, `
		UPDATE word_set_words
		SET word = COALESCE($3, word),
		    sort_order = COALESCE($4, sort_order),
		    is_active = COALESCE($5, is_active)
		WHERE word_set_id = $1
		  AND id = $2
		RETURNING id::text, word_set_id::text, word, sort_order, is_active, created_at
	`, params.WordSetID, params.WordID, params.Word, params.SortOrder, params.IsActive)
	word, err := scanWordSetWord(row)
	if err != nil {
		return domain.WordSetWord{}, mapWriteError(err)
	}

	if _, err := tx.Exec(ctx, `UPDATE word_sets SET updated_at = now() WHERE id = $1`, params.WordSetID); err != nil {
		return domain.WordSetWord{}, mapWriteError(err)
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.WordSetWord{}, mapWriteError(err)
	}

	return word, nil
}

func (s *Store) DeleteWordSetWord(ctx context.Context, wordSetID, wordID string) (domain.WordSetWord, error) {
	inactive := false
	return s.UpdateWordSetWord(ctx, UpdateWordSetWordParams{
		WordSetID: wordSetID,
		WordID:    wordID,
		IsActive:  &inactive,
	})
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

	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "word.called", &calledWord.ID, map[string]any{"word": calledWord.Word, "sequence": calledWord.Sequence}); err != nil {
		return domain.CalledWord{}, err
	}
	if _, err := autoMarkInTx(ctx, tx, params.GameRunID, &calledWord.Word, &calledWord.ID); err != nil {
		return domain.CalledWord{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return domain.CalledWord{}, mapWriteError(err)
	}

	return calledWord, nil
}

func (s *Store) AutoMarkGame(ctx context.Context, gameRunID string) (AutoMarkRunResult, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return AutoMarkRunResult{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, gameRunID); err != nil {
		return AutoMarkRunResult{}, mapWriteError(err)
	}
	result, err := autoMarkInTx(ctx, tx, gameRunID, nil, nil)
	if err != nil {
		return AutoMarkRunResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AutoMarkRunResult{}, mapWriteError(err)
	}

	return result, nil
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

	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return domain.BingoCard{}, mapWriteError(err)
	}

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

	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "card.assigned", &card.ID, map[string]any{"cardId": card.ID, "playerId": params.PlayerID}); err != nil {
		return domain.BingoCard{}, err
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
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.BingoCardCell{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	if _, err := getGameRunForUpdate(ctx, tx, params.GameRunID); err != nil {
		return domain.BingoCardCell{}, mapWriteError(err)
	}

	row := tx.QueryRow(ctx, `
		UPDATE bingo_card_cells AS cell
		SET marked_at = CASE WHEN $4 THEN now() ELSE NULL END
		FROM bingo_cards AS card
		WHERE cell.card_id = card.id
		  AND card.game_run_id = $1
		  AND card.player_id = $2
		  AND cell.id = $3
		RETURNING cell.id::text, cell.card_id::text, cell.row_index, cell.col_index, cell.word, cell.is_free_space, cell.marked_at, cell.created_at
	`, params.GameRunID, params.PlayerID, params.CellID, params.Marked)

	cell, err := scanBingoCardCell(row)
	if err != nil {
		return domain.BingoCardCell{}, err
	}

	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "card.cell_marked", &cell.ID, map[string]any{"cellId": cell.ID, "playerId": params.PlayerID, "marked": params.Marked}); err != nil {
		return domain.BingoCardCell{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.BingoCardCell{}, mapWriteError(err)
	}

	return cell, nil
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

type ClaimValidationData struct {
	GameRun     domain.GameRun
	Player      domain.Player
	Card        domain.BingoCard
	CalledWords []domain.CalledWord
}

type ClaimValidationDecision struct {
	Status           string
	ValidationResult json.RawMessage
	Valid            bool
}

type SubmitBingoClaimTxParams struct {
	GameRunID        string
	PlayerID         string
	Pattern          string
	ReviewedByUserID *string
	Validate         func(ClaimValidationData) (ClaimValidationDecision, error)
}

type SubmitBingoClaimTxResult struct {
	Claim  domain.BingoClaim
	Winner *domain.Winner
	Events []events.Event
}

func (s *Store) SubmitBingoClaimTx(ctx context.Context, params SubmitBingoClaimTxParams) (SubmitBingoClaimTxResult, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return SubmitBingoClaimTxResult{}, mapWriteError(err)
	}
	defer tx.Rollback(ctx)

	run, err := getGameRunForUpdate(ctx, tx, params.GameRunID)
	if err != nil {
		return SubmitBingoClaimTxResult{}, mapWriteError(err)
	}

	player, err := getPlayerForUpdate(ctx, tx, params.PlayerID)
	if err != nil {
		return SubmitBingoClaimTxResult{}, mapWriteError(err)
	}

	card, err := getPlayerCardInTx(ctx, tx, params.PlayerID)
	if err != nil {
		return SubmitBingoClaimTxResult{}, mapWriteError(err)
	}

	calledWords, err := listCalledWordsInTx(ctx, tx, params.GameRunID)
	if err != nil {
		return SubmitBingoClaimTxResult{}, mapWriteError(err)
	}

	if params.Validate == nil {
		return SubmitBingoClaimTxResult{}, fmt.Errorf("claim validator is required")
	}
	decision, err := params.Validate(ClaimValidationData{
		GameRun:     run,
		Player:      player,
		Card:        card,
		CalledWords: calledWords,
	})
	if err != nil {
		return SubmitBingoClaimTxResult{}, err
	}
	if decision.Status == "" {
		return SubmitBingoClaimTxResult{}, fmt.Errorf("claim validator returned empty status")
	}
	if len(decision.ValidationResult) == 0 {
		decision.ValidationResult = json.RawMessage(`{}`)
	}

	claimRow := tx.QueryRow(ctx, `
		INSERT INTO bingo_claims (game_run_id, player_id, pattern, status, validation_result, reviewed_by_user_id, reviewed_at)
		VALUES ($1, $2, $3, $4, $5, $6, now())
		RETURNING id::text, game_run_id::text, player_id::text, pattern, status, validation_result, claimed_at, reviewed_by_user_id::text, reviewed_at, created_at, updated_at
	`, params.GameRunID, params.PlayerID, params.Pattern, decision.Status, decision.ValidationResult, params.ReviewedByUserID)
	claim, err := scanBingoClaim(claimRow)
	if err != nil {
		return SubmitBingoClaimTxResult{}, mapWriteError(err)
	}

	result := SubmitBingoClaimTxResult{
		Claim: claim,
		Events: []events.Event{
			{Type: "claim.submitted", EntityID: claim.ID, Payload: map[string]any{"gameRunId": params.GameRunID, "playerId": params.PlayerID, "pattern": params.Pattern}},
			{Type: "claim.validated", EntityID: claim.ID, Payload: map[string]any{"gameRunId": params.GameRunID, "status": claim.Status, "valid": decision.Valid}},
		},
	}

	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "claim.submitted", &claim.ID, map[string]any{"playerId": params.PlayerID, "pattern": params.Pattern}); err != nil {
		return SubmitBingoClaimTxResult{}, err
	}
	if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "claim.validated", &claim.ID, map[string]any{"status": claim.Status, "valid": decision.Valid}); err != nil {
		return SubmitBingoClaimTxResult{}, err
	}

	if err := recordAuditEventInTx(ctx, tx, audit.Event{GameRunID: &params.GameRunID, EventType: "claim.submitted", EntityType: "bingo_claim", EntityID: &claim.ID, Payload: map[string]any{"playerId": params.PlayerID, "pattern": params.Pattern}}); err != nil {
		return SubmitBingoClaimTxResult{}, err
	}
	if err := recordAuditEventInTx(ctx, tx, audit.Event{GameRunID: &params.GameRunID, EventType: "claim.validated", EntityType: "bingo_claim", EntityID: &claim.ID, Payload: map[string]any{"status": claim.Status, "valid": decision.Valid}}); err != nil {
		return SubmitBingoClaimTxResult{}, err
	}

	if decision.Valid {
		if _, err := tx.Exec(ctx, `
			UPDATE players
			SET state = 'confirmed_winner',
			    updated_at = now()
			WHERE id = $1
		`, params.PlayerID); err != nil {
			return SubmitBingoClaimTxResult{}, mapWriteError(err)
		}

		existingWinner, err := getWinnerByPlayerPatternInTx(ctx, tx, params.GameRunID, params.PlayerID, params.Pattern)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return SubmitBingoClaimTxResult{}, err
		}
		if err == nil {
			result.Winner = &existingWinner
		} else {
			winners, err := listWinnersForUpdateInTx(ctx, tx, params.GameRunID)
			if err != nil {
				return SubmitBingoClaimTxResult{}, err
			}
			if len(winners) < 3 {
				placement := len(winners) + 1
				winnerRow := tx.QueryRow(ctx, `
					INSERT INTO winners (game_run_id, player_id, claim_id, placement, pattern)
					VALUES ($1, $2, $3, $4, $5)
					RETURNING id::text, game_run_id::text, player_id::text, claim_id::text, placement, pattern, confirmed_at, created_at
				`, params.GameRunID, params.PlayerID, claim.ID, placement, params.Pattern)
				winner, err := scanWinner(winnerRow)
				if err != nil {
					return SubmitBingoClaimTxResult{}, mapWriteError(err)
				}
				result.Winner = &winner
				result.Events = append(result.Events, events.Event{Type: "winner.created", EntityID: winner.ID, Payload: map[string]any{"gameRunId": params.GameRunID, "playerId": params.PlayerID, "placement": winner.Placement}})
				if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "winner.created", &winner.ID, map[string]any{"playerId": params.PlayerID, "placement": winner.Placement, "pattern": winner.Pattern}); err != nil {
					return SubmitBingoClaimTxResult{}, err
				}
				if err := recordAuditEventInTx(ctx, tx, audit.Event{GameRunID: &params.GameRunID, EventType: "winner.created", EntityType: "winner", EntityID: &winner.ID, Payload: map[string]any{"playerId": params.PlayerID, "placement": winner.Placement}}); err != nil {
					return SubmitBingoClaimTxResult{}, err
				}

				if placement == 3 {
					if _, err := tx.Exec(ctx, `
						UPDATE game_runs
						SET status = 'finished',
						    ended_at = COALESCE(ended_at, now()),
						    updated_at = now()
						WHERE id = $1
					`, params.GameRunID); err != nil {
						return SubmitBingoClaimTxResult{}, mapWriteError(err)
					}
					result.Events = append(result.Events, events.Event{Type: "game.finished", EntityID: params.GameRunID, Payload: map[string]any{"status": "finished", "reason": "third_winner"}})
					if _, err := insertOutboxEventInTx(ctx, tx, params.GameRunID, "game.finished", &params.GameRunID, map[string]any{"status": "finished", "reason": "third_winner"}); err != nil {
						return SubmitBingoClaimTxResult{}, err
					}
					if err := recordAuditEventInTx(ctx, tx, audit.Event{GameRunID: &params.GameRunID, EventType: "game.finished", EntityType: "game_run", EntityID: &params.GameRunID, Payload: map[string]any{"reason": "third_winner"}}); err != nil {
						return SubmitBingoClaimTxResult{}, err
					}
				}
			}
		}
	} else if player.State != "confirmed_winner" {
		if _, err := tx.Exec(ctx, `
			UPDATE players
			SET state = 'rejected_claim',
			    updated_at = now()
			WHERE id = $1
		`, params.PlayerID); err != nil {
			return SubmitBingoClaimTxResult{}, mapWriteError(err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return SubmitBingoClaimTxResult{}, mapWriteError(err)
	}

	return result, nil
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

func (s *Store) ListPlayers(ctx context.Context, gameRunID string) ([]domain.Player, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id::text, game_run_id::text, user_id::text, email, display_name, player_icon, player_avatar_color, player_avatar_label, connection_state, state, joined_at, last_seen_at, created_at, updated_at
		FROM players
		WHERE game_run_id = $1
		ORDER BY joined_at ASC, id ASC
	`, gameRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make([]domain.Player, 0)
	for rows.Next() {
		player, err := scanPlayer(rows)
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

func (s *Store) ListGameEvents(ctx context.Context, gameRunID string, afterSequence int64, limit int) ([]domain.GameEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id::text, game_run_id::text, type, entity_id::text, payload, sequence, created_at
		FROM game_event_outbox
		WHERE game_run_id = $1
		  AND sequence > $2
		ORDER BY sequence ASC
		LIMIT $3
	`, gameRunID, afterSequence, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	gameEvents := make([]domain.GameEvent, 0)
	for rows.Next() {
		event, err := scanGameEvent(rows)
		if err != nil {
			return nil, err
		}
		gameEvents = append(gameEvents, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return gameEvents, nil
}

func (s *Store) ListActivityEvents(ctx context.Context, gameRunID string, limit int) ([]domain.ActivityEvent, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	rows, err := s.pool.Query(ctx, `
		SELECT id::text, game_run_id::text, type, entity_id::text, payload, sequence, created_at
		FROM game_event_outbox
		WHERE game_run_id = $1
		ORDER BY sequence DESC
		LIMIT $2
	`, gameRunID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]domain.ActivityEvent, 0)
	for rows.Next() {
		var event domain.GameEvent
		var entityID sql.NullString
		var payload []byte
		if err := rows.Scan(
			&event.ID,
			&event.GameRunID,
			&event.Type,
			&entityID,
			&payload,
			&event.Sequence,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		event.EntityID = nullableStringPtr(entityID)
		event.Payload = json.RawMessage(payload)
		sequence := event.Sequence
		events = append(events, domain.ActivityEvent{
			ID:        event.ID,
			GameRunID: event.GameRunID,
			Type:      event.Type,
			EntityID:  event.EntityID,
			Payload:   event.Payload,
			Sequence:  &sequence,
			CreatedAt: event.CreatedAt,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return events, nil
}

type scanner interface {
	Scan(dest ...any) error
}

type queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

func ensureGameSettingsInTx(ctx context.Context, tx pgx.Tx, gameRunID string) (domain.GameRunSettings, error) {
	row := tx.QueryRow(ctx, `
		INSERT INTO game_run_settings (game_run_id)
		VALUES ($1)
		ON CONFLICT (game_run_id) DO UPDATE
		SET game_run_id = EXCLUDED.game_run_id
		RETURNING game_run_id::text, marking_mode, allow_player_marking_mode_choice, show_claim_readiness, voice_claim_mode, voice_claim_autoplay, caller_mode, theme_mode, theme_id::text, created_at, updated_at
	`, gameRunID)

	settings, err := scanGameRunSettings(row)
	if err != nil {
		return domain.GameRunSettings{}, mapWriteError(err)
	}

	return settings, nil
}

func ensurePlayerPreferencesInTx(ctx context.Context, tx pgx.Tx, playerID string) (domain.PlayerPreferences, error) {
	row := tx.QueryRow(ctx, `
		INSERT INTO player_preferences (player_id)
		VALUES ($1)
		ON CONFLICT (player_id) DO UPDATE
		SET player_id = EXCLUDED.player_id
		RETURNING player_id::text, marking_mode, created_at, updated_at
	`, playerID)

	preferences, err := scanPlayerPreferences(row)
	if err != nil {
		return domain.PlayerPreferences{}, mapWriteError(err)
	}

	return preferences, nil
}

func autoMarkInTx(ctx context.Context, tx pgx.Tx, gameRunID string, calledWord *string, entityID *string) (AutoMarkRunResult, error) {
	settings, err := ensureGameSettingsInTx(ctx, tx, gameRunID)
	if err != nil {
		return AutoMarkRunResult{}, err
	}
	result := AutoMarkRunResult{Mode: settings.MarkingMode}

	if err := tx.QueryRow(ctx, `
		SELECT count(*)::int
		FROM called_words
		WHERE game_run_id = $1
		  AND ($2::text IS NULL OR lower(btrim(word)) = lower(btrim($2)))
	`, gameRunID, calledWord).Scan(&result.CalledWordsScanned); err != nil {
		return AutoMarkRunResult{}, mapWriteError(err)
	}

	if !settings.AllowPlayerMarkingModeChoice && settings.MarkingMode != "auto_mark" {
		result.SkippedReason = "marking_mode_not_auto_mark"
		return result, nil
	}

	if err := tx.QueryRow(ctx, `
		SELECT count(DISTINCT p.id)::int
		FROM players AS p
		JOIN bingo_cards AS card ON card.player_id = p.id
		LEFT JOIN player_preferences AS pref ON pref.player_id = p.id
		WHERE p.game_run_id = $1
		  AND CASE
		    WHEN $2::boolean THEN COALESCE(pref.marking_mode, $3) = 'auto_mark'
		    ELSE $3 = 'auto_mark'
		  END
	`, gameRunID, settings.AllowPlayerMarkingModeChoice, settings.MarkingMode).Scan(&result.PlayersScanned); err != nil {
		return AutoMarkRunResult{}, mapWriteError(err)
	}
	if result.PlayersScanned == 0 {
		result.SkippedReason = "no_players_using_auto_mark"
		return result, nil
	}
	if result.CalledWordsScanned == 0 {
		result.SkippedReason = "no_called_words"
		return result, nil
	}

	row := tx.QueryRow(ctx, `
		WITH effective_players AS (
		  SELECT p.id
		  FROM players AS p
		  JOIN bingo_cards AS card ON card.player_id = p.id
		  LEFT JOIN player_preferences AS pref ON pref.player_id = p.id
		  WHERE p.game_run_id = $1
		    AND CASE
		      WHEN $2::boolean THEN COALESCE(pref.marking_mode, $3) = 'auto_mark'
		      ELSE $3 = 'auto_mark'
		    END
		),
		called AS (
		  SELECT DISTINCT lower(btrim(word)) AS normalized_word
		  FROM called_words
		  WHERE game_run_id = $1
		    AND ($4::text IS NULL OR lower(btrim(word)) = lower(btrim($4)))
		),
		marked AS (
		  UPDATE bingo_card_cells AS cell
		  SET marked_at = now()
		  FROM bingo_cards AS card, effective_players AS player, called
		  WHERE cell.card_id = card.id
		    AND card.player_id = player.id
		    AND card.game_run_id = $1
		    AND cell.is_free_space = false
		    AND cell.marked_at IS NULL
		    AND lower(btrim(cell.word)) = called.normalized_word
		  RETURNING card.player_id, cell.id
		)
		SELECT count(DISTINCT player_id)::int, count(*)::int FROM marked
	`, gameRunID, settings.AllowPlayerMarkingModeChoice, settings.MarkingMode, calledWord)
	if err := row.Scan(&result.PlayersMarked, &result.CellsMarked); err != nil {
		return AutoMarkRunResult{}, mapWriteError(err)
	}

	if result.CellsMarked > 0 {
		payload := map[string]any{
			"playersScanned":     result.PlayersScanned,
			"playersMarked":      result.PlayersMarked,
			"calledWordsScanned": result.CalledWordsScanned,
			"cellsMarked":        result.CellsMarked,
			"mode":               result.Mode,
		}
		if calledWord != nil {
			payload["word"] = *calledWord
		}
		if _, err := insertOutboxEventInTx(ctx, tx, gameRunID, "card.auto_marked", entityID, payload); err != nil {
			return AutoMarkRunResult{}, err
		}
		if err := recordAuditEventInTx(ctx, tx, audit.Event{
			GameRunID:  &gameRunID,
			EventType:  "card.auto_marked",
			EntityType: "bingo_card",
			EntityID:   entityID,
			Payload:    payload,
		}); err != nil {
			return AutoMarkRunResult{}, err
		}
	}

	return result, nil
}

func getGameRunForUpdate(ctx context.Context, q queryer, gameRunID string) (domain.GameRun, error) {
	row := q.QueryRow(ctx, `
		SELECT id::text, template_id::text, host_user_id::text, word_set_id::text, code, name, status, scheduled_start_at, started_at, ended_at, current_called_word_id::text, winning_pattern, created_at, updated_at
		FROM game_runs
		WHERE id = $1
		FOR UPDATE
	`, gameRunID)

	return scanGameRun(row)
}

func getPlayerForUpdate(ctx context.Context, q queryer, playerID string) (domain.Player, error) {
	row := q.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, user_id::text, email, display_name, player_icon, player_avatar_color, player_avatar_label, connection_state, state, joined_at, last_seen_at, created_at, updated_at
		FROM players
		WHERE id = $1
		FOR UPDATE
	`, playerID)

	return scanPlayer(row)
}

func getWordSetForUpdate(ctx context.Context, q queryer, wordSetID string) (domain.WordSet, error) {
	row := q.QueryRow(ctx, `
		SELECT id::text, name, status, source, created_at, updated_at
		FROM word_sets
		WHERE id = $1
		FOR UPDATE
	`, wordSetID)

	return scanWordSet(row)
}

func getGeneratedGameContentForUpdate(ctx context.Context, q queryer, gameRunID string) (domain.GeneratedGameContent, error) {
	row := q.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, generation_job_id::text, status, topic, summary, generated_words, current_words, caller_style, theme_prompt, review_window_opens_at, review_window_closes_at, locked_at, locked_word_set_id::text, generation_provider, generation_error, created_at, updated_at
		FROM generated_game_content
		WHERE game_run_id = $1
		FOR UPDATE
	`, gameRunID)

	return scanGeneratedGameContent(row)
}

func listGameCallDeckInTx(ctx context.Context, q queryer, gameRunID string) ([]domain.GameCallDeckItem, error) {
	rows, err := q.Query(ctx, `
		SELECT id::text, game_run_id::text, word_set_word_id::text, word, sequence, shuffle_seed, shuffle_version, locked_at, called_word_id::text, created_at
		FROM game_call_deck
		WHERE game_run_id = $1
		ORDER BY sequence ASC
	`, gameRunID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanGameCallDeckRows(rows)
}

func getCallerAssetByDeckItemInTx(ctx context.Context, q queryer, deckItemID string) (domain.CallerAsset, error) {
	row := q.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, call_deck_item_id::text, word, sequence, line, audio_url, storage_key, voice_name, provider, status, error_reason, created_at, updated_at
		FROM caller_assets
		WHERE call_deck_item_id = $1
	`, deckItemID)
	return scanCallerAsset(row)
}

func getThemeForUpdate(ctx context.Context, q queryer, themeID string) (domain.Theme, error) {
	row := q.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, generation_job_id::text, name, summary, tokens, status, provider, created_by_user_id::text, approved_by_user_id::text, approved_at, rejected_at, created_at, updated_at
		FROM themes
		WHERE id = $1
		FOR UPDATE
	`, themeID)
	return scanTheme(row)
}

func getPlayerCardInTx(ctx context.Context, q queryer, playerID string) (domain.BingoCard, error) {
	cardRow := q.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, player_id::text, seed, created_at
		FROM bingo_cards
		WHERE player_id = $1
		FOR UPDATE
	`, playerID)

	card, err := scanBingoCard(cardRow)
	if err != nil {
		return domain.BingoCard{}, err
	}

	rows, err := q.Query(ctx, `
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

func listCalledWordsInTx(ctx context.Context, q queryer, gameRunID string) ([]domain.CalledWord, error) {
	rows, err := q.Query(ctx, `
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

func listWinnersForUpdateInTx(ctx context.Context, q queryer, gameRunID string) ([]domain.Winner, error) {
	rows, err := q.Query(ctx, `
		SELECT id::text, game_run_id::text, player_id::text, claim_id::text, placement, pattern, confirmed_at, created_at
		FROM winners
		WHERE game_run_id = $1
		ORDER BY placement ASC, confirmed_at ASC
		FOR UPDATE
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

func getWinnerByPlayerPatternInTx(ctx context.Context, q queryer, gameRunID, playerID, pattern string) (domain.Winner, error) {
	row := q.QueryRow(ctx, `
		SELECT id::text, game_run_id::text, player_id::text, claim_id::text, placement, pattern, confirmed_at, created_at
		FROM winners
		WHERE game_run_id = $1
		  AND player_id = $2
		  AND pattern = $3
		ORDER BY placement ASC
		LIMIT 1
	`, gameRunID, playerID, pattern)

	return scanWinner(row)
}

func recordAuditEventInTx(ctx context.Context, tx pgx.Tx, event audit.Event) error {
	payload := event.Payload
	if payload == nil {
		payload = map[string]any{}
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal audit payload: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO audit_events (game_run_id, actor_user_id, event_type, entity_type, entity_id, payload)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, event.GameRunID, event.ActorUserID, event.EventType, event.EntityType, event.EntityID, payloadJSON)
	if err != nil {
		return mapWriteError(err)
	}

	return nil
}

func insertOutboxEventInTx(ctx context.Context, tx pgx.Tx, gameRunID, eventType string, entityID *string, payload map[string]any) (domain.GameEvent, error) {
	if payload == nil {
		payload = map[string]any{}
	}

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return domain.GameEvent{}, fmt.Errorf("marshal outbox payload: %w", err)
	}

	row := tx.QueryRow(ctx, `
		INSERT INTO game_event_outbox (game_run_id, type, entity_id, payload, sequence)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			(SELECT COALESCE(MAX(sequence), 0) + 1 FROM game_event_outbox WHERE game_run_id = $1)
		)
		RETURNING id::text, game_run_id::text, type, entity_id::text, payload, sequence, created_at
	`, gameRunID, eventType, entityID, payloadJSON)

	event, err := scanGameEvent(row)
	if err != nil {
		return domain.GameEvent{}, mapWriteError(err)
	}

	return event, nil
}

func statusAllowed(status string, allowed []string) bool {
	for _, candidate := range allowed {
		if status == candidate {
			return true
		}
	}

	return false
}

func gameSettingsPayload(settings domain.GameRunSettings) map[string]any {
	return map[string]any{
		"gameRunId":                    settings.GameRunID,
		"markingMode":                  settings.MarkingMode,
		"allowPlayerMarkingModeChoice": settings.AllowPlayerMarkingModeChoice,
		"showClaimReadiness":           settings.ShowClaimReadiness,
		"voiceClaimMode":               settings.VoiceClaimMode,
		"voiceClaimAutoplay":           settings.VoiceClaimAutoplay,
		"callerMode":                   settings.CallerMode,
		"themeMode":                    settings.ThemeMode,
		"themeId":                      settings.ThemeID,
	}
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

func scanGameRunSettings(row scanner) (domain.GameRunSettings, error) {
	var settings domain.GameRunSettings
	var themeID sql.NullString

	err := row.Scan(
		&settings.GameRunID,
		&settings.MarkingMode,
		&settings.AllowPlayerMarkingModeChoice,
		&settings.ShowClaimReadiness,
		&settings.VoiceClaimMode,
		&settings.VoiceClaimAutoplay,
		&settings.CallerMode,
		&settings.ThemeMode,
		&themeID,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)
	if err != nil {
		return domain.GameRunSettings{}, mapError(err)
	}

	settings.ThemeID = nullableStringPtr(themeID)
	return settings, nil
}

func scanContentGenerationJob(row scanner) (domain.ContentGenerationJob, error) {
	var job domain.ContentGenerationJob
	var errorMessage sql.NullString

	err := row.Scan(
		&job.ID,
		&job.GameRunID,
		&job.JobType,
		&job.Status,
		&job.Provider,
		&errorMessage,
		&job.RetryCount,
		&job.CreatedAt,
		&job.UpdatedAt,
	)
	if err != nil {
		return domain.ContentGenerationJob{}, mapError(err)
	}

	job.ErrorMessage = nullableStringPtr(errorMessage)
	return job, nil
}

func scanGeneratedGameContent(row scanner) (domain.GeneratedGameContent, error) {
	var content domain.GeneratedGameContent
	var generationJobID, callerStyle, themePrompt, lockedWordSetID, generationError sql.NullString
	var reviewWindowOpensAt, reviewWindowClosesAt, lockedAt sql.NullTime
	var generatedWordsJSON, currentWordsJSON []byte

	err := row.Scan(
		&content.ID,
		&content.GameRunID,
		&generationJobID,
		&content.Status,
		&content.Topic,
		&content.Summary,
		&generatedWordsJSON,
		&currentWordsJSON,
		&callerStyle,
		&themePrompt,
		&reviewWindowOpensAt,
		&reviewWindowClosesAt,
		&lockedAt,
		&lockedWordSetID,
		&content.GenerationProvider,
		&generationError,
		&content.CreatedAt,
		&content.UpdatedAt,
	)
	if err != nil {
		return domain.GeneratedGameContent{}, mapError(err)
	}
	if len(generatedWordsJSON) > 0 {
		if err := json.Unmarshal(generatedWordsJSON, &content.GeneratedWords); err != nil {
			return domain.GeneratedGameContent{}, fmt.Errorf("unmarshal generated words: %w", err)
		}
	}
	if len(currentWordsJSON) > 0 {
		if err := json.Unmarshal(currentWordsJSON, &content.CurrentWords); err != nil {
			return domain.GeneratedGameContent{}, fmt.Errorf("unmarshal current words: %w", err)
		}
	}

	content.GenerationJobID = nullableStringPtr(generationJobID)
	content.CallerStyle = nullableStringPtr(callerStyle)
	content.ThemePrompt = nullableStringPtr(themePrompt)
	content.ReviewWindowOpensAt = nullableTimePtr(reviewWindowOpensAt)
	content.ReviewWindowClosesAt = nullableTimePtr(reviewWindowClosesAt)
	content.LockedAt = nullableTimePtr(lockedAt)
	content.LockedWordSetID = nullableStringPtr(lockedWordSetID)
	content.GenerationError = nullableStringPtr(generationError)
	return content, nil
}

func scanPlayerPreferences(row scanner) (domain.PlayerPreferences, error) {
	var preferences domain.PlayerPreferences
	var markingMode sql.NullString

	err := row.Scan(
		&preferences.PlayerID,
		&markingMode,
		&preferences.CreatedAt,
		&preferences.UpdatedAt,
	)
	if err != nil {
		return domain.PlayerPreferences{}, mapError(err)
	}

	preferences.MarkingMode = nullableStringPtr(markingMode)
	return preferences, nil
}

func scanGameRunWithCounts(row scanner, allowedPlayerCount *int, playerCount *int) (domain.GameRun, error) {
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
		allowedPlayerCount,
		playerCount,
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

func scanWordSet(row scanner) (domain.WordSet, error) {
	var wordSet domain.WordSet
	err := row.Scan(&wordSet.ID, &wordSet.Name, &wordSet.Status, &wordSet.Source, &wordSet.CreatedAt, &wordSet.UpdatedAt)
	if err != nil {
		return domain.WordSet{}, mapError(err)
	}

	return wordSet, nil
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
	var userID, icon, avatarColor, avatarLabel sql.NullString

	err := row.Scan(
		&player.ID,
		&player.GameRunID,
		&userID,
		&player.Email,
		&player.DisplayName,
		&icon,
		&avatarColor,
		&avatarLabel,
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
	player.Icon = nullableStringPtr(icon)
	player.AvatarColor = nullableStringPtr(avatarColor)
	player.AvatarLabel = nullableStringPtr(avatarLabel)
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

func scanGameCallDeckRows(rows pgx.Rows) ([]domain.GameCallDeckItem, error) {
	items := make([]domain.GameCallDeckItem, 0)
	for rows.Next() {
		item, err := scanGameCallDeckItem(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func scanGameCallDeckItem(row scanner) (domain.GameCallDeckItem, error) {
	var item domain.GameCallDeckItem
	var wordSetWordID, calledWordID sql.NullString
	err := row.Scan(
		&item.ID,
		&item.GameRunID,
		&wordSetWordID,
		&item.Word,
		&item.Sequence,
		&item.ShuffleSeed,
		&item.ShuffleVersion,
		&item.LockedAt,
		&calledWordID,
		&item.CreatedAt,
	)
	if err != nil {
		return domain.GameCallDeckItem{}, mapError(err)
	}
	item.WordSetWordID = nullableStringPtr(wordSetWordID)
	item.CalledWordID = nullableStringPtr(calledWordID)
	return item, nil
}

func scanCallerAsset(row scanner) (domain.CallerAsset, error) {
	var asset domain.CallerAsset
	var audioURL, storageKey, voiceName, errorReason sql.NullString
	err := row.Scan(
		&asset.ID,
		&asset.GameRunID,
		&asset.CallDeckItemID,
		&asset.Word,
		&asset.Sequence,
		&asset.Line,
		&audioURL,
		&storageKey,
		&voiceName,
		&asset.Provider,
		&asset.Status,
		&errorReason,
		&asset.CreatedAt,
		&asset.UpdatedAt,
	)
	if err != nil {
		return domain.CallerAsset{}, mapError(err)
	}
	asset.AudioURL = nullableStringPtr(audioURL)
	asset.StorageKey = nullableStringPtr(storageKey)
	asset.VoiceName = nullableStringPtr(voiceName)
	asset.ErrorReason = nullableStringPtr(errorReason)
	return asset, nil
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

func scanGameEvent(row scanner) (domain.GameEvent, error) {
	var event domain.GameEvent
	var entityID sql.NullString
	var payload []byte

	err := row.Scan(
		&event.ID,
		&event.GameRunID,
		&event.Type,
		&entityID,
		&payload,
		&event.Sequence,
		&event.CreatedAt,
	)
	if err != nil {
		return domain.GameEvent{}, mapError(err)
	}

	event.EntityID = nullableStringPtr(entityID)
	event.Payload = json.RawMessage(payload)
	return event, nil
}

func scanDeliveryBatch(row scanner) (domain.DeliveryBatch, error) {
	var batch domain.DeliveryBatch
	if err := row.Scan(&batch.ID, &batch.GameRunID, &batch.Channel, &batch.Purpose, &batch.Status, &batch.CreatedAt, &batch.UpdatedAt); err != nil {
		return domain.DeliveryBatch{}, mapError(err)
	}
	return batch, nil
}

func scanDeliveryAttempt(row scanner) (domain.DeliveryAttempt, error) {
	var attempt domain.DeliveryAttempt
	var recipientUserID, errorReason sql.NullString
	var sentAt sql.NullTime
	if err := row.Scan(
		&attempt.ID,
		&attempt.BatchID,
		&attempt.GameRunID,
		&attempt.Channel,
		&attempt.Purpose,
		&attempt.RecipientEmail,
		&recipientUserID,
		&attempt.Subject,
		&attempt.TemplateKey,
		&attempt.BodyPreview,
		&attempt.LinkURL,
		&attempt.GameCode,
		&attempt.Status,
		&errorReason,
		&sentAt,
		&attempt.CreatedAt,
		&attempt.UpdatedAt,
	); err != nil {
		return domain.DeliveryAttempt{}, mapError(err)
	}
	attempt.RecipientUserID = nullableStringPtr(recipientUserID)
	attempt.ErrorReason = nullableStringPtr(errorReason)
	attempt.SentAt = nullableTimePtr(sentAt)
	return attempt, nil
}

func scanThemeGenerationJob(row scanner) (domain.ThemeGenerationJob, error) {
	var job domain.ThemeGenerationJob
	var gameRunID, errorMessage sql.NullString
	if err := row.Scan(&job.ID, &gameRunID, &job.Status, &job.Provider, &job.Prompt, &errorMessage, &job.CreatedAt, &job.UpdatedAt); err != nil {
		return domain.ThemeGenerationJob{}, mapError(err)
	}
	job.GameRunID = nullableStringPtr(gameRunID)
	job.ErrorMessage = nullableStringPtr(errorMessage)
	return job, nil
}

func scanTheme(row scanner) (domain.Theme, error) {
	var theme domain.Theme
	var gameRunID, generationJobID, createdByUserID, approvedByUserID sql.NullString
	var approvedAt, rejectedAt sql.NullTime
	var tokensJSON []byte
	if err := row.Scan(
		&theme.ID,
		&gameRunID,
		&generationJobID,
		&theme.Name,
		&theme.Summary,
		&tokensJSON,
		&theme.Status,
		&theme.Provider,
		&createdByUserID,
		&approvedByUserID,
		&approvedAt,
		&rejectedAt,
		&theme.CreatedAt,
		&theme.UpdatedAt,
	); err != nil {
		return domain.Theme{}, mapError(err)
	}
	if err := json.Unmarshal(tokensJSON, &theme.Tokens); err != nil {
		return domain.Theme{}, err
	}
	theme.GameRunID = nullableStringPtr(gameRunID)
	theme.GenerationJobID = nullableStringPtr(generationJobID)
	theme.CreatedByUserID = nullableStringPtr(createdByUserID)
	theme.ApprovedByUserID = nullableStringPtr(approvedByUserID)
	theme.ApprovedAt = nullableTimePtr(approvedAt)
	theme.RejectedAt = nullableTimePtr(rejectedAt)
	return theme, nil
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
