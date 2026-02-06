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
	StartingCoins  int64
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
		// Award coins for achievement
		achievement, ok := types.GetAchievementByID(achievementID)
		if ok && achievement.CoinReward > 0 {
			if err := h.Store.AddCoins(userID, achievement.CoinReward); err != nil {
				h.Logger.WithError(err).Error("failed to award coins for achievement")
			}
		}
		// Grant item reward if specified
		if ok && achievement.ItemReward != "" {
			if err := h.Store.AddOwnedItem(userID, achievement.ItemReward); err != nil && err != repo.ErrItemAlreadyOwned {
				h.Logger.WithError(err).Error("failed to grant item reward for achievement")
			}
		}
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
	if betCount >= 50 {
		h.grantAchievement(userID, types.AchievementBets50)
	}
	if betCount >= 100 {
		h.grantAchievement(userID, types.AchievementBets100)
	}

	// Bet size achievements
	if betAmount >= 1000 {
		h.grantAchievement(userID, types.AchievementHighRollerBet)
	}
	if betAmount >= 5000 {
		h.grantAchievement(userID, types.AchievementWhaleBet)
	}

	// All in: user has 0 tokens remaining after placing a bet
	user, err := h.Store.GetUser(userID)
	if err == nil && user.Tokens == 0 {
		h.grantAchievement(userID, types.AchievementAllIn)
	}

	// Diversified: bet on 10 different predictions
	predictionSet := map[string]struct{}{}
	for _, b := range bets {
		predictionSet[b.PredictionID] = struct{}{}
	}
	if len(predictionSet) >= 10 {
		h.grantAchievement(userID, types.AchievementDiversified)
	}

	h.checkBetAmountAchievements(userID, betAmount)
}

func (h *Handler) checkBetAmountAchievements(userID string, betAmount int64) {
	// Penny pincher
	if betAmount == 1 {
		h.grantAchievement(userID, types.AchievementPennyPincher)
	}
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

func (h *Handler) checkWinAchievements(userID string, bet types.Bet) {
	user, err := h.Store.GetUser(userID)
	if err != nil {
		return
	}

	// First win
	h.grantAchievement(userID, types.AchievementFirstWin)

	// Long shot: 10:1 or higher odds (payout >= 10x the bet amount)
	if bet.WonAmount >= bet.Amount*10 {
		h.grantAchievement(userID, types.AchievementLongShot)
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
	if bet.WonAmount >= 500 {
		h.grantAchievement(userID, types.AchievementBigWin500)
	}
	if bet.WonAmount >= 1000 {
		h.grantAchievement(userID, types.AchievementBigWin1000)
	}
	if bet.WonAmount >= 5000 {
		h.grantAchievement(userID, types.AchievementBigWin5000)
	}

	// Double up: won at least 2x the bet
	if bet.WonAmount >= bet.Amount*2 {
		h.grantAchievement(userID, types.AchievementDoubleUp)
	}

	// Win streaks, comeback, and total wins
	bets := h.Store.ListBetsByUser(userID)
	// Sort by created_at descending
	sort.Slice(bets, func(i, j int) bool {
		return bets[i].CreatedAt > bets[j].CreatedAt
	})

	// Count consecutive wins from most recent
	streak := 0
	for _, b := range bets {
		if b.Status == types.BetStatusWon {
			streak++
		} else if b.Status == types.BetStatusLost {
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

	// Comeback: most recent resolved bet is a win, and the one before it was a loss
	if len(bets) >= 2 {
		for _, next := range bets[1:] {
			if next.Status == types.BetStatusLost {
				h.grantAchievement(userID, types.AchievementComeback)
				break
			} else if next.Status == types.BetStatusWon {
				break // previous was also a win, no comeback
			}
			// skip placed/voided
		}
	}

	// Total wins milestone
	totalWins := 0
	for _, b := range bets {
		if b.Status == types.BetStatusWon {
			totalWins++
		}
	}
	if totalWins >= 10 {
		h.grantAchievement(userID, types.AchievementWins10)
	}
}

func (h *Handler) checkLossAchievements(userID string, lostAmount int64) {
	// Big loss: lost 100+ tokens in a single bet
	if lostAmount >= 100 {
		h.grantAchievement(userID, types.AchievementBigLoss)
	}

	// Loss streak: count consecutive losses from most recent
	bets := h.Store.ListBetsByUser(userID)
	sort.Slice(bets, func(i, j int) bool {
		return bets[i].CreatedAt > bets[j].CreatedAt
	})
	lossStreak := 0
	for _, b := range bets {
		if b.Status == types.BetStatusLost {
			lossStreak++
		} else if b.Status == types.BetStatusWon {
			break
		}
	}
	if lossStreak >= 3 {
		h.grantAchievement(userID, types.AchievementLossStreak3)
	}

	// Rock bottom: 0 tokens and no pending bets
	user, err := h.Store.GetUser(userID)
	if err != nil {
		return
	}
	if user.Tokens > 0 {
		return
	}
	for _, bet := range bets {
		if bet.Status == types.BetStatusPlaced {
			return
		}
	}
	h.grantAchievement(userID, types.AchievementBroke)
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
			Cosmetics:    h.Store.GetUserCosmetics(u.ID),
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

func (h *Handler) Spin(w http.ResponseWriter, r *http.Request) {
	user, _ := h.getAuthenticatedUser(r)
	spins, err := h.Store.IncrementSpins(user.ID)
	if err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}
	if spins >= 1 {
		h.grantAchievement(user.ID, types.AchievementSpinner)
	}
	if spins >= 10 {
		h.grantAchievement(user.ID, types.AchievementSpinner10)
	}
	if spins >= 100 {
		h.grantAchievement(user.ID, types.AchievementSpinner100)
	}
	w.WriteHeader(http.StatusNoContent)
}

// Minigame endpoints

type ClaimMinigameCoinsRequest struct {
	Score int64 `json:"score"`
}

func (h *Handler) ClaimMinigameCoins(w http.ResponseWriter, r *http.Request) {
	user, _ := h.getAuthenticatedUser(r)

	var req ClaimMinigameCoinsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Score < 0 {
		req.Score = 0
	}

	// 1 coin per 100 points, capped at 5
	coinsEarned := req.Score / 100
	if coinsEarned > 5 {
		coinsEarned = 5
	}

	if coinsEarned > 0 {
		if err := h.Store.AddCoins(user.ID, coinsEarned); err != nil {
			h.Logger.WithError(err).Error("failed to award minigame coins")
			h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}
	}

	// Track plays
	plays, err := h.Store.IncrementMinigamePlays(user.ID)
	if err != nil {
		h.Logger.WithError(err).Error("failed to increment minigame plays")
	}

	// Achievements
	h.grantAchievement(user.ID, types.AchievementMinigame)
	if req.Score >= 500 {
		h.grantAchievement(user.ID, types.AchievementMinigameHighScore)
	}
	if req.Score >= 1000 {
		h.grantAchievement(user.ID, types.AchievementMinigame1000)
	}
	if req.Score >= 2000 {
		h.grantAchievement(user.ID, types.AchievementMinigame2000)
	}
	if plays >= 10 {
		h.grantAchievement(user.ID, types.AchievementMinigamePlays10)
	}
	if plays >= 50 {
		h.grantAchievement(user.ID, types.AchievementMinigamePlays50)
	}

	h.jsonResponse(w, http.StatusOK, map[string]int64{"coins_earned": coinsEarned})
}

// Shop endpoints

func (h *Handler) ListShopItems(w http.ResponseWriter, r *http.Request) {
	h.jsonResponse(w, http.StatusOK, types.AllShopItems)
}

func (h *Handler) BuyShopItem(w http.ResponseWriter, r *http.Request) {
	user, _ := h.getAuthenticatedUser(r)
	itemID := r.PathValue("itemId")

	item, ok := types.GetShopItemByID(itemID)
	if !ok {
		h.errorResponse(w, http.StatusNotFound, "Item not found")
		return
	}

	if item.Locked {
		h.errorResponse(w, http.StatusBadRequest, "This item cannot be purchased")
		return
	}

	// Non-consumable items can only be bought once
	if !item.Consumable && h.Store.UserOwnsItem(user.ID, itemID) {
		h.errorResponse(w, http.StatusConflict, "You already own this item")
		return
	}

	if err := h.Store.DeductCoins(user.ID, item.Price); err == repo.ErrInsufficientCoins {
		h.errorResponse(w, http.StatusBadRequest, "Insufficient coins")
		return
	} else if err != nil {
		h.Logger.WithError(err).Error("failed to deduct coins")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	// Consumable items trigger a global effect immediately
	if item.Consumable {
		h.EventHub.EmitGlobalAction(user.Name, item.Value)
		w.WriteHeader(http.StatusNoContent)
		return
	}

	if err := h.Store.AddOwnedItem(user.ID, itemID); err != nil {
		h.Logger.WithError(err).Error("failed to add owned item")
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) EquipItem(w http.ResponseWriter, r *http.Request) {
	user, _ := h.getAuthenticatedUser(r)
	itemID := r.PathValue("itemId")

	item, ok := types.GetShopItemByID(itemID)
	if !ok {
		h.errorResponse(w, http.StatusNotFound, "Item not found")
		return
	}

	if !h.Store.UserOwnsItem(user.ID, itemID) {
		h.errorResponse(w, http.StatusForbidden, "You don't own this item")
		return
	}

	cosmetics := h.Store.GetUserCosmetics(user.ID)

	switch item.Category {
	case types.ShopItemCategoryAvatarColor:
		cosmetics.AvatarColor = item.Value
	case types.ShopItemCategoryAvatarEmoji:
		cosmetics.AvatarEmoji = item.Value
	case types.ShopItemCategoryNameEmoji:
		cosmetics.NameEmoji = item.Value
	case types.ShopItemCategoryAvatarEffect:
		cosmetics.AvatarEffect = item.Value
	case types.ShopItemCategoryNameEffect:
		cosmetics.NameEffect = item.Value
	case types.ShopItemCategoryNameBold:
		cosmetics.NameBold = true
	case types.ShopItemCategoryNameFont:
		cosmetics.NameFont = item.Value
	case types.ShopItemCategoryTitle:
		cosmetics.Title = item.Value
	case types.ShopItemCategoryHat:
		cosmetics.Hat = item.Value
	case types.ShopItemCategoryAvatarItem:
		cosmetics.AvatarItem = item.Value
	default:
		h.errorResponse(w, http.StatusBadRequest, "This item cannot be equipped")
		return
	}

	if err := h.Store.SetCosmetics(user.ID, cosmetics); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	h.EventHub.EmitLeaderboard()
	h.jsonResponse(w, http.StatusOK, cosmetics)
}

func (h *Handler) UnequipCategory(w http.ResponseWriter, r *http.Request) {
	user, _ := h.getAuthenticatedUser(r)
	category := r.PathValue("category")

	cosmetics := h.Store.GetUserCosmetics(user.ID)

	switch types.ShopItemCategory(category) {
	case types.ShopItemCategoryAvatarColor:
		cosmetics.AvatarColor = ""
	case types.ShopItemCategoryAvatarEmoji:
		cosmetics.AvatarEmoji = ""
	case types.ShopItemCategoryNameEmoji:
		cosmetics.NameEmoji = ""
	case types.ShopItemCategoryAvatarEffect:
		cosmetics.AvatarEffect = ""
	case types.ShopItemCategoryNameEffect:
		cosmetics.NameEffect = ""
	case types.ShopItemCategoryNameBold:
		cosmetics.NameBold = false
	case types.ShopItemCategoryNameFont:
		cosmetics.NameFont = ""
	case types.ShopItemCategoryTitle:
		cosmetics.Title = ""
	case types.ShopItemCategoryHat:
		cosmetics.Hat = ""
	case types.ShopItemCategoryAvatarItem:
		cosmetics.AvatarItem = ""
	default:
		h.errorResponse(w, http.StatusBadRequest, "Invalid category")
		return
	}

	if err := h.Store.SetCosmetics(user.ID, cosmetics); err != nil {
		h.errorResponse(w, http.StatusInternalServerError, "Internal Server Error")
		return
	}

	h.EventHub.EmitLeaderboard()
	h.jsonResponse(w, http.StatusOK, cosmetics)
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
		Coins:   h.StartingCoins,
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

	// Check achievements and award coins for all users who had bets on this prediction
	bets := h.Store.ListBetsByPrediction(id)
	for _, bet := range bets {
		if bet.Status == types.BetStatusWon {
			h.checkWinAchievements(bet.UserID, bet)
			// Award coins: 1 coin per 200 tokens won
			if coinsEarned := bet.WonAmount / 200; coinsEarned > 0 {
				if err := h.Store.AddCoins(bet.UserID, coinsEarned); err != nil {
					h.Logger.WithError(err).Error("failed to award bet win coins")
				}
			}
		}
		if bet.Status == types.BetStatusLost {
			h.checkLossAchievements(bet.UserID, bet.Amount)
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
	mux.HandleFunc("POST /api/spin", h.requireAuth(h.Spin))
	mux.HandleFunc("GET /api/shop", h.ListShopItems)
	mux.HandleFunc("POST /api/shop/buy/{itemId}", h.requireAuth(h.BuyShopItem))
	mux.HandleFunc("PUT /api/shop/equip/{itemId}", h.requireAuth(h.EquipItem))
	mux.HandleFunc("DELETE /api/shop/equip/{category}", h.requireAuth(h.UnequipCategory))
	mux.HandleFunc("POST /api/bets", h.requireAuth(h.PlaceBet))
	mux.HandleFunc("PUT /api/bets/{id}/amount", h.requireAuth(h.IncreaseBetAmount))
	mux.HandleFunc("POST /api/minigame/claim", h.requireAuth(h.ClaimMinigameCoins))

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
