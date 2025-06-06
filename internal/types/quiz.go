package types

type Quiz struct {
	QuizId      string
	GameSession string
}

type Participant struct {
	UserId     string
	UserName   string
	GSessionId string
	IsHost     bool
	Score      int
}

type GameSession struct {
	Creator       *Participant
	QuizId        string
	GameSessionId string
	IsFinished    bool
}

type Question struct {
	QuestionId    string
	Question      string
	Answers       []string
	CorrectAnswer int
	Cost          int
	IsFinished    bool
}
