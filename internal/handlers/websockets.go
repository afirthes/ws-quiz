package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/afirthes/ws-quiz/internal/errors"
	"github.com/afirthes/ws-quiz/internal/services"
	"github.com/afirthes/ws-quiz/internal/types"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"log"
	"net/http"
)

type ParticipantWithConn struct {
	*types.Participant
	Conn *websocket.Conn
}

type WsPayloadWithParticipant struct {
	msg         []byte
	participant *ParticipantWithConn
}

type WsHandlers struct {
	conn         *websocket.Conn
	log          *zap.SugaredLogger
	errorHandler *errors.ErrorHandler
	wsChan       chan *WsPayloadWithParticipant
	clients      map[string]*ParticipantWithConn

	userService *services.UserService
	quizService *services.QuizService
}

func NewWebsocketsHandlers(log *zap.SugaredLogger, eh *errors.ErrorHandler, us *services.UserService, qs *services.QuizService) *WsHandlers {
	var wsChan = make(chan *WsPayloadWithParticipant)
	var clients = make(map[string]*ParticipantWithConn)

	return &WsHandlers{
		log:          log,
		wsChan:       wsChan,
		clients:      clients,
		errorHandler: eh,
		userService:  us,
		quizService:  qs,
	}
}

func (wsh *WsHandlers) WsEndpoint(w http.ResponseWriter, r *http.Request) {
	userUUID := r.URL.Query().Get("user-id")
	userName := r.URL.Query().Get("user-name")

	if userUUID == "" {
		wsh.errorHandler.BadRequestResponse(w, r, errors.ErrorUserIDRequired)
		return
	}

	if userName == "" {
		wsh.errorHandler.BadRequestResponse(w, r, errors.ErrorUserNameRequired)
		return
	}

	p := &types.Participant{
		UserName: userName,
		UserId:   userUUID,
	}

	// Check userid - username already taken by other user
	err := wsh.userService.RegisterUser(p)
	if err != nil {
		wsh.errorHandler.BadRequestResponse(w, r, err)
		return
	}

	// Check if connection already established
	if _, ok := wsh.clients[p.UserId]; ok {
		log.Printf("Client %s already connected", userName)
		wsh.errorHandler.ForbiddenResponse(w, r)
		return
	}

	wsConn, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	pc := &ParticipantWithConn{
		Participant: p,
		Conn:        wsConn,
	}

	wsConn.SetCloseHandler(func(p *ParticipantWithConn) func(code int, text string) error {
		return func(code int, text string) error {
			log.Printf("Websocket connection closed, code: %d, text: %s", code, text)
			delete(wsh.clients, p.UserName)
			err := wsh.quizService.LeaveAllGameSessions(p.Participant)
			if err != nil {
				log.Printf("Error leaving sessions for user %s: %v", p.UserId, err)
			}

			// Broadcast new user leaved
			broadcast := types.WsLeavedQuizBroadcast{
				WsPayload:  types.WsPayload{Action: "LEAVED_QUIZ_BROADCAST"},
				UserId:     p.UserId,
				UserName:   p.UserName,
				GSessionId: p.GSessionId,
			}
			wsh.broadcastAll(broadcast, p.GSessionId)
			return nil
		}
	}(pc))

	log.Printf("Client connected to server: %s %s ", userName, userUUID)

	var response types.WsPayload
	response.Action = "CONNECTED"

	if _, ok := wsh.clients[p.UserId]; !ok {
		wsh.clients[p.UserId] = pc
	} else {
		log.Printf("Client %s already connected", userName)
		return
	}

	err = wsConn.WriteJSON(response)
	if err != nil {
		log.Println(err)
		return
	}

	go wsh.ListenForWs(pc)
}

func (wsh *WsHandlers) ListenForWs(pc *ParticipantWithConn) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Errorn ListenForWs", fmt.Sprintf("%v", r))
		}
	}()

	for {
		messageType, data, err := pc.Conn.ReadMessage()
		if err != nil {
			// do nothing, a lot of garbage logs
			continue
		}

		if messageType == websocket.TextMessage || messageType == websocket.BinaryMessage {
			wsh.wsChan <- &WsPayloadWithParticipant{
				msg:         data,
				participant: pc,
			}
		}
	}
}

func (wsh *WsHandlers) handleMessage(p *WsPayloadWithParticipant) error {
	var base types.WsPayload
	if err := json.Unmarshal(p.msg, &base); err != nil {
		return fmt.Errorf("failed to parse base: %w", err)
	}

	switch base.Action {
	case "START_QUIZ":
		var req types.WsStartQuizRequest
		if err := json.Unmarshal(p.msg, &req); err != nil {
			return fmt.Errorf("failed to parse WsStartQuizRequest: %w", err)
		}
		gameSession, err := wsh.quizService.StartQuiz(req.QuizId, p.participant.Participant)
		if err != nil {
			log.Printf("Failed to start quiz: %v", err)
			return err
		}

		p.participant.GSessionId = gameSession.GameSessionId
		p.participant.IsHost = true

		log.Printf("Starting quiz with id: %+v", gameSession)
		res := &types.WsStartQuizResponse{
			WsPayload:  types.WsPayload{Action: "QUIZ_STARTED"},
			QuizId:     gameSession.QuizId,
			GSessionId: gameSession.GameSessionId,
		}
		err = p.participant.Conn.WriteJSON(res)
		if err != nil {
			return err
		}
		return nil

	case "ENTER_QUIZ":
		var req types.WsEnterQuizRequest
		if err := json.Unmarshal(p.msg, &req); err != nil {
			return fmt.Errorf("failed to parse WsStartQuizRequest: %w", err)
		}
		err := wsh.quizService.JoinGameSession(req.GSessionId, p.participant.Participant)
		if err != nil {
			return err
		}

		p.participant.GSessionId = req.GSessionId
		p.participant.IsHost = false

		log.Printf("User %s joining quiz with id: %s", p.participant.UserName, req.GSessionId)
		res := types.WsEnterQuizResponse{
			WsPayload:  types.WsPayload{Action: "ENTERED_QUIZ"},
			UserId:     p.participant.UserId,
			GSessionId: req.GSessionId,
		}
		err = p.participant.Conn.WriteJSON(res)
		if err != nil {
			return err
		}

		// Broadcast new user entered
		broadcast := types.WsEnteredQuizBroadcast{
			WsPayload:  types.WsPayload{Action: "ENTERED_QUIZ_BROADCAST"},
			UserId:     p.participant.UserId,
			UserName:   p.participant.UserName,
			GSessionId: req.GSessionId,
		}
		wsh.broadcastAll(broadcast, req.GSessionId)
		return nil

	case "NEXT_QUESTION":
		var req types.WsNextQuizQuestionRequest
		if err := json.Unmarshal(p.msg, &req); err != nil {
			return fmt.Errorf("failed to parse WsStartQuizRequest: %w", err)
		}

		wsh.quizService.AddQuestion(&types.Question{
			QuestionId:    req.QuestionId,
			Question:      req.Question,
			Answers:       req.Answers,
			CorrectAnswer: req.CorrectAnswer,
			Cost:          req.Cost,
			IsFinished:    false,
		})

		broadcast := types.WsNextQuizQuestionBroadcast{
			WsPayload:  types.WsPayload{Action: "NEXT_QUESTION_BROADCAST"},
			QuestionId: req.QuestionId,
			GSessionId: req.GSessionId,
			Question:   req.Question,
			Answers:    req.Answers,
			Cost:       req.Cost,
		}
		wsh.broadcastParticipants(broadcast, req.GSessionId)
		return nil

	case "ANSWER_QUESTION":
		var req types.WsAnswerRequest
		if err := json.Unmarshal(p.msg, &req); err != nil {
			return fmt.Errorf("failed to parse WsStartQuizRequest: %w", err)
		}

		broadcast := types.WsParticipantAnsweredBroadcast{
			WsPayload:  types.WsPayload{Action: "QUESTION_ANSWERED_BROADCAST"},
			GSessionId: req.GSessionId,
			QuestionId: req.QuestionId,
			Correct:    wsh.quizService.CheckAnsSaveAnswer(req.QuestionId, req.Answer, p.participant.Participant),
			UserId:     p.participant.UserId,
			UserName:   p.participant.UserName,
		}
		wsh.broadcastAll(broadcast, req.GSessionId)
		return nil

	case "FINISH_QUESTION":
		var req types.WsFinishQuestionRequest
		if err := json.Unmarshal(p.msg, &req); err != nil {
			return fmt.Errorf("failed to parse WsStartQuizRequest: %w", err)
		}

		q := wsh.quizService.FinishQuestion(req.QuestionId)
		answers := wsh.quizService.GetAnswers(req.QuestionId)

		n := len(answers)
		if n == 0 {
			return nil
		}

		broadcast := types.WsFinishQuestionBroadcast{
			WsPayload:  types.WsPayload{Action: "QUESTION_FINISHED_BROADCAST"},
			QuestionId: req.QuestionId,
			GSessionId: req.GSessionId,
			Scores:     make([]types.Score, 0, n),
		}

		step := float64(q.Cost) / float64(n)
		for i, userID := range answers {
			score := float64(q.Cost) - step*float64(i)
			if score < 0 {
				score = 0
			}
			if u, ok := wsh.clients[userID]; ok {
				u.Score += int(score)
				broadcast.Scores = append(broadcast.Scores, types.Score{
					UserName: u.UserName,
					UserID:   u.UserId,
					Score:    int(score),
				})
			}
		}

		wsh.broadcastAll(broadcast, req.GSessionId)
		return nil

	case "FINISH_QUIZ_SESSION":
		var req types.WsFinishQuizRequest
		if err := json.Unmarshal(p.msg, &req); err != nil {
			return fmt.Errorf("failed to parse WsStartQuizRequest: %w", err)
		}

		err := wsh.quizService.FinishQuiz(req.GSessionId)
		if err != nil {
			return fmt.Errorf("failed to finish quiz: %w", err)
		}

		ps := wsh.quizService.GetParticipants(req.GSessionId)

		broadcast := types.WsFinishQuizBroadcast{
			WsPayload:  types.WsPayload{Action: "QUESTION_FINISHED_BROADCAST"},
			GSessionId: req.GSessionId,
			Scores:     make([]types.Score, 0, len(ps)),
		}
		for _, p := range ps {
			if u, ok := wsh.clients[p.UserId]; ok {
				broadcast.Scores = append(broadcast.Scores, types.Score{
					UserName: u.UserName,
					UserID:   u.UserId,
					Score:    u.Score,
				})
			}
		}

		wsh.broadcastAll(broadcast, req.GSessionId)
		return nil

	default:
		return fmt.Errorf("unknown action: %s", base.Action)
	}
}

func (wsh *WsHandlers) broadcastParticipants(json any, gsessionId string) {
	wsh.broadcast(wsh.quizService.GetParticipants(gsessionId), json)
}

func (wsh *WsHandlers) broadcastAll(json any, gsessionId string) {
	piqs, err := wsh.quizService.GetParticipantsWithCreator(gsessionId)
	if err != nil {
		log.Printf("Unable to broadcast %v", err)
	}
	wsh.broadcast(piqs, json)
}

func (wsh *WsHandlers) broadcast(piqs []*types.Participant, json any) {
	for _, piq := range piqs {
		if v, ok := wsh.clients[piq.UserId]; ok {
			err := v.Conn.WriteJSON(json)
			if err != nil {
				log.Printf("Failed to broadcast to user %s: %v", piq.UserName, err)
			}
		}
	}
}

func (wsh *WsHandlers) ListenToWsChannel() {
	for {
		e := <-wsh.wsChan

		err := wsh.handleMessage(e)
		if err != nil {
			wse := types.WsError{
				WsPayload: types.WsPayload{
					Action: "ERROR",
				},
				Error: err.Error(),
			}
			err := e.participant.Conn.WriteJSON(wse)
			if err != nil {
				log.Printf("Failed to send message: %v", err)
			}
		}
	}
}
