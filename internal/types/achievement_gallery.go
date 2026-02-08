package types

// Achievement IDs
const (
	// Betting milestones
	AchievementFirstBet = "first_bet"
	AchievementBets5    = "bets_5"
	AchievementBets10   = "bets_10"
	AchievementBets25   = "bets_25"
	AchievementBets50   = "bets_50"
	AchievementBets100  = "bets_100"

	// Win streaks
	AchievementStreak3  = "streak_3"
	AchievementStreak5  = "streak_5"
	AchievementStreak10 = "streak_10"

	// Loss streaks
	AchievementLossStreak3 = "loss_streak_3"

	// Win milestones
	AchievementWins10 = "wins_10"

	// Token milestones
	AchievementTokens2000  = "tokens_2000"
	AchievementTokens5000  = "tokens_5000"
	AchievementTokens10000 = "tokens_10000"

	// Big wins
	AchievementBigWin500  = "big_win_500"
	AchievementBigWin1000 = "big_win_1000"
	AchievementBigWin5000 = "big_win_5000"

	// Win types
	AchievementDoubleUp = "double_up"
	AchievementComeback = "comeback"

	// Bet size
	AchievementHighRollerBet = "high_roller_bet"
	AchievementWhaleBet      = "whale_bet"
	AchievementAllIn         = "all_in"
	AchievementPennyPincher  = "penny_pincher"

	// Diversity
	AchievementDiversified = "diversified"

	// Item rewards
	AchievementFirstWin = "first_win"
	AchievementLongShot = "long_shot"
	AchievementBroke    = "broke"
	AchievementBigLoss  = "big_loss"

	// Special
	AchievementIncreasedBet = "increased_bet"
	AchievementBet69        = "bet_69"
	AchievementBet420       = "bet_420"
	AchievementBet1337      = "bet_1337"
	AchievementBet8008      = "bet_8008"
	AchievementSpinner      = "spinner"
	AchievementSpinner10    = "spinner_10"
	AchievementSpinner100   = "spinner_100"

	// Minigame
	AchievementMinigame          = "minigame"
	AchievementMinigameHighScore = "minigame_high_score"
	AchievementMinigame1000      = "minigame_1000"
	AchievementMinigame2000      = "minigame_2000"
	AchievementMinigamePlays10   = "minigame_plays_10"
	AchievementMinigamePlays50   = "minigame_plays_50"

	// Extended loss streaks
	AchievementLossStreak5  = "loss_streak_5"
	AchievementLossStreak10 = "loss_streak_10"

	// Degenerate energy
	AchievementSpeedRun     = "speed_run"
	AchievementPaperHands   = "paper_hands"
	AchievementDiamondHands = "diamond_hands"

	// Game day
	AchievementJinx                = "jinx"
	AchievementGatorade            = "gatorade"
	AchievementCoinTossPhilosopher = "coin_toss_big"

	// Social
	AchievementSheep      = "sheep"
	AchievementContrarian = "contrarian"
	AchievementLastSecond = "last_second"
	AchievementSpamFilter = "spam_filter"

	// Post-decision
	AchievementDownBad             = "down_bad"
	AchievementTrustTheProcess     = "trust_the_process"
	AchievementAccountant          = "accountant"
	AchievementSus                 = "sus"
	AchievementParticipationTrophy = "participation_trophy"
)

// AllAchievements is the complete list of available achievements
var AllAchievements = []Achievement{
	// Betting milestones
	{ID: AchievementFirstBet, Name: "First Bet", Description: "Place your first bet", Icon: "🎯", CoinReward: 2},
	{ID: AchievementBets5, Name: "Getting Started", Description: "Place 5 bets", Icon: "🎲", CoinReward: 3},
	{ID: AchievementBets10, Name: "Regular", Description: "Place 10 bets", Icon: "📊", CoinReward: 5},
	{ID: AchievementBets25, Name: "Veteran", Description: "Place 25 bets", Icon: "🏆", CoinReward: 10},
	{ID: AchievementBets50, Name: "Addict", Description: "Place 50 bets", Icon: "🎰", CoinReward: 15},
	{ID: AchievementBets100, Name: "No Life", Description: "Place 100 bets", Icon: "💀", CoinReward: 25},

	// Win streaks
	{ID: AchievementStreak3, Name: "Hat Trick", Description: "Win 3 bets in a row", Icon: "🔥", CoinReward: 5},
	{ID: AchievementStreak5, Name: "On Fire", Description: "Win 5 bets in a row", Icon: "💥", CoinReward: 10},
	{ID: AchievementStreak10, Name: "Unstoppable", Description: "Win 10 bets in a row", Icon: "⚡", CoinReward: 25},

	// Loss streaks
	{ID: AchievementLossStreak3, Name: "Unlucky", Description: "Lose 3 bets in a row", Icon: "🫠", CoinReward: 3},

	// Win milestones
	{ID: AchievementWins10, Name: "Ten Timer", Description: "Win 10 bets", Icon: "🔟", CoinReward: 10},

	// Token milestones
	{ID: AchievementTokens2000, Name: "Comfortable", Description: "Reach 2,000 tokens", Icon: "💰", CoinReward: 3},
	{ID: AchievementTokens5000, Name: "Wealthy", Description: "Reach 5,000 tokens", Icon: "💎", CoinReward: 5},
	{ID: AchievementTokens10000, Name: "Tycoon", Description: "Reach 10,000 tokens", Icon: "👑", CoinReward: 10},

	// Big wins
	{ID: AchievementBigWin500, Name: "Big Winner", Description: "Win 500+ tokens in a single bet", Icon: "🎉", CoinReward: 5},
	{ID: AchievementBigWin1000, Name: "Jackpot", Description: "Win 1,000+ tokens in a single bet", Icon: "🎰", CoinReward: 10},
	{ID: AchievementBigWin5000, Name: "Mega Jackpot", Description: "Win 5,000+ tokens in a single bet", Icon: "💎", CoinReward: 25},

	// Win types
	{ID: AchievementDoubleUp, Name: "Double Up", Description: "Win at least 2x your bet", Icon: "✌️", CoinReward: 3},
	{ID: AchievementComeback, Name: "Comeback Kid", Description: "Win a bet right after losing one", Icon: "💪", CoinReward: 5},

	// Bet size
	{ID: AchievementHighRollerBet, Name: "High Roller", Description: "Bet 1,000+ tokens at once", Icon: "🎲", CoinReward: 5},
	{ID: AchievementWhaleBet, Name: "Whale", Description: "Bet 5,000+ tokens at once", Icon: "🐋", CoinReward: 10},
	{ID: AchievementAllIn, Name: "YOLO", Description: "Bet every last token you have", Icon: "🤪", CoinReward: 5},
	{ID: AchievementPennyPincher, Name: "Penny Pincher", Description: "Bet exactly 1 token", Icon: "🪙", CoinReward: 2},

	// Diversity
	{ID: AchievementDiversified, Name: "Diversified", Description: "Bet on 10 different predictions", Icon: "📈", CoinReward: 5},

	// Special
	{ID: AchievementIncreasedBet, Name: "Double Down", Description: "Increase a bet", Icon: "⬆️", CoinReward: 2},
	{ID: AchievementBet69, Name: "Nice", Description: "Bet exactly 69 tokens", Icon: "😏", CoinReward: 3},
	{ID: AchievementBet420, Name: "Blazing", Description: "Bet exactly 420 tokens", Icon: "🌿", CoinReward: 3},
	{ID: AchievementBet1337, Name: "Elite", Description: "Bet exactly 1337 tokens", Icon: "🤓", CoinReward: 3},
	{ID: AchievementBet8008, Name: "Classic", Description: "Bet exactly 8008 tokens", Icon: "🔢", CoinReward: 3},

	// Item rewards
	{ID: AchievementFirstWin, Name: "Winner", Description: "Win your first bet", Icon: "🏆", CoinReward: 2, ItemReward: "title_winning"},
	{ID: AchievementLongShot, Name: "Long Shot", Description: "Win a bet with 10:1 or higher odds", Icon: "💸", CoinReward: 5, ItemReward: "avatar_effect_cash"},
	{ID: AchievementBroke, Name: "Rock Bottom", Description: "Lose all your tokens", Icon: "💩", CoinReward: 1, ItemReward: "hat_poop"},
	{ID: AchievementBigLoss, Name: "Big Sad", Description: "Lose 100+ tokens in a single bet", Icon: "😭", CoinReward: 1, ItemReward: "avatar_emoji_sad"},

	// Hidden
	{ID: AchievementSpinner, Name: "Fidget Spinner", Description: "Spin your avatar", Icon: "🌀", CoinReward: 1},
	{ID: AchievementSpinner10, Name: "Dizzy", Description: "Spin your avatar 10 times", Icon: "😵‍💫", CoinReward: 3},
	{ID: AchievementSpinner100, Name: "Centrifuge", Description: "Spin your avatar 100 times", Icon: "🌪️", CoinReward: 10},

	// Minigame
	{ID: AchievementMinigame, Name: "Gamer", Description: "Play the minigame", Icon: "🎮", CoinReward: 1},
	{ID: AchievementMinigameHighScore, Name: "Dino Master", Description: "Score 500+ in the minigame", Icon: "🦖", CoinReward: 5, ItemReward: "avatar_emoji_dino"},
	{ID: AchievementMinigame1000, Name: "Fossil Hunter", Description: "Score 1000+ in the minigame", Icon: "🦴", CoinReward: 10, ItemReward: "avatar_item_bone"},
	{ID: AchievementMinigame2000, Name: "Extinction Event", Description: "Score 2000+ in the minigame", Icon: "☄️", CoinReward: 25, ItemReward: "title_dino_whisperer"},
	{ID: AchievementMinigamePlays10, Name: "Addicted", Description: "Play the minigame 10 times", Icon: "🔁", CoinReward: 3, ItemReward: "name_emoji_dino"},
	{ID: AchievementMinigamePlays50, Name: "Jurassic Park", Description: "Play the minigame 50 times", Icon: "🏞️", CoinReward: 10, ItemReward: "hat_dino"},

	// Extended loss streaks
	{ID: AchievementLossStreak5, Name: "Dumpster Fire", Description: "Lose 5 bets in a row", Icon: "🗑️", CoinReward: 5},
	{ID: AchievementLossStreak10, Name: "Cursed", Description: "Lose 10 bets in a row", Icon: "🪦", CoinReward: 10, ItemReward: "hat_gravestone"},

	// Degenerate energy
	{ID: AchievementSpeedRun, Name: "Speed Run", Description: "Go broke within your first 5 bets", Icon: "💨", CoinReward: 3},
	{ID: AchievementPaperHands, Name: "Paper Hands", Description: "Bet exactly 1 token five times", Icon: "🧻", CoinReward: 3},
	{ID: AchievementDiamondHands, Name: "Diamond Hands", Description: "Bet more than half your tokens at once", Icon: "💎", CoinReward: 5},

	// Game day
	{ID: AchievementJinx, Name: "Jinx", Description: "Lose a bet within 5 minutes of placing it", Icon: "🤡", CoinReward: 3},
	{ID: AchievementGatorade, Name: "Gatorade Sommelier", Description: "Bet on the Gatorade shower color", Icon: "🍷", CoinReward: 2},
	{ID: AchievementCoinTossPhilosopher, Name: "Coin Toss Philosopher", Description: "Bet 500+ tokens on a coin toss", Icon: "🪙", CoinReward: 3},

	// Social
	{ID: AchievementSheep, Name: "Sheep", Description: "Place 5 bets on the majority side", Icon: "🐑", CoinReward: 3},
	{ID: AchievementContrarian, Name: "Contrarian", Description: "Place 5 bets against the majority", Icon: "🙃", CoinReward: 5},
	{ID: AchievementLastSecond, Name: "Last Second Larry", Description: "Place a bet in the final 60 seconds", Icon: "⏰", CoinReward: 3},
	{ID: AchievementSpamFilter, Name: "Spam Filter", Description: "Place 10 bets within 5 minutes", Icon: "📬", CoinReward: 5},

	// Post-decision
	{ID: AchievementDownBad, Name: "Down Bad", Description: "Have fewer tokens than you started with after 20 bets", Icon: "📉", CoinReward: 3},
	{ID: AchievementTrustTheProcess, Name: "Trust the Process", Description: "Win a bet after losing 3 in a row", Icon: "🧘", CoinReward: 5},
	{ID: AchievementAccountant, Name: "The Accountant", Description: "Bet exactly 100 tokens five times", Icon: "🧮", CoinReward: 3},
	{ID: AchievementSus, Name: "Sus", Description: "Win 5 bets with at least 2x payout", Icon: "👀", CoinReward: 10},
	{ID: AchievementParticipationTrophy, Name: "Participation Trophy", Description: "Bet on 20+ predictions but win fewer than 5", Icon: "🏅", CoinReward: 3, ItemReward: "title_participation"},
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
