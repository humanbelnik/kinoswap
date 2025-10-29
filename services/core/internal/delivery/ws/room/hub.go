package ws_room

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	usecase_room "github.com/humanbelnik/kinoswap/core/internal/usecase/room"
)

const (
	EventUserJoined       = "USER_JOINED"
	EventLobbyUpdate      = "LOBBY_UPDATE"
	EventStartVoting      = "START_VOTING"
	EventRedirectToVoting = "REDIRECT_TO_VOTING"
	EventVotingFinished   = "VOTING_FINISHED"
	EventError            = "ERROR"
)

type Event struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan Event
	userID   string
	roomCode string
	role     string
}

type roomEvent struct {
	roomCode string
	event    Event
}

type Hub struct {
	usecase    *usecase_room.Usecase
	logger     *slog.Logger
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan roomEvent
	mu         sync.RWMutex
}

func NewHub(usecase *usecase_room.Usecase) *Hub {
	return &Hub{
		usecase:    usecase,
		logger:     slog.Default(),
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan roomEvent),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case roomEvent := <-h.broadcast:
			h.broadcastToRoom(roomEvent.roomCode, roomEvent.event)
		}
	}
}

func (h *Hub) handleRegister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.clients[client] = true
	if _, exists := h.rooms[client.roomCode]; !exists {
		h.rooms[client.roomCode] = make(map[*Client]bool)
	}
	h.rooms[client.roomCode][client] = true

	h.logger.Info("client registered",
		"user_id", client.userID,
		"room", client.roomCode,
		"role", client.role)

	go h.broadcastParticipantsCount(client.roomCode)
}

func (h *Hub) handleUnregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)

		if roomClients, exists := h.rooms[client.roomCode]; exists {
			delete(roomClients, client)
			if len(roomClients) == 0 {
				delete(h.rooms, client.roomCode)
			}
		}
	}

	h.logger.Info("client unregistered",
		"user_id", client.userID,
		"room", client.roomCode)

	if client.roomCode != "" {
		go h.broadcastParticipantsCount(client.roomCode)
	}
}

func (h *Hub) broadcastParticipantsCount(roomCode string) {
	count, err := h.usecase.ParticipantsCount(context.Background(), roomCode)
	if err != nil {
		h.logger.Error("failed to get participants count", "error", err, "room", roomCode)
		return
	}

	h.broadcastToRoom(roomCode, Event{
		Type: EventLobbyUpdate,
		Payload: map[string]interface{}{
			"participants_count": count,
		},
	})
}

func (h *Hub) broadcastToRoom(roomCode string, event Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if roomClients, exists := h.rooms[roomCode]; exists {
		for client := range roomClients {
			select {
			case client.send <- event:
			default:
				close(client.send)
				delete(h.rooms[roomCode], client)
			}
		}
	}
}

func (h *Hub) NotifyUserJoined(roomCode string) {
	h.broadcastParticipantsCount(roomCode)
}

func (h *Hub) StartVoting(roomCode string, userID string) error {
	err := h.usecase.SetStatus(context.Background(), roomCode, model.StatusVoting)
	if err != nil {
		h.logger.Error("failed to set room status", "error", err, "room", roomCode)
		return err
	}

	h.broadcast <- roomEvent{
		roomCode: roomCode,
		event: Event{
			Type: EventRedirectToVoting,
			Payload: map[string]interface{}{
				"initiated_by": userID,
				"room_code":    roomCode,
				"redirect_url": "/voting.html?room=" + roomCode,
			},
		},
	}

	return nil
}

func (h *Hub) NotifyVotingComplete(roomCode string) error {
	h.broadcast <- roomEvent{
		roomCode: roomCode,
		event: Event{
			Type: EventVotingFinished,
			Payload: map[string]interface{}{
				"room_code": roomCode,
				"message":   "All participants have voted",
				"code":      roomCode,
				"timestamp": time.Now().Unix(),
			},
		},
	}

	h.logger.Info("voting complete notification sent",
		"room", roomCode)

	return nil
}
