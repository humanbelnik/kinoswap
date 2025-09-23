package http_room

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	ws_room "github.com/humanbelnik/kinoswap/core/internal/delivery/ws/room"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	usecase_room "github.com/humanbelnik/kinoswap/core/internal/usecase/room"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Controller struct {
	uc  *usecase_room.Usecase
	hub *ws_room.Hub

	logger *slog.Logger
}

type ControllerOption func(*Controller)

func WithLogger(logger *slog.Logger) ControllerOption {
	return func(c *Controller) {
		c.logger = logger
	}
}

func New(uc *usecase_room.Usecase,
	hub *ws_room.Hub,
	opts ...ControllerOption) *Controller {
	c := &Controller{
		uc:     uc,
		hub:    hub,
		logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Controller) RegisterRoutes(router *gin.RouterGroup) {
	rooms := router.Group("/rooms")
	rooms.GET("", c.acquireRoom)

	room := router.Group("rooms/:room_id")
	room.GET("/ws", c.roomWS)
	room.GET("/acquired", c.isRoomAcquired)
	room.POST("/participate", c.participate)
	room.POST("/release", c.release)
}

func (c *Controller) acquireRoom(ctx *gin.Context) {
	roomID, err := c.uc.AcquireRoom(ctx.Request.Context())
	if err != nil {
		c.logger.Error("failed to acquire room",
			slog.String("error", err.Error()),
		)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	ctx.JSON(http.StatusOK, struct {
		RoomID string `json:"room_id"`
	}{
		RoomID: string(roomID),
	})
}

func (c *Controller) isRoomAcquired(ctx *gin.Context) {
	roomID := ctx.Param("room_id")
	ok, err := c.uc.IsRoomAcquired(ctx.Request.Context(), model.RoomID(roomID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	ctx.Status(http.StatusOK)
}

type RequestDTO struct {
	Text string `json:"text"`
}

func (c *Controller) participate(ctx *gin.Context) {
	roomID := ctx.Param("room_id")

	var req RequestDTO
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "incorrect request"})
		return
	}
	err := c.uc.Participate(ctx.Request.Context(), model.RoomID(roomID), model.Preference{
		Text: req.Text,
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	ctx.Status(http.StatusOK)
}

func (c *Controller) release(ctx *gin.Context) {
	roomID := ctx.Param("room_id")
	if err := c.uc.ReleaseRoom(ctx, model.RoomID(roomID)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	ctx.Status(http.StatusOK)
}

func (c *Controller) roomWS(ctx *gin.Context) {
	roomID := ctx.Param("room_id")

	ok, err := c.uc.IsRoomAcquired(ctx.Request.Context(), model.RoomID(roomID))
	if err != nil || !ok {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
		return
	}

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		c.logger.Error("failed to upgrade to websocket",
			slog.String("error", err.Error()),
		)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	client := &ws_room.Client{
		Hub:    c.hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		RoomID: model.RoomID(roomID),
	}

	c.hub.RegisterClient(client)

	go c.hub.StartClientReading(client)
	go c.hub.StartClientWriting(client)
}
