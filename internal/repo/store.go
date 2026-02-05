package repo

import (
	"errors"
	"sync"

	"go.albinodrought.com/creamy-prediction-market/internal/types"
)

type Store struct {
	lock        sync.RWMutex
	users       map[string]types.User
	predictions map[string]types.Prediction
	bets        map[string]types.Bet
	tokenLog    map[string]types.TokenLog
}

var ErrUserNameTaken = errors.New("user name is taken")

func (s *Store) AddUser(u types.User) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	for id := range s.users {
		if u.Name == s.users[id].Name {
			return ErrUserNameTaken
		}
	}

	s.users[u.ID] = u

	return nil
}
