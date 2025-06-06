package types

type QuizStore interface {
	GetGameSessions(userId string) []GameSession
	StartQuiz(quiz Quiz) (GameSession, error)
}
