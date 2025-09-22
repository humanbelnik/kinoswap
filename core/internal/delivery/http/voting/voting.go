package http_voting

import (
	"log/slog"
	"net/http"

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
	voting.POST("/start", c.startVoting)

	voting.POST("/finish", c.finishVoting)
	voting.GET("/results", c.votingResults)
	voting.POST("/restart", c.restartVoting)
	voting.POST("/continue", c.continueVoting)
}

// @Summary Start voting session
// @Description Start a new voting session in the room
// @Tags voting
// @Accept json
// @Produce json
// @Param room_id path string true "Room ID"
// @Success 200 {object} map[string]interface{} "Voting started"
// @Failure 404 {object} map[string]string "Room not found"
// @Failure 409 {object} map[string]string "Voting already started"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /rooms/{room_id}/voting/start [post]
func (c *Controller) startVoting(ctx *gin.Context) {
	roomID := ctx.Param("room_id")

	votingID := roomID

	m := ws_room.Message{
		Type:   ws_room.VotingStarted,
		RoomID: roomID,
		Data: map[string]interface{}{
			"voting_id":   votingID,
			"redirect_to": "/voting/" + votingID,
		},
	}
	c.hub.BroadcastToRoom(roomID, m)

	ctx.JSON(http.StatusOK, gin.H{
		"voting_id": votingID,
		"status":    "started",
		"message":   "Voting session started successfully",
	})
}

// @Summary Finish voting session
// @Description Finish the current voting session
// @Tags voting
// @Accept json
// @Produce json
// @Param room_id path string true "Room ID"
// @Success 200 {object} map[string]string "Voting finished"
// @Failure 404 {object} map[string]string "Room or voting not found"
// @Failure 400 {object} map[string]string "Voting not started"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /rooms/{room_id}/voting/finish [post]
func (c *Controller) finishVoting(ctx *gin.Context) {

	panic("unimplemented")

}

// @Summary Get voting results
// @Description Get results of the current or finished voting session
// @Tags voting
// @Produce json
// @Param room_id path string true "Room ID"
// @Success 200 {object} map[string]interface{} "Voting results"
// @Failure 404 {object} map[string]string "Room or voting not found"
// @Failure 400 {object} map[string]string "No voting results available"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /rooms/{room_id}/voting/results [get]
func (c *Controller) votingResults(ctx *gin.Context) {
	panic("unimplemented")

}

// @Summary Restart voting session
// @Description Restart the voting session with same options
// @Tags voting
// @Accept json
// @Produce json
// @Param room_id path string true "Room ID"
// @Success 200 {object} map[string]interface{} "Voting restarted"
// @Failure 404 {object} map[string]string "Room not found"
// @Failure 400 {object} map[string]string "No active voting to restart"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /rooms/{room_id}/voting/restart [post]
func (c *Controller) restartVoting(ctx *gin.Context) {
	panic("unimplemented")
}

// @Summary Continue voting session
// @Description Continue paused voting session
// @Tags voting
// @Accept json
// @Produce json
// @Param room_id path string true "Room ID"
// @Success 200 {object} map[string]interface{} "Voting continued"
// @Failure 404 {object} map[string]string "Room not found"
// @Failure 400 {object} map[string]string "No paused voting to continue"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /rooms/{room_id}/voting/continue [post]
func (c *Controller) continueVoting(ctx *gin.Context) {
	panic("unimplemented")
}

// @Summary Submit vote
// @Description Submit a vote in the current voting session
// @Tags voting
// @Accept json
// @Produce json
// @Param room_id path string true "Room ID"
// @Param vote body map[string]interface{} true "Vote data"
// @Success 201 {object} map[string]string "Vote submitted"
// @Failure 400 {object} map[string]string "Invalid vote data"
// @Failure 404 {object} map[string]string "Room or voting not found"
// @Failure 409 {object} map[string]string "User already voted"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /rooms/{room_id}/voting/votes [post]
func (c *Controller) vote(ctx *gin.Context) {
	panic("unimplemented")
}

func (c *Controller) cards(ctx *gin.Context) {
	panic("unimplemented")
}
