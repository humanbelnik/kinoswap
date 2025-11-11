package http_vote

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	http_common "github.com/humanbelnik/kinoswap/core/internal/delivery/http/common"
	ws_room "github.com/humanbelnik/kinoswap/core/internal/delivery/ws/room"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	usecase_vote "github.com/humanbelnik/kinoswap/core/internal/usecase/vote"
)

type ParticipantValidator interface {
	IsParticipant(ctx context.Context, code string, userID string) (bool, error)
}

type Controller struct {
	uc *usecase_vote.Usecase
	pv ParticipantValidator

	hub *ws_room.Hub

	logger *slog.Logger
}

type ControllerOption func(*Controller)

func WithLogger(logger *slog.Logger) ControllerOption {
	return func(c *Controller) {
		c.logger = logger
	}
}

func New(
	uc *usecase_vote.Usecase,
	pv ParticipantValidator,
	hub *ws_room.Hub,
	opts ...ControllerOption,
) *Controller {
	c := &Controller{
		uc:     uc,
		pv:     pv,
		hub:    hub,
		logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Controller) RegisterRoutes(router *gin.RouterGroup) {
	room := router.Group("rooms/:room_id")
	room.GET("/movies", c.getMovies)
	room.GET("/results", c.getResults)
	room.PATCH("/results", c.vote)
}

func (c *Controller) validateParticipant(ctx *gin.Context) (string, string, bool) {
	roomID := ctx.Param("room_id")
	userToken := ctx.GetHeader("X-user-token")

	if userToken == "" {
		ctx.JSON(http.StatusUnauthorized, http_common.ErrorResponse{
			Message: "X-user-token header required",
		})
		return "", "", false
	}

	isParticipant, err := c.pv.IsParticipant(ctx, roomID, userToken)
	if err != nil {
		if errors.Is(err, usecase_vote.ErrResourceNotFound) {
			c.logger.Error("room not found", slog.String("room_id", roomID), slog.String("error", err.Error()))
			ctx.JSON(http.StatusNotFound, http_common.ErrorResponse{
				Message: "not found",
			})
			return "", "", false
		}
		c.logger.Error("failed to validate participant", slog.String("room_id", roomID), slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
			Message: "internal error",
		})
		return "", "", false
	}

	if !isParticipant {
		ctx.JSON(http.StatusForbidden, http_common.ErrorResponse{
			Message: "user is not a participant of this room",
		})
		return "", "", false
	}

	return roomID, userToken, true
}

// GetMoviesRequestDTO DTO для запроса фильмов
type GetMoviesRequestDTO struct {
	Count int `form:"count" binding:"required,min=1,max=50"`
}

// GetMoviesResponseDTO DTO для ответа с фильмами
type GetMoviesResponseDTO struct {
	Movies []*model.MovieMeta `json:"movies"`
}

// GetMovies возвращает батч фильмов для голосования
// @Summary Получение батча фильмов для голосования
// @Description Возвращает N фильмов, наиболее подходящих под предпочтения участников комнаты
// @Tags Voting
// @Param room_id path string true "Код комнаты"
// @Param count query int true "Количество фильмов"
// @Success 200 {object} GetMoviesResponseDTO "Батч фильмов успешно получен"
// @Failure 400 {object} http_common.ErrorResponse "Неверный формат запроса"
// @Failure 403 {object} http_common.ErrorResponse "Пользователь не является участником комнаты"
// @Failure 404 {object} http_common.ErrorResponse "Комната не найдена"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Security UserToken
// @Router /rooms/{room_id}/movies [get]
func (c *Controller) getMovies(ctx *gin.Context) {
	roomID, userToken, ok := c.validateParticipant(ctx)
	if !ok {
		return
	}

	var req GetMoviesRequestDTO
	if err := ctx.ShouldBindQuery(&req); err != nil {
		c.logger.Error("invalid request format", slog.String("room_id", roomID), slog.String("user_token", userToken), slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, http_common.ErrorResponse{
			Message: "invalid request format",
		})
		return
	}

	movies, err := c.uc.VotingBatch(ctx, req.Count, roomID)
	if err != nil {
		if errors.Is(err, usecase_vote.ErrResourceNotFound) {
			c.logger.Error("no participants found for voting", slog.String("room_id", roomID), slog.String("error", err.Error()))
			ctx.JSON(http.StatusNotFound, http_common.ErrorResponse{
				Message: "no participants found for voting",
			})
			return
		}
		c.logger.Error("failed to get voting batch", slog.String("room_id", roomID), slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
			Message: "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, GetMoviesResponseDTO{
		Movies: movies,
	})
}

// GetResultsResponseDTO DTO для результатов голосования
type GetResultsResponseDTO struct {
	Results []*model.Result `json:"results"`
}

// GetResults возвращает результаты голосования в комнате
// @Summary Получение результатов голосования
// @Description Возвращает результаты голосования в комнате, отсортированные по количеству лайков
// @Tags Voting
// @Param room_id path string true "Код комнаты"
// @Success 200 {object} GetResultsResponseDTO "Результаты успешно получены"
// @Failure 403 {object} http_common.ErrorResponse "Пользователь не является участником комнаты"
// @Failure 404 {object} http_common.ErrorResponse "Комната не найдена"
// @Security UserToken
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/{room_id}/results [get]
func (c *Controller) getResults(ctx *gin.Context) {
	roomID, _, ok := c.validateParticipant(ctx)
	if !ok {
		return
	}

	results, err := c.uc.Results(ctx, roomID)
	if err != nil {
		if errors.Is(err, usecase_vote.ErrResourceNotFound) {
			c.logger.Error("room not found for results", slog.String("room_id", roomID), slog.String("error", err.Error()))
			ctx.JSON(http.StatusNotFound, http_common.ErrorResponse{
				Message: "not found",
			})
			return
		}
		c.logger.Error("failed to get results", slog.String("room_id", roomID), slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
			Message: "internal error",
		})
		return
	}

	ctx.JSON(http.StatusOK, GetResultsResponseDTO{
		Results: results,
	})
}

// VoteRequestDTO DTO для голосования
type VoteRequestDTO struct {
	Reactions map[uuid.UUID]int `json:"reactions" binding:"required"`
}

// Vote добавляет реакции пользователя к фильмам
// @Summary Добавление реакций к фильмам
// @Description Добавляет реакции пользователя (лайки) к фильмам в рамках комнаты
// @Tags Voting
// @Accept json
// @Param room_id path string true "Код комнаты"
// @Param request body VoteRequestDTO true "Реакции пользователя"
// @Success 302 "Редирект на страницу резульататов"
// @Failure 400 {object} http_common.ErrorResponse "Неверный формат запроса"
// @Failure 403 {object} http_common.ErrorResponse "Пользователь не является участником комнаты"
// @Failure 404 {object} http_common.ErrorResponse "Комната не найдена"
// @Security UserToken
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/{room_id}/results [patch]
func (c *Controller) vote(ctx *gin.Context) {
	roomID, userToken, ok := c.validateParticipant(ctx)
	if !ok {
		return
	}

	var req struct {
		Reactions map[string]int `json:"reactions" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Error("invalid request format",
			slog.String("room_id", roomID),
			slog.String("user_token", userToken),
			slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, http_common.ErrorResponse{
			Message: "invalid request format",
		})
		return
	}

	userUUID, err := uuid.Parse(userToken)
	if err != nil {
		c.logger.Error("invalid user token format",
			slog.String("room_id", roomID),
			slog.String("user_token", userToken),
			slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, http_common.ErrorResponse{
			Message: "invalid user token format",
		})
		return
	}

	reactions := make(map[uuid.UUID]int, len(req.Reactions))

	for movieIDStr, reaction := range req.Reactions {
		if reaction != 0 && reaction != 1 {
			c.logger.Error("invalid reaction value",
				slog.String("room_id", roomID),
				slog.String("movie_id", movieIDStr),
				slog.Int("reaction", reaction))
			ctx.JSON(http.StatusBadRequest, http_common.ErrorResponse{
				Message: "reaction must be 0 or 1",
			})
			return
		}

		movieUUID, err := uuid.Parse(movieIDStr)
		if err != nil {
			c.logger.Error("invalid movie id format",
				slog.String("room_id", roomID),
				slog.String("movie_id", movieIDStr),
				slog.String("error", err.Error()))
			ctx.JSON(http.StatusBadRequest, http_common.ErrorResponse{
				Message: "invalid movie id format",
			})
			return
		}

		reactions[movieUUID] = reaction
	}

	c.logger.Info("processing vote request",
		slog.String("room_id", roomID),
		slog.String("user_id", userUUID.String()),
		slog.Int("reactions_count", len(reactions)),
		slog.Any("reactions", reactions))

	err = c.uc.AddReaction(ctx, roomID, userUUID, model.Reactions{Reactions: reactions})
	if err != nil {
		c.logger.Error(err.Error())
		if errors.Is(err, usecase_vote.ErrResourceNotFound) {
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

	ready, err := c.uc.IsAllReady(ctx, roomID)
	if err != nil {
		c.logger.Error(err.Error())
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
			Message: "internal error",
		})
		return
	}

	if ready {
		c.hub.NotifyVotingComplete(roomID)
	}

	ctx.Status(http.StatusAccepted)
}
