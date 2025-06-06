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
}

type GameSession struct {
	Creator       *Participant
	QuizId        string
	GameSessionId string
}
