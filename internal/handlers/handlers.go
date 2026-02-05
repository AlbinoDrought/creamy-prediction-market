package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.albinodrought.com/creamy-prediction-market/internal/repo"
	"go.albinodrought.com/creamy-prediction-market/internal/types"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	Store          *repo.Store
	Logger         *logrus.Logger
	StartingTokens int64
}

func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (h *Handler) jsonResponse(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h *Handler) errorResponse(w http.ResponseWriter, status int, message string) {
	h.jsonResponse(w, status, map[string]string{"error": message})
}

// Auth middleware

func (h *Handler) getAuthenticatedUser(r *http.Request) (types.User, bool) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return types.User{}, false
	}

	// Strip "Bearer " prefix if present
	if len(token) > 7 && token[:7] == "Bearer " {
		token = token[7:]
	}

	userID, ok := h.Store.GetUserIDBySession(token)
	if !ok {
		return types.User{}, false
	}

	user, err := h.Store.GetUser(userID)
	if err != nil {
		return types.User{}, false
	}

	return user, true
}

func (h *Handler) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := h.getAuthenticatedUser(r); !ok {
			h.errorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next(w, r)
	}
}

func (h *Handler) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := h.getAuthenticatedUser(r)
		if !ok {
			h.errorResponse(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		if !user.Admin {
			h.errorResponse(w, http.StatusForbidden, "admin required")
			return
		}
		next(w, r)
	}
}

// Public endpoints

func (h *Handler) ListPredictions(w http.ResponseWriter, r *http.Request) {
	results := h.Store.ListPredictionsWithOdds()

	sort.Slice(results, func(i, j int) bool {
		return results[i].Prediction.CreatedAt > results[j].Prediction.CreatedAt
	})

	h.jsonResponse(w, http.StatusOK, results)
}

func (h *Handler) GetPrediction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	result, err := h.Store.GetPredictionWithOdds(id)
	if err != nil {
		h.errorResponse(w, http.StatusNotFound, "prediction not found")
		return
	}

	h.jsonResponse(w, http.StatusOK, result)
}

func (h *Handler) ShowLeaderboard(w http.ResponseWriter, r *http.Request) {
	users := h.Store.ListUsers()

	leaderboard := make([]types.LeaderboardUser, 0, len(users))
	for _, u := range users {
		if u.Admin {
			continue // exclude admins from leaderboard
		}
		leaderboard = append(leaderboard, types.LeaderboardUser{
			ID:     u.ID,
			Name:   u.Name,
			Tokens: u.Tokens,
		})
	}

	// Sort by tokens descending
	sort.Slice(leaderboard, func(i, j int) bool {
		if leaderboard[i].Tokens != leaderboard[j].Tokens {
			return leaderboard[i].Tokens > leaderboard[j].Tokens
		}
		return strings.Compare(leaderboard[i].ID, leaderboard[j].ID) == -1
	})

	// Assign ranks
	for i := range leaderboard {
		leaderboard[i].Rank = i + 1
	}

	h.jsonResponse(w, http.StatusOK, leaderboard)
}

// Guest endpoints

type RegisterRequest struct {
	Name string `json:"name"`
	PIN  string `json:"pin"`
}

type AuthResponse struct {
	Token string     `json:"token"`
	User  types.User `json:"user"`
}

var regexValidUsername = regexp.MustCompile("^[A-Za-z0-9]+$")

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		h.errorResponse(w, http.StatusBadRequest, "name is required")
		return
	}
	if len(req.Name) > 20 {
		h.errorResponse(w, http.StatusBadRequest, "name must be less than 20 characters")
		return
	}
	if !regexValidUsername.MatchString(req.Name) {
		h.errorResponse(w, http.StatusBadRequest, "name must only contain A-Z, a-z, 0-9")
		return
	}

	if req.PIN == "" {
		h.errorResponse(w, http.StatusBadRequest, "pin is required")
		return
	}

	pinHash, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.MinCost)
	if err != nil {
		h.Logger.WithError(err).Error("failed to hash pin")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	userID, err := repo.NewID()
	if err != nil {
		h.Logger.WithError(err).Error("failed to generate user ID")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	user := types.User{
		ID:      userID,
		Name:    req.Name,
		PINHash: pinHash,
		Admin:   false,
		Tokens:  0,
	}

	if err := h.Store.AddUser(user, h.StartingTokens); err != nil {
		if err == repo.ErrUserNameTaken {
			h.errorResponse(w, http.StatusConflict, "name is already taken")
			return
		}
		h.Logger.WithError(err).Error("failed to add user")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	sessionToken, err := generateSessionToken()
	if err != nil {
		h.Logger.WithError(err).Error("failed to generate session token")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.Store.CreateSession(sessionToken, user.ID)

	user2, err := h.Store.GetUser(user.ID)
	if err == nil {
		user = user2 // fetch latest tokens. If err, tokens in response is stale.
	}

	h.jsonResponse(w, http.StatusCreated, AuthResponse{
		Token: sessionToken,
		User:  user,
	})
}

type LoginRequest struct {
	Name string `json:"name"`
	PIN  string `json:"pin"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.Store.GetUserByName(req.Name)
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "invalid name or pin")
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.PINHash, []byte(req.PIN)); err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "invalid name or pin")
		return
	}

	sessionToken, err := generateSessionToken()
	if err != nil {
		h.Logger.WithError(err).Error("failed to generate session token")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.Store.CreateSession(sessionToken, user.ID)

	h.jsonResponse(w, http.StatusOK, AuthResponse{
		Token: sessionToken,
		User:  user,
	})
}

// User endpoints

func (h *Handler) GetMe(w http.ResponseWriter, r *http.Request) {
	user, _ := h.getAuthenticatedUser(r)
	h.jsonResponse(w, http.StatusOK, user)
}

func (h *Handler) GetMyBets(w http.ResponseWriter, r *http.Request) {
	user, _ := h.getAuthenticatedUser(r)
	bets := h.Store.ListBetsByUser(user.ID)

	// Sort by created_at descending
	sort.Slice(bets, func(i, j int) bool {
		return bets[i].CreatedAt > bets[j].CreatedAt
	})

	h.jsonResponse(w, http.StatusOK, bets)
}

type PlaceBetRequest struct {
	PredictionID       string `json:"prediction_id"`
	PredictionChoiceID string `json:"prediction_choice_id"`
	Amount             int64  `json:"amount"`
}

func (h *Handler) PlaceBet(w http.ResponseWriter, r *http.Request) {
	user, _ := h.getAuthenticatedUser(r)

	var req PlaceBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	betID, err := repo.NewID()
	if err != nil {
		h.Logger.WithError(err).Error("failed to generate bet ID")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	bet := types.Bet{
		ID:                 betID,
		CreatedAt:          time.Now().Format(time.RFC3339),
		UserID:             user.ID,
		PredictionID:       req.PredictionID,
		PredictionChoiceID: req.PredictionChoiceID,
		Amount:             req.Amount,
		Status:             types.BetStatusPlaced,
	}

	err = h.Store.CreateBet(bet)
	if err == repo.ErrBetAmountMustBePositive {
		h.errorResponse(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	if err == repo.ErrBetAlreadyExistsForPrediction {
		h.errorResponse(w, http.StatusConflict, "you already have a bet on this prediction")
		return
	}
	if err == repo.ErrPredictionNotOpen {
		h.errorResponse(w, http.StatusBadRequest, "prediction is not open for betting")
		return
	}
	if err == repo.ErrPredictionChoiceNotFound {
		h.errorResponse(w, http.StatusBadRequest, "invalid choice")
		return
	}
	if err == repo.ErrTokensWouldBeNegative {
		h.errorResponse(w, http.StatusBadRequest, "insufficient tokens")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to place bet")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.jsonResponse(w, http.StatusCreated, bet)
}

type IncreaseBetRequest struct {
	Amount int64 `json:"amount"`
}

func (h *Handler) IncreaseBetAmount(w http.ResponseWriter, r *http.Request) {
	user, _ := h.getAuthenticatedUser(r)
	betID := r.PathValue("id")

	var req IncreaseBetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	bet, err := h.Store.GetBet(betID)
	if err != nil {
		h.errorResponse(w, http.StatusNotFound, "bet not found")
		return
	}

	if bet.UserID != user.ID {
		h.errorResponse(w, http.StatusForbidden, "not your bet")
		return
	}

	err = h.Store.IncreaseBet(bet.ID, req.Amount)
	if err == repo.ErrBetNotActive {
		h.errorResponse(w, http.StatusBadRequest, "bet is not active")
		return
	}
	if err == repo.ErrPredictionNotFound {
		h.errorResponse(w, http.StatusInternalServerError, "prediction not found")
		return
	}
	if err == repo.ErrPredictionNotOpen {
		h.errorResponse(w, http.StatusInternalServerError, "prediction not found")
		return
	}
	if err == repo.ErrBetAlreadyHigher {
		h.errorResponse(w, http.StatusConflict, "active bet is already higher than specified amount")
		return
	}
	if err == repo.ErrTokensWouldBeNegative {
		h.errorResponse(w, http.StatusBadRequest, "insufficient tokens")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to updateplace bet")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	bet2, err := h.Store.GetBet(betID)
	if err == nil {
		bet = bet2 // if err, bet amount is stale
	}

	h.jsonResponse(w, http.StatusOK, bet)
}

// Admin endpoints

type CreatePredictionRequest struct {
	Name                 string                   `json:"name"`
	Description          string                   `json:"description"`
	ClosesAt             string                   `json:"closes_at"`
	Choices              []types.PredictionChoice `json:"choices"`
	OddsVisibleBeforeBet bool                     `json:"odds_visible_before_bet"`
}

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) {
	users := h.Store.ListUsers()

	h.jsonResponse(w, http.StatusOK, users)
}

func (h *Handler) CreatePrediction(w http.ResponseWriter, r *http.Request) {
	var req CreatePredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		h.errorResponse(w, http.StatusBadRequest, "name is required")
		return
	}

	if len(req.Choices) < 2 {
		h.errorResponse(w, http.StatusBadRequest, "at least 2 choices required")
		return
	}

	predictionID, err := repo.NewID()
	if err != nil {
		h.Logger.WithError(err).Error("failed to generate prediction ID")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	// Generate IDs for choices if not provided
	for i := range req.Choices {
		if req.Choices[i].ID == "" {
			choiceID, err := repo.NewID()
			if err != nil {
				h.Logger.WithError(err).Error("failed to generate choice ID")
				h.errorResponse(w, http.StatusInternalServerError, "internal error")
				return
			}
			req.Choices[i].ID = choiceID
		}
	}

	prediction := types.Prediction{
		ID:                   predictionID,
		CreatedAt:            time.Now().Format(time.RFC3339),
		Name:                 req.Name,
		Description:          req.Description,
		Status:               types.PredictionStatusOpen,
		ClosesAt:             req.ClosesAt,
		Choices:              req.Choices,
		OddsVisibleBeforeBet: req.OddsVisibleBeforeBet,
	}

	if err := h.Store.PutPrediction(prediction); err != nil {
		h.Logger.WithError(err).Error("failed to create prediction")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.jsonResponse(w, http.StatusCreated, prediction)
}

type UpdatePredictionRequest struct {
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	ClosesAt    *string `json:"closes_at,omitempty"`
	// Choices              []types.PredictionChoice `json:"choices,omitempty"`
	OddsVisibleBeforeBet *bool `json:"odds_visible_before_bet,omitempty"`
}

func (h *Handler) UpdatePrediction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	prediction, err := h.Store.GetPrediction(id)
	if err != nil {
		h.errorResponse(w, http.StatusNotFound, "prediction not found")
		return
	}

	var req UpdatePredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name != nil {
		prediction.Name = *req.Name
	}
	if req.Description != nil {
		prediction.Description = *req.Description
	}
	if req.ClosesAt != nil {
		prediction.ClosesAt = *req.ClosesAt
	}
	if req.OddsVisibleBeforeBet != nil {
		prediction.OddsVisibleBeforeBet = *req.OddsVisibleBeforeBet
	}
	// if len(req.Choices) > 0 {
	// 	// Generate IDs for new choices
	// 	for i := range req.Choices {
	// 		if req.Choices[i].ID == "" {
	// 			choiceID, _ := repo.NewID()
	// 			req.Choices[i].ID = choiceID
	// 		}
	// 	}
	// 	prediction.Choices = req.Choices
	// }

	err = h.Store.PutPrediction(prediction)
	if err == repo.ErrPredictionNotOpen {
		h.errorResponse(w, http.StatusBadRequest, "can only update open predictions")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to update prediction")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.jsonResponse(w, http.StatusOK, prediction)
}

func (h *Handler) ClosePrediction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.Store.ClosePrediction(id)
	if err == repo.ErrPredictionNotFound {
		h.errorResponse(w, http.StatusNotFound, "prediction not found")
		return
	}
	if err == repo.ErrPredictionNotOpen {
		h.errorResponse(w, http.StatusBadRequest, "prediction is not open")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to close prediction")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) VoidPrediction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.Store.VoidPrediction(id)
	if err == repo.ErrPredictionNotFound {
		h.errorResponse(w, http.StatusNotFound, "prediction not found")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to void prediction")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type DecidePredictionRequest struct {
	WinningChoiceID string `json:"winning_choice_id"`
}

func (h *Handler) DecidePrediction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req DecidePredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.Store.DecidePrediction(id, req.WinningChoiceID)
	if err == repo.ErrPredictionNotFound {
		h.errorResponse(w, http.StatusNotFound, "prediction not found")
		return
	}
	if err == repo.ErrPredictionNotInClosedState {
		h.errorResponse(w, http.StatusBadRequest, "prediction must be closed to make decision")
		return
	}
	if err == repo.ErrPredictionChoiceNotFound {
		h.errorResponse(w, http.StatusBadRequest, "invalid winning choice")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to decide prediction")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type GiftTokensRequest struct {
	Amount int64 `json:"amount"`
}

func (h *Handler) GiftTokens(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	var req GiftTokensRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	err := h.Store.GiftTokens(userID, req.Amount)
	if err == repo.ErrUserNotFound {
		h.errorResponse(w, http.StatusNotFound, "user not found")
		return
	}
	if err == repo.ErrGiftAmountMustBePositive {
		h.errorResponse(w, http.StatusBadRequest, "amount must be positive")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to gift tokens")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type ResetPINRequest struct {
	NewPIN string `json:"new_pin"`
}

func (h *Handler) ResetPIN(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	var req ResetPINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.NewPIN == "" {
		h.errorResponse(w, http.StatusBadRequest, "new_pin is required")
		return
	}

	pinHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPIN), bcrypt.MinCost)
	if err != nil {
		h.Logger.WithError(err).Error("failed to hash pin")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	err = h.Store.UpdateUserPIN(userID, pinHash)
	if err == repo.ErrUserNotFound {
		h.errorResponse(w, http.StatusNotFound, "user not found")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to update user")
		h.errorResponse(w, http.StatusInternalServerError, "internal error")
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Register routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Public
	mux.HandleFunc("GET /api/predictions", h.ListPredictions)
	mux.HandleFunc("GET /api/predictions/{id}", h.GetPrediction)
	mux.HandleFunc("GET /api/leaderboard", h.ShowLeaderboard)

	// Guest
	mux.HandleFunc("POST /api/register", h.Register)
	mux.HandleFunc("POST /api/login", h.Login)

	// User (authenticated)
	mux.HandleFunc("GET /api/me", h.requireAuth(h.GetMe))
	mux.HandleFunc("GET /api/my-bets", h.requireAuth(h.GetMyBets))
	mux.HandleFunc("POST /api/bets", h.requireAuth(h.PlaceBet))
	mux.HandleFunc("PUT /api/bets/{id}/amount", h.requireAuth(h.IncreaseBetAmount))

	// Admin
	mux.HandleFunc("GET /api/admin/users", h.requireAdmin(h.ListUsers))
	mux.HandleFunc("POST /api/admin/predictions", h.requireAdmin(h.CreatePrediction))
	mux.HandleFunc("PUT /api/admin/predictions/{id}", h.requireAdmin(h.UpdatePrediction))
	mux.HandleFunc("POST /api/admin/predictions/{id}/close", h.requireAdmin(h.ClosePrediction))
	mux.HandleFunc("POST /api/admin/predictions/{id}/void", h.requireAdmin(h.VoidPrediction))
	mux.HandleFunc("POST /api/admin/predictions/{id}/decide", h.requireAdmin(h.DecidePrediction))
	mux.HandleFunc("POST /api/admin/users/{id}/tokens", h.requireAdmin(h.GiftTokens))
	mux.HandleFunc("POST /api/admin/users/{id}/reset-pin", h.requireAdmin(h.ResetPIN))
}
