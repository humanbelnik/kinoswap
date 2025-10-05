package http_room

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	http_common "github.com/humanbelnik/kinoswap/core/internal/delivery/http/common"
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
	room.GET("", c.isRoomAcquired)
	room.POST("/participation", c.participate)
	room.DELETE("", c.release)
}

// AcquireRoomResponseDTO
type AcquireRoomResponseDTO struct {
	RoomID string `json:"room_id" example:"123456"`
}

// @Summary Выделить комнату для голосования
// @Description Выделяет команату для голосования и возвращет ее идентификатор
// @Tags Rooms opertaions
// @Accept json
// @Produce json
// @Success 200 {object} AcquireRoomResponseDTO "Комната успешно создана"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms [get]
func (c *Controller) acquireRoom(ctx *gin.Context) {
	roomID, err := c.uc.AcquireRoom(ctx.Request.Context())
	if err != nil {
		c.logger.Error("failed to acquire room",
			slog.String("error", err.Error()),
		)
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{Message: "internal error"})
		return
	}

	ctx.JSON(http.StatusOK, AcquireRoomResponseDTO{RoomID: string(roomID)})
}

// @Summary Проверяет доступ к комнате
// @Description Проверяет доступ к комнате
// @Tags Rooms opertaions
// @Produce json
// @Param room_id path string true "Идентификатор комнаты" example("123456")
// @Success 200 "Комната существует"
// @Failure 403 {objet} http_common.ErrorResponse "Комната закрыта"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/{room_id} [get]
func (c *Controller) isRoomAcquired(ctx *gin.Context) {
	roomID := ctx.Param("room_id")
	ok, err := c.uc.IsRoomAcquired(ctx.Request.Context(), model.RoomID(roomID))
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	if !ok {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "room not found"})
		return
	}

	ctx.Status(http.StatusOK)
}

// ParticipateRequestDTO
type ParticipateRequestDTO struct {
	Text string `json:"text"`
}

// @Summary Участие в комнате
// @Tags Rooms opertaions
// @Description Позволяет пользователю присоединиться к голосованию с указнием пожеланий
// @Accept json
// @Produce json
// @Param room_id path string true "Идентификатор комнаты" example("123456")
// @Param request body ParticipateRequestDTO true "Предпочтения пользователя"
// @Success 202 "Участник добавлен в пул голосующих"
// @Failure 400 {object} http_common.ErrorResponse "Некорректный формат тела запроса"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/{room_id}/participation [post]
func (c *Controller) participate(ctx *gin.Context) {
	roomID := ctx.Param("room_id")

	var req ParticipateRequestDTO
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

	ctx.Status(http.StatusAccepted)
}

// @Summary Освобождение комнаты
// @Description Освобождает ресурсы комнаты и делает её готовй для последующей резервации
// @Tags Rooms opertaions
// @Produce json
// @Param room_id path string true "Идентификатор комнаты" example("123456")
// @Success 200 "Комната успешно освобождена"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/{room_id} [delete]
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
