package services

import (
	"errors"
	"github.com/afirthes/ws-quiz/internal/types"
	"go.uber.org/zap"
	"sync"
)

var (
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrTypeAssertion     = errors.New("type assertion error")
)

type UserService struct {
	log   *zap.SugaredLogger
	users sync.Map
}

func NewUserService(log *zap.SugaredLogger) *UserService {
	return &UserService{
		log: log,
	}
}

// RegisterUser registers a new user only if a user with the same name does not already exist.
// ErrUserAlreadyExists - if username already taken
// ErrTypeAssertion - type error
func (us *UserService) RegisterUser(p *types.Participant) error {
	if lv, loaded := us.users.LoadOrStore(p.UserName, p); loaded {
		ps, ok := lv.(*types.Participant)
		if !ok {
			return ErrTypeAssertion
		}
		if p.UserId != ps.UserId {
			return ErrUserAlreadyExists
		}
	}
	return nil
}
