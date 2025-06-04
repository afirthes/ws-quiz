package handlers

import (
	"github.com/gorilla/websocket"
	"sync"
)

type WSManager struct {
	conns map[string]*websocket.Conn
	mu    sync.RWMutex
}

func NewConnectionManager() *WSManager {
	return &WSManager{
		conns: make(map[string]*websocket.Conn),
	}
}

func (cm *WSManager) Add(id string, conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.conns[id] = conn
}

func (cm *WSManager) Remove(id string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	delete(cm.conns, id)
}

func (cm *WSManager) Get(id string) (*websocket.Conn, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	conn, ok := cm.conns[id]
	return conn, ok
}

func (cm *WSManager) Broadcast(message []byte) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	for _, conn := range cm.conns {
		err := conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			// TODO: log error
			return
		}
	}
}
