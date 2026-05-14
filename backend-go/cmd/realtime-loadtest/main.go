package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type envelope[T any] struct {
	Data T `json:"data"`
}

type gameRun struct {
	ID string `json:"id"`
}

type player struct {
	ID string `json:"id"`
}

type card struct {
	Cells []cardCell `json:"cells"`
}

type cardCell struct {
	ID          string `json:"id"`
	IsFreeSpace bool   `json:"isFreeSpace"`
}

type client struct {
	baseURL string
	http    *http.Client
}

func main() {
	baseURL := flag.String("base-url", "http://localhost:8080", "Go API base URL")
	wordSetID := flag.String("word-set-id", "00000000-0000-0000-0000-000000000201", "existing approved word set id")
	playerCount := flag.Int("players", 50, "number of SSE clients and players")
	wordCalls := flag.Int("word-calls", 12, "number of word calls to issue")
	flag.Parse()

	if *playerCount < 1 {
		log.Fatal("players must be at least 1")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	api := client{baseURL: strings.TrimRight(*baseURL, "/"), http: &http.Client{Timeout: 10 * time.Second}}
	run := must[gameRun](api.post(ctx, "/api/v1/games", map[string]any{
		"name":      fmt.Sprintf("Realtime Load %d", time.Now().Unix()),
		"code":      fmt.Sprintf("LOAD-%d", time.Now().UnixNano()%1000000),
		"wordSetId": *wordSetID,
	}, gameRun{}))
	log.Printf("created game %s", run.ID)

	players := make([]player, 0, *playerCount)
	cards := make([]card, 0, *playerCount)
	for index := 0; index < *playerCount; index++ {
		email := fmt.Sprintf("load-%02d@example.local", index+1)
		name := fmt.Sprintf("Load Player %02d", index+1)
		must[map[string]any](api.post(ctx, "/api/v1/games/"+run.ID+"/allowed-players", map[string]any{"email": email, "displayName": name}, map[string]any{}))
		joined := must[player](api.post(ctx, "/api/v1/games/"+run.ID+"/players", map[string]any{"email": email, "displayName": name}, player{}))
		assigned := must[card](api.post(ctx, "/api/v1/games/"+run.ID+"/players/"+joined.ID+"/card", nil, card{}))
		players = append(players, joined)
		cards = append(cards, assigned)
	}
	log.Printf("joined and assigned %d players", len(players))

	must[map[string]any](api.post(ctx, "/api/v1/games/"+run.ID+"/start", nil, map[string]any{}))

	streamCtx, stopStreams := context.WithCancel(ctx)
	var eventCount atomic.Int64
	var wg sync.WaitGroup
	for index := 0; index < *playerCount; index++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			api.streamEvents(streamCtx, run.ID, &eventCount)
		}()
	}

	for index, player := range players {
		cellID := firstMarkableCell(cards[index])
		if cellID == "" {
			continue
		}
		must[map[string]any](api.patch(ctx, "/api/v1/games/"+run.ID+"/players/"+player.ID+"/card/cells/"+cellID, map[string]any{"marked": true}, map[string]any{}))
	}
	log.Printf("sent mark burst for %d players", len(players))

	for index := 0; index < *wordCalls; index++ {
		must[map[string]any](api.post(ctx, "/api/v1/games/"+run.ID+"/calls", nil, map[string]any{}))
	}
	log.Printf("called %d words", *wordCalls)

	for index := 0; index < min(3, len(players)); index++ {
		must[map[string]any](api.post(ctx, "/api/v1/games/"+run.ID+"/claims", map[string]any{"playerId": players[index].ID, "pattern": "single_line"}, map[string]any{}))
	}

	must[map[string]any](api.get(ctx, "/api/v1/games/"+run.ID+"/host-snapshot", map[string]any{}))
	must[map[string]any](api.get(ctx, "/api/v1/games/"+run.ID+"/players/"+players[0].ID+"/snapshot", map[string]any{}))
	resumeCtx, cancelResume := context.WithTimeout(ctx, 500*time.Millisecond)
	api.streamEventsWithLastID(resumeCtx, run.ID, "1", &eventCount)
	cancelResume()

	time.Sleep(500 * time.Millisecond)
	stopStreams()
	wg.Wait()
	log.Printf("load helper completed: players=%d streamed_events_seen=%d", *playerCount, eventCount.Load())
}

func (c client) get(ctx context.Context, path string, out any) (any, error) {
	return c.do(ctx, http.MethodGet, path, nil, out)
}

func (c client) post(ctx context.Context, path string, body any, out any) (any, error) {
	return c.do(ctx, http.MethodPost, path, body, out)
}

func (c client) patch(ctx context.Context, path string, body any, out any) (any, error) {
	return c.do(ctx, http.MethodPatch, path, body, out)
}

func (c client) do(ctx context.Context, method, path string, body any, out any) (any, error) {
	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Dev-User-Email", "host@example.local")
	req.Header.Set("X-Dev-User-Name", "Load Test Host")
	req.Header.Set("X-Dev-User-Role", "host")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%s %s returned %d: %s", method, path, resp.StatusCode, string(responseBody))
	}
	if out == nil || len(responseBody) == 0 {
		return out, nil
	}

	wrapped := envelope[json.RawMessage]{}
	if err := json.Unmarshal(responseBody, &wrapped); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(wrapped.Data, out); err != nil {
		return nil, err
	}

	return out, nil
}

func (c client) streamEvents(ctx context.Context, gameID string, count *atomic.Int64) {
	c.streamEventsWithLastID(ctx, gameID, "", count)
}

func (c client) streamEventsWithLastID(ctx context.Context, gameID, lastEventID string, count *atomic.Int64) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/v1/games/"+gameID+"/events", nil)
	if err != nil {
		return
	}
	req.Header.Set("X-Dev-User-Email", "host@example.local")
	req.Header.Set("X-Dev-User-Role", "host")
	if lastEventID != "" {
		req.Header.Set("Last-Event-ID", lastEventID)
	}

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "event: ") {
			count.Add(1)
		}
	}
}

func firstMarkableCell(value card) string {
	for _, cell := range value.Cells {
		if !cell.IsFreeSpace {
			return cell.ID
		}
	}
	return ""
}

func must[T any](value any, err error) T {
	if err != nil {
		log.Fatal(err)
	}
	typed, ok := value.(T)
	if !ok {
		log.Fatalf("unexpected response type %T", value)
	}
	return typed
}
