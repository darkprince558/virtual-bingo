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

func (s *Server) joinPlayer(w http.ResponseWriter, r *http.Request) {
	if !requireDatabase(w, s.service) {
		return
	}

	var req joinPlayerRequest
	if err := decodeJSON(r, &req); err != nil {
		writeAPIError(w, http.StatusBadRequest, "validation_error", "request body must be valid JSON")
		return
	}

	player, err := s.service.JoinPlayer(r.Context(), game.JoinPlayerInput{
		GameRunID:   r.PathValue("gameID"),
		Email:       req.Email,
		DisplayName: req.DisplayName,
	})
	if err != nil {
		mapServiceError(w, err)
		return
	}

	writeData(w, http.StatusCreated, playerResponseFromDomain(player))
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

type allowedPlayerResponse struct {
	ID          string    `json:"id"`
	GameRunID   string    `json:"gameRunId"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	Source      string    `json:"source"`
	CreatedAt   time.Time `json:"createdAt"`
}

type playerResponse struct {
	ID              string    `json:"id"`
	GameRunID       string    `json:"gameRunId"`
	UserID          *string   `json:"userId,omitempty"`
	Email           string    `json:"email"`
	DisplayName     string    `json:"displayName"`
	ConnectionState string    `json:"connectionState"`
	State           string    `json:"state"`
	JoinedAt        time.Time `json:"joinedAt"`
	LastSeenAt      time.Time `json:"lastSeenAt"`
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
	ID             string    `json:"id"`
	GameRunID      string    `json:"gameRunId"`
	WordSetWordID  *string   `json:"wordSetWordId,omitempty"`
	Word           string    `json:"word"`
	CalledByUserID *string   `json:"calledByUserId,omitempty"`
	Sequence       int       `json:"sequence"`
	CalledAt       time.Time `json:"calledAt"`
	CreatedAt      time.Time `json:"createdAt"`
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
	GameRun        gameRunResponse      `json:"gameRun"`
	Status         string               `json:"status"`
	CurrentWord    *calledWordResponse  `json:"currentWord,omitempty"`
	WinningPattern string               `json:"winningPattern"`
	PlayerCount    int                  `json:"playerCount"`
	Players        []playerResponse     `json:"players"`
	CalledWords    []calledWordResponse `json:"calledWords"`
	Claims         []claimResponse      `json:"claims"`
	Winners        []winnerResponse     `json:"winners"`
}

type playerSnapshotResponse struct {
	GameRun        gameRunResponse      `json:"gameRun"`
	Status         string               `json:"status"`
	CurrentWord    *calledWordResponse  `json:"currentWord,omitempty"`
	WinningPattern string               `json:"winningPattern"`
	Player         playerResponse       `json:"player"`
	Card           *cardResponse        `json:"card,omitempty"`
	CalledWords    []calledWordResponse `json:"calledWords"`
	Claims         []claimResponse      `json:"claims"`
	Winners        []winnerResponse     `json:"winners"`
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
		ConnectionState: player.ConnectionState,
		State:           player.State,
		JoinedAt:        player.JoinedAt,
		LastSeenAt:      player.LastSeenAt,
	}
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
	return calledWordResponse{
		ID:             calledWord.ID,
		GameRunID:      calledWord.GameRunID,
		WordSetWordID:  calledWord.WordSetWordID,
		Word:           calledWord.Word,
		CalledByUserID: calledWord.CalledByUserID,
		Sequence:       calledWord.Sequence,
		CalledAt:       calledWord.CalledAt,
		CreatedAt:      calledWord.CreatedAt,
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

	return hostSnapshotResponse{
		GameRun:        gameRunResponseFromDomain(game.GameRunWithCounts{GameRun: snapshot.GameRun}),
		Status:         snapshot.Status,
		CurrentWord:    currentWord,
		WinningPattern: snapshot.Pattern,
		PlayerCount:    snapshot.PlayerCount,
		Players:        players,
		CalledWords:    calledWords,
		Claims:         claims,
		Winners:        winners,
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
	var card *cardResponse
	if snapshot.Card != nil {
		response := cardResponseFromDomain(*snapshot.Card)
		card = &response
	}

	return playerSnapshotResponse{
		GameRun:        gameRunResponseFromDomain(game.GameRunWithCounts{GameRun: snapshot.GameRun}),
		Status:         snapshot.Status,
		CurrentWord:    currentWord,
		WinningPattern: snapshot.Pattern,
		Player:         playerResponseFromDomain(snapshot.Player),
		Card:           card,
		CalledWords:    calledWords,
		Claims:         claims,
		Winners:        winners,
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
