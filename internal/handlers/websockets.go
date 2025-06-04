package handlers

import (
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type WsHandlers struct {
	conn    *websocket.Conn
	log     *zap.SugaredLogger
	wsChan  chan WsPayload
	clients map[*websocket.Conn]string
}

type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

type WsPayload struct {
	Action   string          `json:"action"`
	Username string          `json:"username"`
	Message  string          `json:"message"`
	Conn     *websocket.Conn `json:"-"`
}

//func NewWebsocketsHandlers(log *zap.SugaredLogger) *WsHandlers {
//	var wsChan = make(chan WsPayload)
//	var clients = make(map[WebSocketConnection]string)
//
//	return &WsHandlers{
//		log:     log,
//		wsChan:  wsChan,
//		clients: clients,
//	}
//}
//
//type WebSocketConnection struct {
//	*websocket.Conn
//}
//

//
//func WsEndpoint(w http.ResponseWriter, r *http.Request) {
//	ws, err := upgradeConnection.Upgrade(w, r, nil)
//	if err != nil {
//		log.Println(err)
//		return
//	}
//
//	log.Println("Client connected to endpoint")
//
//	var response WsJsonResponse
//	response.Message = `<em><small>Connected to server</small></em>`
//
//	conn := WebSocketConnection{Conn: ws}
//	clients[conn] = ""
//
//	err = ws.WriteJSON(response)
//	if err != nil {
//		log.Println(err)
//		return
//	}
//
//	go ListenForWs(&conn)
//}
//
//func ListenForWs(conn *WebSocketConnection) {
//	defer func() {
//		if r := recover(); r != nil {
//			log.Println("Errorn ListenForWs", fmt.Sprintf("%v", r))
//		}
//	}()
//
//	var payload WsPayload
//
//	for {
//		err := conn.ReadJSON(&payload)
//		if err != nil {
//			// do nothing
//		} else {
//			payload.Conn = *conn
//			wsChan <- payload
//		}
//	}
//}
//
//func ListenToWsChannel() {
//	var response WsJsonResponse
//	for {
//		e := <-wsChan
//
//		switch e.Action {
//		case "username":
//			clients[e.Conn] = e.Username
//			users := getUserList()
//			response.Action = "list_users"
//			response.ConnectedUsers = users
//			broadcastToAll(response)
//
//		case "left":
//			response.Action = "list_users"
//			delete(clients, e.Conn)
//			users := getUserList()
//			response.ConnectedUsers = users
//			broadcastToAll(response)
//
//		case "broadcast":
//			response.Action = "broadcast"
//			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.Username, e.Message)
//			broadcastToAll(response)
//		}
//	}
//}
//
//func getUserList() []string {
//	var userList []string
//	for _, c := range clients {
//		if c != "" {
//			userList = append(userList, c)
//		}
//	}
//	sort.Strings(userList)
//	return userList
//}
//
//func broadcastToAll(response WsJsonResponse) {
//	for client := range clients {
//		err := client.WriteJSON(response)
//		if err != nil {
//			log.Println("websocket err")
//			_ = client.Close()
//			delete(clients, client)
//		}
//	}
//}
//
//func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
//	view, err := views.GetTemplate(tmpl)
//	if err != nil {
//		log.Println(err)
//		return err
//	}
//
//	err = view.Execute(w, data, nil)
//	if err != nil {
//		log.Println(err)
//		return err
//	}
//
//	return nil
//}
