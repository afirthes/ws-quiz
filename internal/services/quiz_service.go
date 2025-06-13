package services

import (
	"fmt"
	"github.com/afirthes/ws-quiz/internal/types"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
	"sync"
)

type QuizService struct {
	log              *zap.SugaredLogger
	mu               *sync.Mutex
	sessions         map[string]*types.GameSession   //UserId -> GameSession
	participants     map[string][]*types.Participant //GameSession -> Participants
	questions        map[string]*types.Question
	questionsAnswers map[string][]string
}

func NewQuizService(log *zap.SugaredLogger) *QuizService {
	return &QuizService{
		log:              log,
		mu:               &sync.Mutex{},
		sessions:         make(map[string]*types.GameSession),
		participants:     make(map[string][]*types.Participant),
		questions:        make(map[string]*types.Question),
		questionsAnswers: make(map[string][]string),
	}
}

func (qs *QuizService) UpdateScores(scores []types.Score, gsession string) []types.Score {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	result := make([]types.Score, 0)

	qs.log.Info("Calculating scores:", scores)

	participants := qs.participants[gsession]
	for _, p := range participants {
		found := false
		for _, score := range scores {
			if p.UserId == score.UserID {
				found = true
				qs.log.Info("p.Score before", p.Score)
				qs.log.Info("adding score.Score", score.Score)
				p.Score += score.Score
				qs.log.Info("p.Score after", p.Score)
				result = append(result, types.Score{
					UserID:   score.UserID,
					UserName: score.UserName,
					Score:    p.Score,
				})
				qs.log.Infof("Updated score for user %s (session %s): +%d => %d",
					p.UserName, gsession, score.Score, p.Score)
				break
			}
		}
		if !found {
			result = append(result, types.Score{
				UserID:   p.UserId,
				UserName: p.UserName,
				Score:    p.Score,
			})
		}
	}
	return result
}

func (qs *QuizService) AddQuestion(q *types.Question) {
	qs.questions[q.QuestionId] = q
}

func (qs *QuizService) FinishQuestion(questionId string) *types.Question {
	if v, ok := qs.questions[questionId]; ok {
		v.IsFinished = true
		return v
	}
	return nil
}

func (qs *QuizService) CheckAnsSaveAnswer(questionId string, answer int, p *types.Participant) bool {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	if a, ok := qs.questions[questionId]; ok {
		if a.IsFinished {
			return false
		}

		if a.CorrectAnswer == answer {

			// Если уже отвечал - возвращаем true
			for _, pa := range qs.questionsAnswers[questionId] {
				if pa == p.UserId {
					return true
				}
			}

			// добавляем в список ответивших
			qs.questionsAnswers[questionId] = append(qs.questionsAnswers[questionId], p.UserId)
			return true
		}
	}

	return false
}

func (qs *QuizService) GetAnswers(questionId string) []string {
	return qs.questionsAnswers[questionId]
}

func (qs *QuizService) StartQuiz(quizId string, creator *types.Participant) (*types.GameSession, error) {
	if _, ok := qs.sessions[creator.UserId]; ok {
		return nil, fmt.Errorf("user %s has already started quiz %s", creator.UserName, quizId)
	}
	gs := &types.GameSession{
		QuizId:        quizId,
		Creator:       creator,
		GameSessionId: ksuid.New().String(),
		IsFinished:    false,
	}
	qs.sessions[creator.UserId] = gs
	qs.log.Info("Game session created:", gs)
	return gs, nil
}

func (qs *QuizService) FinishQuiz(gsessionId, userId string) error {
	qs.mu.Lock()
	defer qs.mu.Unlock()

	if gs, ok := qs.sessions[userId]; ok && !gs.IsFinished {
		if gs.GameSessionId == gsessionId {
			gs.IsFinished = true
			return nil
		} else {
			return fmt.Errorf("gsession %s cant be found, other session started", gsessionId)
		}
	} else if !ok {
		return fmt.Errorf("gsession %s cant be found", gsessionId)
	} else {
		return fmt.Errorf("gsession %s already finished", gsessionId)
	}
}

func (qs *QuizService) JoinGameSession(gsessionId string, p *types.Participant) (quizId string, err error) {
	found := false
	for _, session := range qs.sessions {
		if session.GameSessionId == gsessionId {
			quizId = session.QuizId
			found = true
			break
		}
	}
	if !found {
		return "", fmt.Errorf("game session %s not found", gsessionId)
	}

	// добавляем участника, если его ещё нет
	participants := qs.participants[gsessionId]
	for _, existing := range participants {
		if existing.UserId == p.UserId {
			return "", fmt.Errorf("user %s already joined session %s", p.UserId, gsessionId)
		}
	}

	qs.participants[gsessionId] = append(qs.participants[gsessionId], p)
	qs.log.Infof("User %s joined session %s", p.UserName, gsessionId)
	return quizId, nil
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
	qs.mu.Lock()
	defer qs.mu.Unlock()

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
	return nil, fmt.Errorf("could not find creator for game session %s", gsessionId)
}
