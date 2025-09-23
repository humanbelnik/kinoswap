package ws_room

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/humanbelnik/kinoswap/core/internal/model"
)

type MessageType string

const (
	VotingStarted MessageType = "voting_started"
)

type Message struct {
	Type   MessageType            `json:"type"`
	RoomID string                 `json:"room_id"`
	Data   map[string]interface{} `json:"data,omitempty"`
}

type Client struct {
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	RoomID model.RoomID
}

type Hub struct {
	mu sync.RWMutex

	// Keep track of sets of Clinets within each room
	rooms map[model.RoomID]map[*Client]bool

	logger *slog.Logger
}

func New(logger *slog.Logger) *Hub {
	return &Hub{
		rooms:  make(map[model.RoomID]map[*Client]bool),
		logger: logger,
	}
}

func (h *Hub) RegisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.rooms[client.RoomID]; !ok {
		h.rooms[client.RoomID] = make(map[*Client]bool)
	}
	h.rooms[client.RoomID][client] = true

	h.logger.Info("client registered", "room_id", client.RoomID)
}

func (h *Hub) RemoveClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if room, ok := h.rooms[client.RoomID]; ok {
		delete(room, client)
		if len(room) == 0 {
			delete(h.rooms, client.RoomID)
		}
	}
	h.logger.Info("client unregistered", "room_id", client.RoomID)
}

func (h *Hub) BroadcastToRoom(roomID model.RoomID, message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	messageBytes, _ := json.Marshal(message)

	if clients, ok := h.rooms[roomID]; ok {
		for client := range clients {
			select {
			case client.Send <- messageBytes:
			default:
				close(client.Send)
				delete(h.rooms[roomID], client)
			}
		}
	}
}

func (h *Hub) StartClientReading(client *Client) {
	defer func() {
		h.RemoveClient(client)
		client.Conn.Close()
	}()

	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (h *Hub) StartClientWriting(client *Client) {
	defer client.Conn.Close()

	for message := range client.Send {
		err := client.Conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
}
