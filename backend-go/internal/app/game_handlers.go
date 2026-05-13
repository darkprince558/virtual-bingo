package app

import (
	"net/http"
	"time"

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

type gameRunResponse struct {
	ID                 string     `json:"id"`
	TemplateID         *string    `json:"templateId,omitempty"`
	HostUserID         string     `json:"hostUserId"`
	WordSetID          *string    `json:"wordSetId,omitempty"`
	Code               string     `json:"code"`
	Name               string     `json:"name"`
	Status             string     `json:"status"`
	ScheduledStartAt   *time.Time `json:"scheduledStartAt,omitempty"`
	StartedAt          *time.Time `json:"startedAt,omitempty"`
	EndedAt            *time.Time `json:"endedAt,omitempty"`
	WinningPattern     *string    `json:"winningPattern,omitempty"`
	AllowedPlayerCount int        `json:"allowedPlayerCount"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
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

func gameRunResponseFromDomain(run game.GameRunWithCounts) gameRunResponse {
	return gameRunResponse{
		ID:                 run.GameRun.ID,
		TemplateID:         run.GameRun.TemplateID,
		HostUserID:         run.GameRun.HostUserID,
		WordSetID:          run.GameRun.WordSetID,
		Code:               run.GameRun.Code,
		Name:               run.GameRun.Name,
		Status:             run.GameRun.Status,
		ScheduledStartAt:   run.GameRun.ScheduledStartAt,
		StartedAt:          run.GameRun.StartedAt,
		EndedAt:            run.GameRun.EndedAt,
		WinningPattern:     run.GameRun.WinningPattern,
		AllowedPlayerCount: run.AllowedPlayerCount,
		CreatedAt:          run.GameRun.CreatedAt,
		UpdatedAt:          run.GameRun.UpdatedAt,
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

func cardResponseFromDomain(card domain.BingoCard) cardResponse {
	cells := make([]cardCellResponse, 0, len(card.Cells))
	for _, cell := range card.Cells {
		cells = append(cells, cardCellResponse{
			ID:          cell.ID,
			RowIndex:    cell.RowIndex,
			ColIndex:    cell.ColIndex,
			Word:        cell.Word,
			IsFreeSpace: cell.IsFreeSpace,
			MarkedAt:    cell.MarkedAt,
		})
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
