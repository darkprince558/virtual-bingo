package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestMigrationsCanRun(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)

	if err := RunMigrationsFromPath(databaseURL, "migrations", MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool := openTestPool(t, ctx, databaseURL)
	defer pool.Close()

	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'game_runs')`).Scan(&exists); err != nil {
		t.Fatalf("check migrated table: %v", err)
	}
	if !exists {
		t.Fatal("expected game_runs table to exist")
	}
}

func TestStoreGameRunCreation(t *testing.T) {
	ctx := context.Background()
	store, pool := setupMigratedStore(t, ctx)
	defer pool.Close()

	hostID := insertTestHost(t, ctx, pool)
	wordSetID := insertTestWordSet(t, ctx, pool, hostID)
	scheduledAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Microsecond)

	run, err := store.CreateGameRun(ctx, CreateGameRunParams{
		HostUserID:       hostID,
		WordSetID:        &wordSetID,
		Code:             "TEST-RUN",
		Name:             "Test Run",
		Status:           "lobby_open",
		ScheduledStartAt: &scheduledAt,
	})
	if err != nil {
		t.Fatalf("create game run: %v", err)
	}

	got, err := store.GetGameRun(ctx, run.ID)
	if err != nil {
		t.Fatalf("get game run: %v", err)
	}

	if got.ID != run.ID || got.Code != "TEST-RUN" || got.Status != "lobby_open" {
		t.Fatalf("unexpected game run: %+v", got)
	}
	if got.WordSetID == nil || *got.WordSetID != wordSetID {
		t.Fatalf("expected word set %s, got %+v", wordSetID, got.WordSetID)
	}
}

func TestStoreAllowedPlayers(t *testing.T) {
	ctx := context.Background()
	store, pool := setupMigratedStore(t, ctx)
	defer pool.Close()

	runID := insertTestRun(t, ctx, pool)

	if _, err := store.AddAllowedPlayer(ctx, AddAllowedPlayerParams{
		GameRunID:   runID,
		Email:       "alex@example.local",
		DisplayName: "Alex Demo",
	}); err != nil {
		t.Fatalf("add first allowed player: %v", err)
	}
	if _, err := store.AddAllowedPlayer(ctx, AddAllowedPlayerParams{
		GameRunID:   runID,
		Email:       "sam@example.local",
		DisplayName: "Sam Demo",
		Source:      "seed",
	}); err != nil {
		t.Fatalf("add second allowed player: %v", err)
	}

	players, err := store.ListAllowedPlayers(ctx, runID)
	if err != nil {
		t.Fatalf("list allowed players: %v", err)
	}

	if len(players) != 2 {
		t.Fatalf("expected 2 allowed players, got %d", len(players))
	}
	if players[0].Email != "alex@example.local" || players[1].Email != "sam@example.local" {
		t.Fatalf("unexpected allowed players: %+v", players)
	}
}

func TestStoreCardCreation(t *testing.T) {
	ctx := context.Background()
	store, pool := setupMigratedStore(t, ctx)
	defer pool.Close()

	runID := insertTestRun(t, ctx, pool)
	player, err := store.CreatePlayer(ctx, CreatePlayerParams{
		GameRunID:   runID,
		Email:       "player@example.local",
		DisplayName: "Player Demo",
		State:       "playing",
	})
	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	card, err := store.CreateCard(ctx, CreateCardParams{
		GameRunID: runID,
		PlayerID:  player.ID,
		Seed:      "test-seed-1",
		Cells:     makeTestCardCells(),
	})
	if err != nil {
		t.Fatalf("create card: %v", err)
	}
	if len(card.Cells) != 25 {
		t.Fatalf("expected 25 cells from create, got %d", len(card.Cells))
	}

	got, err := store.GetPlayerCard(ctx, player.ID)
	if err != nil {
		t.Fatalf("get player card: %v", err)
	}
	if got.ID != card.ID || got.PlayerID != player.ID || len(got.Cells) != 25 {
		t.Fatalf("unexpected card: %+v", got)
	}
	if !got.Cells[12].IsFreeSpace {
		t.Fatalf("expected center cell to be free space: %+v", got.Cells[12])
	}
}

func TestStorePlayerConnectionState(t *testing.T) {
	ctx := context.Background()
	store, pool := setupMigratedStore(t, ctx)
	defer pool.Close()

	runID := insertTestRun(t, ctx, pool)
	player, err := store.CreatePlayer(ctx, CreatePlayerParams{
		GameRunID:       runID,
		Email:           "connection@example.local",
		DisplayName:     "Connection Player",
		ConnectionState: "offline",
		State:           "joined",
	})
	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	updated, err := store.UpdatePlayerConnectionState(ctx, UpdatePlayerConnectionStateParams{
		GameRunID:       runID,
		PlayerID:        player.ID,
		ConnectionState: "online",
		EventType:       "player.reconnected",
	})
	if err != nil {
		t.Fatalf("update player connection state: %v", err)
	}
	if updated.ConnectionState != "online" || !updated.LastSeenAt.After(player.LastSeenAt) {
		t.Fatalf("unexpected updated player: %+v after %+v", updated, player)
	}
}

func setupMigratedStore(t *testing.T, ctx context.Context) (*Store, *Pool) {
	t.Helper()

	databaseURL := createTestDatabase(t)
	if err := RunMigrationsFromPath(databaseURL, "migrations", MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool := openTestPool(t, ctx, databaseURL)
	return NewStore(pool), pool
}

func createTestDatabase(t *testing.T) string {
	t.Helper()

	baseURL := os.Getenv("TEST_DATABASE_URL")
	if baseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping Postgres integration test")
	}

	parsed, err := url.Parse(baseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}

	testName := strings.ToLower(regexp.MustCompile(`[^a-zA-Z0-9]+`).ReplaceAllString(t.Name(), "_"))
	databaseName := fmt.Sprintf("virtual_bingo_%s_%d", testName, time.Now().UnixNano())

	adminURL := *parsed
	adminURL.Path = "/postgres"

	adminDB, err := sql.Open("postgres", adminURL.String())
	if err != nil {
		t.Fatalf("open admin database: %v", err)
	}
	t.Cleanup(func() {
		adminDB.Close()
	})

	if _, err := adminDB.Exec("CREATE DATABASE " + quoteIdentifier(databaseName)); err != nil {
		t.Fatalf("create test database: %v", err)
	}

	t.Cleanup(func() {
		_, _ = adminDB.Exec(`SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1`, databaseName)
		_, _ = adminDB.Exec("DROP DATABASE IF EXISTS " + quoteIdentifier(databaseName))
	})

	testURL := *parsed
	testURL.Path = "/" + databaseName
	return testURL.String()
}

func openTestPool(t *testing.T, ctx context.Context, databaseURL string) *Pool {
	t.Helper()

	pool, err := Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open test pool: %v", err)
	}

	return pool
}

func insertTestHost(t *testing.T, ctx context.Context, pool *Pool) string {
	t.Helper()

	var id string
	if err := pool.QueryRow(ctx, `
		INSERT INTO users (display_name, email, role)
		VALUES ('Test Host', $1, 'host')
		RETURNING id::text
	`, fmt.Sprintf("host-%d@example.local", time.Now().UnixNano())).Scan(&id); err != nil {
		t.Fatalf("insert test host: %v", err)
	}

	return id
}

func insertTestWordSet(t *testing.T, ctx context.Context, pool *Pool, hostID string) string {
	t.Helper()

	var id string
	if err := pool.QueryRow(ctx, `
		INSERT INTO word_sets (name, status, source, created_by_user_id, approved_by_user_id, approved_at)
		VALUES ('Test Word Set', 'approved', 'manual', $1, $1, now())
		RETURNING id::text
	`, hostID).Scan(&id); err != nil {
		t.Fatalf("insert test word set: %v", err)
	}

	return id
}

func insertTestRun(t *testing.T, ctx context.Context, pool *Pool) string {
	t.Helper()

	hostID := insertTestHost(t, ctx, pool)
	wordSetID := insertTestWordSet(t, ctx, pool, hostID)

	var id string
	if err := pool.QueryRow(ctx, `
		INSERT INTO game_runs (host_user_id, word_set_id, code, name, status)
		VALUES ($1, $2, $3, 'Test Run', 'lobby_open')
		RETURNING id::text
	`, hostID, wordSetID, fmt.Sprintf("RUN-%d", time.Now().UnixNano())).Scan(&id); err != nil {
		t.Fatalf("insert test run: %v", err)
	}

	return id
}

func makeTestCardCells() []CreateCardCellParams {
	cells := make([]CreateCardCellParams, 0, 25)
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			isFreeSpace := row == 2 && col == 2
			word := fmt.Sprintf("Word %d-%d", row, col)
			if isFreeSpace {
				word = "FREE"
			}
			cells = append(cells, CreateCardCellParams{
				RowIndex:    row,
				ColIndex:    col,
				Word:        word,
				IsFreeSpace: isFreeSpace,
			})
		}
	}

	return cells
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}
