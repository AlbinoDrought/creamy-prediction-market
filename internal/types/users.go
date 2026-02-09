package types

// UserCosmetics represents the currently equipped cosmetic values for a user.
type UserCosmetics struct {
	AvatarColor  string `json:"avatar_color,omitempty"`  // hex pair "from,to" e.g. "#EF4444,#F97316"
	AvatarEmoji  string `json:"avatar_emoji,omitempty"`  // emoji replaces letter
	NameEmoji    string `json:"name_emoji,omitempty"`    // emoji next to name
	AvatarEffect string `json:"avatar_effect,omitempty"` // CSS class: glow, sparkle, fire, rainbow
	NameEffect   string `json:"name_effect,omitempty"`   // CSS class: glow, sparkle, rainbow
	NameBold     bool   `json:"name_bold,omitempty"`
	NameFont     string `json:"name_font,omitempty"`   // serif, mono, cursive
	Title        string `json:"title,omitempty"`       // replaces "Player"
	Hat          string `json:"hat,omitempty"`         // emoji displayed above avatar
	AvatarItem   string `json:"avatar_item,omitempty"` // emoji displayed at bottom-right of avatar
}

type User struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// PINHash is a bcrypt hash of the user's pin.
	// This isn't meant to be secure at all really - we expect the pin to simply be four digits, like 0000.
	PINHash []byte `json:"pin_hash,omitempty"`
	Admin   bool   `json:"admin"`

	Tokens            int64 `json:"tokens"`
	Spins             int64 `json:"spins"`
	MinigamePlays     int64 `json:"minigame_plays"`
	MinigameHighScore int64 `json:"minigame_high_score"`
	SheepBets         int64 `json:"sheep_bets,omitempty"`
	ContrarianBets    int64 `json:"contrarian_bets,omitempty"`

	Coins      int64         `json:"coins"`
	OwnedItems []string      `json:"owned_items"`
	Cosmetics  UserCosmetics `json:"cosmetics"`
}

type LeaderboardUser struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Tokens       int64         `json:"tokens"`
	Score        int64         `json:"score"`
	Rank         int           `json:"rank"`
	Achievements []string      `json:"achievements"`
	Cosmetics    UserCosmetics `json:"cosmetics"`
}

type TokenChangeCause string

const (
	// TokenChangeCauseStart means these tokens were given when the user signed up
	TokenChangeCauseStart = TokenChangeCause("start")
	// TokenChangeCauseBetPlaced means these tokens were taken as the user placed a bet
	TokenChangeCauseBetPlaced = TokenChangeCause("bet-placed")
	// TokenChangeCauseBetWon means these tokens were won after the user placed a bet
	TokenChangeCauseBetWon = TokenChangeCause("bet-won")
	// TokenChangeCauseBetVoided means these tokens were refunded after the user placed a bet, but the prediction was voided
	TokenChangeCauseBetVoided = TokenChangeCause("bet-voided")
	// TokenChangeCauseGift means these tokens were given as a gift by the hosts (probably because the user ran out of fake money :) )
	TokenChangeCauseGift = TokenChangeCause("gift")
)

type TokenLog struct {
	ID        string           `json:"id"`
	CreatedAt string           `json:"created_at"`
	UserID    string           `json:"user_id"`
	Change    int64            `json:"change"`
	Cause     TokenChangeCause `json:"cause"`

	// BetID and PredictionID are set if cause is TokenChangeCauseBetPlaced, TokenChangeCauseBetWon, or TokenChangeCauseBetVoided
	BetID        string `json:"bet_id"`
	PredictionID string `json:"prediction_id"`
}
