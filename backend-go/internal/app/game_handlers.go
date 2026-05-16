package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/auth"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/domain"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/game"
)

type createGameRequest struct {
	Name             string     `json:"name"`
	Code             string     `json:"code"`
	WordSetID        *string    `json:"wordSetId"`
	ScheduledStartAt *time.Time `json:"scheduledStartAt"`
	WinningPattern   *string    `json:"winningPattern"`
}

type addAllowedPlayerRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type joinPlayerRequest struct {
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
}

type markCardCellRequest struct {
	Marked bool `json:"marked"`
}

type submitBingoClaimRequest struct {
	PlayerID string `json:"playerId"`
	Pattern  string `json:"pattern"`
}

type updateGameSettingsRequest struct {
	MarkingMode                  *string `json:"markingMode"`
	AllowPlayerMarkingModeChoice *bool   `json:"allowPlayerMarkingModeChoice"`
	ShowClaimReadiness           *bool   `json:"showClaimReadiness"`
	VoiceClaimMode               *string `json:"voiceClaimMode"`
	VoiceClaimAutoplay           *bool   `json:"voiceClaimAutoplay"`
	CallerMode                   *string `json:"callerMode"`
	ThemeMode                    *string `json:"themeMode"`
	ThemeID                      *string `json:"themeId"`
}

type updateGameContentRequest struct {
	Topic       *string   `json:"topic"`
	Summary     *string   `json:"summary"`
	Words       *[]string `json:"words"`
	CallerStyle *string   `json:"callerStyle"`
}

type updatePlayerProfileRequest struct {
	Icon        string `json:"icon"`
	AvatarColor string `json:"avatarColor"`
	AvatarLabel string `json:"avatarLabel"`
}

type generateThemeRequest struct {
	GameRunID *string `json:"gameRunId"`
	Prompt    string  `json:"prompt"`
	Tone      string  `json:"tone"`
}

type updateThemeRequest struct {
	Name          *string        `json:"name"`
	Summary       *string        `json:"summary"`
	Palette       map[string]any `json:"palette"`
	Icons         []string       `json:"icons"`
	Decorations   []string       `json:"decorations"`
	Motion        string         `json:"motion"`
	CallerTone    string         `json:"callerTone"`
	Accessibility map[string]any `json:"accessibility"`
}

type applyThemeRequest struct {
	ThemeID string `json:"themeId"`
}

type updatePlayerPreferencesRequest struct {
	MarkingMode *string `json:"markingMode"`
}

func (s *Server) createGame(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	var req createGameRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}

	run, err := s.service.CreateGameRun(r.Context(), principal, game.CreateGameRunInput{
		Name:             req.Name,
		Code:             req.Code,
		WordSetID:        req.WordSetID,
		ScheduledStartAt: req.ScheduledStartAt,
		WinningPattern:   req.WinningPattern,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusCreated, gameRunResponseFromDomain(run))
}

func (s *Server) getGame(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	run, err := s.service.GetGameRun(r.Context(), r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, gameRunResponseFromDomain(run))
}

func (s *Server) startGame(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	run, err := s.service.StartGame(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, gameRunResponseFromDomain(run))
}

func (s *Server) pauseGame(w http.ResponseWriter, r *http.Request) {
	s.lifecycleGame(w, r, s.service.PauseGame)
}

func (s *Server) resumeGame(w http.ResponseWriter, r *http.Request) {
	s.lifecycleGame(w, r, s.service.ResumeGame)
}

func (s *Server) finishGame(w http.ResponseWriter, r *http.Request) {
	s.lifecycleGame(w, r, s.service.FinishGame)
}

func (s *Server) cancelGame(w http.ResponseWriter, r *http.Request) {
	s.lifecycleGame(w, r, s.service.CancelGame)
}

func (s *Server) lifecycleGame(w http.ResponseWriter, r *http.Request, action func(context.Context, auth.Principal, string) (game.GameRunWithCounts, error)) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	run, err := action(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, gameRunResponseFromDomain(run))
}

func (s *Server) callNextWord(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	calledWord, err := s.service.CallNextWord(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusCreated, calledWordResponseFromDomain(calledWord))
}

func (s *Server) listCalledWords(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	calledWords, err := s.service.ListCalledWords(r.Context(), r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	response := make([]calledWordResponse, 0, len(calledWords))
	for _, calledWord := range calledWords {
		response = append(response, calledWordResponseFromDomain(calledWord))
	}

	writeData(w, http.StatusOK, response)
}

func (s *Server) addAllowedPlayer(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	var req addAllowedPlayerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}

	player, err := s.service.AddAllowedPlayer(r.Context(), principal, game.AddAllowedPlayerInput{
		GameRunID:   r.PathValue("gameID"),
		Email:       req.Email,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusCreated, allowedPlayerResponseFromDomain(player))
}

func (s *Server) listAllowedPlayers(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	players, err := s.service.ListAllowedPlayers(r.Context(), r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	response := make([]allowedPlayerResponse, 0, len(players))
	for _, player := range players {
		response = append(response, allowedPlayerResponseFromDomain(player))
	}

	writeData(w, http.StatusOK, response)
}

func (s *Server) getGameSettings(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	settings, err := s.service.GetGameSettings(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, gameSettingsResponseFromDomain(settings))
}

func (s *Server) updateGameSettings(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	var req updateGameSettingsRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}

	settings, err := s.service.UpdateGameSettings(r.Context(), principal, game.UpdateGameSettingsInput{
		GameRunID:                    r.PathValue("gameID"),
		MarkingMode:                  req.MarkingMode,
		AllowPlayerMarkingModeChoice: req.AllowPlayerMarkingModeChoice,
		ShowClaimReadiness:           req.ShowClaimReadiness,
		VoiceClaimMode:               req.VoiceClaimMode,
		VoiceClaimAutoplay:           req.VoiceClaimAutoplay,
		CallerMode:                   req.CallerMode,
		ThemeMode:                    req.ThemeMode,
		ThemeID:                      req.ThemeID,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, gameSettingsResponseFromDomain(settings))
}

func (s *Server) getCurrentPlayerPreferences(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	preferences, err := s.service.GetCurrentPlayerPreferences(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, playerPreferencesResponseFromDomain(preferences))
}

func (s *Server) updateCurrentPlayerPreferences(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	var req updatePlayerPreferencesRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}

	preferences, err := s.service.UpdateCurrentPlayerPreferences(r.Context(), principal, game.UpdatePlayerPreferencesInput{
		GameRunID:   r.PathValue("gameID"),
		MarkingMode: req.MarkingMode,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, playerPreferencesResponseFromDomain(preferences))
}

func (s *Server) runAutoMark(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	result, err := s.service.AutoMarkGame(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, autoMarkRunResponse{
		PlayersScanned:     result.PlayersScanned,
		PlayersMarked:      result.PlayersMarked,
		CalledWordsScanned: result.CalledWordsScanned,
		CellsMarked:        result.CellsMarked,
		Mode:               result.Mode,
		SkippedReason:      result.SkippedReason,
	})
}

func (s *Server) prepareGameContent(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	content, err := s.service.PrepareGameContentForHost(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusCreated, gameContentResponseFromDomain(content))
}

func (s *Server) getGameContent(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	content, err := s.service.GetGeneratedGameContent(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, gameContentResponseFromDomain(content))
}

func (s *Server) updateGameContent(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	var req updateGameContentRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}
	words := []string(nil)
	hasWordPatch := req.Words != nil
	if req.Words != nil {
		words = *req.Words
	}

	content, err := s.service.UpdateGeneratedGameContent(r.Context(), principal, game.UpdateGeneratedContentInput{
		GameRunID:    r.PathValue("gameID"),
		Topic:        req.Topic,
		Summary:      req.Summary,
		Words:        words,
		CallerStyle:  req.CallerStyle,
		HasWordPatch: hasWordPatch,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, gameContentResponseFromDomain(content))
}

func (s *Server) lockGameContent(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	content, err := s.service.LockGameContentForHost(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, gameContentResponseFromDomain(content))
}

func (s *Server) generateCallerAssets(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	assets, err := s.service.GenerateCallerAssets(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}
	response := make([]callerAssetResponse, 0, len(assets))
	for _, asset := range assets {
		response = append(response, callerAssetResponseFromDomain(asset))
	}
	writeData(w, http.StatusOK, response)
}

func (s *Server) sendPlayerInvites(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	attempts, err := s.service.SendMockPlayerInvites(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeData(w, http.StatusCreated, deliveryAttemptsResponseFromDomain(attempts))
}

func (s *Server) listDeliveries(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	attempts, err := s.service.ListDeliveries(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, deliveryAttemptsResponseFromDomain(attempts))
}

func (s *Server) retryDelivery(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	attempt, err := s.service.RetryDelivery(r.Context(), principal, r.PathValue("deliveryID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, deliveryAttemptResponseFromDomain(attempt))
}

func (s *Server) openLobby(w http.ResponseWriter, r *http.Request) {
	s.lifecycleGame(w, r, s.service.OpenLobby)
}

func (s *Server) updateCurrentPlayerProfile(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	var req updatePlayerProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}
	player, err := s.service.UpdateCurrentPlayerProfile(r.Context(), principal, game.UpdatePlayerProfileInput{GameRunID: r.PathValue("gameID"), Icon: req.Icon, AvatarColor: req.AvatarColor, AvatarLabel: req.AvatarLabel})
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, playerResponseFromDomain(player))
}

func (s *Server) generateTheme(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	var req generateThemeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}
	theme, err := s.service.GenerateTheme(r.Context(), principal, game.GenerateThemeInput{GameRunID: req.GameRunID, Prompt: req.Prompt, Tone: req.Tone})
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeData(w, http.StatusCreated, themeResponseFromDomain(theme))
}

func (s *Server) getTheme(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	theme, err := s.service.GetTheme(r.Context(), principal, r.PathValue("themeID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, themeResponseFromDomain(theme))
}

func (s *Server) updateTheme(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	existing, err := s.service.GetTheme(r.Context(), principal, r.PathValue("themeID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}
	var req updateThemeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}
	tokens := existing.Tokens
	if req.Palette != nil {
		tokens.Palette = req.Palette
	}
	if req.Icons != nil {
		tokens.Icons = req.Icons
	}
	if req.Decorations != nil {
		tokens.Decorations = req.Decorations
	}
	if req.Motion != "" {
		tokens.Motion = req.Motion
	}
	if req.CallerTone != "" {
		tokens.CallerTone = req.CallerTone
	}
	if req.Accessibility != nil {
		tokens.Accessibility = req.Accessibility
	}
	theme, err := s.service.UpdateTheme(r.Context(), principal, r.PathValue("themeID"), tokens, req.Name, req.Summary)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, themeResponseFromDomain(theme))
}

func (s *Server) approveTheme(w http.ResponseWriter, r *http.Request) {
	s.setThemeApproval(w, r, true)
}

func (s *Server) rejectTheme(w http.ResponseWriter, r *http.Request) {
	s.setThemeApproval(w, r, false)
}

func (s *Server) setThemeApproval(w http.ResponseWriter, r *http.Request, approved bool) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	theme, err := s.service.SetThemeApproval(r.Context(), principal, r.PathValue("themeID"), approved)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, themeResponseFromDomain(theme))
}

func (s *Server) applyThemeToGame(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	var req applyThemeRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}
	settings, err := s.service.ApplyThemeToGame(r.Context(), principal, r.PathValue("gameID"), req.ThemeID)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	writeData(w, http.StatusOK, gameSettingsResponseFromDomain(settings))
}

func (s *Server) listThemeAssets(w http.ResponseWriter, r *http.Request) {
	writeData(w, http.StatusOK, game.ThemeAssetIDs())
}

func (s *Server) getCurrentPlayerClaimReadiness(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	readiness, err := s.service.GetCurrentPlayerClaimReadiness(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, claimReadinessResponseFromDomain(readiness))
}

func (s *Server) joinPlayer(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	var req joinPlayerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}

	update, err := s.service.JoinPlayer(r.Context(), game.JoinPlayerInput{
		GameRunID:   r.PathValue("gameID"),
		Email:       req.Email,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusCreated, playerConnectionResponseFromDomain(update))
}

func (s *Server) assignPlayerCard(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	card, err := s.service.AssignPlayerCard(r.Context(), r.PathValue("gameID"), r.PathValue("playerID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusCreated, cardResponseFromDomain(card))
}

func (s *Server) getPlayerCard(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	card, err := s.service.GetPlayerCard(r.Context(), r.PathValue("gameID"), r.PathValue("playerID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, cardResponseFromDomain(card))
}

func (s *Server) markCardCell(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	var req markCardCellRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}

	cell, err := s.service.MarkCardCell(r.Context(), game.MarkCardCellInput{
		GameRunID: r.PathValue("gameID"),
		PlayerID:  r.PathValue("playerID"),
		CellID:    r.PathValue("cellID"),
		Marked:    req.Marked,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, cardCellResponseFromDomain(cell))
}

func (s *Server) submitBingoClaim(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	var req submitBingoClaimRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}

	result, err := s.service.SubmitBingoClaim(r.Context(), game.SubmitBingoClaimInput{
		GameRunID: r.PathValue("gameID"),
		PlayerID:  req.PlayerID,
		Pattern:   req.Pattern,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusCreated, claimSubmissionResponseFromDomain(result))
}

func (s *Server) listBingoClaims(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	if !auth.HasRole(principal, "admin", "host") {
		mapServiceError(w, game.ErrForbidden)
		return
	}

	claims, err := s.service.ListBingoClaims(r.Context(), r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	response := make([]claimResponse, 0, len(claims))
	for _, claim := range claims {
		response = append(response, claimResponseFromDomain(claim))
	}

	writeData(w, http.StatusOK, response)
}

func (s *Server) getGameSummary(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	summary, err := s.service.GetGameSummary(r.Context(), r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, gameSummaryResponseFromDomain(summary))
}

func (s *Server) getHostSnapshot(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	snapshot, err := s.service.GetHostSnapshot(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, hostSnapshotResponseFromDomain(snapshot))
}

func (s *Server) getPlayerSnapshot(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	snapshot, err := s.service.GetPlayerSnapshot(r.Context(), principal, r.PathValue("gameID"), r.PathValue("playerID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, playerSnapshotResponseFromDomain(snapshot))
}

func (s *Server) heartbeatPlayer(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	update, err := s.service.HeartbeatPlayer(r.Context(), principal, r.PathValue("gameID"), r.PathValue("playerID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, playerConnectionResponseFromDomain(update))
}

func (s *Server) streamGameEvents(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	if _, err := s.service.Authenticate(r); err != nil {
		mapServiceError(w, err)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeAPIError(w, http.StatusInternalServerError, "streaming_unsupported", "response writer does not support streaming")
		return
	}

	afterSequence := parseLastEventSequence(r)
	pollInterval := 500 * time.Millisecond
	heartbeatInterval := 15 * time.Second
	if value := r.URL.Query().Get("heartbeatMs"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil && parsed >= 25 {
			heartbeatInterval = time.Duration(parsed) * time.Millisecond
		}
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()
	heartbeatTicker := time.NewTicker(heartbeatInterval)
	defer heartbeatTicker.Stop()

	writeSSEComment(w, "connected")
	flusher.Flush()

	for {
		events, err := s.service.ListGameEvents(r.Context(), r.PathValue("gameID"), afterSequence, 100)
		if err != nil {
			writeSSEComment(w, "error")
			flusher.Flush()
			return
		}
		for _, event := range events {
			if err := writeSSEEvent(w, eventResponseFromDomain(event)); err != nil {
				return
			}
			afterSequence = event.Sequence
		}
		if len(events) > 0 {
			flusher.Flush()
			continue
		}

		select {
		case <-r.Context().Done():
			return
		case <-pollTicker.C:
		case <-heartbeatTicker.C:
			writeSSEComment(w, "heartbeat")
			flusher.Flush()
		}
	}
}

func parseLastEventSequence(r *http.Request) int64 {
	value := r.Header.Get("Last-Event-ID")
	if value == "" {
		value = r.URL.Query().Get("lastEventId")
	}
	if value == "" {
		return 0
	}

	sequence, err := strconv.ParseInt(value, 10, 64)
	if err != nil || sequence < 0 {
		return 0
	}

	return sequence
}

func writeSSEComment(w http.ResponseWriter, value string) {
	_, _ = fmt.Fprintf(w, ": %s\n\n", value)
}

func writeSSEEvent(w http.ResponseWriter, event eventResponse) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(w, "id: %d\nevent: %s\ndata: %s\n\n", event.Sequence, event.Type, payload)
	return err
}

type gameRunResponse struct {
	ID                  string     `json:"id"`
	TemplateID          *string    `json:"templateId,omitempty"`
	HostUserID          string     `json:"hostUserId"`
	WordSetID           *string    `json:"wordSetId,omitempty"`
	Code                string     `json:"code"`
	Name                string     `json:"name"`
	Status              string     `json:"status"`
	ScheduledStartAt    *time.Time `json:"scheduledStartAt,omitempty"`
	StartedAt           *time.Time `json:"startedAt,omitempty"`
	EndedAt             *time.Time `json:"endedAt,omitempty"`
	CurrentCalledWordID *string    `json:"currentCalledWordId,omitempty"`
	WinningPattern      *string    `json:"winningPattern,omitempty"`
	AllowedPlayerCount  int        `json:"allowedPlayerCount"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type gameSettingsResponse struct {
	GameRunID                    string    `json:"gameRunId"`
	MarkingMode                  string    `json:"markingMode"`
	AllowPlayerMarkingModeChoice bool      `json:"allowPlayerMarkingModeChoice"`
	ShowClaimReadiness           bool      `json:"showClaimReadiness"`
	VoiceClaimMode               string    `json:"voiceClaimMode"`
	VoiceClaimAutoplay           bool      `json:"voiceClaimAutoplay"`
	CallerMode                   string    `json:"callerMode"`
	ThemeMode                    string    `json:"themeMode"`
	ThemeID                      *string   `json:"themeId,omitempty"`
	CreatedAt                    time.Time `json:"createdAt"`
	UpdatedAt                    time.Time `json:"updatedAt"`
}

type playerPreferencesResponse struct {
	PlayerID    string    `json:"playerId"`
	MarkingMode *string   `json:"markingMode"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type autoMarkRunResponse struct {
	PlayersScanned     int    `json:"playersScanned"`
	PlayersMarked      int    `json:"playersMarked"`
	CalledWordsScanned int    `json:"calledWordsScanned"`
	CellsMarked        int    `json:"cellsMarked"`
	Mode               string `json:"mode"`
	SkippedReason      string `json:"skippedReason,omitempty"`
}

type gameContentResponse struct {
	ID                   string     `json:"id"`
	GameRunID            string     `json:"gameRunId"`
	GenerationJobID      *string    `json:"generationJobId,omitempty"`
	Status               string     `json:"status"`
	Topic                string     `json:"topic"`
	Summary              string     `json:"summary"`
	Words                []string   `json:"words"`
	GeneratedWords       []string   `json:"generatedWords,omitempty"`
	CallerStyle          *string    `json:"callerStyle,omitempty"`
	ThemePrompt          *string    `json:"themePrompt,omitempty"`
	ReviewWindowOpensAt  *time.Time `json:"reviewWindowOpensAt,omitempty"`
	ReviewWindowClosesAt *time.Time `json:"reviewWindowClosesAt,omitempty"`
	LockedAt             *time.Time `json:"lockedAt,omitempty"`
	LockedWordSetID      *string    `json:"lockedWordSetId,omitempty"`
	GenerationProvider   string     `json:"generationProvider"`
	GenerationError      *string    `json:"generationError,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

type claimReadinessResponse struct {
	Ready             bool               `json:"ready"`
	SupportedPatterns []string           `json:"supportedPatterns"`
	ReadyPatterns     []string           `json:"readyPatterns"`
	BestPattern       string             `json:"bestPattern"`
	MatchedCells      []cardCellResponse `json:"matchedCells"`
	MissingCells      []cardCellResponse `json:"missingCells"`
	Reason            string             `json:"reason"`
}

type allowedPlayerResponse struct {
	ID          string    `json:"id"`
	GameRunID   string    `json:"gameRunId"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"createdAt"`
}

type playerResponse struct {
	ID              string                   `json:"id"`
	GameRunID       string                   `json:"gameRunId"`
	UserID          *string                  `json:"userId,omitempty"`
	Email           string                   `json:"email"`
	DisplayName     string                   `json:"displayName"`
	Icon            *string                  `json:"icon,omitempty"`
	AvatarColor     *string                  `json:"avatarColor,omitempty"`
	AvatarLabel     *string                  `json:"avatarLabel,omitempty"`
	ConnectionState string                   `json:"connectionState"`
	State           string                   `json:"state"`
	JoinedAt        time.Time                `json:"joinedAt"`
	LastSeenAt      time.Time                `json:"lastSeenAt"`
	ReconnectNotice *reconnectNoticeResponse `json:"reconnectNotice,omitempty"`
}

type cardResponse struct {
	ID        string             `json:"id"`
	GameRunID string             `json:"gameRunId"`
	PlayerID  string             `json:"playerId"`
	Seed      string             `json:"seed"`
	Cells     []cardCellResponse `json:"cells"`
	CreatedAt time.Time          `json:"createdAt"`
}

type cardCellResponse struct {
	ID          string     `json:"id"`
	RowIndex    int        `json:"rowIndex"`
	ColIndex    int        `json:"colIndex"`
	Word        string     `json:"word"`
	IsFreeSpace bool       `json:"isFreeSpace"`
	MarkedAt    *time.Time `json:"markedAt,omitempty"`
}

type calledWordResponse struct {
	ID             string               `json:"id"`
	GameRunID      string               `json:"gameRunId"`
	WordSetWordID  *string              `json:"wordSetWordId,omitempty"`
	Word           string               `json:"word"`
	CalledByUserID *string              `json:"calledByUserId,omitempty"`
	Sequence       int                  `json:"sequence"`
	CalledAt       time.Time            `json:"calledAt"`
	CreatedAt      time.Time            `json:"createdAt"`
	CallerAsset    *callerAssetResponse `json:"callerAsset,omitempty"`
}

type callerAssetResponse struct {
	ID             string  `json:"id"`
	GameRunID      string  `json:"gameRunId"`
	CallDeckItemID string  `json:"callDeckItemId"`
	Word           string  `json:"word"`
	Sequence       int     `json:"sequence"`
	Line           string  `json:"line"`
	AudioURL       *string `json:"audioUrl,omitempty"`
	StorageKey     *string `json:"storageKey,omitempty"`
	VoiceName      *string `json:"voiceName,omitempty"`
	Provider       string  `json:"provider"`
	Status         string  `json:"status"`
	ErrorReason    *string `json:"errorReason,omitempty"`
}

type deliveryAttemptResponse struct {
	ID             string     `json:"id"`
	GameRunID      string     `json:"gameRunId"`
	Channel        string     `json:"channel"`
	Purpose        string     `json:"purpose"`
	RecipientEmail string     `json:"recipientEmail"`
	Subject        string     `json:"subject"`
	TemplateKey    string     `json:"templateKey"`
	BodyPreview    string     `json:"bodyPreview"`
	LinkURL        string     `json:"linkUrl"`
	GameCode       string     `json:"gameCode"`
	Status         string     `json:"status"`
	ErrorReason    *string    `json:"errorReason,omitempty"`
	SentAt         *time.Time `json:"sentAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

type themeResponse struct {
	ID        string             `json:"id"`
	GameRunID *string            `json:"gameRunId,omitempty"`
	Name      string             `json:"name"`
	Summary   string             `json:"summary"`
	Tokens    domain.ThemeTokens `json:"tokens"`
	Status    string             `json:"status"`
	Provider  string             `json:"provider"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}

type claimResponse struct {
	ID               string          `json:"id"`
	GameRunID        string          `json:"gameRunId"`
	PlayerID         string          `json:"playerId"`
	Pattern          string          `json:"pattern"`
	Status           string          `json:"status"`
	ValidationResult json.RawMessage `json:"validationResult"`
	ClaimedAt        time.Time       `json:"claimedAt"`
	ReviewedByUserID *string         `json:"reviewedByUserId,omitempty"`
	ReviewedAt       *time.Time      `json:"reviewedAt,omitempty"`
	CreatedAt        time.Time       `json:"createdAt"`
	UpdatedAt        time.Time       `json:"updatedAt"`
}

type claimSubmissionResponse struct {
	Claim  claimResponse   `json:"claim"`
	Winner *winnerResponse `json:"winner,omitempty"`
}

type winnerResponse struct {
	ID          string    `json:"id"`
	GameRunID   string    `json:"gameRunId"`
	PlayerID    string    `json:"playerId"`
	ClaimID     *string   `json:"claimId,omitempty"`
	Placement   int       `json:"placement"`
	Pattern     string    `json:"pattern"`
	ConfirmedAt time.Time `json:"confirmedAt"`
	CreatedAt   time.Time `json:"createdAt"`
}

type gameSummaryResponse struct {
	GameRun         gameRunResponse      `json:"gameRun"`
	PlayerCount     int                  `json:"playerCount"`
	CalledWordCount int                  `json:"calledWordCount"`
	CurrentWord     *calledWordResponse  `json:"currentWord,omitempty"`
	Claims          []claimResponse      `json:"claims"`
	Winners         []winnerResponse     `json:"winners"`
	Players         []playerResponse     `json:"players"`
	CalledWords     []calledWordResponse `json:"calledWords"`
	Status          string               `json:"status"`
}

type hostSnapshotResponse struct {
	GameRun            gameRunResponse      `json:"gameRun"`
	Settings           gameSettingsResponse `json:"settings"`
	Status             string               `json:"status"`
	CurrentWord        *calledWordResponse  `json:"currentWord,omitempty"`
	CurrentCallerAsset *callerAssetResponse `json:"currentCallerAsset,omitempty"`
	AppliedTheme       *themeResponse       `json:"appliedTheme,omitempty"`
	WinningPattern     string               `json:"winningPattern"`
	PlayerCount        int                  `json:"playerCount"`
	Players            []playerResponse     `json:"players"`
	CalledWords        []calledWordResponse `json:"calledWords"`
	Claims             []claimResponse      `json:"claims"`
	Winners            []winnerResponse     `json:"winners"`
}

type playerSnapshotResponse struct {
	GameRun                      gameRunResponse          `json:"gameRun"`
	MarkingMode                  string                   `json:"markingMode"`
	AllowPlayerMarkingModeChoice bool                     `json:"allowPlayerMarkingModeChoice"`
	ShowClaimReadiness           bool                     `json:"showClaimReadiness"`
	Status                       string                   `json:"status"`
	CurrentWord                  *calledWordResponse      `json:"currentWord,omitempty"`
	CurrentCallerAsset           *callerAssetResponse     `json:"currentCallerAsset,omitempty"`
	AppliedTheme                 *themeResponse           `json:"appliedTheme,omitempty"`
	WinningPattern               string                   `json:"winningPattern"`
	Player                       playerResponse           `json:"player"`
	Card                         *cardResponse            `json:"card,omitempty"`
	CalledWords                  []calledWordResponse     `json:"calledWords"`
	Claims                       []claimResponse          `json:"claims"`
	Winners                      []winnerResponse         `json:"winners"`
	ReconnectNotice              *reconnectNoticeResponse `json:"reconnectNotice,omitempty"`
}

type eventResponse struct {
	ID        string          `json:"id"`
	GameRunID string          `json:"gameRunId"`
	Type      string          `json:"type"`
	EntityID  *string         `json:"entityId,omitempty"`
	Payload   json.RawMessage `json:"payload"`
	Sequence  int64           `json:"sequence"`
	CreatedAt time.Time       `json:"createdAt"`
}

type reconnectNoticeResponse struct {
	LastSeenAt        time.Time            `json:"lastSeenAt"`
	MissedCalledWords []calledWordResponse `json:"missedCalledWords"`
}

func gameRunResponseFromDomain(run game.GameRunWithCounts) gameRunResponse {
	return gameRunResponse{
		ID:                  run.GameRun.ID,
		TemplateID:          run.GameRun.TemplateID,
		HostUserID:          run.GameRun.HostUserID,
		WordSetID:           run.GameRun.WordSetID,
		Code:                run.GameRun.Code,
		Name:                run.GameRun.Name,
		Status:              run.GameRun.Status,
		ScheduledStartAt:    run.GameRun.ScheduledStartAt,
		StartedAt:           run.GameRun.StartedAt,
		EndedAt:             run.GameRun.EndedAt,
		CurrentCalledWordID: run.GameRun.CurrentCalledWordID,
		WinningPattern:      run.GameRun.WinningPattern,
		AllowedPlayerCount:  run.AllowedPlayerCount,
		CreatedAt:           run.GameRun.CreatedAt,
		UpdatedAt:           run.GameRun.UpdatedAt,
	}
}

func gameSettingsResponseFromDomain(settings domain.GameRunSettings) gameSettingsResponse {
	return gameSettingsResponse{
		GameRunID:                    settings.GameRunID,
		MarkingMode:                  settings.MarkingMode,
		AllowPlayerMarkingModeChoice: settings.AllowPlayerMarkingModeChoice,
		ShowClaimReadiness:           settings.ShowClaimReadiness,
		VoiceClaimMode:               settings.VoiceClaimMode,
		VoiceClaimAutoplay:           settings.VoiceClaimAutoplay,
		CallerMode:                   settings.CallerMode,
		ThemeMode:                    settings.ThemeMode,
		ThemeID:                      settings.ThemeID,
		CreatedAt:                    settings.CreatedAt,
		UpdatedAt:                    settings.UpdatedAt,
	}
}

func playerPreferencesResponseFromDomain(preferences domain.PlayerPreferences) playerPreferencesResponse {
	return playerPreferencesResponse{
		PlayerID:    preferences.PlayerID,
		MarkingMode: preferences.MarkingMode,
		CreatedAt:   preferences.CreatedAt,
		UpdatedAt:   preferences.UpdatedAt,
	}
}

func gameContentResponseFromDomain(content domain.GeneratedGameContent) gameContentResponse {
	return gameContentResponse{
		ID:                   content.ID,
		GameRunID:            content.GameRunID,
		GenerationJobID:      content.GenerationJobID,
		Status:               content.Status,
		Topic:                content.Topic,
		Summary:              content.Summary,
		Words:                content.CurrentWords,
		GeneratedWords:       content.GeneratedWords,
		CallerStyle:          content.CallerStyle,
		ThemePrompt:          content.ThemePrompt,
		ReviewWindowOpensAt:  content.ReviewWindowOpensAt,
		ReviewWindowClosesAt: content.ReviewWindowClosesAt,
		LockedAt:             content.LockedAt,
		LockedWordSetID:      content.LockedWordSetID,
		GenerationProvider:   content.GenerationProvider,
		GenerationError:      content.GenerationError,
		CreatedAt:            content.CreatedAt,
		UpdatedAt:            content.UpdatedAt,
	}
}

func claimReadinessResponseFromDomain(readiness domain.ClaimReadiness) claimReadinessResponse {
	matchedCells := make([]cardCellResponse, 0, len(readiness.MatchedCells))
	for _, cell := range readiness.MatchedCells {
		matchedCells = append(matchedCells, cardCellResponseFromDomain(cell))
	}
	missingCells := make([]cardCellResponse, 0, len(readiness.MissingCells))
	for _, cell := range readiness.MissingCells {
		missingCells = append(missingCells, cardCellResponseFromDomain(cell))
	}

	return claimReadinessResponse{
		Ready:             readiness.Ready,
		SupportedPatterns: readiness.SupportedPatterns,
		ReadyPatterns:     readiness.ReadyPatterns,
		BestPattern:       readiness.BestPattern,
		MatchedCells:      matchedCells,
		MissingCells:      missingCells,
		Reason:            readiness.Reason,
	}
}

func allowedPlayerResponseFromDomain(player domain.AllowedPlayer) allowedPlayerResponse {
	return allowedPlayerResponse{
		ID:          player.ID,
		GameRunID:   player.GameRunID,
		Email:       player.Email,
		DisplayName: player.DisplayName,
		Source:      player.Source,
		CreatedAt:   player.CreatedAt,
	}
}

func playerResponseFromDomain(player domain.Player) playerResponse {
	return playerResponse{
		ID:              player.ID,
		GameRunID:       player.GameRunID,
		UserID:          player.UserID,
		Email:           player.Email,
		DisplayName:     player.DisplayName,
		Icon:            player.Icon,
		AvatarColor:     player.AvatarColor,
		AvatarLabel:     player.AvatarLabel,
		ConnectionState: player.ConnectionState,
		State:           player.State,
		JoinedAt:        player.JoinedAt,
		LastSeenAt:      player.LastSeenAt,
	}
}

func playerConnectionResponseFromDomain(update game.PlayerConnectionUpdate) playerResponse {
	response := playerResponseFromDomain(update.Player)
	response.ReconnectNotice = reconnectNoticeResponseFromDomain(update.ReconnectNotice)
	return response
}

func cardCellResponseFromDomain(cell domain.BingoCardCell) cardCellResponse {
	return cardCellResponse{
		ID:          cell.ID,
		RowIndex:    cell.RowIndex,
		ColIndex:    cell.ColIndex,
		Word:        cell.Word,
		IsFreeSpace: cell.IsFreeSpace,
		MarkedAt:    cell.MarkedAt,
	}
}

func cardResponseFromDomain(card domain.BingoCard) cardResponse {
	cells := make([]cardCellResponse, 0, len(card.Cells))
	for _, cell := range card.Cells {
		cells = append(cells, cardCellResponseFromDomain(cell))
	}

	return cardResponse{
		ID:        card.ID,
		GameRunID: card.GameRunID,
		PlayerID:  card.PlayerID,
		Seed:      card.Seed,
		Cells:     cells,
		CreatedAt: card.CreatedAt,
	}
}

func calledWordResponseFromDomain(calledWord domain.CalledWord) calledWordResponse {
	response := calledWordResponse{
		ID:             calledWord.ID,
		GameRunID:      calledWord.GameRunID,
		WordSetWordID:  calledWord.WordSetWordID,
		Word:           calledWord.Word,
		CalledByUserID: calledWord.CalledByUserID,
		Sequence:       calledWord.Sequence,
		CalledAt:       calledWord.CalledAt,
		CreatedAt:      calledWord.CreatedAt,
	}
	if calledWord.CallerAsset != nil {
		asset := callerAssetResponseFromDomain(*calledWord.CallerAsset)
		response.CallerAsset = &asset
	}
	return response
}

func callerAssetResponseFromDomain(asset domain.CallerAsset) callerAssetResponse {
	return callerAssetResponse{
		ID:             asset.ID,
		GameRunID:      asset.GameRunID,
		CallDeckItemID: asset.CallDeckItemID,
		Word:           asset.Word,
		Sequence:       asset.Sequence,
		Line:           asset.Line,
		AudioURL:       asset.AudioURL,
		StorageKey:     asset.StorageKey,
		VoiceName:      asset.VoiceName,
		Provider:       asset.Provider,
		Status:         asset.Status,
		ErrorReason:    asset.ErrorReason,
	}
}

func deliveryAttemptsResponseFromDomain(attempts []domain.DeliveryAttempt) []deliveryAttemptResponse {
	response := make([]deliveryAttemptResponse, 0, len(attempts))
	for _, attempt := range attempts {
		response = append(response, deliveryAttemptResponseFromDomain(attempt))
	}
	return response
}

func deliveryAttemptResponseFromDomain(attempt domain.DeliveryAttempt) deliveryAttemptResponse {
	return deliveryAttemptResponse{
		ID:             attempt.ID,
		GameRunID:      attempt.GameRunID,
		Channel:        attempt.Channel,
		Purpose:        attempt.Purpose,
		RecipientEmail: attempt.RecipientEmail,
		Subject:        attempt.Subject,
		TemplateKey:    attempt.TemplateKey,
		BodyPreview:    attempt.BodyPreview,
		LinkURL:        attempt.LinkURL,
		GameCode:       attempt.GameCode,
		Status:         attempt.Status,
		ErrorReason:    attempt.ErrorReason,
		SentAt:         attempt.SentAt,
		CreatedAt:      attempt.CreatedAt,
	}
}

func themeResponseFromDomain(theme domain.Theme) themeResponse {
	return themeResponse{
		ID:        theme.ID,
		GameRunID: theme.GameRunID,
		Name:      theme.Name,
		Summary:   theme.Summary,
		Tokens:    theme.Tokens,
		Status:    theme.Status,
		Provider:  theme.Provider,
		CreatedAt: theme.CreatedAt,
		UpdatedAt: theme.UpdatedAt,
	}
}

func claimResponseFromDomain(claim domain.BingoClaim) claimResponse {
	validationResult := claim.ValidationResult
	if len(validationResult) == 0 {
		validationResult = json.RawMessage(`{}`)
	}

	return claimResponse{
		ID:               claim.ID,
		GameRunID:        claim.GameRunID,
		PlayerID:         claim.PlayerID,
		Pattern:          claim.Pattern,
		Status:           claim.Status,
		ValidationResult: validationResult,
		ClaimedAt:        claim.ClaimedAt,
		ReviewedByUserID: claim.ReviewedByUserID,
		ReviewedAt:       claim.ReviewedAt,
		CreatedAt:        claim.CreatedAt,
		UpdatedAt:        claim.UpdatedAt,
	}
}

func claimSubmissionResponseFromDomain(result game.BingoClaimResult) claimSubmissionResponse {
	response := claimSubmissionResponse{
		Claim: claimResponseFromDomain(result.Claim),
	}
	if result.Winner != nil {
		winner := winnerResponseFromDomain(*result.Winner)
		response.Winner = &winner
	}

	return response
}

func winnerResponseFromDomain(winner domain.Winner) winnerResponse {
	return winnerResponse{
		ID:          winner.ID,
		GameRunID:   winner.GameRunID,
		PlayerID:    winner.PlayerID,
		ClaimID:     winner.ClaimID,
		Placement:   winner.Placement,
		Pattern:     winner.Pattern,
		ConfirmedAt: winner.ConfirmedAt,
		CreatedAt:   winner.CreatedAt,
	}
}

func gameSummaryResponseFromDomain(summary domain.GameSummary) gameSummaryResponse {
	claims := make([]claimResponse, 0, len(summary.Claims))
	for _, claim := range summary.Claims {
		claims = append(claims, claimResponseFromDomain(claim))
	}

	winners := make([]winnerResponse, 0, len(summary.Winners))
	for _, winner := range summary.Winners {
		winners = append(winners, winnerResponseFromDomain(winner))
	}
	players := make([]playerResponse, 0, len(summary.Players))
	for _, player := range summary.Players {
		players = append(players, playerResponseFromDomain(player))
	}
	calledWords := make([]calledWordResponse, 0, len(summary.CalledWords))
	for _, word := range summary.CalledWords {
		calledWords = append(calledWords, calledWordResponseFromDomain(word))
	}

	var currentWord *calledWordResponse
	if summary.CurrentWord != nil {
		word := calledWordResponseFromDomain(*summary.CurrentWord)
		currentWord = &word
	}

	return gameSummaryResponse{
		GameRun:         gameRunResponseFromDomain(game.GameRunWithCounts{GameRun: summary.GameRun}),
		PlayerCount:     summary.PlayerCount,
		CalledWordCount: summary.CalledWordCount,
		CurrentWord:     currentWord,
		Claims:          claims,
		Winners:         winners,
		Players:         players,
		CalledWords:     calledWords,
		Status:          summary.Status,
	}
}

func hostSnapshotResponseFromDomain(snapshot domain.HostSnapshot) hostSnapshotResponse {
	players := make([]playerResponse, 0, len(snapshot.Players))
	for _, player := range snapshot.Players {
		players = append(players, playerResponseFromDomain(player))
	}
	calledWords := make([]calledWordResponse, 0, len(snapshot.CalledWords))
	for _, word := range snapshot.CalledWords {
		calledWords = append(calledWords, calledWordResponseFromDomain(word))
	}
	claims := make([]claimResponse, 0, len(snapshot.Claims))
	for _, claim := range snapshot.Claims {
		claims = append(claims, claimResponseFromDomain(claim))
	}
	winners := make([]winnerResponse, 0, len(snapshot.Winners))
	for _, winner := range snapshot.Winners {
		winners = append(winners, winnerResponseFromDomain(winner))
	}

	var currentWord *calledWordResponse
	if snapshot.CurrentWord != nil {
		word := calledWordResponseFromDomain(*snapshot.CurrentWord)
		currentWord = &word
	}
	var currentCallerAsset *callerAssetResponse
	if snapshot.CurrentCallerAsset != nil {
		asset := callerAssetResponseFromDomain(*snapshot.CurrentCallerAsset)
		currentCallerAsset = &asset
	}
	var appliedTheme *themeResponse
	if snapshot.AppliedTheme != nil {
		theme := themeResponseFromDomain(*snapshot.AppliedTheme)
		appliedTheme = &theme
	}

	return hostSnapshotResponse{
		GameRun:            gameRunResponseFromDomain(game.GameRunWithCounts{GameRun: snapshot.GameRun}),
		Settings:           gameSettingsResponseFromDomain(snapshot.Settings),
		Status:             snapshot.Status,
		CurrentWord:        currentWord,
		CurrentCallerAsset: currentCallerAsset,
		AppliedTheme:       appliedTheme,
		WinningPattern:     snapshot.Pattern,
		PlayerCount:        snapshot.PlayerCount,
		Players:            players,
		CalledWords:        calledWords,
		Claims:             claims,
		Winners:            winners,
	}
}

func playerSnapshotResponseFromDomain(snapshot domain.PlayerSnapshot) playerSnapshotResponse {
	calledWords := make([]calledWordResponse, 0, len(snapshot.CalledWords))
	for _, word := range snapshot.CalledWords {
		calledWords = append(calledWords, calledWordResponseFromDomain(word))
	}
	claims := make([]claimResponse, 0, len(snapshot.Claims))
	for _, claim := range snapshot.Claims {
		claims = append(claims, claimResponseFromDomain(claim))
	}
	winners := make([]winnerResponse, 0, len(snapshot.Winners))
	for _, winner := range snapshot.Winners {
		winners = append(winners, winnerResponseFromDomain(winner))
	}

	var currentWord *calledWordResponse
	if snapshot.CurrentWord != nil {
		word := calledWordResponseFromDomain(*snapshot.CurrentWord)
		currentWord = &word
	}
	var currentCallerAsset *callerAssetResponse
	if snapshot.CurrentCallerAsset != nil {
		asset := callerAssetResponseFromDomain(*snapshot.CurrentCallerAsset)
		currentCallerAsset = &asset
	}
	var appliedTheme *themeResponse
	if snapshot.AppliedTheme != nil {
		theme := themeResponseFromDomain(*snapshot.AppliedTheme)
		appliedTheme = &theme
	}
	var card *cardResponse
	if snapshot.Card != nil {
		response := cardResponseFromDomain(*snapshot.Card)
		card = &response
	}

	return playerSnapshotResponse{
		GameRun:                      gameRunResponseFromDomain(game.GameRunWithCounts{GameRun: snapshot.GameRun}),
		MarkingMode:                  snapshot.MarkingMode,
		AllowPlayerMarkingModeChoice: snapshot.Settings.AllowPlayerMarkingModeChoice,
		ShowClaimReadiness:           snapshot.Settings.ShowClaimReadiness,
		Status:                       snapshot.Status,
		CurrentWord:                  currentWord,
		CurrentCallerAsset:           currentCallerAsset,
		AppliedTheme:                 appliedTheme,
		WinningPattern:               snapshot.Pattern,
		Player:                       playerResponseFromDomain(snapshot.Player),
		Card:                         card,
		CalledWords:                  calledWords,
		Claims:                       claims,
		Winners:                      winners,
		ReconnectNotice:              reconnectNoticeResponseFromDomain(snapshot.ReconnectNotice),
	}
}

func reconnectNoticeResponseFromDomain(notice *domain.ReconnectNotice) *reconnectNoticeResponse {
	if notice == nil {
		return nil
	}

	missed := make([]calledWordResponse, 0, len(notice.MissedCalledWords))
	for _, word := range notice.MissedCalledWords {
		missed = append(missed, calledWordResponseFromDomain(word))
	}

	return &reconnectNoticeResponse{
		LastSeenAt:        notice.LastSeenAt,
		MissedCalledWords: missed,
	}
}

func eventResponseFromDomain(event domain.GameEvent) eventResponse {
	payload := event.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}

	return eventResponse{
		ID:        event.ID,
		GameRunID: event.GameRunID,
		Type:      event.Type,
		EntityID:  event.EntityID,
		Payload:   payload,
		Sequence:  event.Sequence,
		CreatedAt: event.CreatedAt,
	}
}
