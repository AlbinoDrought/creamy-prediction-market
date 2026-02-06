package types

type PredictionStatus string

const (
	// PredictionStatusOpen means players can still place bets
	PredictionStatusOpen = PredictionStatus("open")
	// PredictionStatusClosed means players can no longer place bets, but the outcome isn't decided yet.
	PredictionStatusClosed = PredictionStatus("closed")
	// PredictionStatusDecided means the outcome has been decided and players have been paid out!
	PredictionStatusDecided = PredictionStatus("decided")
	// PredictionStatusVoid means that the prediction was invalidated by something (ex: emergency, loss of power, etc) and all player bets have been refunded
	PredictionStatusVoid = PredictionStatus("void")
)

type Prediction struct {
	ID          string             `json:"id"`
	CreatedAt   string             `json:"created_at"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Status      PredictionStatus   `json:"status"`
	ClosesAt    string             `json:"closes_at"`
	Choices     []PredictionChoice `json:"choices"`
	WinningChoiceID string             `json:"winning_choice_id"`

	OddsVisibleBeforeBet bool `json:"odds_visible_before_bet"`
}

type PredictionChoice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type PredictionOdds struct {
	TotalTokensPlaced int64                  `json:"total_tokens_placed"`
	TotalBetsPlaced   int                    `json:"total_bets_placed"`
	Choices           []PredictionChoiceOdds `json:"choices"`
}

type PredictionChoiceOdds struct {
	PredictionChoiceID string `json:"prediction_choice_id"`
	TokensPlaced       int64  `json:"tokens_placed"`
	BetsPlaced         int    `json:"bets_placed"`

	// OddsBasisPoints is the payout multiplier in basis points (100 = 1x, 250 = 2.5x, 400 = 4x).
	// 0 means no bets have been placed on this choice.
	OddsBasisPoints int64 `json:"odds_basis_points"`
}

func (p Prediction) Odds(bets []Bet) PredictionOdds {
	choicesMap := make(map[string]PredictionChoiceOdds, len(p.Choices))
	for i := range p.Choices {
		choicesMap[p.Choices[i].ID] = PredictionChoiceOdds{
			PredictionChoiceID: p.Choices[i].ID,
		}
	}

	var totalTokensPlaced int64
	for _, bet := range bets {
		totalTokensPlaced += bet.Amount
		choiceOdds := choicesMap[bet.PredictionChoiceID]
		choiceOdds.TokensPlaced += bet.Amount
		choiceOdds.BetsPlaced += 1
		choicesMap[bet.PredictionChoiceID] = choiceOdds
	}

	for choiceID := range choicesMap {
		choiceOdds := choicesMap[choiceID]
		if choiceOdds.TokensPlaced > 0 {
			choiceOdds.OddsBasisPoints = (totalTokensPlaced * 100) / choiceOdds.TokensPlaced
		}
		choicesMap[choiceID] = choiceOdds
	}

	choices := make([]PredictionChoiceOdds, len(p.Choices))
	for i := range p.Choices {
		choices[i] = choicesMap[p.Choices[i].ID]
	}

	return PredictionOdds{
		TotalTokensPlaced: totalTokensPlaced,
		TotalBetsPlaced:   len(bets),
		Choices:           choices,
	}
}

type PredictionWithOdds struct {
	Prediction Prediction     `json:"prediction"`
	Odds       PredictionOdds `json:"odds"`
}

type BetStatus string

const (
	BetStatusPlaced = BetStatus("placed")
	BetStatusWon    = BetStatus("won")
	BetStatusLost   = BetStatus("lost")
	BetStatusVoided = BetStatus("voided")
)

type Bet struct {
	ID                 string    `json:"id"`
	CreatedAt          string    `json:"created_at"`
	UserID             string    `json:"user_id"`
	PredictionID       string    `json:"prediction_id"`
	PredictionChoiceID string    `json:"prediction_choice_id"`
	Amount             int64     `json:"amount"`
	Status             BetStatus `json:"status"`

	WonAmount int64 `json:"won_amount"`
}
