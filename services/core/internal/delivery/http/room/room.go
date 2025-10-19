package http_room

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	http_common "github.com/humanbelnik/kinoswap/core/internal/delivery/http/common"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	usecase_room "github.com/humanbelnik/kinoswap/core/internal/usecase/room"
)

type Controller struct {
	usecase *usecase_room.Usecase
	logger  *slog.Logger
}

func New(usecase *usecase_room.Usecase) *Controller {
	return &Controller{
		usecase: usecase,
		logger:  slog.Default(),
	}
}

func (c *Controller) RegisterRoutes(router *gin.RouterGroup) {
	rooms := router.Group("/rooms")
	{
		rooms.POST("", c.book)
		rooms.GET("/:room_id/status", c.status)
		rooms.POST("/:room_id/participations", c.participate)
		rooms.DELETE("/:room_id", c.free)
	}
}

// BookResponseDTO DTO для ответа создания комнаты
type BookResponseDTO struct {
	RoomCode string `json:"room_code"`
}

// Book создает новую комнату
// @Summary Создание комнаты
// @Description Создает новую комнату для выбора фильмов
// @Tags Rooms
// @Accept json
// @Produce json
// @Success 201 "Комната успешно создана"
// @Header 201 {string} X-user-token "Токен владельца комнаты"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Failure 503 {object} http_common.ErrorResponse "Ресур недоступен"
// @Router /rooms [post]
func (c *Controller) book(ctx *gin.Context) {
	roomCode, ownerToken, err := c.usecase.Book(ctx)
	if err != nil {
		c.logger.Error("failed to book room", slog.String("error", err.Error()))
		switch err {
		case usecase_room.ErrInternal:
			ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
				Message: "internal error",
			})
		case usecase_room.ErrRoomsUnavailable:
			ctx.JSON(http.StatusServiceUnavailable, http_common.ErrorResponse{
				Message: "unavailable",
			})
		}
		return
	}

	ctx.Header("X-user-token", ownerToken)
	ctx.JSON(http.StatusCreated, BookResponseDTO{
		RoomCode: roomCode,
	})
}

type StatusResponseDTO struct {
	Status string `json:"status"`
}

// Status возвращает статус комнаты
// @Summary Получение статуса комнаты
// @Description Возвращает текущий статус комнаты
// @Tags Rooms
// @Param code path string true "Код комнаты"
// @Success 200 {object} StatusResponseDTO "Статус комнаты"
// @Failure 404 {object} ErrorResponseDTO "Комната не найдена"
// @Failure 500 {object} ErrorResponseDTO "Внутренняя ошибка сервера"
// @Router /rooms/{code}/status [get]
func (c *Controller) status(ctx *gin.Context) {
	code := ctx.Param("room_id")

	status, err := c.usecase.Status(ctx, code)
	if err != nil {
		c.logger.Error("failed to get status", slog.String("error", err.Error()))
		if errors.Is(err, usecase_room.ErrResourceNotFound) {
			ctx.JSON(http.StatusNotFound, http_common.ErrorResponse{
				Message: "not found",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
			Message: "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, StatusResponseDTO{
		Status: status,
	})
}

// Free освобождает комнату
// @Summary Удаление комнаты
// @Description Удаляет комнату по коду
// @Tags Rooms
// @Param code path string true "Код комнаты"
// @Success 201 "Комната успешно удалена"
// @Failure 404 {object} ErrorResponseDTO "Комната не найдена"
// @Failure 401 {object} http_common.ErrorResponse "Не авторизован"
// @Failure 500 {object} ErrorResponseDTO "Внутренняя ошибка сервера"
// @Security UserToken
// @Router /rooms/{code} [delete]
func (c *Controller) free(ctx *gin.Context) {
	code := ctx.Param("room_id")

	userToken := ctx.GetHeader("X-user-token")
	if userToken == "" {
		ctx.JSON(http.StatusUnauthorized, http_common.ErrorResponse{
			Message: "X-user-token not found",
		})
		return
	}
	isOwner, err := c.usecase.IsOwner(ctx, code, userToken)
	if err != nil {
		if errors.Is(err, usecase_room.ErrResourceNotFound) {
			c.logger.Error("failed to free room", slog.String("error", err.Error()))
			ctx.JSON(http.StatusNotFound, http_common.ErrorResponse{
				Message: "not found",
			})
			return
		}
		c.logger.Error("failed to free room", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
			Message: "internal error",
		})
		return
	}

	if !isOwner {
		ctx.JSON(http.StatusUnauthorized, http_common.ErrorResponse{
			Message: "unauthorized",
		})
		return
	}

	err = c.usecase.Free(ctx, code)
	if err != nil {
		if errors.Is(err, usecase_room.ErrResourceNotFound) {
			c.logger.Error("failed to free room", slog.String("error", err.Error()))
			ctx.JSON(http.StatusNotFound, http_common.ErrorResponse{
				Message: "not found",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
			Message: "internal error",
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

// ParticipateRequestDTO DTO для участия в комнате
type ParticipateRequestDTO struct {
	Preference model.Preference `json:"preference" binding:"required"`
}

// ParticipateResponseDTO DTO для ответа участия
type ParticipateResponseDTO struct {
	UserID  string `json:"user_id"`
	Message string `json:"message"`
}

// Participate добавляет участника в комнату
// @Summary Участие в комнате
// @Description Добавляет участника с предпочтениями в комнату
// @Tags Rooms
// @Accept json
// @Param room_id path string true "Код комнаты"
// @Param request body ParticipateRequestDTO true "Данные участника"
// @Success 201 {object} ParticipateResponseDTO "Участник успешно добавлен"
// @Header 201 {string} X-user-token "Токен пользователя"
// @Failure 404 {object} http_common.ErrorResponse "Комната не найдена"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/{room_id}/participate [post]
func (c *Controller) participate(ctx *gin.Context) {
	code := ctx.Param("room_id")
	var req ParticipateRequestDTO

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, http_common.ErrorResponse{
			Message: "invalid request format",
		})
		return
	}

	userToken := ctx.GetHeader("X-user-token")
	var userIDPtr *string
	if userToken != "" {
		userIDPtr = &userToken
	}

	returnedUserID, err := c.usecase.Participate(ctx, code, req.Preference, userIDPtr)
	if err != nil {
		if errors.Is(err, usecase_room.ErrResourceNotFound) {
			c.logger.Error("failed to participate in room", slog.String("error", err.Error()))
			ctx.JSON(http.StatusNotFound, http_common.ErrorResponse{
				Message: "not found",
			})
			return
		}
		c.logger.Error("failed to participate in room", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
			Message: "internal error",
		})
		return
	}

	if userToken == "" {
		ctx.Header("X-user-token", returnedUserID)
	}

	ctx.Status(http.StatusCreated)
}
