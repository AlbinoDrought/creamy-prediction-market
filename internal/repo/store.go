package repo

import (
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.albinodrought.com/creamy-prediction-market/internal/types"
)

func NewID() (string, error) {
	valUUID, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return valUUID.String(), nil
}

type Store struct {
	lock        sync.RWMutex
	dirty       bool
	users       map[string]types.User
	predictions map[string]types.Prediction
	bets        map[string]types.Bet
	tokenLog    map[string]types.TokenLog
	sessions    map[string]string // session token -> user ID
}

func NewStore() *Store {
	return &Store{
		users:       make(map[string]types.User),
		predictions: make(map[string]types.Prediction),
		bets:        make(map[string]types.Bet),
		tokenLog:    make(map[string]types.TokenLog),
		sessions:    make(map[string]string),
	}
}

type storeCopy struct {
	Users       map[string]types.User
	Predictions map[string]types.Prediction
	Bets        map[string]types.Bet
	TokenLog    map[string]types.TokenLog
	Sessions    map[string]string
}

func (s *Store) Save(w io.Writer) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.dirty = false

	return json.NewEncoder(w).Encode(&storeCopy{
		Users:       s.users,
		Predictions: s.predictions,
		Bets:        s.bets,
		TokenLog:    s.tokenLog,
		Sessions:    s.sessions,
	})
}

func (s *Store) Load(r io.Reader) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	var copy storeCopy
	if err := json.NewDecoder(r).Decode(&copy); err != nil {
		return err
	}
	if copy.Users == nil {
		copy.Users = make(map[string]types.User)
	}
	if copy.Predictions == nil {
		copy.Predictions = make(map[string]types.Prediction)
	}
	if copy.Bets == nil {
		copy.Bets = make(map[string]types.Bet)
	}
	if copy.TokenLog == nil {
		copy.TokenLog = make(map[string]types.TokenLog)
	}
	if copy.Sessions == nil {
		copy.Sessions = make(map[string]string)
	}

	s.dirty = false
	s.users = copy.Users
	s.predictions = copy.Predictions
	s.bets = copy.Bets
	s.tokenLog = copy.TokenLog
	s.sessions = copy.Sessions

	return nil
}

func (s *Store) IsDirty() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.dirty
}

var ErrUserNameTaken = errors.New("user name is taken")
var ErrUserMustBePassedWithZeroTokens = errors.New("user must be passed with 0 tokens")

func (s *Store) AddUser(u types.User, startingTokens int64) error {
	if u.Tokens != 0 {
		return ErrUserMustBePassedWithZeroTokens
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	lcase := strings.ToLower(u.Name)

	for id := range s.users {
		if lcase == strings.ToLower(s.users[id].Name) {
			return ErrUserNameTaken
		}
	}

	logID, err := NewID()
	if err != nil {
		return err
	}

	s.dirty = true

	s.users[u.ID] = u

	err = s.applyTokenLogLocked(types.TokenLog{
		ID:        logID,
		CreatedAt: time.Now().Format(time.RFC3339),
		UserID:    u.ID,
		Change:    startingTokens,
		Cause:     types.TokenChangeCauseStart,
	})
	if err != nil {
		return err
	}

	return nil
}

var ErrPredictionNotOpen = errors.New("prediction exists but is not open")

func (s *Store) PutPrediction(p types.Prediction) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if existing, ok := s.predictions[p.ID]; ok {
		if existing.Status != types.PredictionStatusOpen {
			return ErrPredictionNotOpen
		}
	}

	s.dirty = true

	s.predictions[p.ID] = p

	return nil
}

var ErrTokensWouldBeNegative = errors.New("token log change would make tokens negative, refusing")

func (s *Store) applyTokenLogLocked(tc types.TokenLog) error {
	user, ok := s.users[tc.UserID]
	if !ok {
		return ErrUserNotFound
	}

	newTokenValue := user.Tokens + tc.Change
	if newTokenValue < 0 {
		return ErrTokensWouldBeNegative
	}

	s.dirty = true

	user.Tokens = newTokenValue
	s.users[tc.UserID] = user

	s.tokenLog[tc.ID] = tc

	return nil
}

var ErrGiftAmountMustBePositive = errors.New("gift token amount must be positive")

func (s *Store) GiftTokens(userID string, amount int64) error {
	if amount <= 0 {
		return ErrGiftAmountMustBePositive
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	logID, err := NewID()
	if err != nil {
		return err
	}

	s.dirty = true

	err = s.applyTokenLogLocked(types.TokenLog{
		ID:        logID,
		CreatedAt: time.Now().Format(time.RFC3339),
		UserID:    userID,
		Change:    amount,
		Cause:     types.TokenChangeCauseGift,
	})
	if err != nil {
		return err
	}

	return nil
}

// User methods

var ErrUserNotFound = errors.New("user not found")

func (s *Store) GetUser(id string) (types.User, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	user, ok := s.users[id]
	if !ok {
		return types.User{}, ErrUserNotFound
	}
	return user, nil
}

func (s *Store) GetUserByName(name string) (types.User, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	lcase := strings.ToLower(name)
	for _, user := range s.users {
		if lcase == strings.ToLower(user.Name) {
			return user, nil
		}
	}
	return types.User{}, ErrUserNotFound
}

func (s *Store) ListUsers() []types.User {
	s.lock.RLock()
	defer s.lock.RUnlock()

	users := make([]types.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}
	return users
}

func (s *Store) UpdateUserPIN(id string, hash []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	user, ok := s.users[id]
	if !ok {
		return ErrUserNotFound
	}

	s.dirty = true

	user.PINHash = hash
	s.users[user.ID] = user

	return nil
}

// Session methods

func (s *Store) CreateSession(token, userID string) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.dirty = true
	s.sessions[token] = userID
}

func (s *Store) GetUserIDBySession(token string) (string, bool) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	userID, ok := s.sessions[token]
	return userID, ok
}

// func (s *Store) DeleteSession(token string) {
// 	s.lock.Lock()
// 	defer s.lock.Unlock()
// 	delete(s.sessions, token)
// }

// Prediction methods

var ErrPredictionNotFound = errors.New("prediction not found")

func (s *Store) getPredictionLocked(id string) (types.Prediction, error) {
	prediction, ok := s.predictions[id]
	if !ok {
		return types.Prediction{}, ErrPredictionNotFound
	}
	return prediction, nil
}

func (s *Store) GetPrediction(id string) (types.Prediction, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.getPredictionLocked(id)
}

func (s *Store) GetPredictionWithOdds(id string) (types.PredictionWithOdds, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	prediction, err := s.getPredictionLocked(id)
	if err != nil {
		return types.PredictionWithOdds{}, err
	}
	bets := s.listBetsByPredictionLocked(id)

	return types.PredictionWithOdds{
		Prediction: prediction,
		Odds:       prediction.Odds(bets),
	}, nil
}

func (s *Store) ListPredictions() []types.Prediction {
	s.lock.RLock()
	defer s.lock.RUnlock()

	predictions := make([]types.Prediction, 0, len(s.predictions))
	for _, p := range s.predictions {
		predictions = append(predictions, p)
	}
	return predictions
}

func (s *Store) ListPredictionsWithOdds() []types.PredictionWithOdds {
	s.lock.RLock()
	defer s.lock.RUnlock()

	predictions := make([]types.Prediction, 0, len(s.predictions))
	for k := range s.predictions {
		predictions = append(predictions, s.predictions[k]) // done this way because it's a map
	}

	betsByPrediction := map[string][]types.Bet{}
	for k := range s.bets {
		betsForPrediction := betsByPrediction[s.bets[k].PredictionID]
		betsForPrediction = append(betsForPrediction, s.bets[k])
		betsByPrediction[s.bets[k].PredictionID] = betsForPrediction
	}

	predictionsWithBets := make([]types.PredictionWithOdds, len(predictions))
	for i := range predictions {
		predictionsWithBets[i].Prediction = predictions[i]
		predictionsWithBets[i].Odds = predictions[i].Odds(betsByPrediction[predictions[i].ID])
	}

	return predictionsWithBets
}

var ErrPredictionNotInClosedState = errors.New("prediction not in closed state")

func (s *Store) DecidePrediction(id, choice string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	p, ok := s.predictions[id]
	if !ok {
		return ErrPredictionNotFound
	}

	if p.Status != types.PredictionStatusClosed {
		return ErrPredictionNotInClosedState
	}

	validChoice := false
	for _, c := range p.Choices {
		if c.ID == choice {
			validChoice = true
			break
		}
	}
	if !validChoice {
		return ErrPredictionChoiceNotFound
	}

	bets := s.listBetsByPredictionLocked(id)
	odds := p.Odds(bets)

	var winningOdds int64
	for _, choiceOdds := range odds.Choices {
		if choiceOdds.PredictionChoiceID == choice {
			winningOdds = choiceOdds.OddsBasisPoints
			break
		}
	}
	if winningOdds < 100 {
		winningOdds = 100
	}

	s.dirty = true

	tcs := []types.TokenLog{}

	for i := range bets {
		if bets[i].Status != types.BetStatusPlaced {
			continue
		}

		if bets[i].PredictionChoiceID == choice {
			// winner
			payout := (bets[i].Amount * winningOdds) / 100

			logID, err := NewID()
			if err != nil {
				return err
			}
			tcs = append(tcs, types.TokenLog{
				ID:           logID,
				CreatedAt:    time.Now().Format(time.RFC3339),
				UserID:       bets[i].UserID,
				Change:       payout,
				Cause:        types.TokenChangeCauseBetWon,
				BetID:        bets[i].ID,
				PredictionID: id,
			})

			bets[i].Status = types.BetStatusWon
			bets[i].WonAmount = payout
		} else {
			bets[i].Status = types.BetStatusLost
		}
	}

	// apply all token changes
	for i := range tcs {
		if err := s.applyTokenLogLocked(tcs[i]); err != nil {
			return err
		}
	}

	// apply all bet changes
	for i := range bets {
		s.bets[bets[i].ID] = bets[i]
	}

	// update prediction
	p.Status = types.PredictionStatusDecided
	s.predictions[p.ID] = p

	return nil
}

func (s *Store) VoidPrediction(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	p, ok := s.predictions[id]
	if !ok {
		return ErrPredictionNotFound
	}

	if p.Status == types.PredictionStatusVoid {
		// already voided
		return nil
	}

	logs := []types.TokenLog{}
	for _, tc := range s.tokenLog {
		if tc.PredictionID != id {
			continue
		}
		logs = append(logs, tc)
	}

	sort.Slice(logs, func(i, j int) bool {
		// put any events that remove tokens at the front
		// this is so, as we revert these changes, we increase a user's balance before decreasing it
		// (avoiding negative token balance or similar)
		return logs[i].Change < logs[j].Change
	})

	reverseLogs := make([]types.TokenLog, len(logs))
	for i := range logs {
		logID, err := NewID()
		if err != nil {
			return err
		}
		reverseLogs[i] = types.TokenLog{
			ID:           logID,
			CreatedAt:    time.Now().Format(time.RFC3339),
			UserID:       logs[i].UserID,
			Change:       -logs[i].Change,
			Cause:        types.TokenChangeCauseBetVoided,
			BetID:        logs[i].BetID,
			PredictionID: logs[i].PredictionID,
		}
	}

	s.dirty = true

	for i := range reverseLogs {
		if err := s.applyTokenLogLocked(reverseLogs[i]); err != nil {
			return err // this would be a really bad error to have
		}
	}

	// mark all bets as voided
	bets := s.listBetsByPredictionLocked(id)
	for i := range bets {
		bet := s.bets[bets[i].ID]
		bet.Status = types.BetStatusVoided
		s.bets[bets[i].ID] = bet
	}

	return nil
}

func (s *Store) ClosePrediction(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	p, ok := s.predictions[id]
	if !ok {
		return ErrPredictionNotFound
	}

	if p.Status != types.PredictionStatusOpen {
		return ErrPredictionNotOpen
	}

	s.dirty = true

	p.Status = types.PredictionStatusClosed
	s.predictions[id] = p

	return nil
}

// Bet methods

var ErrBetNotFound = errors.New("bet not found")
var ErrBetAmountMustBePositive = errors.New("bet amount must be positive")
var ErrBetAlreadyExistsForPrediction = errors.New("a bet already exists by this user for this prediction")
var ErrPredictionChoiceNotFound = errors.New("prediction found but choice does not exist")

func (s *Store) CreateBet(bet types.Bet) error {
	if bet.Amount <= 0 {
		return ErrBetAmountMustBePositive
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	if _, exists := s.getUserBetOnPredictionLocked(bet.UserID, bet.PredictionID); exists {
		return ErrBetAlreadyExistsForPrediction
	}

	prediction, err := s.getPredictionLocked(bet.PredictionID)
	if err != nil {
		return err
	}
	if prediction.Status != types.PredictionStatusOpen {
		return ErrPredictionNotOpen
	}

	validChoice := false
	for _, c := range prediction.Choices {
		if c.ID == bet.PredictionChoiceID {
			validChoice = true
			break
		}
	}
	if !validChoice {
		return ErrPredictionChoiceNotFound
	}

	logID, err := NewID()
	if err != nil {
		return err
	}

	s.dirty = true

	err = s.applyTokenLogLocked(types.TokenLog{
		ID:           logID,
		CreatedAt:    time.Now().Format(time.RFC3339),
		UserID:       bet.UserID,
		Change:       -bet.Amount,
		Cause:        types.TokenChangeCauseBetPlaced,
		BetID:        bet.ID,
		PredictionID: bet.PredictionID,
	})
	if err != nil {
		return err
	}

	s.bets[bet.ID] = bet

	return nil
}

func (s *Store) GetBet(id string) (types.Bet, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	bet, ok := s.bets[id]
	if !ok {
		return types.Bet{}, ErrBetNotFound
	}
	return bet, nil
}

var ErrBetNotActive = errors.New("bet not active")
var ErrBetAlreadyHigher = errors.New("bet is already higher than specified amount")

func (s *Store) IncreaseBet(betID string, to int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	bet, ok := s.bets[betID]
	if !ok {
		return ErrBetNotFound
	}

	if bet.Status != types.BetStatusPlaced {
		return ErrBetNotActive
	}

	prediction, ok := s.predictions[bet.PredictionID]
	if !ok {
		return ErrPredictionNotFound
	}
	if prediction.Status != types.PredictionStatusOpen {
		return ErrPredictionNotOpen
	}

	if bet.Amount == to {
		// no change needed
		return nil
	}
	if bet.Amount > to {
		return ErrBetAlreadyHigher
	}

	difference := to - bet.Amount
	if difference <= 0 {
		return ErrBetAlreadyHigher // sanity check
	}

	logID, err := NewID()
	if err != nil {
		return err
	}

	s.dirty = true

	err = s.applyTokenLogLocked(types.TokenLog{
		ID:           logID,
		CreatedAt:    time.Now().Format(time.RFC3339),
		UserID:       bet.UserID,
		Change:       -difference,
		Cause:        types.TokenChangeCauseBetPlaced,
		BetID:        bet.ID,
		PredictionID: bet.PredictionID,
	})
	if err != nil {
		return err
	}

	bet.Amount = to
	s.bets[bet.ID] = bet

	return nil
}

func (s *Store) listBetsByPredictionLocked(predictionID string) []types.Bet {
	bets := make([]types.Bet, 0)
	for _, bet := range s.bets {
		if bet.PredictionID == predictionID {
			bets = append(bets, bet)
		}
	}
	return bets
}

func (s *Store) ListBetsByUser(userID string) []types.Bet {
	s.lock.RLock()
	defer s.lock.RUnlock()

	bets := make([]types.Bet, 0)
	for _, bet := range s.bets {
		if bet.UserID == userID {
			bets = append(bets, bet)
		}
	}
	return bets
}

func (s *Store) getUserBetOnPredictionLocked(userID, predictionID string) (types.Bet, bool) {
	for _, bet := range s.bets {
		if bet.UserID == userID && bet.PredictionID == predictionID {
			return bet, true
		}
	}
	return types.Bet{}, false
}
