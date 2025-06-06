package types

type WsError struct {
	WsPayload
	Error string `json:"error"`
}

type WsPayload struct {
	Action string `json:"action"`
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
