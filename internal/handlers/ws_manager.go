package handlers

import (
	"github.com/gorilla/websocket"
	"sync"
)

type WSManager struct {
	conns sync.Map
}

func NewConnectionManager() *WSManager {
	return &WSManager{}
}

func (cm *WSManager) Add(id string, conn *websocket.Conn) {
	cm.conns.Store(id, conn)
}

func (cm *WSManager) Remove(id string) {
	cm.conns.Delete(id)
}

func (cm *WSManager) Get(id string) (*websocket.Conn, bool) {
	conn, ok := cm.conns.Load(id)
	if !ok {
		return nil, false
	}
	return conn.(*websocket.Conn), true
}

func (cm *WSManager) Broadcast(message []byte) {
	cm.conns.Range(func(_, value any) bool {
		conn := value.(*websocket.Conn)
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			// TODO: log error
			return false // stop iteration on error (optional)
		}
		return true
	})
}
