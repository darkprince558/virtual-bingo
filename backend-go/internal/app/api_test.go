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
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/config"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/db"
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

	claimsResp := doRequest(t, handler, http.MethodGet, "/api/v1/games/"+gameID+"/claims", nil)
	if claimsResp.StatusCode != http.StatusOK {
		t.Fatalf("expected claims list 200, got %d: %s", claimsResp.StatusCode, claimsResp.Body)
	}
	if claims := claimsFromData(t, claimsResp.Body); len(claims) != 2 {
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

type apiCardCell struct {
	ID          string     `json:"id"`
	RowIndex    int        `json:"rowIndex"`
	ColIndex    int        `json:"colIndex"`
	Word        string     `json:"word"`
	IsFreeSpace bool       `json:"isFreeSpace"`
	MarkedAt    *time.Time `json:"markedAt"`
}

type apiCard struct {
	ID        string        `json:"id"`
	GameRunID string        `json:"gameRunId"`
	PlayerID  string        `json:"playerId"`
	Cells     []apiCardCell `json:"cells"`
}

type apiCalledWord struct {
	ID       string `json:"id"`
	Word     string `json:"word"`
	Sequence int    `json:"sequence"`
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
	ID    string `json:"id"`
	State string `json:"state"`
}

type apiHostSnapshot struct {
	Status      string          `json:"status"`
	PlayerCount int             `json:"playerCount"`
	CurrentWord *apiCalledWord  `json:"currentWord"`
	Players     []apiPlayer     `json:"players"`
	CalledWords []apiCalledWord `json:"calledWords"`
	Claims      []apiClaim      `json:"claims"`
	Winners     []apiWinner     `json:"winners"`
}

type apiPlayerSnapshot struct {
	Status      string          `json:"status"`
	CurrentWord *apiCalledWord  `json:"currentWord"`
	Player      apiPlayer       `json:"player"`
	Card        *apiCard        `json:"card"`
	CalledWords []apiCalledWord `json:"calledWords"`
	Claims      []apiClaim      `json:"claims"`
	Winners     []apiWinner     `json:"winners"`
}

type apiGameEvent struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Sequence int64  `json:"sequence"`
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
		"name":      name,
		"code":      code,
		"wordSetId": wordSetID,
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
