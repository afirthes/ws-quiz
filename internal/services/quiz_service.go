package services

type Quiz struct {
	UUID string
}

type QuizService interface {
	StartQuiz(uuid string)
	NextQuestion(uuid string)
	FinishQuiz(uuid string)
}
