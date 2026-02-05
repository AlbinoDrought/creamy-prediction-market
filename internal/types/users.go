package types

type User struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	PINHash string `json:"-"`
	Admin   bool   `json:"admin"`

	Tokens int64 `json:"tokens"`
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

	// BetID is set if cause is TokenChangeCauseBetPlaced, TokenChangeCauseBetWon, or TokenChangeCauseBetVoided
	BetID string `json:"bet_id"`
}
