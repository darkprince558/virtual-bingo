package app

import (
	"bufio"
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/config"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/db"
	gamesvc "github.com/darkprince558/virtual-bingo/backend-go/internal/game"
	_ "github.com/lib/pq"
)

func TestAPIHealthAndVersionWithoutDatabase(t *testing.T) {
	handler := NewServer(config.Config{AppEnv: "test"}, testLogger(), nil).Handler

	versionResp := doRequest(t, handler, http.MethodGet, "/api/v1/version", nil)
	if versionResp.StatusCode != http.StatusOK {
		t.Fatalf("expected version 200, got %d: %s", versionResp.StatusCode, versionResp.Body)
	}
	if !strings.Contains(versionResp.Body, `"data"`) {
		t.Fatalf("expected enveloped version response, got %s", versionResp.Body)
	}

	configResp := doRequest(t, handler, http.MethodGet, "/api/v1/config", nil)
	if configResp.StatusCode != http.StatusOK {
		t.Fatalf("expected config 200, got %d: %s", configResp.StatusCode, configResp.Body)
	}
	if !strings.Contains(configResp.Body, `"sseEvents":true`) || !strings.Contains(configResp.Body, `"rewards":false`) {
		t.Fatalf("expected local capability flags in config, got %s", configResp.Body)
	}
	for _, expected := range []string{
		`"gameSettings":true`,
		`"autoMark":true`,
		`"voiceClaims":false`,
		`"aiCaller":true`,
		`"themeGenerator":true`,
		`"automation":true`,
		`"teamsApp":false`,
	} {
		if !strings.Contains(configResp.Body, expected) {
			t.Fatalf("expected config capability %s, got %s", expected, configResp.Body)
		}
	}

	readyResp := doRequest(t, handler, http.MethodGet, "/readyz", nil)
	if readyResp.StatusCode != http.StatusOK {
		t.Fatalf("expected readyz 200, got %d: %s", readyResp.StatusCode, readyResp.Body)
	}

	createResp := doRequest(t, handler, http.MethodPost, "/api/v1/games", map[string]any{"name": "No DB"})
	if createResp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected create game without DB to return 503, got %d: %s", createResp.StatusCode, createResp.Body)
	}
}

func TestAPIAuthModeSelectionAndUnauthorizedEnvelope(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	wordSetID := insertAPIWordSet(t, ctx, pool)
	devHandler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL, AuthMode: "dev"}, testLogger(), pool).Handler
	devResp := doRequest(t, devHandler, http.MethodPost, "/api/v1/games", map[string]any{
		"name":      "Dev Auth Game",
		"code":      "DEV-AUTH",
		"wordSetId": wordSetID,
	})
	if devResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected dev auth create game 201, got %d: %s", devResp.StatusCode, devResp.Body)
	}

	entraReadyHandler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL, AuthMode: "entra-ready"}, testLogger(), pool).Handler
	unauthorizedResp := doRequestWithoutDevAuth(t, entraReadyHandler, http.MethodPost, "/api/v1/games", map[string]any{
		"name":      "Blocked Auth Game",
		"code":      "BLOCKED-AUTH",
		"wordSetId": wordSetID,
	})
	if unauthorizedResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected entra-ready without verifier/token to return 401, got %d: %s", unauthorizedResp.StatusCode, unauthorizedResp.Body)
	}
	if !strings.Contains(unauthorizedResp.Body, `"error"`) || !strings.Contains(unauthorizedResp.Body, `"code":"unauthorized"`) {
		t.Fatalf("expected unauthorized error envelope, got %s", unauthorizedResp.Body)
	}
}

func TestAPIIdentityGameManagementAllowlistAndActivity(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	wordSetID := insertAPIWordSet(t, ctx, pool)
	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL, AuthMode: "dev"}, testLogger(), pool).Handler

	meResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/me", nil, "manager@example.local", "Manager User", "host")
	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("expected /me 200, got %d: %s", meResp.StatusCode, meResp.Body)
	}
	me := currentUserFromData(t, meResp.Body)
	if me.Email != "manager@example.local" || me.Role != "host" || me.AuthMode != "dev" {
		t.Fatalf("unexpected current user response: %+v", me)
	}

	gameID := createAPIGameWithDevAuth(t, handler, "Management Game", "MANAGE-CODE", wordSetID, nil, "manager@example.local", "Manager User", "host")
	codeResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/code/manage-code", nil)
	if codeResp.StatusCode != http.StatusOK || stringFromData(t, codeResp.Body, "id") != gameID {
		t.Fatalf("expected case-insensitive code lookup to return game %s, got %d: %s", gameID, codeResp.StatusCode, codeResp.Body)
	}

	hostListResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games?scope=host", nil, "manager@example.local", "Manager User", "host")
	if hostListResp.StatusCode != http.StatusOK {
		t.Fatalf("expected host list 200, got %d: %s", hostListResp.StatusCode, hostListResp.Body)
	}
	if summaries := gameRunSummariesFromData(t, hostListResp.Body); len(summaries) != 1 || summaries[0].ID != gameID {
		t.Fatalf("expected one hosted game, got %+v", summaries)
	}

	bulkResp := doRequestWithDevAuth(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/allowed-players/bulk", []map[string]any{
		{"email": "player-list@example.local", "displayName": "Player List"},
		{"email": "second-list@example.local", "displayName": "Second List"},
	}, "manager@example.local", "Manager User", "host")
	if bulkResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected bulk allowlist 201, got %d: %s", bulkResp.StatusCode, bulkResp.Body)
	}
	allowed := allowedPlayersFromData(t, bulkResp.Body)
	if len(allowed) != 2 || allowed[0].Email != "player-list@example.local" {
		t.Fatalf("unexpected bulk allowlist response: %+v", allowed)
	}

	duplicateResp := doRequestWithDevAuth(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/allowed-players/bulk", []map[string]any{
		{"email": "dupe@example.local", "displayName": "Dupe One"},
		{"email": "DUPE@example.local", "displayName": "Dupe Two"},
	}, "manager@example.local", "Manager User", "host")
	if duplicateResp.StatusCode != http.StatusConflict {
		t.Fatalf("expected duplicate bulk allowlist 409, got %d: %s", duplicateResp.StatusCode, duplicateResp.Body)
	}

	playerListResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games?scope=player", nil, "player-list@example.local", "Player List", "player")
	if playerListResp.StatusCode != http.StatusOK {
		t.Fatalf("expected player list 200, got %d: %s", playerListResp.StatusCode, playerListResp.Body)
	}
	if summaries := gameRunSummariesFromData(t, playerListResp.Body); len(summaries) != 1 || summaries[0].ID != gameID || summaries[0].AllowedPlayerCount != 2 {
		t.Fatalf("expected player-scoped game with counts, got %+v", summaries)
	}

	patchResp := doRequestWithDevAuth(t, handler, http.MethodPatch, "/api/v1/games/"+gameID, map[string]any{
		"name":           "Updated Management Game",
		"winningPattern": "four_corners",
	}, "manager@example.local", "Manager User", "host")
	if patchResp.StatusCode != http.StatusOK {
		t.Fatalf("expected draft/lobby update 200, got %d: %s", patchResp.StatusCode, patchResp.Body)
	}
	if got := stringFromData(t, patchResp.Body, "name"); got != "Updated Management Game" {
		t.Fatalf("expected updated name, got %s", got)
	}
	if got := stringFromData(t, patchResp.Body, "winningPattern"); got != "four_corners" {
		t.Fatalf("expected normalized winning pattern, got %s", got)
	}

	forbiddenPatchResp := doRequestWithDevAuth(t, handler, http.MethodPatch, "/api/v1/games/"+gameID, map[string]any{"name": "Wrong Host"}, "wrong-host@example.local", "Wrong Host", "host")
	if forbiddenPatchResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected wrong host patch 403, got %d: %s", forbiddenPatchResp.StatusCode, forbiddenPatchResp.Body)
	}

	deleteResp := doRequestWithDevAuth(t, handler, http.MethodDelete, "/api/v1/games/"+gameID+"/allowed-players/"+allowed[1].ID, nil, "manager@example.local", "Manager User", "host")
	if deleteResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected delete allowed player 204, got %d: %s", deleteResp.StatusCode, deleteResp.Body)
	}
	wrongGameDeleteResp := doRequestWithDevAuth(t, handler, http.MethodDelete, "/api/v1/games/00000000-0000-0000-0000-000000000000/allowed-players/"+allowed[0].ID, nil, "manager@example.local", "Manager User", "host")
	if wrongGameDeleteResp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected wrong game delete 404, got %d: %s", wrongGameDeleteResp.StatusCode, wrongGameDeleteResp.Body)
	}

	if resp := doRequestWithDevAuth(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/start", nil, "manager@example.local", "Manager User", "host"); resp.StatusCode != http.StatusOK {
		t.Fatalf("expected start 200, got %d: %s", resp.StatusCode, resp.Body)
	}
	livePatchResp := doRequestWithDevAuth(t, handler, http.MethodPatch, "/api/v1/games/"+gameID, map[string]any{"name": "Too Late"}, "manager@example.local", "Manager User", "host")
	if livePatchResp.StatusCode != http.StatusConflict {
		t.Fatalf("expected live game patch 409, got %d: %s", livePatchResp.StatusCode, livePatchResp.Body)
	}

	activityResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/activity", nil, "manager@example.local", "Manager User", "host")
	if activityResp.StatusCode != http.StatusOK {
		t.Fatalf("expected activity 200, got %d: %s", activityResp.StatusCode, activityResp.Body)
	}
	if events := activityEventsFromData(t, activityResp.Body); len(events) == 0 {
		t.Fatalf("expected committed activity events, got %+v", events)
	}
	adminListResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games?scope=admin&status=live", nil, "admin@example.local", "Admin User", "admin")
	if adminListResp.StatusCode != http.StatusOK {
		t.Fatalf("expected admin list 200, got %d: %s", adminListResp.StatusCode, adminListResp.Body)
	}
	if summaries := gameRunSummariesFromData(t, adminListResp.Body); len(summaries) != 1 || summaries[0].Status != "live" {
		t.Fatalf("expected admin live list to include started game, got %+v", summaries)
	}
}

func TestAPIGameAllowlistJoinAndCardFlow(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	wordSetID := insertAPIWordSet(t, ctx, pool)
	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL}, testLogger(), pool).Handler

	createResp := doRequest(t, handler, http.MethodPost, "/api/v1/games", map[string]any{
		"name":      "API Test Game",
		"code":      "API-TEST",
		"wordSetId": wordSetID,
	})
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create game 201, got %d: %s", createResp.StatusCode, createResp.Body)
	}
	gameID := stringFromData(t, createResp.Body, "id")

	getResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID, nil)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected get game 200, got %d: %s", getResp.StatusCode, getResp.Body)
	}

	allowedResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/allowed-players", map[string]any{
		"email":       "alex@example.local",
		"displayName": "Alex Demo",
	})
	if allowedResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected add allowed player 201, got %d: %s", allowedResp.StatusCode, allowedResp.Body)
	}

	listResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/allowed-players", nil)
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected list allowed players 200, got %d: %s", listResp.StatusCode, listResp.Body)
	}

	blockedJoinResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players", map[string]any{
		"email":       "not-allowed@example.local",
		"displayName": "Not Allowed",
	})
	if blockedJoinResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected blocked join 403, got %d: %s", blockedJoinResp.StatusCode, blockedJoinResp.Body)
	}

	joinResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players", map[string]any{
		"email":       "alex@example.local",
		"displayName": "Alex Demo",
	})
	if joinResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected join player 201, got %d: %s", joinResp.StatusCode, joinResp.Body)
	}
	playerID := stringFromData(t, joinResp.Body, "id")

	rejoinResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players", map[string]any{
		"email":       "alex@example.local",
		"displayName": "Alex Demo",
	})
	if rejoinResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected rejoin player 201, got %d: %s", rejoinResp.StatusCode, rejoinResp.Body)
	}
	if got := stringFromData(t, rejoinResp.Body, "id"); got != playerID {
		t.Fatalf("expected rejoin to return existing player %s, got %s", playerID, got)
	}

	cardResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players/"+playerID+"/card", nil)
	if cardResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected assign card 201, got %d: %s", cardResp.StatusCode, cardResp.Body)
	}
	cardID := stringFromData(t, cardResp.Body, "id")
	if cells := cellsFromData(t, cardResp.Body); len(cells) != 25 {
		t.Fatalf("expected 25 cells, got %d", len(cells))
	}

	secondCardResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players/"+playerID+"/card", nil)
	if secondCardResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected second assign card 201, got %d: %s", secondCardResp.StatusCode, secondCardResp.Body)
	}
	if got := stringFromData(t, secondCardResp.Body, "id"); got != cardID {
		t.Fatalf("expected second card assign to return existing card %s, got %s", cardID, got)
	}

	getCardResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/"+playerID+"/card", nil)
	if getCardResp.StatusCode != http.StatusOK {
		t.Fatalf("expected get card 200, got %d: %s", getCardResp.StatusCode, getCardResp.Body)
	}
	if cells := cellsFromData(t, getCardResp.Body); len(cells) != 25 || !cells[12].IsFreeSpace {
		t.Fatalf("expected persisted card cells with center free space, got %+v", cells)
	}
}

func TestAPIWordSetManagement(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL}, testLogger(), pool).Handler

	createResp := doRequest(t, handler, http.MethodPost, "/api/v1/word-sets", map[string]any{
		"name":   "Manual Word Set",
		"status": "draft",
		"source": "manual",
		"words": []map[string]any{
			{"word": "Discovery"},
			{"word": "Planning"},
		},
	})
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create word set 201, got %d: %s", createResp.StatusCode, createResp.Body)
	}
	wordSet := wordSetFromData(t, createResp.Body)
	if wordSet.Name != "Manual Word Set" || len(wordSet.Words) != 2 || !wordSet.Words[0].IsActive {
		t.Fatalf("unexpected created word set: %+v", wordSet)
	}

	listResp := doRequest(t, handler, http.MethodGet, "/api/v1/word-sets", nil)
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("expected list word sets 200, got %d: %s", listResp.StatusCode, listResp.Body)
	}
	if wordSets := wordSetsFromData(t, listResp.Body); len(wordSets) != 1 || wordSets[0].ID != wordSet.ID {
		t.Fatalf("expected created word set in list, got %+v", wordSets)
	}

	patchResp := doRequest(t, handler, http.MethodPatch, "/api/v1/word-sets/"+wordSet.ID, map[string]any{
		"name":   "Approved Manual Word Set",
		"status": "approved",
	})
	if patchResp.StatusCode != http.StatusOK {
		t.Fatalf("expected update word set 200, got %d: %s", patchResp.StatusCode, patchResp.Body)
	}
	updated := wordSetFromData(t, patchResp.Body)
	if updated.Name != "Approved Manual Word Set" || updated.Status != "approved" {
		t.Fatalf("unexpected updated word set: %+v", updated)
	}

	createWordResp := doRequest(t, handler, http.MethodPost, "/api/v1/word-sets/"+wordSet.ID+"/words", map[string]any{
		"word":      "Launch",
		"sortOrder": 3,
	})
	if createWordResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create word 201, got %d: %s", createWordResp.StatusCode, createWordResp.Body)
	}
	word := wordSetWordFromData(t, createWordResp.Body)
	if word.Word != "Launch" || word.SortOrder != 3 {
		t.Fatalf("unexpected created word: %+v", word)
	}

	patchWordResp := doRequest(t, handler, http.MethodPatch, "/api/v1/word-sets/"+wordSet.ID+"/words/"+word.ID, map[string]any{
		"word": "Launch Plan",
	})
	if patchWordResp.StatusCode != http.StatusOK {
		t.Fatalf("expected update word 200, got %d: %s", patchWordResp.StatusCode, patchWordResp.Body)
	}
	if patchedWord := wordSetWordFromData(t, patchWordResp.Body); patchedWord.Word != "Launch Plan" {
		t.Fatalf("expected patched word text, got %+v", patchedWord)
	}

	deleteWordResp := doRequest(t, handler, http.MethodDelete, "/api/v1/word-sets/"+wordSet.ID+"/words/"+word.ID, nil)
	if deleteWordResp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected delete word 204, got %d: %s", deleteWordResp.StatusCode, deleteWordResp.Body)
	}
	detailResp := doRequest(t, handler, http.MethodGet, "/api/v1/word-sets/"+wordSet.ID, nil)
	if detailResp.StatusCode != http.StatusOK {
		t.Fatalf("expected detail word set 200, got %d: %s", detailResp.StatusCode, detailResp.Body)
	}
	detail := wordSetFromData(t, detailResp.Body)
	if len(detail.Words) != 3 || activeWordCount(detail.Words) != 2 {
		t.Fatalf("expected soft-deleted inactive word in management detail, got %+v", detail.Words)
	}
}

func TestAPIPlayerReconnectHeartbeatAndAuthorization(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	wordSetID := insertAPIWordSet(t, ctx, pool)
	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL}, testLogger(), pool).Handler
	gameID := createAPIGame(t, handler, "Reconnect Test Game", "RECONNECT-TEST", wordSetID, nil)
	playerID, _ := allowJoinAndAssignCard(t, handler, gameID, "reconnect@example.local", "Reconnect Player")
	initialPlayer := playerFromData(t, doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players", map[string]any{
		"email":       "reconnect@example.local",
		"displayName": "Reconnect Player",
	}).Body)

	if _, err := pool.Exec(ctx, `
		UPDATE players
		SET connection_state = 'offline',
		    last_seen_at = now() - interval '1 hour',
		    updated_at = now()
		WHERE id = $1
	`, playerID); err != nil {
		t.Fatalf("force player offline: %v", err)
	}

	rejoinResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players", map[string]any{
		"email":       "reconnect@example.local",
		"displayName": "Reconnect Player",
	})
	if rejoinResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected rejoin 201, got %d: %s", rejoinResp.StatusCode, rejoinResp.Body)
	}
	rejoined := playerFromData(t, rejoinResp.Body)
	if rejoined.ID != playerID || rejoined.ConnectionState != "online" || !rejoined.LastSeenAt.After(initialPlayer.LastSeenAt) {
		t.Fatalf("expected rejoin to refresh existing online player, got %+v from initial %+v", rejoined, initialPlayer)
	}

	if _, err := pool.Exec(ctx, `
		UPDATE players
		SET connection_state = 'offline',
		    last_seen_at = now() - interval '1 hour',
		    updated_at = now()
		WHERE id = $1
	`, playerID); err != nil {
		t.Fatalf("force player offline before heartbeat: %v", err)
	}

	heartbeatResp := doRequestWithDevAuth(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players/"+playerID+"/heartbeat", nil, "reconnect@example.local", "Reconnect Player", "player")
	if heartbeatResp.StatusCode != http.StatusOK {
		t.Fatalf("expected heartbeat 200, got %d: %s", heartbeatResp.StatusCode, heartbeatResp.Body)
	}
	heartbeatPlayer := playerFromData(t, heartbeatResp.Body)
	if heartbeatPlayer.ConnectionState != "online" || !heartbeatPlayer.LastSeenAt.After(rejoined.LastSeenAt) {
		t.Fatalf("expected heartbeat to refresh player, got %+v after %+v", heartbeatPlayer, rejoined)
	}

	blockedHeartbeatResp := doRequestWithDevAuth(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players/"+playerID+"/heartbeat", nil, "other@example.local", "Other Player", "player")
	if blockedHeartbeatResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected other player heartbeat 403, got %d: %s", blockedHeartbeatResp.StatusCode, blockedHeartbeatResp.Body)
	}

	if _, err := pool.Exec(ctx, `
		UPDATE players
		SET connection_state = 'offline',
		    last_seen_at = now() - interval '1 hour',
		    updated_at = now()
		WHERE id = $1
	`, playerID); err != nil {
		t.Fatalf("force player offline before snapshot: %v", err)
	}

	playerSnapshotResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/"+playerID+"/snapshot", nil, "reconnect@example.local", "Reconnect Player", "player")
	if playerSnapshotResp.StatusCode != http.StatusOK {
		t.Fatalf("expected player snapshot 200, got %d: %s", playerSnapshotResp.StatusCode, playerSnapshotResp.Body)
	}
	playerSnapshot := playerSnapshotFromData(t, playerSnapshotResp.Body)
	if playerSnapshot.Player.ConnectionState != "online" || !playerSnapshot.Player.LastSeenAt.After(heartbeatPlayer.LastSeenAt) {
		t.Fatalf("expected player snapshot to reconnect player, got %+v", playerSnapshot.Player)
	}

	blockedSnapshotResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/"+playerID+"/snapshot", nil, "other@example.local", "Other Player", "player")
	if blockedSnapshotResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected other player snapshot 403, got %d: %s", blockedSnapshotResp.StatusCode, blockedSnapshotResp.Body)
	}

	hostSnapshotResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/host-snapshot", nil)
	if hostSnapshotResp.StatusCode != http.StatusOK {
		t.Fatalf("expected host snapshot 200, got %d: %s", hostSnapshotResp.StatusCode, hostSnapshotResp.Body)
	}
	hostSnapshot := hostSnapshotFromData(t, hostSnapshotResp.Body)
	if len(hostSnapshot.Players) != 1 || hostSnapshot.Players[0].ConnectionState != "online" || hostSnapshot.Players[0].LastSeenAt.IsZero() {
		t.Fatalf("expected host snapshot connection fields, got %+v", hostSnapshot.Players)
	}

	events := gameEventsFromDB(t, ctx, pool, gameID)
	assertEventTypes(t, events, []string{"player.joined", "player.reconnected"})
}

func TestAPICurrentPlayerRoutesAndClaimDetailAuthorization(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	wordSetID := insertAPIWordSet(t, ctx, pool)
	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL}, testLogger(), pool).Handler
	gameID := createAPIGame(t, handler, "Current Player Game", "CURRENT-PLAYER", wordSetID, nil)

	if resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/allowed-players", map[string]any{
		"email":       "self@example.local",
		"displayName": "Self Player",
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected allowed self player 201, got %d: %s", resp.StatusCode, resp.Body)
	}
	joinResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players", map[string]any{
		"email":       "self@example.local",
		"displayName": "Self Player",
	})
	if joinResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected self player join 201, got %d: %s", joinResp.StatusCode, joinResp.Body)
	}
	playerID := stringFromData(t, joinResp.Body, "id")

	selfCardResp := doRequestWithDevAuth(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players/me/card", nil, "self@example.local", "Self Player", "player")
	if selfCardResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected self card assign 201, got %d: %s", selfCardResp.StatusCode, selfCardResp.Body)
	}
	card := cardFromData(t, selfCardResp.Body)
	if card.PlayerID != playerID || len(card.Cells) != 25 {
		t.Fatalf("expected self card for current player, got %+v", card)
	}
	selfGetCardResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/me/card", nil, "self@example.local", "Self Player", "player")
	if selfGetCardResp.StatusCode != http.StatusOK {
		t.Fatalf("expected self get card 200, got %d: %s", selfGetCardResp.StatusCode, selfGetCardResp.Body)
	}

	cell := firstNonFreeCell(t, card.Cells)
	markResp := doRequestWithDevAuth(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/players/me/card/cells/"+cell.ID, map[string]any{"marked": true}, "self@example.local", "Self Player", "player")
	if markResp.StatusCode != http.StatusOK {
		t.Fatalf("expected self mark 200, got %d: %s", markResp.StatusCode, markResp.Body)
	}
	heartbeatResp := doRequestWithDevAuth(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players/me/heartbeat", nil, "self@example.local", "Self Player", "player")
	if heartbeatResp.StatusCode != http.StatusOK {
		t.Fatalf("expected self heartbeat 200, got %d: %s", heartbeatResp.StatusCode, heartbeatResp.Body)
	}

	if resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/start", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("expected start 200, got %d: %s", resp.StatusCode, resp.Body)
	}
	claimResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/claims", map[string]any{
		"playerId": playerID,
		"pattern":  "single_line",
	})
	if claimResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected claim 201, got %d: %s", claimResp.StatusCode, claimResp.Body)
	}
	claim := claimSubmissionFromData(t, claimResp.Body).Claim

	selfSnapshotResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/me/snapshot", nil, "self@example.local", "Self Player", "player")
	if selfSnapshotResp.StatusCode != http.StatusOK {
		t.Fatalf("expected self snapshot 200, got %d: %s", selfSnapshotResp.StatusCode, selfSnapshotResp.Body)
	}
	if snapshot := playerSnapshotFromData(t, selfSnapshotResp.Body); snapshot.Player.ID != playerID || snapshot.Card == nil || len(snapshot.Claims) != 1 {
		t.Fatalf("unexpected self snapshot: %+v", snapshot)
	}

	ownClaimResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/claims/"+claim.ID, nil, "self@example.local", "Self Player", "player")
	if ownClaimResp.StatusCode != http.StatusOK {
		t.Fatalf("expected own claim detail 200, got %d: %s", ownClaimResp.StatusCode, ownClaimResp.Body)
	}
	otherClaimResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/claims/"+claim.ID, nil, "other-claim@example.local", "Other Claim", "player")
	if otherClaimResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected other player claim detail 403, got %d: %s", otherClaimResp.StatusCode, otherClaimResp.Body)
	}
	hostClaimResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/claims/"+claim.ID, nil)
	if hostClaimResp.StatusCode != http.StatusOK {
		t.Fatalf("expected host claim detail 200, got %d: %s", hostClaimResp.StatusCode, hostClaimResp.Body)
	}
}

func TestAPISettingsPreferencesAutoMarkAndClaimReadiness(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	wordSetID := insertAPIWordSet(t, ctx, pool)
	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL}, testLogger(), pool).Handler

	gameID := createAPIGame(t, handler, "Settings Auto Mark Game", "SETTINGS-AUTO", wordSetID, nil)
	playerID, card := allowJoinAndAssignCard(t, handler, gameID, "settings-player@example.local", "Settings Player")
	targets := sortedNonFreeCellsByWordNumber(t, card.Cells)
	firstTarget := targets[0]
	secondTarget := targets[1]

	defaultSettingsResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/settings", nil)
	if defaultSettingsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected default settings 200, got %d: %s", defaultSettingsResp.StatusCode, defaultSettingsResp.Body)
	}
	defaultSettings := gameSettingsFromData(t, defaultSettingsResp.Body)
	if defaultSettings.MarkingMode != "manual" || !defaultSettings.ShowClaimReadiness || defaultSettings.VoiceClaimMode != "off" || defaultSettings.CallerMode != "off" || defaultSettings.ThemeMode != "default" {
		t.Fatalf("unexpected default settings: %+v", defaultSettings)
	}

	invalidSettingsResp := doRequest(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/settings", map[string]any{"markingMode": "robot"})
	if invalidSettingsResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected invalid marking mode 400, got %d: %s", invalidSettingsResp.StatusCode, invalidSettingsResp.Body)
	}
	playerPatchSettingsResp := doRequestWithDevAuth(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/settings", map[string]any{"markingMode": "auto_mark"}, "settings-player@example.local", "Settings Player", "player")
	if playerPatchSettingsResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected player settings patch 403, got %d: %s", playerPatchSettingsResp.StatusCode, playerPatchSettingsResp.Body)
	}

	patchSettingsResp := doRequest(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/settings", map[string]any{
		"markingMode":                  "auto_mark",
		"allowPlayerMarkingModeChoice": false,
		"showClaimReadiness":           true,
		"voiceClaimMode":               "optional",
		"voiceClaimAutoplay":           true,
		"callerMode":                   "text_only",
		"themeMode":                    "manual",
	})
	if patchSettingsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected patch settings 200, got %d: %s", patchSettingsResp.StatusCode, patchSettingsResp.Body)
	}
	patchedSettings := gameSettingsFromData(t, patchSettingsResp.Body)
	if patchedSettings.MarkingMode != "auto_mark" || patchedSettings.AllowPlayerMarkingModeChoice || patchedSettings.VoiceClaimMode != "optional" || !patchedSettings.VoiceClaimAutoplay {
		t.Fatalf("unexpected patched settings: %+v", patchedSettings)
	}

	prefsResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/me/preferences", nil, "settings-player@example.local", "Settings Player", "player")
	if prefsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected default preferences 200, got %d: %s", prefsResp.StatusCode, prefsResp.Body)
	}
	if prefs := playerPreferencesFromData(t, prefsResp.Body); prefs.PlayerID != playerID || prefs.MarkingMode != nil {
		t.Fatalf("unexpected default preferences: %+v", prefs)
	}
	manualPrefResp := doRequestWithDevAuth(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/players/me/preferences", map[string]any{"markingMode": "manual"}, "settings-player@example.local", "Settings Player", "player")
	if manualPrefResp.StatusCode != http.StatusOK {
		t.Fatalf("expected manual preference 200, got %d: %s", manualPrefResp.StatusCode, manualPrefResp.Body)
	}

	settingWinsSnapshotResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/me/snapshot", nil, "settings-player@example.local", "Settings Player", "player")
	if settingWinsSnapshotResp.StatusCode != http.StatusOK {
		t.Fatalf("expected player snapshot 200, got %d: %s", settingWinsSnapshotResp.StatusCode, settingWinsSnapshotResp.Body)
	}
	if snapshot := playerSnapshotFromData(t, settingWinsSnapshotResp.Body); snapshot.MarkingMode != "auto_mark" || snapshot.AllowPlayerMarkingModeChoice {
		t.Fatalf("expected game setting to win while player choice disabled, got %+v", snapshot)
	}

	allowChoiceResp := doRequest(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/settings", map[string]any{"allowPlayerMarkingModeChoice": true})
	if allowChoiceResp.StatusCode != http.StatusOK {
		t.Fatalf("expected allow-choice settings patch 200, got %d: %s", allowChoiceResp.StatusCode, allowChoiceResp.Body)
	}
	playerWinsSnapshotResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/me/snapshot", nil, "settings-player@example.local", "Settings Player", "player")
	if playerWinsSnapshotResp.StatusCode != http.StatusOK {
		t.Fatalf("expected player snapshot after allow choice 200, got %d: %s", playerWinsSnapshotResp.StatusCode, playerWinsSnapshotResp.Body)
	}
	if snapshot := playerSnapshotFromData(t, playerWinsSnapshotResp.Body); snapshot.MarkingMode != "manual" || !snapshot.AllowPlayerMarkingModeChoice || !snapshot.ShowClaimReadiness {
		t.Fatalf("expected player manual preference to win when choice enabled, got %+v", snapshot)
	}

	if resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/start", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("expected start 200, got %d: %s", resp.StatusCode, resp.Body)
	}
	callUntilWord(t, handler, gameID, firstTarget.Word)
	cardAfterManualCall := fetchCard(t, handler, gameID, playerID)
	if cellByID(t, cardAfterManualCall.Cells, firstTarget.ID).MarkedAt != nil {
		t.Fatalf("expected manual effective mode not to auto-mark %s", firstTarget.Word)
	}
	skippedRunResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/auto-mark/run", nil)
	if skippedRunResp.StatusCode != http.StatusOK {
		t.Fatalf("expected skipped auto-mark run 200, got %d: %s", skippedRunResp.StatusCode, skippedRunResp.Body)
	}
	if skipped := autoMarkRunFromData(t, skippedRunResp.Body); skipped.CellsMarked != 0 || skipped.SkippedReason != "no_players_using_auto_mark" {
		t.Fatalf("expected skipped auto-mark run for manual preference, got %+v", skipped)
	}

	autoPrefResp := doRequestWithDevAuth(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/players/me/preferences", map[string]any{"markingMode": "auto_mark"}, "settings-player@example.local", "Settings Player", "player")
	if autoPrefResp.StatusCode != http.StatusOK {
		t.Fatalf("expected auto preference 200, got %d: %s", autoPrefResp.StatusCode, autoPrefResp.Body)
	}
	backfillResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/auto-mark/run", nil)
	if backfillResp.StatusCode != http.StatusOK {
		t.Fatalf("expected backfill auto-mark run 200, got %d: %s", backfillResp.StatusCode, backfillResp.Body)
	}
	backfill := autoMarkRunFromData(t, backfillResp.Body)
	if backfill.CellsMarked == 0 || backfill.PlayersScanned != 1 || backfill.PlayersMarked != 1 {
		t.Fatalf("expected backfill to mark missed called cells, got %+v", backfill)
	}
	cardAfterBackfill := fetchCard(t, handler, gameID, playerID)
	if cellByID(t, cardAfterBackfill.Cells, firstTarget.ID).MarkedAt == nil {
		t.Fatalf("expected backfill to mark %s", firstTarget.Word)
	}
	idempotentResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/auto-mark/run", nil)
	if idempotentResp.StatusCode != http.StatusOK {
		t.Fatalf("expected idempotent auto-mark run 200, got %d: %s", idempotentResp.StatusCode, idempotentResp.Body)
	}
	if idempotent := autoMarkRunFromData(t, idempotentResp.Body); idempotent.CellsMarked != 0 {
		t.Fatalf("expected second auto-mark run to mark 0 cells, got %+v", idempotent)
	}

	callUntilWord(t, handler, gameID, secondTarget.Word)
	cardAfterAutoCall := fetchCard(t, handler, gameID, playerID)
	if cellByID(t, cardAfterAutoCall.Cells, secondTarget.ID).MarkedAt == nil {
		t.Fatalf("expected auto_mark effective mode to mark %s after call", secondTarget.Word)
	}
	assertEventTypes(t, gameEventsFromDB(t, ctx, pool, gameID), []string{"game.settings_updated", "player.preferences_updated", "card.auto_marked"})

	assistGameID := createAPIGame(t, handler, "Assist Readiness Game", "ASSIST-READY", wordSetID, nil)
	assistPlayerID, assistCard := allowJoinAndAssignCard(t, handler, assistGameID, "assist-player@example.local", "Assist Player")
	assistTarget := sortedNonFreeCellsByWordNumber(t, assistCard.Cells)[0]
	if resp := doRequest(t, handler, http.MethodPatch, "/api/v1/games/"+assistGameID+"/settings", map[string]any{"markingMode": "assist"}); resp.StatusCode != http.StatusOK {
		t.Fatalf("expected assist settings patch 200, got %d: %s", resp.StatusCode, resp.Body)
	}
	if resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+assistGameID+"/start", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("expected assist game start 200, got %d: %s", resp.StatusCode, resp.Body)
	}
	callUntilWord(t, handler, assistGameID, assistTarget.Word)
	assistCardAfterCall := fetchCard(t, handler, assistGameID, assistPlayerID)
	if cellByID(t, assistCardAfterCall.Cells, assistTarget.ID).MarkedAt != nil {
		t.Fatalf("expected assist mode not to auto-mark %s", assistTarget.Word)
	}
	callUntilExhausted(t, handler, assistGameID)
	readinessResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+assistGameID+"/players/me/claim-readiness", nil, "assist-player@example.local", "Assist Player", "player")
	if readinessResp.StatusCode != http.StatusOK {
		t.Fatalf("expected claim-readiness 200, got %d: %s", readinessResp.StatusCode, readinessResp.Body)
	}
	readiness := claimReadinessFromData(t, readinessResp.Body)
	if !readiness.Ready || readiness.BestPattern != "single_line" || len(readiness.ReadyPatterns) == 0 || len(readiness.MatchedCells) == 0 || len(readiness.MissingCells) != 0 {
		t.Fatalf("expected ready claim-readiness after all words called, got %+v", readiness)
	}
}

func TestAPIGeneratedContentPrepareReviewLockAndCardAssignment(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL, AuthMode: "dev"}, testLogger(), pool).Handler

	gameID := createAPIGame(t, handler, "Generated Content Game", "CONTENT-GAME", "", nil)
	prepareResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/content/prepare", nil)
	if prepareResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected content prepare 201, got %d: %s", prepareResp.StatusCode, prepareResp.Body)
	}
	prepared := gameContentFromData(t, prepareResp.Body)
	if prepared.Status != "generated" || prepared.Topic == "" || len(prepared.Words) < 24 {
		t.Fatalf("unexpected prepared content: %+v", prepared)
	}

	getResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/content", nil)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("expected content get 200, got %d: %s", getResp.StatusCode, getResp.Body)
	}
	if got := gameContentFromData(t, getResp.Body); got.ID != prepared.ID {
		t.Fatalf("expected get to return prepared content %s, got %+v", prepared.ID, got)
	}

	editedWords := make([]string, 0, 26)
	for index := 1; index <= 26; index++ {
		editedWords = append(editedWords, fmt.Sprintf("Edited Word %02d", index))
	}
	patchResp := doRequest(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/content", map[string]any{
		"topic":   "Edited Topic",
		"summary": "Edited Summary",
		"words":   editedWords,
	})
	if patchResp.StatusCode != http.StatusOK {
		t.Fatalf("expected content patch 200, got %d: %s", patchResp.StatusCode, patchResp.Body)
	}
	edited := gameContentFromData(t, patchResp.Body)
	if edited.Status != "edited" || edited.Topic != "Edited Topic" || len(edited.Words) != 26 {
		t.Fatalf("unexpected edited content: %+v", edited)
	}

	lockResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/content/lock", nil)
	if lockResp.StatusCode != http.StatusOK {
		t.Fatalf("expected content lock 200, got %d: %s", lockResp.StatusCode, lockResp.Body)
	}
	locked := gameContentFromData(t, lockResp.Body)
	if locked.Status != "locked" || locked.LockedAt == nil || locked.LockedWordSetID == nil {
		t.Fatalf("unexpected locked content: %+v", locked)
	}
	deck := callDeckRowsFromDB(t, ctx, pool, gameID)
	if len(deck) != len(editedWords) {
		t.Fatalf("expected locked deck with %d words, got %d", len(editedWords), len(deck))
	}

	postLockPatchResp := doRequest(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/content", map[string]any{"topic": "Too Late"})
	if postLockPatchResp.StatusCode != http.StatusConflict {
		t.Fatalf("expected post-lock patch 409, got %d: %s", postLockPatchResp.StatusCode, postLockPatchResp.Body)
	}

	if resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/allowed-players", map[string]any{
		"email":       "content-player@example.local",
		"displayName": "Content Player",
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected allowed player 201, got %d: %s", resp.StatusCode, resp.Body)
	}
	playerResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players", map[string]any{
		"email":       "content-player@example.local",
		"displayName": "Content Player",
	})
	if playerResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected join 201, got %d: %s", playerResp.StatusCode, playerResp.Body)
	}
	playerID := stringFromData(t, playerResp.Body, "id")
	cardResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players/"+playerID+"/card", nil)
	if cardResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected locked content card assignment 201, got %d: %s", cardResp.StatusCode, cardResp.Body)
	}
	if cells := cellsFromData(t, cardResp.Body); len(cells) != 25 {
		t.Fatalf("expected card from locked content words, got %d cells", len(cells))
	}

	if resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/allowed-players", map[string]any{
		"email":       "invite-two@example.local",
		"displayName": "Invite Two",
	}); resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected second allowed player 201, got %d: %s", resp.StatusCode, resp.Body)
	}
	inviteResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/deliveries/player-invites", nil)
	if inviteResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected mock invite delivery 201, got %d: %s", inviteResp.StatusCode, inviteResp.Body)
	}
	invites := deliveryAttemptsFromData(t, inviteResp.Body)
	if len(invites) != 2 {
		t.Fatalf("expected two invite attempts, got %+v", invites)
	}
	for _, invite := range invites {
		if invite.Status != "sent" || invite.GameCode != "CONTENT-GAME" || invite.LinkURL != "/join?code=CONTENT-GAME" {
			t.Fatalf("unexpected mock invite attempt: %+v", invite)
		}
	}
	listDeliveriesResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/deliveries", nil)
	if listDeliveriesResp.StatusCode != http.StatusOK {
		t.Fatalf("expected delivery list 200, got %d: %s", listDeliveriesResp.StatusCode, listDeliveriesResp.Body)
	}
	if listed := deliveryAttemptsFromData(t, listDeliveriesResp.Body); len(listed) != 2 {
		t.Fatalf("expected listed deliveries to include two attempts, got %+v", listed)
	}

	if _, err := pool.Exec(ctx, `UPDATE game_runs SET status = 'scheduled', updated_at = now() WHERE id = $1`, gameID); err != nil {
		t.Fatalf("force scheduled status before lobby open: %v", err)
	}
	openLobbyResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/lobby/open", nil)
	if openLobbyResp.StatusCode != http.StatusOK {
		t.Fatalf("expected lobby open 200, got %d: %s", openLobbyResp.StatusCode, openLobbyResp.Body)
	}
	if !strings.Contains(openLobbyResp.Body, `"status":"lobby_open"`) {
		t.Fatalf("expected lobby open response status, got %s", openLobbyResp.Body)
	}

	profileResp := doRequestWithDevAuth(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/players/me/profile", map[string]any{
		"icon":        "rocket",
		"avatarColor": "#1F7A8C",
		"avatarLabel": "CP",
	}, "content-player@example.local", "Content Player", "player")
	if profileResp.StatusCode != http.StatusOK {
		t.Fatalf("expected profile update 200, got %d: %s", profileResp.StatusCode, profileResp.Body)
	}
	profile := playerFromData(t, profileResp.Body)
	if profile.Icon == nil || *profile.Icon != "rocket" || profile.AvatarColor == nil || *profile.AvatarColor != "#1F7A8C" || profile.AvatarLabel == nil || *profile.AvatarLabel != "CP" {
		t.Fatalf("unexpected player profile response: %+v", profile)
	}

	themeResp := doRequest(t, handler, http.MethodPost, "/api/v1/themes/generate", map[string]any{
		"gameRunId": gameID,
		"prompt":    "Make a restrained internal celebration theme",
		"tone":      "warm",
	})
	if themeResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected theme generate 201, got %d: %s", themeResp.StatusCode, themeResp.Body)
	}
	theme := themeFromData(t, themeResp.Body)
	if theme.ID == "" || theme.Status != "draft" || theme.Tokens.Palette["primary"] == "" {
		t.Fatalf("unexpected generated theme: %+v", theme)
	}
	updateThemeResp := doRequest(t, handler, http.MethodPatch, "/api/v1/themes/"+theme.ID, map[string]any{
		"name":        "Internal Celebration",
		"decorations": []string{"confetti", "star"},
	})
	if updateThemeResp.StatusCode != http.StatusOK {
		t.Fatalf("expected theme update 200, got %d: %s", updateThemeResp.StatusCode, updateThemeResp.Body)
	}
	approveResp := doRequest(t, handler, http.MethodPost, "/api/v1/themes/"+theme.ID+"/approve", nil)
	if approveResp.StatusCode != http.StatusOK {
		t.Fatalf("expected theme approve 200, got %d: %s", approveResp.StatusCode, approveResp.Body)
	}
	if approved := themeFromData(t, approveResp.Body); approved.Status != "approved" {
		t.Fatalf("expected approved theme, got %+v", approved)
	}
	applyThemeResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/theme", map[string]any{"themeId": theme.ID})
	if applyThemeResp.StatusCode != http.StatusOK {
		t.Fatalf("expected theme apply 200, got %d: %s", applyThemeResp.StatusCode, applyThemeResp.Body)
	}
	if settings := gameSettingsFromData(t, applyThemeResp.Body); settings.ThemeID == nil || *settings.ThemeID != theme.ID || settings.ThemeMode != "ai_generated" {
		t.Fatalf("unexpected applied theme settings: %+v", settings)
	}

	assetsResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/caller-assets/generate", nil)
	if assetsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected caller assets generate 200, got %d: %s", assetsResp.StatusCode, assetsResp.Body)
	}
	assets := callerAssetsFromData(t, assetsResp.Body)
	if len(assets) != len(deck) {
		t.Fatalf("expected one caller asset per deck item, got %d assets for %d deck items", len(assets), len(deck))
	}
	if assets[0].Status != "ready" || assets[0].Line == "" || assets[0].CallDeckItemID != deck[0].ID {
		t.Fatalf("unexpected first caller asset: %+v for deck %+v", assets[0], deck[0])
	}

	playerSnapshotResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/me/snapshot", nil, "content-player@example.local", "Content Player", "player")
	if playerSnapshotResp.StatusCode != http.StatusOK {
		t.Fatalf("expected player snapshot 200, got %d: %s", playerSnapshotResp.StatusCode, playerSnapshotResp.Body)
	}
	playerSnapshot := playerSnapshotFromData(t, playerSnapshotResp.Body)
	if playerSnapshot.Player.Icon == nil || *playerSnapshot.Player.Icon != "rocket" || playerSnapshot.AppliedTheme == nil || playerSnapshot.AppliedTheme.ID != theme.ID {
		t.Fatalf("expected profile and applied theme in player snapshot, got %+v", playerSnapshot)
	}

	startResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/start", nil)
	if startResp.StatusCode != http.StatusOK {
		t.Fatalf("expected start after lobby open 200, got %d: %s", startResp.StatusCode, startResp.Body)
	}
	callResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
	if callResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected deck-backed call 201, got %d: %s", callResp.StatusCode, callResp.Body)
	}
	called := calledWordFromData(t, callResp.Body)
	if called.Word != deck[0].Word || called.Sequence != deck[0].Sequence || called.CallerAsset == nil || called.CallerAsset.Line == "" {
		t.Fatalf("expected first call to follow locked deck with caller asset, got called=%+v deck=%+v", called, deck[0])
	}
	hostSnapshotResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/host-snapshot", nil)
	if hostSnapshotResp.StatusCode != http.StatusOK {
		t.Fatalf("expected host snapshot 200, got %d: %s", hostSnapshotResp.StatusCode, hostSnapshotResp.Body)
	}
	hostSnapshot := hostSnapshotFromData(t, hostSnapshotResp.Body)
	if hostSnapshot.CurrentCallerAsset == nil || hostSnapshot.CurrentCallerAsset.Word != called.Word || hostSnapshot.AppliedTheme == nil || hostSnapshot.AppliedTheme.ID != theme.ID {
		t.Fatalf("expected caller metadata and theme in host snapshot, got %+v", hostSnapshot)
	}
}

func TestAPIPlayerTimeoutDisconnectAndReconnectNotice(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	wordSetID := insertAPIWordSet(t, ctx, pool)
	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL}, testLogger(), pool).Handler
	gameID := createAPIGame(t, handler, "Timeout Test Game", "TIMEOUT-TEST", wordSetID, nil)
	playerID, _ := allowJoinAndAssignCard(t, handler, gameID, "timeout@example.local", "Timeout Player")

	if resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/start", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("expected start 200, got %d: %s", resp.StatusCode, resp.Body)
	}

	staleSeenAt := time.Now().Add(-2 * time.Hour).UTC().Truncate(time.Microsecond)
	if _, err := pool.Exec(ctx, `
		UPDATE players
		SET connection_state = 'online',
		    last_seen_at = $2,
		    updated_at = now()
		WHERE id = $1
	`, playerID, staleSeenAt); err != nil {
		t.Fatalf("force stale online player: %v", err)
	}

	firstCallResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
	if firstCallResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected first call 201, got %d: %s", firstCallResp.StatusCode, firstCallResp.Body)
	}
	secondCallResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
	if secondCallResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected second call 201, got %d: %s", secondCallResp.StatusCode, secondCallResp.Body)
	}

	service := gamesvc.NewService(gamesvc.ServiceConfig{Store: db.NewStore(pool)})
	disconnected, err := service.SweepStalePlayerConnections(ctx, 30*time.Minute, 10)
	if err != nil {
		t.Fatalf("sweep stale players: %v", err)
	}
	if len(disconnected) != 1 || disconnected[0].ID != playerID || disconnected[0].ConnectionState != "disconnected" || disconnected[0].State != "disconnected" {
		t.Fatalf("expected stale player to be disconnected, got %+v", disconnected)
	}

	hostSnapshotResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/host-snapshot", nil)
	if hostSnapshotResp.StatusCode != http.StatusOK {
		t.Fatalf("expected host snapshot 200, got %d: %s", hostSnapshotResp.StatusCode, hostSnapshotResp.Body)
	}
	hostSnapshot := hostSnapshotFromData(t, hostSnapshotResp.Body)
	if len(hostSnapshot.Players) != 1 || hostSnapshot.Players[0].ConnectionState != "disconnected" {
		t.Fatalf("expected host to see disconnected player, got %+v", hostSnapshot.Players)
	}

	playerSnapshotResp := doRequestWithDevAuth(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/"+playerID+"/snapshot", nil, "timeout@example.local", "Timeout Player", "player")
	if playerSnapshotResp.StatusCode != http.StatusOK {
		t.Fatalf("expected reconnect snapshot 200, got %d: %s", playerSnapshotResp.StatusCode, playerSnapshotResp.Body)
	}
	playerSnapshot := playerSnapshotFromData(t, playerSnapshotResp.Body)
	if playerSnapshot.Player.ConnectionState != "online" || playerSnapshot.Player.State != "playing" {
		t.Fatalf("expected reconnect to restore online playing state, got %+v", playerSnapshot.Player)
	}
	if playerSnapshot.ReconnectNotice == nil || len(playerSnapshot.ReconnectNotice.MissedCalledWords) != 2 {
		t.Fatalf("expected reconnect notice with 2 missed called words, got %+v", playerSnapshot.ReconnectNotice)
	}

	events := gameEventsFromDB(t, ctx, pool, gameID)
	assertEventTypes(t, events, []string{"player.disconnected", "player.reconnected"})
}

func TestAPIGameplayFlow(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	wordSetID := insertAPIWordSet(t, ctx, pool)
	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL}, testLogger(), pool).Handler

	createResp := doRequest(t, handler, http.MethodPost, "/api/v1/games", map[string]any{
		"name":      "Gameplay Test Game",
		"code":      "GAMEPLAY-TEST",
		"wordSetId": wordSetID,
	})
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create game 201, got %d: %s", createResp.StatusCode, createResp.Body)
	}
	gameID := stringFromData(t, createResp.Body, "id")

	callBeforeStartResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
	if callBeforeStartResp.StatusCode != http.StatusConflict {
		t.Fatalf("expected call before start 409, got %d: %s", callBeforeStartResp.StatusCode, callBeforeStartResp.Body)
	}

	alexID, alexCard := allowJoinAndAssignCard(t, handler, gameID, "alex-gameplay@example.local", "Alex Gameplay")
	samID, _ := allowJoinAndAssignCard(t, handler, gameID, "sam-gameplay@example.local", "Sam Gameplay")

	startResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/start", nil)
	if startResp.StatusCode != http.StatusOK {
		t.Fatalf("expected start game 200, got %d: %s", startResp.StatusCode, startResp.Body)
	}
	if status := stringFromData(t, startResp.Body, "status"); status != "live" {
		t.Fatalf("expected live status, got %s", status)
	}

	secondStartResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/start", nil)
	if secondStartResp.StatusCode != http.StatusOK {
		t.Fatalf("expected second start game 200, got %d: %s", secondStartResp.StatusCode, secondStartResp.Body)
	}
	if status := stringFromData(t, secondStartResp.Body, "status"); status != "live" {
		t.Fatalf("expected second start to stay live, got %s", status)
	}

	firstCallResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
	if firstCallResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected first call 201, got %d: %s", firstCallResp.StatusCode, firstCallResp.Body)
	}
	firstCall := calledWordFromData(t, firstCallResp.Body)
	if firstCall.Sequence != 1 {
		t.Fatalf("expected first sequence 1, got %+v", firstCall)
	}

	secondCallResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
	if secondCallResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected second call 201, got %d: %s", secondCallResp.StatusCode, secondCallResp.Body)
	}
	secondCall := calledWordFromData(t, secondCallResp.Body)
	if secondCall.Sequence != 2 {
		t.Fatalf("expected second sequence 2, got %+v", secondCall)
	}

	callsResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/calls", nil)
	if callsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected calls list 200, got %d: %s", callsResp.StatusCode, callsResp.Body)
	}
	calls := calledWordsFromData(t, callsResp.Body)
	if len(calls) != 2 || calls[0].Sequence != 1 || calls[1].Sequence != 2 {
		t.Fatalf("expected ordered calls with sequences 1 and 2, got %+v", calls)
	}

	markCell := firstNonFreeCell(t, alexCard.Cells)
	markResp := doRequest(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/players/"+alexID+"/card/cells/"+markCell.ID, map[string]any{"marked": true})
	if markResp.StatusCode != http.StatusOK {
		t.Fatalf("expected mark cell 200, got %d: %s", markResp.StatusCode, markResp.Body)
	}

	getMarkedCardResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/"+alexID+"/card", nil)
	if getMarkedCardResp.StatusCode != http.StatusOK {
		t.Fatalf("expected marked card fetch 200, got %d: %s", getMarkedCardResp.StatusCode, getMarkedCardResp.Body)
	}
	markedCard := cardFromData(t, getMarkedCardResp.Body)
	if cellByID(t, markedCard.Cells, markCell.ID).MarkedAt == nil {
		t.Fatalf("expected markedAt to be set after mark")
	}

	unmarkResp := doRequest(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/players/"+alexID+"/card/cells/"+markCell.ID, map[string]any{"marked": false})
	if unmarkResp.StatusCode != http.StatusOK {
		t.Fatalf("expected unmark cell 200, got %d: %s", unmarkResp.StatusCode, unmarkResp.Body)
	}

	getUnmarkedCardResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/"+alexID+"/card", nil)
	if getUnmarkedCardResp.StatusCode != http.StatusOK {
		t.Fatalf("expected unmarked card fetch 200, got %d: %s", getUnmarkedCardResp.StatusCode, getUnmarkedCardResp.Body)
	}
	unmarkedCard := cardFromData(t, getUnmarkedCardResp.Body)
	if cellByID(t, unmarkedCard.Cells, markCell.ID).MarkedAt != nil {
		t.Fatalf("expected markedAt to clear after unmark")
	}

	invalidClaimResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/claims", map[string]any{
		"playerId": samID,
		"pattern":  "single_line",
	})
	if invalidClaimResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected invalid claim submission 201, got %d: %s", invalidClaimResp.StatusCode, invalidClaimResp.Body)
	}
	invalidClaim := claimSubmissionFromData(t, invalidClaimResp.Body)
	if invalidClaim.Claim.Status != "invalid" || invalidClaim.Winner != nil {
		t.Fatalf("expected invalid claim without winner, got %+v", invalidClaim)
	}

	calledSet := map[string]struct{}{
		strings.ToLower(firstCall.Word):  {},
		strings.ToLower(secondCall.Word): {},
	}
	for !hasCompleteLine(alexCard.Cells, calledSet) {
		nextCallResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
		if nextCallResp.StatusCode != http.StatusCreated {
			t.Fatalf("expected next call while looking for line 201, got %d: %s", nextCallResp.StatusCode, nextCallResp.Body)
		}
		nextCall := calledWordFromData(t, nextCallResp.Body)
		calledSet[strings.ToLower(nextCall.Word)] = struct{}{}
		if len(calledSet) > 26 {
			t.Fatalf("expected to complete a line before exhausting word set")
		}
	}

	validClaimResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/claims", map[string]any{
		"playerId": alexID,
		"pattern":  "single_line",
	})
	if validClaimResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected valid claim submission 201, got %d: %s", validClaimResp.StatusCode, validClaimResp.Body)
	}
	validClaim := claimSubmissionFromData(t, validClaimResp.Body)
	if validClaim.Claim.Status != "confirmed" || validClaim.Winner == nil || validClaim.Winner.Placement != 1 {
		t.Fatalf("expected confirmed claim with first-place winner, got %+v", validClaim)
	}

	playerAckResp := doRequestWithDevAuth(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/claims/"+validClaim.Claim.ID+"/acknowledge", map[string]any{
		"decision": "approve",
	}, "alex-gameplay@example.local", "Alex Gameplay", "player")
	if playerAckResp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected player claim acknowledge 403, got %d: %s", playerAckResp.StatusCode, playerAckResp.Body)
	}

	mismatchedAckResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/claims/"+invalidClaim.Claim.ID+"/acknowledge", map[string]any{
		"decision": "approve",
	})
	if mismatchedAckResp.StatusCode != http.StatusConflict {
		t.Fatalf("expected mismatched claim acknowledge 409, got %d: %s", mismatchedAckResp.StatusCode, mismatchedAckResp.Body)
	}

	rejectAckResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/claims/"+invalidClaim.Claim.ID+"/acknowledge", map[string]any{
		"decision": "reject",
		"note":     "Backend validation rejected the claim.",
	})
	if rejectAckResp.StatusCode != http.StatusCreated || stringFromData(t, rejectAckResp.Body, "type") != "claim.acknowledged" {
		t.Fatalf("expected invalid claim acknowledge event, got %d: %s", rejectAckResp.StatusCode, rejectAckResp.Body)
	}

	approveAckResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/claims/"+validClaim.Claim.ID+"/acknowledge", map[string]any{
		"decision": "approve",
		"note":     "Winner order reviewed by host.",
	})
	if approveAckResp.StatusCode != http.StatusCreated || stringFromData(t, approveAckResp.Body, "type") != "claim.acknowledged" {
		t.Fatalf("expected valid claim acknowledge event, got %d: %s", approveAckResp.StatusCode, approveAckResp.Body)
	}

	claimsResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/claims", nil)
	if claimsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected claims list 200, got %d: %s", claimsResp.StatusCode, claimsResp.Body)
	}
	if claims := claimsFromData(t, claimsResp.Body); len(claims) != 2 || claims[0].Status == "acknowledged" || claims[1].Status == "acknowledged" {
		t.Fatalf("expected 2 claims, got %+v", claims)
	}

	summaryResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/summary", nil)
	if summaryResp.StatusCode != http.StatusOK {
		t.Fatalf("expected summary 200, got %d: %s", summaryResp.StatusCode, summaryResp.Body)
	}
	summary := summaryFromData(t, summaryResp.Body)
	if summary.Status != "live" || summary.PlayerCount != 2 || summary.CalledWordCount != len(calledSet) || len(summary.Claims) != 2 || len(summary.Winners) != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	activityResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/activity", nil)
	if activityResp.StatusCode != http.StatusOK {
		t.Fatalf("expected activity 200, got %d: %s", activityResp.StatusCode, activityResp.Body)
	}
	if !hasActivityEventType(activityEventsFromData(t, activityResp.Body), "claim.acknowledged") {
		t.Fatalf("expected claim.acknowledged activity, got %s", activityResp.Body)
	}
	if state := playerState(summary.Players, alexID); state != "confirmed_winner" {
		t.Fatalf("expected Alex to be confirmed_winner, got %s", state)
	}
	if state := playerState(summary.Players, samID); state != "rejected_claim" {
		t.Fatalf("expected Sam to be rejected_claim, got %s", state)
	}
	if len(summary.CalledWords) != summary.CalledWordCount || summary.CurrentWord == nil {
		t.Fatalf("expected summary called-word details, got %+v", summary)
	}
}

func TestAPILifecycleFlow(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	wordSetID := insertAPIWordSet(t, ctx, pool)
	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL}, testLogger(), pool).Handler
	gameID := createAPIGame(t, handler, "Lifecycle Test Game", "LIFECYCLE-TEST", wordSetID, nil)
	playerID, _ := allowJoinAndAssignCard(t, handler, gameID, "life@example.local", "Life Cycle")

	startResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/start", nil)
	if startResp.StatusCode != http.StatusOK {
		t.Fatalf("expected start 200, got %d: %s", startResp.StatusCode, startResp.Body)
	}
	startSummary := fetchSummary(t, handler, gameID)
	if state := playerState(startSummary.Players, playerID); state != "playing" {
		t.Fatalf("expected joined player to move to playing, got %s", state)
	}

	pauseResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/pause", nil)
	if pauseResp.StatusCode != http.StatusOK || stringFromData(t, pauseResp.Body, "status") != "paused" {
		t.Fatalf("expected pause 200 paused, got %d: %s", pauseResp.StatusCode, pauseResp.Body)
	}
	callPausedResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
	if callPausedResp.StatusCode != http.StatusConflict {
		t.Fatalf("expected call while paused 409, got %d: %s", callPausedResp.StatusCode, callPausedResp.Body)
	}
	pausedClaimResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/claims", map[string]any{
		"playerId": playerID,
		"pattern":  "single_line",
	})
	if pausedClaimResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected paused claim to be accepted for validation, got %d: %s", pausedClaimResp.StatusCode, pausedClaimResp.Body)
	}

	resumeResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/resume", nil)
	if resumeResp.StatusCode != http.StatusOK || stringFromData(t, resumeResp.Body, "status") != "live" {
		t.Fatalf("expected resume 200 live, got %d: %s", resumeResp.StatusCode, resumeResp.Body)
	}
	finishResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/finish", nil)
	if finishResp.StatusCode != http.StatusOK || stringFromData(t, finishResp.Body, "status") != "finished" {
		t.Fatalf("expected finish 200 finished, got %d: %s", finishResp.StatusCode, finishResp.Body)
	}
	startFinishedResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/start", nil)
	if startFinishedResp.StatusCode != http.StatusConflict {
		t.Fatalf("expected start after finished 409, got %d: %s", startFinishedResp.StatusCode, startFinishedResp.Body)
	}

	cancelGameID := createAPIGame(t, handler, "Cancel Test Game", "CANCEL-TEST", wordSetID, nil)
	cancelResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+cancelGameID+"/cancel", nil)
	if cancelResp.StatusCode != http.StatusOK || stringFromData(t, cancelResp.Body, "status") != "cancelled" {
		t.Fatalf("expected cancel 200 cancelled, got %d: %s", cancelResp.StatusCode, cancelResp.Body)
	}
	startCancelledResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+cancelGameID+"/start", nil)
	if startCancelledResp.StatusCode != http.StatusConflict {
		t.Fatalf("expected start after cancelled 409, got %d: %s", startCancelledResp.StatusCode, startCancelledResp.Body)
	}
}

func TestAPIWinningPatternAndWordExhaustion(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	wordSetID := insertAPIWordSet(t, ctx, pool)
	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL}, testLogger(), pool).Handler

	defaultGameID := createAPIGame(t, handler, "Default Pattern Game", "DEFAULT-PATTERN", wordSetID, nil)
	unsupportedResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+defaultGameID+"/claims", map[string]any{
		"playerId": "00000000-0000-0000-0000-000000000000",
		"pattern":  "postage_stamp",
	})
	if unsupportedResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected unsupported pattern 400, got %d: %s", unsupportedResp.StatusCode, unsupportedResp.Body)
	}
	defaultPlayerID, _ := allowJoinAndAssignCard(t, handler, defaultGameID, "default-pattern@example.local", "Default Pattern")
	disallowedResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+defaultGameID+"/claims", map[string]any{
		"playerId": defaultPlayerID,
		"pattern":  "four_corners",
	})
	if disallowedResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected disallowed default pattern 400, got %d: %s", disallowedResp.StatusCode, disallowedResp.Body)
	}

	pattern := "four_corners"
	fourCornersGameID := createAPIGame(t, handler, "Four Corners Game", "FOUR-CORNERS", wordSetID, &pattern)
	playerID, _ := allowJoinAndAssignCard(t, handler, fourCornersGameID, "corners@example.local", "Corners Player")
	if resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+fourCornersGameID+"/start", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("expected start 200, got %d: %s", resp.StatusCode, resp.Body)
	}
	called := callAllWords(t, handler, fourCornersGameID)
	if len(called) != 26 {
		t.Fatalf("expected 26 called words, got %d", len(called))
	}
	extraCallResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+fourCornersGameID+"/calls", nil)
	if extraCallResp.StatusCode != http.StatusConflict {
		t.Fatalf("expected exhausted word call 409, got %d: %s", extraCallResp.StatusCode, extraCallResp.Body)
	}
	validResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+fourCornersGameID+"/claims", map[string]any{
		"playerId": playerID,
		"pattern":  "four_corners",
	})
	if validResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected four-corners claim 201, got %d: %s", validResp.StatusCode, validResp.Body)
	}
	if claim := claimSubmissionFromData(t, validResp.Body); claim.Claim.Status != "confirmed" || claim.Winner == nil {
		t.Fatalf("expected confirmed four-corners winner, got %+v", claim)
	}
	summary := fetchSummary(t, handler, fourCornersGameID)
	if summary.Status != "live" || summary.CalledWordCount != 26 || summary.CurrentWord == nil {
		t.Fatalf("expected exhausted game to stay live with current word, got %+v", summary)
	}
}

func TestAPIWinnerHardening(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	wordSetID := insertAPIWordSet(t, ctx, pool)
	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL}, testLogger(), pool).Handler
	gameID := createAPIGame(t, handler, "Winner Hardening Game", "WINNER-HARDENING", wordSetID, nil)
	alexID, _ := allowJoinAndAssignCard(t, handler, gameID, "winner-a@example.local", "Winner A")
	samID, _ := allowJoinAndAssignCard(t, handler, gameID, "winner-b@example.local", "Winner B")
	taylorID, _ := allowJoinAndAssignCard(t, handler, gameID, "winner-c@example.local", "Winner C")
	if resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/start", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("expected start 200, got %d: %s", resp.StatusCode, resp.Body)
	}
	callAllWords(t, handler, gameID)

	firstResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/claims", map[string]any{
		"playerId": alexID,
		"pattern":  "single_line",
	})
	if firstResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected first claim 201, got %d: %s", firstResp.StatusCode, firstResp.Body)
	}
	repeatResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/claims", map[string]any{
		"playerId": alexID,
		"pattern":  "single_line",
	})
	if repeatResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected repeated valid claim 201, got %d: %s", repeatResp.StatusCode, repeatResp.Body)
	}
	if summary := fetchSummary(t, handler, gameID); len(summary.Winners) != 1 {
		t.Fatalf("expected repeated same player/pattern claim not to duplicate winner, got %+v", summary.Winners)
	}

	var wg sync.WaitGroup
	responses := make([]testResponse, 2)
	for index, playerID := range []string{samID, taylorID} {
		wg.Add(1)
		go func(index int, playerID string) {
			defer wg.Done()
			responses[index] = doRequestForConcurrent(handler, http.MethodPost, "/api/v1/games/"+gameID+"/claims", map[string]any{
				"playerId": playerID,
				"pattern":  "single_line",
			})
		}(index, playerID)
	}
	wg.Wait()
	for _, resp := range responses {
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected concurrent claim 201, got %d: %s", resp.StatusCode, resp.Body)
		}
	}

	summary := fetchSummary(t, handler, gameID)
	if summary.Status != "finished" || len(summary.Winners) != 3 {
		t.Fatalf("expected third winner to finish game with 3 winners, got %+v", summary)
	}
	placements := map[int]bool{}
	for _, winner := range summary.Winners {
		if placements[winner.Placement] {
			t.Fatalf("duplicate winner placement in %+v", summary.Winners)
		}
		placements[winner.Placement] = true
	}
}

func TestAPIRealtimeSnapshotsOutboxAndSSE(t *testing.T) {
	ctx := context.Background()
	databaseURL := createTestDatabase(t)
	if err := db.RunMigrationsFromPath(databaseURL, "../db/migrations", db.MigrationUp); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	pool, err := db.Open(ctx, databaseURL)
	if err != nil {
		t.Fatalf("open pool: %v", err)
	}
	defer pool.Close()

	wordSetID := insertAPIWordSet(t, ctx, pool)
	handler := NewServer(config.Config{AppEnv: "test", DatabaseURL: databaseURL}, testLogger(), pool).Handler
	gameID := createAPIGame(t, handler, "Realtime Backbone Game", "REALTIME-BACKBONE", wordSetID, nil)
	playerID, card := allowJoinAndAssignCard(t, handler, gameID, "realtime@example.local", "Realtime Player")

	if resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/start", nil); resp.StatusCode != http.StatusOK {
		t.Fatalf("expected start 200, got %d: %s", resp.StatusCode, resp.Body)
	}
	firstCallResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
	if firstCallResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected first call 201, got %d: %s", firstCallResp.StatusCode, firstCallResp.Body)
	}
	markCell := firstNonFreeCell(t, card.Cells)
	if resp := doRequest(t, handler, http.MethodPatch, "/api/v1/games/"+gameID+"/players/"+playerID+"/card/cells/"+markCell.ID, map[string]any{"marked": true}); resp.StatusCode != http.StatusOK {
		t.Fatalf("expected mark 200, got %d: %s", resp.StatusCode, resp.Body)
	}
	for index := 0; index < 25; index++ {
		resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected remaining call %d to return 201, got %d: %s", index+1, resp.StatusCode, resp.Body)
		}
	}
	validClaimResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/claims", map[string]any{
		"playerId": playerID,
		"pattern":  "single_line",
	})
	if validClaimResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected valid claim 201, got %d: %s", validClaimResp.StatusCode, validClaimResp.Body)
	}

	hostSnapshotResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/host-snapshot", nil)
	if hostSnapshotResp.StatusCode != http.StatusOK {
		t.Fatalf("expected host snapshot 200, got %d: %s", hostSnapshotResp.StatusCode, hostSnapshotResp.Body)
	}
	hostSnapshot := hostSnapshotFromData(t, hostSnapshotResp.Body)
	if hostSnapshot.Status != "live" || hostSnapshot.PlayerCount != 1 || len(hostSnapshot.Players) != 1 || len(hostSnapshot.Claims) != 1 || len(hostSnapshot.Winners) != 1 || hostSnapshot.CurrentWord == nil {
		t.Fatalf("unexpected host snapshot: %+v", hostSnapshot)
	}

	playerSnapshotResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/"+playerID+"/snapshot", nil)
	if playerSnapshotResp.StatusCode != http.StatusOK {
		t.Fatalf("expected player snapshot 200, got %d: %s", playerSnapshotResp.StatusCode, playerSnapshotResp.Body)
	}
	playerSnapshot := playerSnapshotFromData(t, playerSnapshotResp.Body)
	if playerSnapshot.Status != "live" || playerSnapshot.Player.ID != playerID || playerSnapshot.Card == nil || len(playerSnapshot.Card.Cells) != 25 || len(playerSnapshot.Claims) != 1 || len(playerSnapshot.Winners) != 1 {
		t.Fatalf("unexpected player snapshot: %+v", playerSnapshot)
	}

	events := gameEventsFromDB(t, ctx, pool, gameID)
	assertEventTypes(t, events, []string{"game.created", "player.joined", "card.assigned", "game.started", "word.called", "card.cell_marked", "claim.submitted", "claim.validated", "winner.created"})
	for index := 1; index < len(events); index++ {
		if events[index].Sequence <= events[index-1].Sequence {
			t.Fatalf("expected ordered outbox sequences, got %+v", events)
		}
	}

	server := httptest.NewServer(handler)
	defer server.Close()
	streamed := readSSEEvents(t, server.URL+"/api/v1/games/"+gameID+"/events", "", 4)
	if len(streamed) != 4 || streamed[0].Sequence != 1 || streamed[1].Sequence != 2 {
		t.Fatalf("expected initial SSE stream from sequence 1, got %+v", streamed)
	}
	resumed := readSSEEvents(t, server.URL+"/api/v1/games/"+gameID+"/events", "2", 2)
	if len(resumed) != 2 || resumed[0].Sequence != 3 {
		t.Fatalf("expected resumed SSE stream after sequence 2, got %+v", resumed)
	}
	readSSEHeartbeat(t, server.URL+"/api/v1/games/"+gameID+"/events?heartbeatMs=25", events[len(events)-1].Sequence)
}

type testResponse struct {
	StatusCode int
	Body       string
}

func doRequest(t *testing.T, handler http.Handler, method, path string, body any) testResponse {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Dev-User-Email", "host@example.local")
	req.Header.Set("X-Dev-User-Name", "Host User")
	req.Header.Set("X-Dev-User-Role", "host")
	req.Header.Set("X-Request-ID", "test-request-id")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	resp := recorder.Result()
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	if got := resp.Header.Get("X-Request-ID"); got != "test-request-id" {
		t.Fatalf("expected request id response header, got %q", got)
	}

	return testResponse{StatusCode: resp.StatusCode, Body: string(responseBody)}
}

func doRequestWithoutDevAuth(t *testing.T, handler http.Handler, method, path string, body any) testResponse {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Request-ID", "test-request-id")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	resp := recorder.Result()
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	return testResponse{StatusCode: resp.StatusCode, Body: string(responseBody)}
}

func doRequestWithDevAuth(t *testing.T, handler http.Handler, method, path string, body any, email, name, role string) testResponse {
	t.Helper()

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal request: %v", err)
		}
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Dev-User-Email", email)
	req.Header.Set("X-Dev-User-Name", name)
	req.Header.Set("X-Dev-User-Role", role)
	req.Header.Set("X-Request-ID", "test-request-id")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	resp := recorder.Result()
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}

	return testResponse{StatusCode: resp.StatusCode, Body: string(responseBody)}
}

func doRequestForConcurrent(handler http.Handler, method, path string, body any) testResponse {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return testResponse{StatusCode: http.StatusInternalServerError, Body: err.Error()}
		}
		reader = bytes.NewReader(payload)
	}

	req := httptest.NewRequest(method, path, reader)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-Dev-User-Email", "host@example.local")
	req.Header.Set("X-Dev-User-Name", "Host User")
	req.Header.Set("X-Dev-User-Role", "host")
	req.Header.Set("X-Request-ID", "test-request-id")

	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	resp := recorder.Result()
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return testResponse{StatusCode: http.StatusInternalServerError, Body: err.Error()}
	}

	return testResponse{StatusCode: resp.StatusCode, Body: string(responseBody)}
}

func stringFromData(t *testing.T, body, key string) string {
	t.Helper()

	var envelope struct {
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	value, ok := envelope.Data[key].(string)
	if !ok || value == "" {
		t.Fatalf("expected data.%s string in %s", key, body)
	}

	return value
}

func currentUserFromData(t *testing.T, body string) apiCurrentUser {
	t.Helper()

	var envelope struct {
		Data apiCurrentUser `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode current user response: %v", err)
	}

	return envelope.Data
}

func gameRunSummariesFromData(t *testing.T, body string) []apiGameRunSummary {
	t.Helper()

	var envelope struct {
		Data []apiGameRunSummary `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode game summaries response: %v", err)
	}

	return envelope.Data
}

func allowedPlayersFromData(t *testing.T, body string) []apiAllowedPlayer {
	t.Helper()

	var envelope struct {
		Data []apiAllowedPlayer `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode allowed players response: %v", err)
	}

	return envelope.Data
}

func wordSetFromData(t *testing.T, body string) apiWordSet {
	t.Helper()

	var envelope struct {
		Data apiWordSet `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode word set response: %v", err)
	}

	return envelope.Data
}

func wordSetsFromData(t *testing.T, body string) []apiWordSet {
	t.Helper()

	var envelope struct {
		Data []apiWordSet `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode word sets response: %v", err)
	}

	return envelope.Data
}

func wordSetWordFromData(t *testing.T, body string) apiWordSetWord {
	t.Helper()

	var envelope struct {
		Data apiWordSetWord `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode word set word response: %v", err)
	}

	return envelope.Data
}

func activityEventsFromData(t *testing.T, body string) []apiActivityEvent {
	t.Helper()

	var envelope struct {
		Data []apiActivityEvent `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode activity events response: %v", err)
	}

	return envelope.Data
}

type apiCardCell struct {
	ID          string     `json:"id"`
	RowIndex    int        `json:"rowIndex"`
	ColIndex    int        `json:"colIndex"`
	Word        string     `json:"word"`
	IsFreeSpace bool       `json:"isFreeSpace"`
	MarkedAt    *time.Time `json:"markedAt"`
}

type apiCurrentUser struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	AuthMode string `json:"authMode"`
}

type apiGameRunSummary struct {
	ID                 string `json:"id"`
	Code               string `json:"code"`
	Name               string `json:"name"`
	Status             string `json:"status"`
	AllowedPlayerCount int    `json:"allowedPlayerCount"`
	PlayerCount        int    `json:"playerCount"`
}

type apiAllowedPlayer struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type apiWordSet struct {
	ID     string           `json:"id"`
	Name   string           `json:"name"`
	Status string           `json:"status"`
	Source string           `json:"source"`
	Words  []apiWordSetWord `json:"words"`
}

type apiWordSetWord struct {
	ID        string `json:"id"`
	WordSetID string `json:"wordSetId"`
	Word      string `json:"word"`
	SortOrder int    `json:"sortOrder"`
	IsActive  bool   `json:"isActive"`
}

type apiCard struct {
	ID        string        `json:"id"`
	GameRunID string        `json:"gameRunId"`
	PlayerID  string        `json:"playerId"`
	Cells     []apiCardCell `json:"cells"`
}

type apiCalledWord struct {
	ID          string          `json:"id"`
	Word        string          `json:"word"`
	Sequence    int             `json:"sequence"`
	CallerAsset *apiCallerAsset `json:"callerAsset"`
}

type apiCallerAsset struct {
	ID             string  `json:"id"`
	GameRunID      string  `json:"gameRunId"`
	CallDeckItemID string  `json:"callDeckItemId"`
	Word           string  `json:"word"`
	Sequence       int     `json:"sequence"`
	Line           string  `json:"line"`
	AudioURL       *string `json:"audioUrl"`
	StorageKey     *string `json:"storageKey"`
	VoiceName      *string `json:"voiceName"`
	Provider       string  `json:"provider"`
	Status         string  `json:"status"`
	ErrorReason    *string `json:"errorReason"`
}

type apiDeliveryAttempt struct {
	ID             string `json:"id"`
	GameRunID      string `json:"gameRunId"`
	Channel        string `json:"channel"`
	Purpose        string `json:"purpose"`
	RecipientEmail string `json:"recipientEmail"`
	Subject        string `json:"subject"`
	TemplateKey    string `json:"templateKey"`
	BodyPreview    string `json:"bodyPreview"`
	LinkURL        string `json:"linkUrl"`
	GameCode       string `json:"gameCode"`
	Status         string `json:"status"`
}

type apiTheme struct {
	ID     string         `json:"id"`
	Status string         `json:"status"`
	Tokens apiThemeTokens `json:"tokens"`
}

type apiThemeTokens struct {
	Palette     map[string]string `json:"palette"`
	Icons       []string          `json:"icons"`
	Decorations []string          `json:"decorations"`
	Motion      string            `json:"motion"`
	CallerTone  string            `json:"callerTone"`
}

type apiCallDeckRow struct {
	ID       string
	Word     string
	Sequence int
}

type apiGameSettings struct {
	GameRunID                    string  `json:"gameRunId"`
	MarkingMode                  string  `json:"markingMode"`
	AllowPlayerMarkingModeChoice bool    `json:"allowPlayerMarkingModeChoice"`
	ShowClaimReadiness           bool    `json:"showClaimReadiness"`
	VoiceClaimMode               string  `json:"voiceClaimMode"`
	VoiceClaimAutoplay           bool    `json:"voiceClaimAutoplay"`
	CallerMode                   string  `json:"callerMode"`
	ThemeMode                    string  `json:"themeMode"`
	ThemeID                      *string `json:"themeId"`
}

type apiPlayerPreferences struct {
	PlayerID    string  `json:"playerId"`
	MarkingMode *string `json:"markingMode"`
}

type apiAutoMarkRun struct {
	PlayersScanned     int    `json:"playersScanned"`
	PlayersMarked      int    `json:"playersMarked"`
	CalledWordsScanned int    `json:"calledWordsScanned"`
	CellsMarked        int    `json:"cellsMarked"`
	Mode               string `json:"mode"`
	SkippedReason      string `json:"skippedReason"`
}

type apiGameContent struct {
	ID                   string     `json:"id"`
	GameRunID            string     `json:"gameRunId"`
	Status               string     `json:"status"`
	Topic                string     `json:"topic"`
	Summary              string     `json:"summary"`
	Words                []string   `json:"words"`
	ReviewWindowClosesAt *time.Time `json:"reviewWindowClosesAt"`
	LockedAt             *time.Time `json:"lockedAt"`
	LockedWordSetID      *string    `json:"lockedWordSetId"`
	GenerationProvider   string     `json:"generationProvider"`
}

type apiClaimReadiness struct {
	Ready             bool          `json:"ready"`
	SupportedPatterns []string      `json:"supportedPatterns"`
	ReadyPatterns     []string      `json:"readyPatterns"`
	BestPattern       string        `json:"bestPattern"`
	MatchedCells      []apiCardCell `json:"matchedCells"`
	MissingCells      []apiCardCell `json:"missingCells"`
	Reason            string        `json:"reason"`
}

type apiClaim struct {
	ID       string `json:"id"`
	PlayerID string `json:"playerId"`
	Pattern  string `json:"pattern"`
	Status   string `json:"status"`
}

type apiWinner struct {
	ID        string  `json:"id"`
	PlayerID  string  `json:"playerId"`
	ClaimID   *string `json:"claimId"`
	Placement int     `json:"placement"`
}

type apiClaimSubmission struct {
	Claim  apiClaim   `json:"claim"`
	Winner *apiWinner `json:"winner"`
}

type apiSummary struct {
	Status          string          `json:"status"`
	PlayerCount     int             `json:"playerCount"`
	CalledWordCount int             `json:"calledWordCount"`
	CurrentWord     *apiCalledWord  `json:"currentWord"`
	Claims          []apiClaim      `json:"claims"`
	Winners         []apiWinner     `json:"winners"`
	Players         []apiPlayer     `json:"players"`
	CalledWords     []apiCalledWord `json:"calledWords"`
}

type apiPlayer struct {
	ID              string    `json:"id"`
	Icon            *string   `json:"icon"`
	AvatarColor     *string   `json:"avatarColor"`
	AvatarLabel     *string   `json:"avatarLabel"`
	ConnectionState string    `json:"connectionState"`
	State           string    `json:"state"`
	LastSeenAt      time.Time `json:"lastSeenAt"`
}

type apiHostSnapshot struct {
	Status             string          `json:"status"`
	PlayerCount        int             `json:"playerCount"`
	CurrentWord        *apiCalledWord  `json:"currentWord"`
	CurrentCallerAsset *apiCallerAsset `json:"currentCallerAsset"`
	AppliedTheme       *apiTheme       `json:"appliedTheme"`
	Players            []apiPlayer     `json:"players"`
	CalledWords        []apiCalledWord `json:"calledWords"`
	Claims             []apiClaim      `json:"claims"`
	Winners            []apiWinner     `json:"winners"`
}

type apiPlayerSnapshot struct {
	Status                       string              `json:"status"`
	MarkingMode                  string              `json:"markingMode"`
	AllowPlayerMarkingModeChoice bool                `json:"allowPlayerMarkingModeChoice"`
	ShowClaimReadiness           bool                `json:"showClaimReadiness"`
	CurrentWord                  *apiCalledWord      `json:"currentWord"`
	CurrentCallerAsset           *apiCallerAsset     `json:"currentCallerAsset"`
	AppliedTheme                 *apiTheme           `json:"appliedTheme"`
	Player                       apiPlayer           `json:"player"`
	Card                         *apiCard            `json:"card"`
	CalledWords                  []apiCalledWord     `json:"calledWords"`
	Claims                       []apiClaim          `json:"claims"`
	Winners                      []apiWinner         `json:"winners"`
	ReconnectNotice              *apiReconnectNotice `json:"reconnectNotice"`
}

type apiReconnectNotice struct {
	LastSeenAt        time.Time       `json:"lastSeenAt"`
	MissedCalledWords []apiCalledWord `json:"missedCalledWords"`
}

type apiGameEvent struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Sequence int64  `json:"sequence"`
}

type apiActivityEvent struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Sequence *int64 `json:"sequence"`
}

func cellsFromData(t *testing.T, body string) []apiCardCell {
	t.Helper()

	var envelope struct {
		Data struct {
			Cells []apiCardCell `json:"cells"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode card response: %v", err)
	}

	return envelope.Data.Cells
}

func cardFromData(t *testing.T, body string) apiCard {
	t.Helper()

	var envelope struct {
		Data apiCard `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode card response: %v", err)
	}

	return envelope.Data
}

func playerFromData(t *testing.T, body string) apiPlayer {
	t.Helper()

	var envelope struct {
		Data apiPlayer `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode player response: %v", err)
	}

	return envelope.Data
}

func calledWordFromData(t *testing.T, body string) apiCalledWord {
	t.Helper()

	var envelope struct {
		Data apiCalledWord `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode called word response: %v", err)
	}

	return envelope.Data
}

func calledWordsFromData(t *testing.T, body string) []apiCalledWord {
	t.Helper()

	var envelope struct {
		Data []apiCalledWord `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode called words response: %v", err)
	}

	return envelope.Data
}

func gameSettingsFromData(t *testing.T, body string) apiGameSettings {
	t.Helper()

	var envelope struct {
		Data apiGameSettings `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode game settings response: %v", err)
	}

	return envelope.Data
}

func playerPreferencesFromData(t *testing.T, body string) apiPlayerPreferences {
	t.Helper()

	var envelope struct {
		Data apiPlayerPreferences `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode player preferences response: %v", err)
	}

	return envelope.Data
}

func autoMarkRunFromData(t *testing.T, body string) apiAutoMarkRun {
	t.Helper()

	var envelope struct {
		Data apiAutoMarkRun `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode auto-mark response: %v", err)
	}

	return envelope.Data
}

func gameContentFromData(t *testing.T, body string) apiGameContent {
	t.Helper()

	var envelope struct {
		Data apiGameContent `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode game content response: %v", err)
	}

	return envelope.Data
}

func callerAssetsFromData(t *testing.T, body string) []apiCallerAsset {
	t.Helper()

	var envelope struct {
		Data []apiCallerAsset `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode caller assets response: %v", err)
	}

	return envelope.Data
}

func deliveryAttemptsFromData(t *testing.T, body string) []apiDeliveryAttempt {
	t.Helper()

	var envelope struct {
		Data []apiDeliveryAttempt `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode delivery attempts response: %v", err)
	}

	return envelope.Data
}

func themeFromData(t *testing.T, body string) apiTheme {
	t.Helper()

	var envelope struct {
		Data apiTheme `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode theme response: %v", err)
	}

	return envelope.Data
}

func claimReadinessFromData(t *testing.T, body string) apiClaimReadiness {
	t.Helper()

	var envelope struct {
		Data apiClaimReadiness `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode claim-readiness response: %v", err)
	}

	return envelope.Data
}

func claimSubmissionFromData(t *testing.T, body string) apiClaimSubmission {
	t.Helper()

	var envelope struct {
		Data apiClaimSubmission `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode claim submission response: %v", err)
	}

	return envelope.Data
}

func claimsFromData(t *testing.T, body string) []apiClaim {
	t.Helper()

	var envelope struct {
		Data []apiClaim `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode claims response: %v", err)
	}

	return envelope.Data
}

func summaryFromData(t *testing.T, body string) apiSummary {
	t.Helper()

	var envelope struct {
		Data apiSummary `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode summary response: %v", err)
	}

	return envelope.Data
}

func hostSnapshotFromData(t *testing.T, body string) apiHostSnapshot {
	t.Helper()

	var envelope struct {
		Data apiHostSnapshot `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode host snapshot response: %v", err)
	}

	return envelope.Data
}

func playerSnapshotFromData(t *testing.T, body string) apiPlayerSnapshot {
	t.Helper()

	var envelope struct {
		Data apiPlayerSnapshot `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &envelope); err != nil {
		t.Fatalf("decode player snapshot response: %v", err)
	}

	return envelope.Data
}

func allowJoinAndAssignCard(t *testing.T, handler http.Handler, gameID, email, displayName string) (string, apiCard) {
	t.Helper()

	allowedResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/allowed-players", map[string]any{
		"email":       email,
		"displayName": displayName,
	})
	if allowedResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected add allowed player 201, got %d: %s", allowedResp.StatusCode, allowedResp.Body)
	}

	joinResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players", map[string]any{
		"email":       email,
		"displayName": displayName,
	})
	if joinResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected join player 201, got %d: %s", joinResp.StatusCode, joinResp.Body)
	}
	playerID := stringFromData(t, joinResp.Body, "id")

	cardResp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/players/"+playerID+"/card", nil)
	if cardResp.StatusCode != http.StatusCreated {
		t.Fatalf("expected assign card 201, got %d: %s", cardResp.StatusCode, cardResp.Body)
	}

	return playerID, cardFromData(t, cardResp.Body)
}

func createAPIGame(t *testing.T, handler http.Handler, name, code, wordSetID string, winningPattern *string) string {
	t.Helper()

	body := map[string]any{
		"name": name,
		"code": code,
	}
	if wordSetID != "" {
		body["wordSetId"] = wordSetID
	}
	if winningPattern != nil {
		body["winningPattern"] = *winningPattern
	}

	resp := doRequest(t, handler, http.MethodPost, "/api/v1/games", body)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create game 201, got %d: %s", resp.StatusCode, resp.Body)
	}

	return stringFromData(t, resp.Body, "id")
}

func createAPIGameWithDevAuth(t *testing.T, handler http.Handler, name, code, wordSetID string, winningPattern *string, email, displayName, role string) string {
	t.Helper()

	body := map[string]any{
		"name": name,
		"code": code,
	}
	if wordSetID != "" {
		body["wordSetId"] = wordSetID
	}
	if winningPattern != nil {
		body["winningPattern"] = *winningPattern
	}

	resp := doRequestWithDevAuth(t, handler, http.MethodPost, "/api/v1/games", body, email, displayName, role)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected create game 201, got %d: %s", resp.StatusCode, resp.Body)
	}

	return stringFromData(t, resp.Body, "id")
}

func callAllWords(t *testing.T, handler http.Handler, gameID string) []apiCalledWord {
	t.Helper()

	words := make([]apiCalledWord, 0, 26)
	for index := 0; index < 26; index++ {
		resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected call %d to return 201, got %d: %s", index+1, resp.StatusCode, resp.Body)
		}
		words = append(words, calledWordFromData(t, resp.Body))
	}

	return words
}

func callUntilWord(t *testing.T, handler http.Handler, gameID, targetWord string) apiCalledWord {
	t.Helper()

	for index := 0; index < 26; index++ {
		resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected call while looking for %s to return 201, got %d: %s", targetWord, resp.StatusCode, resp.Body)
		}
		called := calledWordFromData(t, resp.Body)
		if strings.EqualFold(strings.TrimSpace(called.Word), strings.TrimSpace(targetWord)) {
			return called
		}
	}

	t.Fatalf("expected to call target word %s before word exhaustion", targetWord)
	return apiCalledWord{}
}

func callUntilExhausted(t *testing.T, handler http.Handler, gameID string) []apiCalledWord {
	t.Helper()

	words := make([]apiCalledWord, 0, 26)
	for index := 0; index < 27; index++ {
		resp := doRequest(t, handler, http.MethodPost, "/api/v1/games/"+gameID+"/calls", nil)
		if resp.StatusCode == http.StatusConflict {
			return words
		}
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("expected call until exhaustion to return 201/409, got %d: %s", resp.StatusCode, resp.Body)
		}
		words = append(words, calledWordFromData(t, resp.Body))
	}

	t.Fatal("expected word exhaustion conflict")
	return words
}

func fetchCard(t *testing.T, handler http.Handler, gameID, playerID string) apiCard {
	t.Helper()

	resp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/players/"+playerID+"/card", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected card fetch 200, got %d: %s", resp.StatusCode, resp.Body)
	}

	return cardFromData(t, resp.Body)
}

func fetchSummary(t *testing.T, handler http.Handler, gameID string) apiSummary {
	t.Helper()

	resp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/summary", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected summary 200, got %d: %s", resp.StatusCode, resp.Body)
	}

	return summaryFromData(t, resp.Body)
}

func playerState(players []apiPlayer, playerID string) string {
	for _, player := range players {
		if player.ID == playerID {
			return player.State
		}
	}

	return ""
}

func activeWordCount(words []apiWordSetWord) int {
	count := 0
	for _, word := range words {
		if word.IsActive {
			count++
		}
	}

	return count
}

func firstNonFreeCell(t *testing.T, cells []apiCardCell) apiCardCell {
	t.Helper()

	for _, cell := range cells {
		if !cell.IsFreeSpace {
			return cell
		}
	}

	t.Fatal("expected at least one non-free cell")
	return apiCardCell{}
}

func sortedNonFreeCellsByWordNumber(t *testing.T, cells []apiCardCell) []apiCardCell {
	t.Helper()

	filtered := make([]apiCardCell, 0, len(cells))
	for _, cell := range cells {
		if !cell.IsFreeSpace {
			filtered = append(filtered, cell)
		}
	}
	sort.Slice(filtered, func(i, j int) bool {
		return apiWordNumber(filtered[i].Word) < apiWordNumber(filtered[j].Word)
	})
	if len(filtered) < 2 {
		t.Fatal("expected at least two non-free cells")
	}

	return filtered
}

func apiWordNumber(word string) int {
	var number int
	if _, err := fmt.Sscanf(word, "API Word %02d", &number); err != nil {
		return 999
	}

	return number
}

func cellByID(t *testing.T, cells []apiCardCell, id string) apiCardCell {
	t.Helper()

	for _, cell := range cells {
		if cell.ID == id {
			return cell
		}
	}

	t.Fatalf("expected cell %s in card", id)
	return apiCardCell{}
}

func hasCompleteLine(cells []apiCardCell, called map[string]struct{}) bool {
	cellsByPosition := make(map[[2]int]apiCardCell, len(cells))
	for _, cell := range cells {
		cellsByPosition[[2]int{cell.RowIndex, cell.ColIndex}] = cell
	}

	for _, line := range testSingleLinePositions() {
		complete := true
		for _, position := range line {
			cell, ok := cellsByPosition[position]
			if !ok {
				complete = false
				break
			}
			if cell.IsFreeSpace {
				continue
			}
			if _, ok := called[strings.ToLower(cell.Word)]; !ok {
				complete = false
				break
			}
		}
		if complete {
			return true
		}
	}

	return false
}

func testSingleLinePositions() [][][2]int {
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

	return append(lines, firstDiagonal, secondDiagonal)
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

func insertAPIWordSet(t *testing.T, ctx context.Context, pool *db.Pool) string {
	t.Helper()

	var wordSetID string
	if err := pool.QueryRow(ctx, `
		INSERT INTO word_sets (name, status, source)
		VALUES ('API Test Word Set', 'approved', 'manual')
		RETURNING id::text
	`).Scan(&wordSetID); err != nil {
		t.Fatalf("insert API word set: %v", err)
	}

	for index := 1; index <= 26; index++ {
		if _, err := pool.Exec(ctx, `
			INSERT INTO word_set_words (word_set_id, word, sort_order)
			VALUES ($1, $2, $3)
		`, wordSetID, fmt.Sprintf("API Word %02d", index), index); err != nil {
			t.Fatalf("insert API word set word %d: %v", index, err)
		}
	}

	return wordSetID
}

func callDeckRowsFromDB(t *testing.T, ctx context.Context, pool *db.Pool, gameID string) []apiCallDeckRow {
	t.Helper()

	rows, err := pool.Query(ctx, `
		SELECT id::text, word, sequence
		FROM game_call_deck
		WHERE game_run_id = $1
		ORDER BY sequence ASC
	`, gameID)
	if err != nil {
		t.Fatalf("query game call deck: %v", err)
	}
	defer rows.Close()

	deck := make([]apiCallDeckRow, 0)
	for rows.Next() {
		var item apiCallDeckRow
		if err := rows.Scan(&item.ID, &item.Word, &item.Sequence); err != nil {
			t.Fatalf("scan call deck row: %v", err)
		}
		deck = append(deck, item)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate call deck rows: %v", err)
	}

	return deck
}

func gameEventsFromDB(t *testing.T, ctx context.Context, pool *db.Pool, gameID string) []apiGameEvent {
	t.Helper()

	rows, err := pool.Query(ctx, `
		SELECT id::text, type, sequence
		FROM game_event_outbox
		WHERE game_run_id = $1
		ORDER BY sequence ASC
	`, gameID)
	if err != nil {
		t.Fatalf("query game event outbox: %v", err)
	}
	defer rows.Close()

	events := make([]apiGameEvent, 0)
	for rows.Next() {
		var event apiGameEvent
		if err := rows.Scan(&event.ID, &event.Type, &event.Sequence); err != nil {
			t.Fatalf("scan game event: %v", err)
		}
		events = append(events, event)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate game events: %v", err)
	}

	return events
}

func assertEventTypes(t *testing.T, events []apiGameEvent, expected []string) {
	t.Helper()

	seen := make(map[string]bool, len(events))
	for _, event := range events {
		seen[event.Type] = true
	}
	for _, eventType := range expected {
		if !seen[eventType] {
			t.Fatalf("expected event type %s in %+v", eventType, events)
		}
	}
}

func hasActivityEventType(events []apiActivityEvent, eventType string) bool {
	for _, event := range events {
		if event.Type == eventType {
			return true
		}
	}
	return false
}

func readSSEEvents(t *testing.T, url string, lastEventID string, limit int) []apiGameEvent {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("create SSE request: %v", err)
	}
	req.Header.Set("X-Dev-User-Email", "host@example.local")
	req.Header.Set("X-Dev-User-Role", "host")
	if lastEventID != "" {
		req.Header.Set("Last-Event-ID", lastEventID)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("open SSE stream: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected SSE status 200, got %d: %s", resp.StatusCode, string(body))
	}

	events := make([]apiGameEvent, 0, limit)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var event apiGameEvent
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &event); err != nil {
			t.Fatalf("decode SSE event: %v", err)
		}
		events = append(events, event)
		if len(events) == limit {
			cancel()
			return events
		}
	}
	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		t.Fatalf("scan SSE stream: %v", err)
	}

	return events
}

func readSSEHeartbeat(t *testing.T, url string, lastSequence int64) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		t.Fatalf("create heartbeat SSE request: %v", err)
	}
	req.Header.Set("X-Dev-User-Email", "host@example.local")
	req.Header.Set("X-Dev-User-Role", "host")
	req.Header.Set("Last-Event-ID", fmt.Sprintf("%d", lastSequence))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("open heartbeat SSE stream: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected heartbeat SSE status 200, got %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if scanner.Text() == ": heartbeat" {
			cancel()
			return
		}
	}
	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		t.Fatalf("scan heartbeat SSE stream: %v", err)
	}
	t.Fatal("expected heartbeat comment before stream closed")
}

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
