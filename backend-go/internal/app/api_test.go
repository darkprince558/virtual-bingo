package app

import (
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
	RowIndex    int    `json:"rowIndex"`
	ColIndex    int    `json:"colIndex"`
	Word        string `json:"word"`
	IsFreeSpace bool   `json:"isFreeSpace"`
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

func quoteIdentifier(value string) string {
	return `"` + strings.ReplaceAll(value, `"`, `""`) + `"`
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
