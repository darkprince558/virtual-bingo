package app

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/darkprince558/virtual-bingo/backend-go/internal/domain"
	"github.com/darkprince558/virtual-bingo/backend-go/internal/game"
)

type updateGameRequest struct {
	Name             *string    `json:"name"`
	Code             *string    `json:"code"`
	WordSetID        *string    `json:"wordSetId"`
	ScheduledStartAt *time.Time `json:"scheduledStartAt"`
	WinningPattern   *string    `json:"winningPattern"`
}

type bulkAllowedPlayersRequest struct {
	Players []addAllowedPlayerRequest `json:"players"`
}

func (r *bulkAllowedPlayersRequest) UnmarshalJSON(data []byte) error {
	var rows []addAllowedPlayerRequest
	if err := json.Unmarshal(data, &rows); err == nil {
		r.Players = rows
		return nil
	}

	var wrapped struct {
		Players []addAllowedPlayerRequest `json:"players"`
	}
	if err := json.Unmarshal(data, &wrapped); err != nil {
		return err
	}
	r.Players = wrapped.Players
	return nil
}

type updateAllowedPlayerRequest struct {
	Email       *string `json:"email"`
	DisplayName *string `json:"displayName"`
}

type createWordSetRequest struct {
	Name   string           `json:"name"`
	Status string           `json:"status"`
	Source string           `json:"source"`
	Words  []wordSetWordReq `json:"words"`
}

type updateWordSetRequest struct {
	Name   *string `json:"name"`
	Status *string `json:"status"`
	Source *string `json:"source"`
}

type wordSetWordReq struct {
	Word      string `json:"word"`
	SortOrder *int   `json:"sortOrder"`
	IsActive  *bool  `json:"isActive"`
}

type updateWordSetWordRequest struct {
	Word      *string `json:"word"`
	SortOrder *int    `json:"sortOrder"`
	IsActive  *bool   `json:"isActive"`
}

func (s *Server) getCurrentUser(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	user, err := s.service.CurrentUser(r.Context(), principal)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, currentUserResponseFromDomain(user, s.authMode()))
}

func (s *Server) listGames(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}

	runs, err := s.service.ListGameRuns(r.Context(), principal, r.URL.Query().Get("scope"), r.URL.Query().Get("status"))
	if err != nil {
		mapServiceError(w, err)
		return
	}
	response := make([]gameRunSummaryResponse, 0, len(runs))
	for _, run := range runs {
		response = append(response, gameRunSummaryResponseFromDomain(run))
	}

	writeData(w, http.StatusOK, response)
}

func (s *Server) getGameByCode(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	run, err := s.service.GetGameRunByCode(r.Context(), r.PathValue("code"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, gameRunResponseFromDomain(run))
}

func (s *Server) updateGame(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	var req updateGameRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}

	run, err := s.service.UpdateGameRun(r.Context(), principal, game.UpdateGameRunInput{
		GameRunID:        r.PathValue("gameID"),
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

	writeData(w, http.StatusOK, gameRunResponseFromDomain(run))
}

func (s *Server) bulkAddAllowedPlayers(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	var req bulkAllowedPlayersRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}

	inputs := make([]game.AddAllowedPlayerInput, 0, len(req.Players))
	for _, player := range req.Players {
		inputs = append(inputs, game.AddAllowedPlayerInput{
			GameRunID:   r.PathValue("gameID"),
			Email:       player.Email,
			DisplayName: player.DisplayName,
		})
	}
	players, err := s.service.BulkAddAllowedPlayers(r.Context(), principal, r.PathValue("gameID"), inputs)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	response := make([]allowedPlayerResponse, 0, len(players))
	for _, player := range players {
		response = append(response, allowedPlayerResponseFromDomain(player))
	}

	writeData(w, http.StatusCreated, response)
}

func (s *Server) updateAllowedPlayer(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	var req updateAllowedPlayerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}

	player, err := s.service.UpdateAllowedPlayer(r.Context(), principal, game.UpdateAllowedPlayerInput{
		GameRunID:       r.PathValue("gameID"),
		AllowedPlayerID: r.PathValue("allowedPlayerID"),
		Email:           req.Email,
		DisplayName:     req.DisplayName,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, allowedPlayerResponseFromDomain(player))
}

func (s *Server) deleteAllowedPlayer(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	if err := s.service.DeleteAllowedPlayer(r.Context(), principal, r.PathValue("gameID"), r.PathValue("allowedPlayerID")); err != nil {
		mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listWordSets(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	wordSets, err := s.service.ListWordSets(r.Context(), principal)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	response := make([]wordSetResponse, 0, len(wordSets))
	for _, wordSet := range wordSets {
		response = append(response, wordSetResponseFromDomain(wordSet))
	}

	writeData(w, http.StatusOK, response)
}

func (s *Server) createWordSet(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	var req createWordSetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}
	wordSet, err := s.service.CreateWordSet(r.Context(), principal, game.CreateWordSetInput{
		Name:   req.Name,
		Status: req.Status,
		Source: req.Source,
		Words:  wordSetWordInputs(req.Words),
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusCreated, wordSetResponseFromDomain(wordSet))
}

func (s *Server) getWordSet(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	wordSet, err := s.service.GetWordSet(r.Context(), principal, r.PathValue("wordSetID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, wordSetResponseFromDomain(wordSet))
}

func (s *Server) updateWordSet(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	var req updateWordSetRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}
	wordSet, err := s.service.UpdateWordSet(r.Context(), principal, game.UpdateWordSetInput{
		WordSetID: r.PathValue("wordSetID"),
		Name:      req.Name,
		Status:    req.Status,
		Source:    req.Source,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, wordSetResponseFromDomain(wordSet))
}

func (s *Server) createWordSetWord(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	var req wordSetWordReq
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}
	word, err := s.service.CreateWordSetWord(r.Context(), principal, r.PathValue("wordSetID"), wordSetWordInput(req))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusCreated, wordSetWordResponseFromDomain(word))
}

func (s *Server) updateWordSetWord(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	var req updateWordSetWordRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}
	word, err := s.service.UpdateWordSetWord(r.Context(), principal, game.UpdateWordSetWordInput{
		WordSetID: r.PathValue("wordSetID"),
		WordID:    r.PathValue("wordID"),
		Word:      req.Word,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, wordSetWordResponseFromDomain(word))
}

func (s *Server) deleteWordSetWord(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	if err := s.service.DeleteWordSetWord(r.Context(), principal, r.PathValue("wordSetID"), r.PathValue("wordID")); err != nil {
		mapServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) getCurrentPlayerSnapshot(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	snapshot, err := s.service.GetCurrentPlayerSnapshot(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, playerSnapshotResponseFromDomain(snapshot))
}

func (s *Server) assignCurrentPlayerCard(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	card, err := s.service.AssignCurrentPlayerCard(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusCreated, cardResponseFromDomain(card))
}

func (s *Server) getCurrentPlayerCard(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	card, err := s.service.GetCurrentPlayerCard(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, cardResponseFromDomain(card))
}

func (s *Server) markCurrentPlayerCard(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	var req markCardCellRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}
	cell, err := s.service.MarkCurrentPlayerCardCell(r.Context(), principal, game.MarkCardCellInput{
		GameRunID: r.PathValue("gameID"),
		CellID:    r.PathValue("cellID"),
		Marked:    req.Marked,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, cardCellResponseFromDomain(cell))
}

func (s *Server) heartbeatCurrentPlayer(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	update, err := s.service.HeartbeatCurrentPlayer(r.Context(), principal, r.PathValue("gameID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, playerConnectionResponseFromDomain(update))
}

func (s *Server) getBingoClaim(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	claim, err := s.service.GetBingoClaim(r.Context(), principal, r.PathValue("gameID"), r.PathValue("claimID"))
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusOK, claimResponseFromDomain(claim))
}

func (s *Server) listActivityEvents(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}
	principal, err := s.service.Authenticate(r)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	limit := 100
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			limit = parsed
		}
	}
	events, err := s.service.ListActivityEvents(r.Context(), principal, r.PathValue("gameID"), limit)
	if err != nil {
		mapServiceError(w, err)
		return
	}
	response := make([]activityEventResponse, 0, len(events))
	for _, event := range events {
		response = append(response, activityEventResponseFromDomain(event))
	}

	writeData(w, http.StatusOK, response)
}

type currentUserResponse struct {
	ID              string  `json:"id"`
	ExternalSubject *string `json:"externalSubject,omitempty"`
	Email           string  `json:"email"`
	DisplayName     string  `json:"displayName"`
	Role            string  `json:"role"`
	AuthMode        string  `json:"authMode"`
}

type gameRunSummaryResponse struct {
	gameRunResponse
	PlayerCount int `json:"playerCount"`
}

type wordSetResponse struct {
	ID        string                `json:"id"`
	Name      string                `json:"name"`
	Status    string                `json:"status"`
	Source    string                `json:"source"`
	Words     []wordSetWordResponse `json:"words,omitempty"`
	CreatedAt time.Time             `json:"createdAt"`
	UpdatedAt time.Time             `json:"updatedAt"`
}

type wordSetWordResponse struct {
	ID        string    `json:"id"`
	WordSetID string    `json:"wordSetId"`
	Word      string    `json:"word"`
	SortOrder int       `json:"sortOrder"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

type activityEventResponse struct {
	ID          string          `json:"id"`
	GameRunID   string          `json:"gameRunId"`
	Type        string          `json:"type"`
	EntityType  *string         `json:"entityType,omitempty"`
	EntityID    *string         `json:"entityId,omitempty"`
	ActorUserID *string         `json:"actorUserId,omitempty"`
	Payload     json.RawMessage `json:"payload"`
	Sequence    *int64          `json:"sequence,omitempty"`
	CreatedAt   time.Time       `json:"createdAt"`
}

func currentUserResponseFromDomain(user domain.User, authMode string) currentUserResponse {
	return currentUserResponse{
		ID:              user.ID,
		ExternalSubject: user.ExternalSubject,
		Email:           user.Email,
		DisplayName:     user.DisplayName,
		Role:            user.Role,
		AuthMode:        authMode,
	}
}

func gameRunSummaryResponseFromDomain(summary domain.GameRunSummary) gameRunSummaryResponse {
	return gameRunSummaryResponse{
		gameRunResponse: gameRunResponseFromDomain(game.GameRunWithCounts{
			GameRun:            summary.GameRun,
			AllowedPlayerCount: summary.AllowedPlayerCount,
		}),
		PlayerCount: summary.PlayerCount,
	}
}

func wordSetResponseFromDomain(wordSet domain.WordSetWithWords) wordSetResponse {
	words := make([]wordSetWordResponse, 0, len(wordSet.Words))
	for _, word := range wordSet.Words {
		words = append(words, wordSetWordResponseFromDomain(word))
	}
	return wordSetResponse{
		ID:        wordSet.WordSet.ID,
		Name:      wordSet.WordSet.Name,
		Status:    wordSet.WordSet.Status,
		Source:    wordSet.WordSet.Source,
		Words:     words,
		CreatedAt: wordSet.WordSet.CreatedAt,
		UpdatedAt: wordSet.WordSet.UpdatedAt,
	}
}

func wordSetWordResponseFromDomain(word domain.WordSetWord) wordSetWordResponse {
	return wordSetWordResponse{
		ID:        word.ID,
		WordSetID: word.WordSetID,
		Word:      word.Word,
		SortOrder: word.SortOrder,
		IsActive:  word.IsActive,
		CreatedAt: word.CreatedAt,
	}
}

func activityEventResponseFromDomain(event domain.ActivityEvent) activityEventResponse {
	payload := event.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	return activityEventResponse{
		ID:          event.ID,
		GameRunID:   event.GameRunID,
		Type:        event.Type,
		EntityType:  event.EntityType,
		EntityID:    event.EntityID,
		ActorUserID: event.ActorUserID,
		Payload:     payload,
		Sequence:    event.Sequence,
		CreatedAt:   event.CreatedAt,
	}
}

func wordSetWordInputs(requests []wordSetWordReq) []game.WordSetWordInput {
	inputs := make([]game.WordSetWordInput, 0, len(requests))
	for _, req := range requests {
		inputs = append(inputs, wordSetWordInput(req))
	}
	return inputs
}

func wordSetWordInput(req wordSetWordReq) game.WordSetWordInput {
	return game.WordSetWordInput{
		Word:      req.Word,
		SortOrder: req.SortOrder,
		IsActive:  req.IsActive,
	}
}
