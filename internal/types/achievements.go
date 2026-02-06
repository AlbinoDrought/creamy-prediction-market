package types

// Achievement represents a possible achievement
type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	CoinReward  int64  `json:"coin_reward"`
	ItemReward  string `json:"item_reward,omitempty"` // shop item ID granted when earned
}

// UserAchievement represents an earned achievement
type UserAchievement struct {
	UserID        string `json:"user_id"`
	AchievementID string `json:"achievement_id"`
	EarnedAt      string `json:"earned_at"`
}
