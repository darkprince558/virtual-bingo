package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestDisabledClientGeneratesValidLocalGamePrep(t *testing.T) {
	output, err := (DisabledClient{}).GenerateGamePrep(context.Background(), GamePrepInput{
		GameRunID: "game-1",
		WordCount: 10,
	})
	if err != nil {
		t.Fatalf("generate disabled prep: %v", err)
	}
	if output.Provider != "local-disabled" {
		t.Fatalf("expected local-disabled provider, got %q", output.Provider)
	}
	if len(output.Words) != MinGamePrepWords {
		t.Fatalf("expected disabled client to enforce %d words, got %d", MinGamePrepWords, len(output.Words))
	}
}

func TestNormalizeGamePrepOutputTrimsAndDedupesWords(t *testing.T) {
	words := make([]string, 0, MinGamePrepWords+2)
	words = append(words, "  Standup  ", "standup")
	for index := 1; index <= MinGamePrepWords; index++ {
		words = append(words, "Word "+string(rune('A'+index)))
	}

	output, err := NormalizeGamePrepOutput(GamePrepOutput{
		Topic:   "  Team Week ",
		Summary: " Summary ",
		Words:   words,
	})
	if err != nil {
		t.Fatalf("normalize output: %v", err)
	}
	if output.Topic != "Team Week" || output.Summary != "Summary" {
		t.Fatalf("expected trimmed topic/summary, got %+v", output)
	}
	if output.Words[0] != "Standup" {
		t.Fatalf("expected trimmed first word, got %q", output.Words[0])
	}
	for _, word := range output.Words[1:] {
		if strings.EqualFold(word, "standup") {
			t.Fatalf("expected case-insensitive duplicate removed, got %+v", output.Words)
		}
	}
}

func TestNormalizeGamePrepOutputRejectsInvalidWords(t *testing.T) {
	if _, err := NormalizeGamePrepOutput(GamePrepOutput{Topic: "Topic", Summary: "Summary", Words: []string{"ok", " "}}); err == nil {
		t.Fatal("expected blank word to fail")
	}

	words := make([]string, 0, MinGamePrepWords-1)
	for index := 1; index < MinGamePrepWords; index++ {
		words = append(words, "Word")
	}
	if _, err := NormalizeGamePrepOutput(GamePrepOutput{Topic: "Topic", Summary: "Summary", Words: words}); err == nil {
		t.Fatal("expected too few unique words to fail")
	}
}

func TestHTTPClientCallsGamePrepEndpoint(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/ai/v1/game-prep" || r.Method != http.MethodPost {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		var input GamePrepInput
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if input.GameRunID != "game-1" || input.WordCount != 24 {
			t.Fatalf("unexpected request body: %+v", input)
		}
		words := make([]string, 0, MinGamePrepWords)
		for index := 1; index <= MinGamePrepWords; index++ {
			words = append(words, "AI Word "+string(rune('A'+index)))
		}
		_ = json.NewEncoder(w).Encode(GamePrepOutput{
			Topic:   "AI Topic",
			Summary: "AI Summary",
			Words:   words,
		})
	}))
	defer server.Close()

	output, err := NewHTTPClient(server.URL, time.Second).GenerateGamePrep(context.Background(), GamePrepInput{
		GameRunID: "game-1",
		WordCount: MinGamePrepWords,
	})
	if err != nil {
		t.Fatalf("generate HTTP prep: %v", err)
	}
	if output.Provider != "python-ai-service" || len(output.Words) != MinGamePrepWords {
		t.Fatalf("unexpected HTTP output: %+v", output)
	}
}

func TestDisabledClientCallerAssetsAndTheme(t *testing.T) {
	client := DisabledClient{}
	assets, err := client.GenerateCallerAssetsBulk(context.Background(), CallerAssetsBulkInput{
		GameRunID: "game-1",
		Deck: []CallDeckItemInput{
			{CallDeckItemID: "deck-1", Word: "Synergy", Sequence: 1},
		},
	})
	if err != nil {
		t.Fatalf("generate disabled caller assets: %v", err)
	}
	if len(assets.Assets) != 1 || assets.Assets[0].Status != "ready" || assets.Assets[0].Line != "Next word is Synergy." {
		t.Fatalf("unexpected disabled caller assets: %+v", assets)
	}

	theme, err := client.GenerateTheme(context.Background(), ThemeInput{Prompt: "Launch Week"})
	if err != nil {
		t.Fatalf("generate disabled theme: %v", err)
	}
	if theme.Name != "Launch Week" || theme.Provider != "local-disabled" || len(theme.Palette) == 0 {
		t.Fatalf("unexpected disabled theme: %+v", theme)
	}
}

func TestHTTPClientCallsCallerAssetsAndThemeEndpoints(t *testing.T) {
	seen := make(map[string]bool)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen[r.URL.Path] = true
		switch r.URL.Path {
		case "/ai/v1/caller-assets/bulk":
			_ = json.NewEncoder(w).Encode(CallerAssetsBulkOutput{GameRunID: "game-1", Assets: []CallerAssetOutput{{CallDeckItemID: "deck-1", Word: "Roadmap", Sequence: 1, Line: "Roadmap is up.", Status: "ready"}}})
		case "/ai/v1/themes/generate":
			_ = json.NewEncoder(w).Encode(ThemeOutput{Name: "Roadmap Theme", Summary: "A safe theme", Palette: map[string]any{"primary": "#1F7A8C"}, Icons: []string{"sparkles"}, Decorations: []string{"confetti"}, Motion: "subtle", CallerTone: "fun"})
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewHTTPClient(server.URL, time.Second)
	if _, err := client.GenerateCallerAssetsBulk(context.Background(), CallerAssetsBulkInput{GameRunID: "game-1", Deck: []CallDeckItemInput{{CallDeckItemID: "deck-1", Word: "Roadmap", Sequence: 1}}}); err != nil {
		t.Fatalf("generate caller assets: %v", err)
	}
	if _, err := client.GenerateTheme(context.Background(), ThemeInput{Prompt: "Roadmap"}); err != nil {
		t.Fatalf("generate theme: %v", err)
	}
	if !seen["/ai/v1/caller-assets/bulk"] || !seen["/ai/v1/themes/generate"] {
		t.Fatalf("expected both endpoints to be called, saw %+v", seen)
	}
}
