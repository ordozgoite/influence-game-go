package realtime

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	GameID   string
	PlayerID string
}

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string][]*Client // gameID -> clients
}

var Manager = NewRoomManager()

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string][]*Client),
	}
}

func (m *RoomManager) AddClient(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.rooms[c.GameID] = append(m.rooms[c.GameID], c)
}

func (m *RoomManager) RemoveClient(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	clients := m.rooms[c.GameID]
	newList := make([]*Client, 0, len(clients))
	for _, cl := range clients {
		if cl != c {
			newList = append(newList, cl)
		}
	}
	if len(newList) == 0 {
		delete(m.rooms, c.GameID)
	} else {
		m.rooms[c.GameID] = newList
	}
}

func (m *RoomManager) Broadcast(gameID string, msg []byte) {
	m.mu.RLock()
	clients := m.rooms[gameID]
	m.mu.RUnlock()

	for _, c := range clients {
		_ = c.Conn.WriteMessage(websocket.TextMessage, msg)
	}
}
