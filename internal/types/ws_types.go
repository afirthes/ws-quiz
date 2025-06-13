package types

type WsError struct {
	WsPayload
	Error string `json:"error"`
}

type WsPayload struct {
	Action string `json:"action"`
}

type WsConnectedResponse struct {
	WsPayload
	UserId   string `json:"user_id"`
	UserName string `json:"user_name"`
}

type WsStartQuizRequest struct {
	WsPayload
	QuizId string `json:"quiz_id"`
}

type WsStartQuizResponse struct {
	WsPayload
	QuizId     string `json:"quiz_id"`
	GSessionId string `json:"gsession_id"`
}

type WsEnterQuizRequest struct {
	WsPayload
	GSessionId string `json:"gsession_id"`
}

type WsEnterQuizResponse struct {
	WsPayload
	UserId     string `json:"user_id"`
	QuizId     string `json:"quiz_id"`
	GSessionId string `json:"gsession_id"`
}

type WsEnteredQuizBroadcast struct {
	WsPayload
	UserId     string `json:"user_id"`
	UserName   string `json:"user_name"`
	GSessionId string `json:"gsession_id"`
}

type WsLeavedQuizBroadcast struct {
	WsPayload
	UserId     string `json:"user_id"`
	UserName   string `json:"user_name"`
	GSessionId string `json:"gsession_id"`
}

type WsNextQuizQuestionRequest struct {
	WsPayload
	QuestionId    string   `json:"question_id"`
	GSessionId    string   `json:"gsession_id"`
	Question      string   `json:"question"`
	Answers       []string `json:"answers"`
	CorrectAnswer int      `json:"correct_answer"`
	Cost          int      `json:"cost"`
}

type WsNextQuizQuestionBroadcast struct {
	WsPayload
	QuestionId string   `json:"question_id"`
	GSessionId string   `json:"gsession_id"`
	Question   string   `json:"question"`
	Answers    []string `json:"answers"`
	Cost       int      `json:"cost"`
}

type WsAnswerRequest struct {
	WsPayload
	QuestionId string `json:"question_id"`
	GSessionId string `json:"gsession_id"`
	Answer     int    `json:"answer"`
}

type WsParticipantAnsweredBroadcast struct {
	WsPayload
	GSessionId string `json:"gsession_id"`
	QuestionId string `json:"question_id"`
	UserName   string `json:"user_name"`
	UserId     string `json:"user_id"`
	Correct    bool   `json:"correct"`
}

type WsFinishQuestionRequest struct {
	WsPayload
	QuestionId string `json:"question_id"`
	GSessionId string `json:"gsession_id"`
}

type Score struct {
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
	Score    int    `json:"score"`
}

type WsFinishQuestionBroadcast struct {
	WsPayload
	QuestionId string  `json:"question_id"`
	GSessionId string  `json:"gsession_id"`
	Scores     []Score `json:"scores"`
}

type WsFinishQuizRequest struct {
	WsPayload
	GSessionId string `json:"gsession_id"`
	UserID     string `json:"user_id"`
}

type WsFinishQuizBroadcast struct {
	WsPayload
	GSessionId string  `json:"gsession_id"`
	Scores     []Score `json:"scores"`
}
