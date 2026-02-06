package types

// Achievement IDs
const (
	// Betting milestones
	AchievementFirstBet = "first_bet"
	AchievementBets5    = "bets_5"
	AchievementBets10   = "bets_10"
	AchievementBets25   = "bets_25"

	// Win streaks
	AchievementStreak3  = "streak_3"
	AchievementStreak5  = "streak_5"
	AchievementStreak10 = "streak_10"

	// Token milestones
	AchievementTokens2000  = "tokens_2000"
	AchievementTokens5000  = "tokens_5000"
	AchievementTokens10000 = "tokens_10000"

	// Big wins
	AchievementBigWin500  = "big_win_500"
	AchievementBigWin1000 = "big_win_1000"

	// Special
	AchievementIncreasedBet = "increased_bet"
	AchievementBet69        = "bet_69"
	AchievementBet420       = "bet_420"
	AchievementBet1337      = "bet_1337"
	AchievementBet8008      = "bet_8008"
)

// Achievement represents a possible achievement
type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

// UserAchievement represents an earned achievement
type UserAchievement struct {
	UserID        string `json:"user_id"`
	AchievementID string `json:"achievement_id"`
	EarnedAt      string `json:"earned_at"`
}

// AllAchievements is the complete list of available achievements
var AllAchievements = []Achievement{
	// Betting milestones
	{ID: AchievementFirstBet, Name: "First Bet", Description: "Place your first bet", Icon: "🎯"},
	{ID: AchievementBets5, Name: "Getting Started", Description: "Place 5 bets", Icon: "🎲"},
	{ID: AchievementBets10, Name: "Regular", Description: "Place 10 bets", Icon: "📊"},
	{ID: AchievementBets25, Name: "Veteran", Description: "Place 25 bets", Icon: "🏆"},

	// Win streaks
	{ID: AchievementStreak3, Name: "Hat Trick", Description: "Win 3 bets in a row", Icon: "🔥"},
	{ID: AchievementStreak5, Name: "On Fire", Description: "Win 5 bets in a row", Icon: "💥"},
	{ID: AchievementStreak10, Name: "Unstoppable", Description: "Win 10 bets in a row", Icon: "⚡"},

	// Token milestones
	{ID: AchievementTokens2000, Name: "Comfortable", Description: "Reach 2,000 tokens", Icon: "💰"},
	{ID: AchievementTokens5000, Name: "Wealthy", Description: "Reach 5,000 tokens", Icon: "💎"},
	{ID: AchievementTokens10000, Name: "Tycoon", Description: "Reach 10,000 tokens", Icon: "👑"},

	// Big wins
	{ID: AchievementBigWin500, Name: "Big Winner", Description: "Win 500+ tokens in a single bet", Icon: "🎉"},
	{ID: AchievementBigWin1000, Name: "Jackpot", Description: "Win 1,000+ tokens in a single bet", Icon: "🎰"},

	// Special
	{ID: AchievementIncreasedBet, Name: "Double Down", Description: "Increase a bet", Icon: "⬆️"},
	{ID: AchievementBet69, Name: "Nice", Description: "Bet exactly 69 tokens", Icon: "😏"},
	{ID: AchievementBet420, Name: "Blazing", Description: "Bet exactly 420 tokens", Icon: "🌿"},
	{ID: AchievementBet1337, Name: "Elite", Description: "Bet exactly 1337 tokens", Icon: "🤓"},
	{ID: AchievementBet8008, Name: "Classic", Description: "Bet exactly 8008 tokens", Icon: "🔢"},
}

// GetAchievementByID returns an achievement by its ID
func GetAchievementByID(id string) (Achievement, bool) {
	for _, a := range AllAchievements {
		if a.ID == id {
			return a, true
		}
	}
	return Achievement{}, false
}
