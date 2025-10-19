package ws_room

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Controller struct {
	hub *Hub
}

func NewController(hub *Hub) *Controller {
	return &Controller{
		hub: hub,
	}
}

func (c *Controller) RegisterRoutes(router *gin.RouterGroup) {
	ws := router.Group("/ws")
	{
		ws.GET("/rooms/:room_id", c.connect)
	}
}

type ConnectRequest struct {
	UserToken string `header:"X-user-token" binding:"required"`
}

func (c *Controller) connect(ctx *gin.Context) {
	roomCode := ctx.Param("room_id")

	userToken := ctx.Query("token")
	if userToken == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "token query parameter required"})
		return
	}

	status, err := c.hub.usecase.Status(ctx, roomCode)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	isOwner, _ := c.hub.usecase.IsOwner(ctx, roomCode, userToken)
	role := "participant"
	if isOwner {
		role = "owner"
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		c.hub.logger.Error("failed to upgrade connection", "error", err)
		return
	}

	client := &Client{
		hub:      c.hub,
		conn:     conn,
		send:     make(chan Event, 256),
		userID:   userToken,
		roomCode: roomCode,
		role:     role,
	}

	c.hub.register <- client
	fmt.Println("6")

	go client.writePump()
	go client.readPump()

	c.hub.logger.Info("WebSocket connection established",
		"user_id", userToken,
		"room", roomCode,
		"role", role,
		"status", status)
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	for {
		var event Event
		err := c.conn.ReadJSON(&event)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.hub.logger.Error("WebSocket read error", "error", err)
			}
			break
		}

		c.handleEvent(event)
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()

	for event := range c.send {
		if err := c.conn.WriteJSON(event); err != nil {
			c.hub.logger.Error("WebSocket write error", "error", err)
			return
		}
	}

	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

func (c *Client) handleEvent(event Event) {
	switch event.Type {
	case "START_VOTING":

		if c.role != "owner" {
			c.send <- Event{
				Type: EventError,
				Payload: map[string]interface{}{
					"message": "Only room owner can start voting",
				},
			}
			return
		}

		err := c.hub.StartVoting(c.roomCode, c.userID)
		if err != nil {
			c.send <- Event{
				Type: EventError,
				Payload: map[string]interface{}{
					"message": "Failed to start voting: " + err.Error(),
				},
			}
			return
		}

		c.hub.logger.Info("voting started",
			"room", c.roomCode,
			"initiated_by", c.userID)

	default:
		c.send <- Event{
			Type: EventError,
			Payload: map[string]interface{}{
				"message": "Unknown event type: " + event.Type,
			},
		}
	}
}
