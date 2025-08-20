package http_room

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	usecase_room "github.com/humanbelnik/kinoswap/core/internal/usecase/room"
)

type Contoller struct {
	uc     *usecase_room.Usecase
	logger *slog.Logger
}

type ControllerOption func(*Contoller)

func WithLogger(logger *slog.Logger) ControllerOption {
	return func(c *Contoller) {
		c.logger = logger
	}
}

func New(uc *usecase_room.Usecase, opts ...ControllerOption) *Contoller {
	c := &Contoller{
		uc:     uc,
		logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Contoller) RegisterRoutes(router *gin.RouterGroup) {
	rooms := router.Group("/rooms", c.acquireRoom)
	rooms.GET("")
}

func (c *Contoller) acquireRoom(ctx *gin.Context) {
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
