package services

import (
	"fmt"
	"github.com/afirthes/ws-quiz/internal/types"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type QuizService struct {
	log          *zap.SugaredLogger
	sessions     map[string]*types.GameSession   //UserId -> GameSession
	participants map[string][]*types.Participant //GameSession -> Participants
}

func NewQuizService(log *zap.SugaredLogger) *QuizService {
	return &QuizService{
		log:          log,
		sessions:     make(map[string]*types.GameSession),
		participants: make(map[string][]*types.Participant),
	}
}

func (qs *QuizService) StartQuiz(quizId string, creator *types.Participant) (*types.GameSession, error) {
	if _, ok := qs.sessions[creator.UserId]; ok {
		return nil, fmt.Errorf("user %s has already started quiz %s", creator.UserName, quizId)
	}
	gs := &types.GameSession{
		QuizId:        quizId,
		Creator:       creator,
		GameSessionId: ksuid.New().String(),
	}
	qs.sessions[creator.UserId] = gs
	return gs, nil
}

func (qs *QuizService) JoinGameSession(gsessionId string, p *types.Participant) error {
	found := false
	for _, session := range qs.sessions {
		if session.GameSessionId == gsessionId {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("game session %s not found", gsessionId)
	}

	// добавляем участника, если его ещё нет
	participants := qs.participants[gsessionId]
	for _, existing := range participants {
		if existing.UserId == p.UserId {
			return fmt.Errorf("user %s already joined session %s", p.UserId, gsessionId)
		}
	}

	qs.participants[gsessionId] = append(qs.participants[gsessionId], p)
	qs.log.Infof("User %s joined session %s", p.UserName, gsessionId)
	return nil
}

func (qs *QuizService) LeaveAllGameSessions(p *types.Participant) error {
	leftCount := 0

	for gsessionId, participants := range qs.participants {
		updated := make([]*types.Participant, 0, len(participants))
		wasInSession := false

		for _, existing := range participants {
			if existing.UserId == p.UserId {
				wasInSession = true
				continue // исключаем этого участника
			}
			updated = append(updated, existing)
		}

		if wasInSession {
			qs.participants[gsessionId] = updated
			leftCount++
			qs.log.Infof("User %s left session %s", p.UserName, gsessionId)
		}
	}

	if leftCount == 0 {
		return fmt.Errorf("user %s was not part of any sessions", p.UserName)
	}
	return nil
}

func (qs *QuizService) LeaveGameSession(gsessionId string, p *types.Participant) error {
	participants, ok := qs.participants[gsessionId]
	if !ok {
		return fmt.Errorf("game session %s not found", gsessionId)
	}

	updated := make([]*types.Participant, 0, len(participants))
	found := false
	for _, existing := range participants {
		if existing.UserId == p.UserId {
			found = true
			continue // пропускаем этого участника
		}
		updated = append(updated, existing)
	}

	if !found {
		return fmt.Errorf("user %s not found in session %s", p.UserId, gsessionId)
	}

	qs.participants[gsessionId] = updated
	qs.log.Infof("User %s left session %s", p.UserName, gsessionId)
	return nil
}

func (qs *QuizService) NextQuestion(uuid string) error {
	return nil
}

func (qs *QuizService) FinishQuiz(uuid string) error {
	return nil
}

func (qs *QuizService) GetParticipants(gsessionId string) []*types.Participant {
	return qs.participants[gsessionId]
}

func (qs *QuizService) GetParticipantsWithCreator(gsessionId string) ([]*types.Participant, error) {
	for _, session := range qs.sessions {
		if session.GameSessionId == gsessionId {
			creator := &types.Participant{
				UserName: session.Creator.UserName,
				UserId:   session.Creator.UserId,
			}
			return append(qs.participants[gsessionId], creator), nil
		}
	}
	return nil, fmt.Errorf("Could not find creator for game session %s", gsessionId)
}
