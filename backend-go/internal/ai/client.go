package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const MinGamePrepWords = 24

var ErrInvalidGamePrep = errors.New("invalid game prep response")

type GamePrepInput struct {
	GameRunID     string            `json:"gameRunId"`
	TopicPrompt   string            `json:"topicPrompt,omitempty"`
	WordCount     int               `json:"wordCount"`
	Tone          string            `json:"tone,omitempty"`
	Audience      string            `json:"audience,omitempty"`
	ExcludedWords []string          `json:"excludedWords"`
	Settings      map[string]string `json:"settings,omitempty"`
}

type GamePrepOutput struct {
	Topic       string   `json:"topic"`
	Summary     string   `json:"summary"`
	Words       []string `json:"words"`
	CallerStyle string   `json:"callerStyle,omitempty"`
	ThemePrompt string   `json:"themePrompt,omitempty"`
	Provider    string   `json:"provider,omitempty"`
}

type Client interface {
	GenerateGamePrep(context.Context, GamePrepInput) (GamePrepOutput, error)
	GenerateCallerAssetsBulk(context.Context, CallerAssetsBulkInput) (CallerAssetsBulkOutput, error)
	GenerateCallerAsset(context.Context, CallerAssetInput) (CallerAssetOutput, error)
	GenerateTheme(context.Context, ThemeInput) (ThemeOutput, error)
}

type CallDeckItemInput struct {
	CallDeckItemID string `json:"callDeckItemId"`
	Word           string `json:"word"`
	Sequence       int    `json:"sequence"`
}

type CallerAssetsBulkInput struct {
	GameRunID string              `json:"gameRunId"`
	VoiceName string              `json:"voiceName,omitempty"`
	Tone      string              `json:"tone,omitempty"`
	Deck      []CallDeckItemInput `json:"deck"`
}

type CallerAssetsBulkOutput struct {
	GameRunID string              `json:"gameRunId"`
	Assets    []CallerAssetOutput `json:"assets"`
	Provider  string              `json:"provider,omitempty"`
}

type CallerAssetInput struct {
	GameRunID      string `json:"gameRunId"`
	CallDeckItemID string `json:"callDeckItemId"`
	Word           string `json:"word"`
	Sequence       int    `json:"sequence"`
	VoiceName      string `json:"voiceName,omitempty"`
	Tone           string `json:"tone,omitempty"`
}

type CallerAssetOutput struct {
	CallDeckItemID string `json:"callDeckItemId"`
	Word           string `json:"word"`
	Sequence       int    `json:"sequence"`
	Line           string `json:"line"`
	AudioURL       string `json:"audioUrl,omitempty"`
	StorageKey     string `json:"storageKey,omitempty"`
	Status         string `json:"status"`
	ErrorReason    string `json:"errorReason,omitempty"`
	Provider       string `json:"provider,omitempty"`
}

type ThemeInput struct {
	GameRunID     string         `json:"gameRunId,omitempty"`
	Prompt        string         `json:"prompt"`
	Tone          string         `json:"tone,omitempty"`
	Accessibility map[string]any `json:"accessibility,omitempty"`
	AllowedAssets []string       `json:"allowedAssets,omitempty"`
	ExistingTheme map[string]any `json:"existingTheme,omitempty"`
}

type ThemeOutput struct {
	Name          string         `json:"name"`
	Summary       string         `json:"summary"`
	Palette       map[string]any `json:"palette"`
	Icons         []string       `json:"icons"`
	Decorations   []string       `json:"decorations"`
	Motion        string         `json:"motion"`
	CallerTone    string         `json:"callerTone"`
	Accessibility map[string]any `json:"accessibility"`
	Provider      string         `json:"provider,omitempty"`
}

type HTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
	Provider   string
}

func NewHTTPClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		BaseURL: strings.TrimRight(strings.TrimSpace(baseURL), "/"),
		HTTPClient: &http.Client{
			Timeout: timeout,
		},
		Provider: "python-ai-service",
	}
}

func (c *HTTPClient) GenerateGamePrep(ctx context.Context, input GamePrepInput) (GamePrepOutput, error) {
	if c == nil || strings.TrimSpace(c.BaseURL) == "" {
		return GamePrepOutput{}, fmt.Errorf("%w: AI service base URL is empty", ErrInvalidGamePrep)
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	payload, err := json.Marshal(input)
	if err != nil {
		return GamePrepOutput{}, fmt.Errorf("marshal game prep request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/ai/v1/game-prep", bytes.NewReader(payload))
	if err != nil {
		return GamePrepOutput{}, fmt.Errorf("build game prep request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return GamePrepOutput{}, fmt.Errorf("call AI game prep service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return GamePrepOutput{}, fmt.Errorf("AI game prep service returned status %d", resp.StatusCode)
	}

	var output GamePrepOutput
	if err := json.NewDecoder(resp.Body).Decode(&output); err != nil {
		return GamePrepOutput{}, fmt.Errorf("decode AI game prep response: %w", err)
	}
	output.Provider = firstNonBlank(output.Provider, c.Provider)

	return NormalizeGamePrepOutput(output)
}

func (c *HTTPClient) GenerateCallerAssetsBulk(ctx context.Context, input CallerAssetsBulkInput) (CallerAssetsBulkOutput, error) {
	var output CallerAssetsBulkOutput
	if err := c.post(ctx, "/ai/v1/caller-assets/bulk", input, &output); err != nil {
		return CallerAssetsBulkOutput{}, err
	}
	output.Provider = firstNonBlank(output.Provider, c.Provider)
	for index := range output.Assets {
		output.Assets[index] = normalizeCallerAsset(output.Assets[index], output.Provider)
	}
	return output, nil
}

func (c *HTTPClient) GenerateCallerAsset(ctx context.Context, input CallerAssetInput) (CallerAssetOutput, error) {
	var output CallerAssetOutput
	if err := c.post(ctx, "/ai/v1/caller-assets", input, &output); err != nil {
		return CallerAssetOutput{}, err
	}
	return normalizeCallerAsset(output, firstNonBlank(output.Provider, c.Provider)), nil
}

func (c *HTTPClient) GenerateTheme(ctx context.Context, input ThemeInput) (ThemeOutput, error) {
	var output ThemeOutput
	if err := c.post(ctx, "/ai/v1/themes/generate", input, &output); err != nil {
		return ThemeOutput{}, err
	}
	output.Provider = firstNonBlank(output.Provider, c.Provider)
	return NormalizeThemeOutput(output)
}

func (c *HTTPClient) post(ctx context.Context, path string, input any, output any) error {
	if c == nil || strings.TrimSpace(c.BaseURL) == "" {
		return fmt.Errorf("%w: AI service base URL is empty", ErrInvalidGamePrep)
	}
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	payload, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("marshal AI request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("build AI request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call AI service: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("AI service returned status %d", resp.StatusCode)
	}
	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return fmt.Errorf("decode AI response: %w", err)
	}
	return nil
}

type DisabledClient struct{}

func (DisabledClient) GenerateGamePrep(ctx context.Context, input GamePrepInput) (GamePrepOutput, error) {
	count := input.WordCount
	if count < MinGamePrepWords {
		count = MinGamePrepWords
	}
	words := make([]string, 0, count)
	for index := 1; index <= count; index++ {
		words = append(words, fmt.Sprintf("Local Prep Word %02d", index))
	}

	return NormalizeGamePrepOutput(GamePrepOutput{
		Topic:       "Local Generated Bingo",
		Summary:     "Local disabled AI response for content review and lock testing.",
		Words:       words,
		CallerStyle: firstNonBlank(input.Settings["callerStyle"], "light workplace caller"),
		Provider:    "local-disabled",
	})
}

func (DisabledClient) GenerateCallerAssetsBulk(ctx context.Context, input CallerAssetsBulkInput) (CallerAssetsBulkOutput, error) {
	assets := make([]CallerAssetOutput, 0, len(input.Deck))
	for _, item := range input.Deck {
		assets = append(assets, normalizeCallerAsset(CallerAssetOutput{
			CallDeckItemID: item.CallDeckItemID,
			Word:           item.Word,
			Sequence:       item.Sequence,
			Line:           fmt.Sprintf("Next word is %s.", strings.TrimSpace(item.Word)),
			Status:         "ready",
			Provider:       "local-disabled",
		}, "local-disabled"))
	}
	return CallerAssetsBulkOutput{
		GameRunID: input.GameRunID,
		Assets:    assets,
		Provider:  "local-disabled",
	}, nil
}

func (DisabledClient) GenerateCallerAsset(ctx context.Context, input CallerAssetInput) (CallerAssetOutput, error) {
	return normalizeCallerAsset(CallerAssetOutput{
		CallDeckItemID: input.CallDeckItemID,
		Word:           input.Word,
		Sequence:       input.Sequence,
		Line:           fmt.Sprintf("Next word is %s.", strings.TrimSpace(input.Word)),
		Status:         "ready",
		Provider:       "local-disabled",
	}, "local-disabled"), nil
}

func (DisabledClient) GenerateTheme(ctx context.Context, input ThemeInput) (ThemeOutput, error) {
	name := strings.TrimSpace(input.Prompt)
	if name == "" {
		name = "Local Theme"
	}
	return NormalizeThemeOutput(ThemeOutput{
		Name:    name,
		Summary: "Local disabled AI theme draft using safe structured tokens.",
		Palette: map[string]any{
			"primary":    "#1F7A8C",
			"accent":     "#F25F5C",
			"background": "#F7F9FB",
		},
		Icons:         []string{"sparkles", "briefcase"},
		Decorations:   []string{"confetti"},
		Motion:        "subtle",
		CallerTone:    "upbeat workplace host",
		Accessibility: map[string]any{"contrast": "AA"},
		Provider:      "local-disabled",
	})
}

func NormalizeGamePrepOutput(output GamePrepOutput) (GamePrepOutput, error) {
	output.Topic = strings.TrimSpace(output.Topic)
	output.Summary = strings.TrimSpace(output.Summary)
	output.CallerStyle = strings.TrimSpace(output.CallerStyle)
	output.ThemePrompt = strings.TrimSpace(output.ThemePrompt)
	output.Provider = strings.TrimSpace(output.Provider)
	if output.Topic == "" {
		return GamePrepOutput{}, fmt.Errorf("%w: topic is required", ErrInvalidGamePrep)
	}
	if output.Summary == "" {
		return GamePrepOutput{}, fmt.Errorf("%w: summary is required", ErrInvalidGamePrep)
	}

	normalized := make([]string, 0, len(output.Words))
	seen := make(map[string]struct{}, len(output.Words))
	for index, raw := range output.Words {
		word := strings.TrimSpace(raw)
		if word == "" {
			return GamePrepOutput{}, fmt.Errorf("%w: word %d cannot be blank", ErrInvalidGamePrep, index+1)
		}
		key := strings.ToLower(word)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		normalized = append(normalized, word)
	}
	if len(normalized) < MinGamePrepWords {
		return GamePrepOutput{}, fmt.Errorf("%w: at least %d unique words are required", ErrInvalidGamePrep, MinGamePrepWords)
	}
	output.Words = normalized

	return output, nil
}

func NormalizeThemeOutput(output ThemeOutput) (ThemeOutput, error) {
	output.Name = strings.TrimSpace(output.Name)
	output.Summary = strings.TrimSpace(output.Summary)
	output.Motion = strings.TrimSpace(output.Motion)
	output.CallerTone = strings.TrimSpace(output.CallerTone)
	output.Provider = strings.TrimSpace(output.Provider)
	if output.Name == "" {
		return ThemeOutput{}, fmt.Errorf("%w: theme name is required", ErrInvalidGamePrep)
	}
	if output.Summary == "" {
		return ThemeOutput{}, fmt.Errorf("%w: theme summary is required", ErrInvalidGamePrep)
	}
	if len(output.Palette) == 0 {
		return ThemeOutput{}, fmt.Errorf("%w: theme palette is required", ErrInvalidGamePrep)
	}
	if unsafeStructuredTheme(output) {
		return ThemeOutput{}, fmt.Errorf("%w: theme contains unsafe executable or external asset content", ErrInvalidGamePrep)
	}
	return output, nil
}

func normalizeCallerAsset(asset CallerAssetOutput, provider string) CallerAssetOutput {
	asset.Word = strings.TrimSpace(asset.Word)
	asset.Line = strings.TrimSpace(asset.Line)
	asset.Status = strings.ToLower(strings.TrimSpace(asset.Status))
	if asset.Status == "" {
		asset.Status = "ready"
	}
	if asset.Line == "" {
		asset.Line = fmt.Sprintf("Next word is %s.", asset.Word)
	}
	asset.Provider = firstNonBlank(asset.Provider, provider)
	return asset
}

func unsafeStructuredTheme(output ThemeOutput) bool {
	payload, err := json.Marshal(output)
	if err != nil {
		return true
	}
	value := strings.ToLower(string(payload))
	return strings.Contains(value, "<script") ||
		strings.Contains(value, "javascript:") ||
		strings.Contains(value, "http://") ||
		strings.Contains(value, "https://") ||
		strings.Contains(value, "url(")
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}

	return ""
}
