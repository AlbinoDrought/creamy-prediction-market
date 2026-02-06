package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"go.albinodrought.com/creamy-prediction-market/internal/events"
	"go.albinodrought.com/creamy-prediction-market/internal/repo"
	"go.albinodrought.com/creamy-prediction-market/internal/types"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	GracefulCtx    context.Context
	Store          *repo.Store
	Logger         *logrus.Logger
	StartingTokens int64
	EventHub       *events.Hub
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
			h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		next(w, r)
	}
}

func (h *Handler) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := h.getAuthenticatedUser(r)
		if !ok {
			h.errorResponse(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if !user.Admin {
			h.errorResponse(w, http.StatusForbidden, "Admin required")
			return
		}
		next(w, r)
	}
}

// Achievement checking helpers

func newDebouncer(dur time.Duration) func(fn func()) {
    d := &debouncer{
        dur: dur,
    }

    return func(fn func()) {
        d.reset(fn)
    }
}

type debouncer struct {
    mu    sync.Mutex
    dur   time.Duration
    delay *time.Timer
}

func (d *debouncer) reset(fn func()) {
    d.mu.Lock()
    defer d.mu.Unlock()

    if d.delay != nil {
        d.delay.Stop()
    }

    d.delay = time.AfterFunc(d.dur, fn)
}

var grantAchievementLeaderboardDebouncer = newDebouncer(10 * time.Second)

func (h *Handler) grantAchievement(userID, achievementID string) {
	granted, err := h.Store.GrantAchievement(userID, achievementID, time.Now().Format(time.RFC3339))
	if err != nil {
		h.Logger.WithError(err).Error("failed to grant achievement")
		return
	}
	if granted {
		h.EventHub.EmitAchievement(userID, achievementID)
		 // Update leaderboard to show new achievement
		 // Try to avoid emitting a ton of these though: at most once every 10s
		grantAchievementLeaderboardDebouncer(h.EventHub.EmitLeaderboard)
	}
}

func (h *Handler) checkBetAchievements(userID string, betAmount int64) {
	bets := h.Store.ListBetsByUser(userID)

	// Betting milestones
	betCount := len(bets)
	if betCount >= 1 {
		h.grantAchievement(userID, types.AchievementFirstBet)
	}
	if betCount >= 5 {
		h.grantAchievement(userID, types.AchievementBets5)
	}
	if betCount >= 10 {
		h.grantAchievement(userID, types.AchievementBets10)
	}
	if betCount >= 25 {
		h.grantAchievement(userID, types.AchievementBets25)
	}

	h.checkBetAmountAchievements(userID, betAmount)
}

func (h *Handler) checkBetAmountAchievements(userID string, betAmount int64) {
	// Special bet amounts
	if betAmount == 69 {
		h.grantAchievement(userID, types.AchievementBet69)
	}
	if betAmount == 420 {
		h.grantAchievement(userID, types.AchievementBet420)
	}
	if betAmount == 1337 {
		h.grantAchievement(userID, types.AchievementBet1337)
	}
	if betAmount == 8008 {
		h.grantAchievement(userID, types.AchievementBet8008)
	}
}

func (h *Handler) checkWinAchievements(userID string, wonAmount int64) {
	user, err := h.Store.GetUser(userID)
	if err != nil {
		return
	}

	// Token milestones
	if user.Tokens >= 2000 {
		h.grantAchievement(userID, types.AchievementTokens2000)
	}
	if user.Tokens >= 5000 {
		h.grantAchievement(userID, types.AchievementTokens5000)
	}
	if user.Tokens >= 10000 {
		h.grantAchievement(userID, types.AchievementTokens10000)
	}

	// Big wins
	if wonAmount >= 500 {
		h.grantAchievement(userID, types.AchievementBigWin500)
	}
	if wonAmount >= 1000 {
		h.grantAchievement(userID, types.AchievementBigWin1000)
	}

	// Win streaks - count consecutive wins from most recent
	bets := h.Store.ListBetsByUser(userID)
	// Sort by created_at descending
	sort.Slice(bets, func(i, j int) bool {
		return bets[i].CreatedAt > bets[j].CreatedAt
	})

	streak := 0
	for _, bet := range bets {
		if bet.Status == types.BetStatusWon {
			streak++
		} else if bet.Status == types.BetStatusLost {
			break // streak broken
		}
		// Skip placed/voided bets
	}

	if streak >= 3 {
		h.grantAchievement(userID, types.AchievementStreak3)
	}
	if streak >= 5 {
		h.grantAchievement(userID, types.AchievementStreak5)
	}
	if streak >= 10 {
		h.grantAchievement(userID, types.AchievementStreak10)
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
		h.errorResponse(w, http.StatusNotFound, "Prediction not found")
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
			ID:           u.ID,
			Name:         u.Name,
			Tokens:       u.Tokens,
			Achievements: h.Store.GetUserAchievementIDs(u.ID),
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

func (h *Handler) GetAchievements(w http.ResponseWriter, r *http.Request) {
	h.jsonResponse(w, http.StatusOK, types.AllAchievements)
}

func (h *Handler) GetMyAchievements(w http.ResponseWriter, r *http.Request) {
	user, _ := h.getAuthenticatedUser(r)
	achievements := h.Store.GetUserAchievements(user.ID)
	h.jsonResponse(w, http.StatusOK, achievements)
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
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		h.errorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}
	if len(req.Name) > 20 {
		h.errorResponse(w, http.StatusBadRequest, "Name must be less than 20 characters")
		return
	}
	if !regexValidUsername.MatchString(req.Name) {
		h.errorResponse(w, http.StatusBadRequest, "Name must only contain A-Z, a-z, 0-9")
		return
	}

	if req.PIN == "" {
		h.errorResponse(w, http.StatusBadRequest, "Pin is required")
		return
	}

	pinHash, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.MinCost)
	if err != nil {
		h.Logger.WithError(err).Error("failed to hash pin")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	userID, err := repo.NewID()
	if err != nil {
		h.Logger.WithError(err).Error("failed to generate user ID")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
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
			h.errorResponse(w, http.StatusConflict, "Name is already taken")
			return
		}
		h.Logger.WithError(err).Error("failed to add user")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	sessionToken, err := generateSessionToken()
	if err != nil {
		h.Logger.WithError(err).Error("failed to generate session token")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
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
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.Store.GetUserByName(req.Name)
	if err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Invalid name or pin")
		return
	}

	if err := bcrypt.CompareHashAndPassword(user.PINHash, []byte(req.PIN)); err != nil {
		h.errorResponse(w, http.StatusUnauthorized, "Invalid name or pin")
		return
	}

	sessionToken, err := generateSessionToken()
	if err != nil {
		h.Logger.WithError(err).Error("failed to generate session token")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
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
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	betID, err := repo.NewID()
	if err != nil {
		h.Logger.WithError(err).Error("failed to generate bet ID")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
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
		h.errorResponse(w, http.StatusBadRequest, "Amount must be positive")
		return
	}
	if err == repo.ErrBetAlreadyExistsForPrediction {
		h.errorResponse(w, http.StatusConflict, "You already have a bet on this prediction")
		return
	}
	if err == repo.ErrPredictionNotOpen {
		h.errorResponse(w, http.StatusBadRequest, "Prediction is not open for betting")
		return
	}
	if err == repo.ErrPredictionChoiceNotFound {
		h.errorResponse(w, http.StatusBadRequest, "Invalid choice")
		return
	}
	if err == repo.ErrTokensWouldBeNegative {
		h.errorResponse(w, http.StatusBadRequest, "Insufficient tokens")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to place bet")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Emit events
	h.EventHub.EmitPredictions()
	h.EventHub.EmitLeaderboard()
	h.EventHub.EmitBets(user.ID)

	// Check achievements
	h.checkBetAchievements(user.ID, bet.Amount)

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
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	bet, err := h.Store.GetBet(betID)
	if err != nil {
		h.errorResponse(w, http.StatusNotFound, "Bet not found")
		return
	}

	if bet.UserID != user.ID {
		h.errorResponse(w, http.StatusForbidden, "Not your bet")
		return
	}

	err = h.Store.IncreaseBet(bet.ID, req.Amount)
	if err == repo.ErrBetNotActive {
		h.errorResponse(w, http.StatusBadRequest, "Bet is not active")
		return
	}
	if err == repo.ErrPredictionNotFound {
		h.errorResponse(w, http.StatusInternalServerError, "Prediction not found")
		return
	}
	if err == repo.ErrPredictionNotOpen {
		h.errorResponse(w, http.StatusInternalServerError, "Prediction not found")
		return
	}
	if err == repo.ErrBetAlreadyHigher {
		h.errorResponse(w, http.StatusConflict, "Active bet is already higher than specified amount")
		return
	}
	if err == repo.ErrTokensWouldBeNegative {
		h.errorResponse(w, http.StatusBadRequest, "Insufficient tokens")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to updateplace bet")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	bet2, err := h.Store.GetBet(betID)
	if err == nil {
		bet = bet2 // if err, bet amount is stale
	}

	// Emit events
	h.EventHub.EmitPredictions()
	h.EventHub.EmitLeaderboard()
	h.EventHub.EmitBets(user.ID)

	// Check achievements
	h.grantAchievement(user.ID, types.AchievementIncreasedBet)
	h.checkBetAmountAchievements(user.ID, req.Amount)

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
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		h.errorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}

	if len(req.Choices) < 2 {
		h.errorResponse(w, http.StatusBadRequest, "At least 2 choices required")
		return
	}

	predictionID, err := repo.NewID()
	if err != nil {
		h.Logger.WithError(err).Error("failed to generate prediction ID")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Generate IDs for choices if not provided
	for i := range req.Choices {
		if req.Choices[i].ID == "" {
			choiceID, err := repo.NewID()
			if err != nil {
				h.Logger.WithError(err).Error("failed to generate choice ID")
				h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
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
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	h.EventHub.EmitPredictions()

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
		h.errorResponse(w, http.StatusNotFound, "Prediction not found")
		return
	}

	var req UpdatePredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
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
		h.errorResponse(w, http.StatusBadRequest, "Can only update open predictions")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to update prediction")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	h.EventHub.EmitPredictions()

	h.jsonResponse(w, http.StatusOK, prediction)
}

// Sweep closes any open predictions whose ClosesAt time has passed.
func (h *Handler) Sweep() {
	now := time.Now()
	predictions := h.Store.ListPredictions()
	closed := 0
	for _, p := range predictions {
		if p.Status != types.PredictionStatusOpen || p.ClosesAt == "" {
			continue
		}
		closesAt, err := time.Parse(time.RFC3339, p.ClosesAt)
		if err != nil {
			h.Logger.WithError(err).WithField("prediction_id", p.ID).Warn("failed to parse closes_at")
			continue
		}
		if now.Before(closesAt) {
			continue
		}
		if err := h.Store.ClosePrediction(p.ID); err != nil {
			h.Logger.WithError(err).WithField("prediction_id", p.ID).Warn("sweep: failed to close prediction")
			continue
		}
		h.Logger.WithField("prediction_id", p.ID).Info("sweep: closed prediction")
		closed++
	}
	if closed > 0 {
		h.EventHub.EmitPredictions()
	}
}

func (h *Handler) ClosePrediction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.Store.ClosePrediction(id)
	if err == repo.ErrPredictionNotFound {
		h.errorResponse(w, http.StatusNotFound, "Prediction not found")
		return
	}
	if err == repo.ErrPredictionNotOpen {
		h.errorResponse(w, http.StatusBadRequest, "Prediction is not open")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to close prediction")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	h.EventHub.EmitPredictions()

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) VoidPrediction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	err := h.Store.VoidPrediction(id)
	if err == repo.ErrPredictionNotFound {
		h.errorResponse(w, http.StatusNotFound, "Prediction not found")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to void prediction")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	h.EventHub.EmitPredictions()
	h.EventHub.EmitLeaderboard()

	w.WriteHeader(http.StatusNoContent)
}

type DecidePredictionRequest struct {
	WinningChoiceID string `json:"winning_choice_id"`
}

func (h *Handler) DecidePrediction(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var req DecidePredictionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.Store.DecidePrediction(id, req.WinningChoiceID)
	if err == repo.ErrPredictionNotFound {
		h.errorResponse(w, http.StatusNotFound, "Prediction not found")
		return
	}
	if err == repo.ErrPredictionNotInClosedState {
		h.errorResponse(w, http.StatusBadRequest, "Prediction must be closed to make decision")
		return
	}
	if err == repo.ErrPredictionChoiceNotFound {
		h.errorResponse(w, http.StatusBadRequest, "Invalid winning choice")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to decide prediction")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	h.EventHub.EmitPredictions()
	h.EventHub.EmitLeaderboard()
	h.EventHub.EmitBetsAll()

	// Check win achievements for all users who had bets on this prediction
	bets := h.Store.ListBetsByPrediction(id)
	for _, bet := range bets {
		if bet.Status == types.BetStatusWon {
			h.checkWinAchievements(bet.UserID, bet.WonAmount)
		}
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
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	err := h.Store.GiftTokens(userID, req.Amount)
	if err == repo.ErrUserNotFound {
		h.errorResponse(w, http.StatusNotFound, "User not found")
		return
	}
	if err == repo.ErrGiftAmountMustBePositive {
		h.errorResponse(w, http.StatusBadRequest, "Amount must be positive")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to gift tokens")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	h.EventHub.EmitLeaderboard()

	w.WriteHeader(http.StatusNoContent)
}

type ResetPINRequest struct {
	NewPIN string `json:"new_pin"`
}

func (h *Handler) ResetPIN(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("id")

	var req ResetPINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.NewPIN == "" {
		h.errorResponse(w, http.StatusBadRequest, "New PIN is required")
		return
	}

	pinHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPIN), bcrypt.MinCost)
	if err != nil {
		h.Logger.WithError(err).Error("failed to hash pin")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	err = h.Store.UpdateUserPIN(userID, pinHash)
	if err == repo.ErrUserNotFound {
		h.errorResponse(w, http.StatusNotFound, "User not found")
		return
	}
	if err != nil {
		h.Logger.WithError(err).Error("failed to update user")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	h.jsonResponse(w, http.StatusOK, map[string]string{"status": "ok"})
}

// SSE endpoint
func (h *Handler) Events(w http.ResponseWriter, r *http.Request) {
	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Get user ID if authenticated (for user-specific events)
	// Try header first, then query param (EventSource doesn't support headers)
	var userID string
	if user, ok := h.getAuthenticatedUser(r); ok {
		userID = user.ID
	} else if token := r.URL.Query().Get("token"); token != "" {
		if uid, ok := h.Store.GetUserIDBySession(token); ok {
			userID = uid
		}
	}

	// Create client
	client := &events.Client{
		ID:     uuid.New().String(),
		UserID: userID,
		Send:   make(chan []byte, 256),
	}

	h.EventHub.Register(client)
	defer h.EventHub.Unregister(client)

	// Send initial ping
	w.Write([]byte(": ping\n\n"))
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Listen for events or client disconnect or server shutdown
	for {
		select {
		case message, ok := <-client.Send:
			if !ok {
				return
			}
			w.Write(message)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		case <-r.Context().Done():
			return
		case <-h.GracefulCtx.Done():
			return
		}
	}
}

// Register routes
func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// SSE
	mux.HandleFunc("GET /api/events", h.Events)

	// Public
	mux.HandleFunc("GET /api/predictions", h.ListPredictions)
	mux.HandleFunc("GET /api/predictions/{id}", h.GetPrediction)
	mux.HandleFunc("GET /api/leaderboard", h.ShowLeaderboard)
	mux.HandleFunc("GET /api/achievements", h.GetAchievements)

	// Guest
	mux.HandleFunc("POST /api/register", h.Register)
	mux.HandleFunc("POST /api/login", h.Login)

	// User (authenticated)
	mux.HandleFunc("GET /api/me", h.requireAuth(h.GetMe))
	mux.HandleFunc("GET /api/my-bets", h.requireAuth(h.GetMyBets))
	mux.HandleFunc("GET /api/my-achievements", h.requireAuth(h.GetMyAchievements))
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
