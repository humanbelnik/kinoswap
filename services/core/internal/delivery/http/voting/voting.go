package http_voting

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	ws_room "github.com/humanbelnik/kinoswap/core/internal/delivery/ws/room"
	usecase_vote "github.com/humanbelnik/kinoswap/core/internal/usecase/vote"
)

type Controller struct {
	uc  *usecase_vote.Usecase
	hub *ws_room.Hub

	logger *slog.Logger
}

type ControllerOption func(*Controller)

func WithLogger(logger *slog.Logger) ControllerOption {
	return func(c *Controller) {
		c.logger = logger
	}
}

func New(uc *usecase_vote.Usecase,
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
	voting := router.Group("rooms/:room_id/voting")
	voting.POST("/votes", c.vote)
	voting.PATCH("/status", c.changeStatus)
	voting.GET("/results", c.votingResults)
}

// VoteRequestDTO
type VoteRequestDTO struct {
	Votes []MovieVoteDTO `json:"votes"`
}

// MovieVoteDTO
type MovieVoteDTO struct {
	MovieID string `json:"movie_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Liked   bool   `json:"liked" example:"true"`
}

// VotingStatusRequestDTO
type VotingStatusRequestDTO struct {
	Status string `json:"status" example:"active" enums:"start,restart,finish"`
}

// MovieResultDTO
type MovieResultDTO struct {
	MovieMetaDTO
	LikesCount int `json:"likes_count" example:"5"`
}

// MovieMetaDTO
type MovieMetaDTO struct {
	ID         string   `json:"id" example:"550e8400-e29b-41d4-a716-446655440000" swaggertype:"string"`
	PosterLink string   `json:"poster_link" example:"https://example.com/poster.jpg"`
	Title      string   `json:"title" example:"Интерстеллар"`
	Genres     []string `json:"genres" example:"фантастика,драма,приключения"`
	Year       int      `json:"year" example:"2014"`
	Rating     float64  `json:"rating" example:"8.6"`
	Overview   string   `json:"overview" example:"Захватывающая история..."`
}

// @Summary Отправить результаты голосования
// @Description Принимает массив голосов (movie_id, liked) и регистрирует их для текущей сессии голосования
// @Tags Voting operations
// @Accept json
// @Produce json
// @Param room_id path string true "Идентификатор комнаты" example("123456")
// @Param request body VoteRequestDTO true "Массив голосов за фильмы"
// @Success 200 "Голоса успешно приняты"
// @Failure 400 {object} http_common.ErrorResponse "Некорректный запрос"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/{room_id}/voting/votes [post]
func (c *Controller) vote(ctx *gin.Context) {
	panic("unimplamented")
}

// @Summary Изменить статус голосования
// @Description Изменяет статус сессии голосования
// @Tags Voting operations
// @Accept json
// @Produce json
// @Param room_id path string true "Идентификатор комнаты" example("123456")
// @Param request body VotingStatusRequestDTO true "Новый статус голосования"
// @Success 200 "Статус успешно обновлен"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/{room_id}/voting/status [patch]
func (c *Controller) changeStatus(ctx *gin.Context) {
	panic("unimplamented")
}

// @Summary Получить результаты голосования
// @Description Возвращает массив фильмов с количеством лайков для текущей или завершенной сессии голосования
// @Tags Voting operations
// @Produce json
// @Param room_id path string true "Идентификатор комнаты" example("123456")
// @Success 200 {array} MovieResultDTO "Результаты голосования"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/{room_id}/voting/results [get]
func (c *Controller) votingResults(ctx *gin.Context) {
	panic("unimplenmented")
}

// // @Summary Start voting session
// // @Description Start a new voting session in the room
// // @Tags voting
// // @Accept json
// // @Produce json
// // @Param room_id path string true "Room ID"
// // @Success 200 {object} map[string]interface{} "Voting started"
// // @Failure 404 {object} map[string]string "Room not found"
// // @Failure 409 {object} map[string]string "Voting already started"
// // @Failure 500 {object} map[string]string "Internal server error"
// // @Router /rooms/{room_id}/voting/start [post]
// func (c *Controller) startVoting(ctx *gin.Context) {
// 	roomID := ctx.Param("room_id")

// 	votingID := roomID

// 	m := ws_room.Message{
// 		Type:   ws_room.VotingStarted,
// 		RoomID: roomID,
// 		Data: map[string]interface{}{
// 			"voting_id":   votingID,
// 			"redirect_to": "/voting/" + votingID,
// 		},
// 	}
// 	c.hub.BroadcastToRoom(model.RoomID(roomID), m)

// 	ctx.JSON(http.StatusOK, gin.H{
// 		"voting_id": votingID,
// 		"status":    "started",
// 		"message":   "Voting session started successfully",
// 	})
// }

// // @Summary Finish voting session
// // @Description Finish the current voting session
// // @Tags voting
// // @Accept json
// // @Produce json
// // @Param room_id path string true "Room ID"
// // @Success 200 {object} map[string]string "Voting finished"
// // @Failure 404 {object} map[string]string "Room or voting not found"
// // @Failure 400 {object} map[string]string "Voting not started"
// // @Failure 500 {object} map[string]string "Internal server error"
// // @Router /rooms/{room_id}/voting/finish [post]
// func (c *Controller) finishVoting(ctx *gin.Context) {

// 	panic("unimplemented")

// }

// // @Summary Get voting results
// // @Description Get results of the current or finished voting session
// // @Tags voting
// // @Produce json
// // @Param room_id path string true "Room ID"
// // @Success 200 {object} map[string]interface{} "Voting results"
// // @Failure 404 {object} map[string]string "Room or voting not found"
// // @Failure 400 {object} map[string]string "No voting results available"
// // @Failure 500 {object} map[string]string "Internal server error"
// // @Router /rooms/{room_id}/voting/results [get]
// func (c *Controller) votingResults(ctx *gin.Context) {
// 	panic("unimplemented")

// }

// // @Summary Restart voting session
// // @Description Restart the voting session with same options
// // @Tags voting
// // @Accept json
// // @Produce json
// // @Param room_id path string true "Room ID"
// // @Success 200 {object} map[string]interface{} "Voting restarted"
// // @Failure 404 {object} map[string]string "Room not found"
// // @Failure 400 {object} map[string]string "No active voting to restart"
// // @Failure 500 {object} map[string]string "Internal server error"
// // @Router /rooms/{room_id}/voting/restart [post]
// func (c *Controller) restartVoting(ctx *gin.Context) {
// 	panic("unimplemented")
// }

// // @Summary Continue voting session
// // @Description Continue paused voting session
// // @Tags voting
// // @Accept json
// // @Produce json
// // @Param room_id path string true "Room ID"
// // @Success 200 {object} map[string]interface{} "Voting continued"
// // @Failure 404 {object} map[string]string "Room not found"
// // @Failure 400 {object} map[string]string "No paused voting to continue"
// // @Failure 500 {object} map[string]string "Internal server error"
// // @Router /rooms/{room_id}/voting/continue [post]
// func (c *Controller) continueVoting(ctx *gin.Context) {
// 	panic("unimplemented")
// }

// // @Summary Submit vote
// // @Description Submit a vote in the current voting session
// // @Tags voting
// // @Accept json
// // @Produce json
// // @Param room_id path string true "Room ID"
// // @Param vote body map[string]interface{} true "Vote data"
// // @Success 201 {object} map[string]string "Vote submitted"
// // @Failure 400 {object} map[string]string "Invalid vote data"
// // @Failure 404 {object} map[string]string "Room or voting not found"
// // @Failure 409 {object} map[string]string "User already voted"
// // @Failure 500 {object} map[string]string "Internal server error"
// // @Router /rooms/{room_id}/voting/votes [post]
// func (c *Controller) vote(ctx *gin.Context) {
// 	panic("unimplemented")
// }

// func (c *Controller) cards(ctx *gin.Context) {
// 	panic("unimplemented")
// }
