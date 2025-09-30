# WebSocket Backend Implementation Guide

## Dependencies

First, add the WebSocket package to your Go project:

```bash
go get github.com/gorilla/websocket
```

## WebSocket Hub Implementation

Create a WebSocket hub to manage connections for each room:

```go
// websocket.go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "sync"

    "github.com/gin-gonic/gin"
    "github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        // Allow connections from any origin in development
        // In production, add proper origin checking
        return true
    },
}

type Hub struct {
    // Registered clients for each room
    rooms map[string]map[*Client]bool
    
    // Register requests from the clients
    register chan *Client
    
    // Unregister requests from clients
    unregister chan *Client
    
    // Inbound messages from the clients
    broadcast chan *Message
    
    mutex sync.RWMutex
}

type Client struct {
    hub    *Hub
    conn   *websocket.Conn
    send   chan []byte
    roomID string
}

type Message struct {
    Type   string      `json:"type"`
    RoomID string      `json:"room_id"`
    Data   interface{} `json:"data"`
}

// WebSocket message types
const (
    MSG_PARTICIPANT_JOINED = "participant_joined"
    MSG_PARTICIPANT_READY  = "participant_ready"
    MSG_VOTING_STARTED     = "voting_started"
    MSG_ROOM_STATUS        = "room_status"
)

func NewHub() *Hub {
    return &Hub{
        rooms:      make(map[string]map[*Client]bool),
        register:   make(chan *Client),
        unregister: make(chan *Client),
        broadcast:  make(chan *Message),
    }
}

func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mutex.Lock()
            if h.rooms[client.roomID] == nil {
                h.rooms[client.roomID] = make(map[*Client]bool)
            }
            h.rooms[client.roomID][client] = true
            h.mutex.Unlock()
            
            log.Printf("Client connected to room %s", client.roomID)

        case client := <-h.unregister:
            h.mutex.Lock()
            if clients, ok := h.rooms[client.roomID]; ok {
                if _, ok := clients[client]; ok {
                    delete(clients, client)
                    close(client.send)
                    
                    // Clean up empty rooms
                    if len(clients) == 0 {
                        delete(h.rooms, client.roomID)
                    }
                }
            }
            h.mutex.Unlock()
            
            log.Printf("Client disconnected from room %s", client.roomID)

        case message := <-h.broadcast:
            h.mutex.RLock()
            clients := h.rooms[message.RoomID]
            h.mutex.RUnlock()
            
            data, err := json.Marshal(message)
            if err != nil {
                log.Printf("Error marshaling message: %v", err)
                continue
            }
            
            for client := range clients {
                select {
                case client.send <- data:
                default:
                    close(client.send)
                    delete(clients, client)
                }
            }
        }
    }
}

// Broadcast message to all clients in a room
func (h *Hub) BroadcastToRoom(roomID string, messageType string, data interface{}) {
    message := &Message{
        Type:   messageType,
        RoomID: roomID,
        Data:   data,
    }
    
    select {
    case h.broadcast <- message:
    default:
        log.Printf("Failed to broadcast message to room %s", roomID)
    }
}

// Handle WebSocket connection
func (c *Client) writePump() {
    defer c.conn.Close()
    
    for {
        select {
        case message, ok := <-c.send:
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            
            c.conn.WriteMessage(websocket.TextMessage, message)
        }
    }
}

func (c *Client) readPump() {
    defer func() {
        c.hub.unregister <- c
        c.conn.Close()
    }()
    
    for {
        // Just read messages to keep connection alive
        // We don't expect clients to send messages via WebSocket
        _, _, err := c.conn.ReadMessage()
        if err != nil {
            break
        }
    }
}
```

## Controller Updates

Update your controller to include WebSocket endpoint and hub integration:

```go
// Add to your Controller struct
type Controller struct {
    uc     YourUseCase
    logger *slog.Logger
    hub    *Hub  // Add this
}

// Add WebSocket endpoint to your routes
func (c *Controller) RegisterRoutes(router *gin.RouterGroup) {
    rooms := router.Group("/rooms")
    rooms.GET("", c.acquireRoom)
    rooms.GET("/:room_id/ws", c.handleWebSocket) // Add this

    room := router.Group("rooms/:room_id")
    room.GET("/acquired", c.isRoomAcquired)
    room.PATCH("/participate", c.participate)
    room.PATCH("/start", c.start)
    room.PATCH("/release", c.release)
}

// WebSocket handler
func (c *Controller) handleWebSocket(ctx *gin.Context) {
    roomID := ctx.Param("room_id")
    
    conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
    if err != nil {
        c.logger.Error("Failed to upgrade to WebSocket", 
            slog.String("error", err.Error()))
        return
    }
    
    client := &Client{
        hub:    c.hub,
        conn:   conn,
        send:   make(chan []byte, 256),
        roomID: roomID,
    }
    
    client.hub.register <- client
    
    // Start goroutines for this client
    go client.writePump()
    go client.readPump()
}

// Update your existing methods to broadcast WebSocket messages

func (c *Controller) participate(ctx *gin.Context) {
    roomID := ctx.Param("room_id")
    var requestDTO struct {
        Text string `json:"text"`
    }
    
    if err := ctx.ShouldBindJSON(&requestDTO); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{"error": "incorrect request"})
        return
    }
    
    err := c.uc.Participate(ctx.Request.Context(), roomID, model.Preference{
        Text: requestDTO.Text,
    })
    if err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
        return
    }
    
    // Broadcast to WebSocket clients
    c.hub.BroadcastToRoom(roomID, MSG_PARTICIPANT_READY, map[string]interface{}{
        "message": "participant ready",
    })
    
    ctx.Status(http.StatusOK)
}

func (c *Controller) start(ctx *gin.Context) {
    roomID := ctx.Param("room_id")
    
    // Your existing start logic here...
    
    // After starting, broadcast to all clients in the room
    c.hub.BroadcastToRoom(roomID, MSG_VOTING_STARTED, map[string]interface{}{
        "message": "voting started",
        "redirect_url": "/voting",
    })
    
    ctx.Status(http.StatusOK)
}
```

## Main Application Setup

Initialize the hub in your main application:

```go
func main() {
    // Initialize hub
    hub := NewHub()
    go hub.Run()
    
    // Pass hub to your controller
    controller := &Controller{
        uc:     yourUseCase,
        logger: yourLogger,
        hub:    hub,
    }
    
    // Setup routes...
    router := gin.Default()
    api := router.Group("/api")
    controller.RegisterRoutes(api)
    
    router.Run(":8080")
}
```

## WebSocket URL

With this setup, your WebSocket endpoint will be:
```
ws://localhost:8080/api/rooms/{room_id}/ws
```

The frontend will connect to this endpoint and receive real-time updates when:
- Someone joins the room
- Someone becomes ready
- Host starts voting (triggering auto-redirect)
