package http_movie

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ws_room "github.com/humanbelnik/kinoswap/core/internal/delivery/ws/room"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	usecase_movie "github.com/humanbelnik/kinoswap/core/internal/usecase/movie"
)

type CreateMovieRequest struct {
	Title      string   `json:"title" binding:"required"`
	Year       int      `json:"year" binding:"required"`
	Rating     float64  `json:"rating" binding:"required"`
	Genres     []string `json:"genres" binding:"required"`
	Overview   string   `json:"overview" binding:"required"`
	PosterLink string   `json:"poster_link" binding:"required"`
}

type UpdateMovieRequest struct {
	Title      string   `json:"title"`
	Year       int      `json:"year"`
	Rating     float64  `json:"rating"`
	Genres     []string `json:"genres"`
	Overview   string   `json:"overview"`
	PosterLink string   `json:"poster_link"`
}

type MovieResponse struct {
	ID         uuid.UUID `json:"id"`
	Title      string    `json:"title"`
	Year       int       `json:"year"`
	Rating     float64   `json:"rating"`
	Genres     []string  `json:"genres"`
	Overview   string    `json:"overview"`
	PosterLink string    `json:"poster_link"`
}

// MoviesListResponse DTO для списка фильмов
type MoviesListResponse struct {
	Movies []MovieResponse `json:"movies"`
	Total  int             `json:"total"`
}

func (r *CreateMovieRequest) ConvertToMovieMeta() model.MovieMeta {
	return model.MovieMeta{
		ID:         uuid.New(),
		Title:      r.Title,
		Year:       r.Year,
		Rating:     r.Rating,
		Genres:     r.Genres,
		Overview:   r.Overview,
		PosterLink: r.PosterLink,
	}
}

func (r *UpdateMovieRequest) ConvertToMovieMeta(id uuid.UUID) model.MovieMeta {
	return model.MovieMeta{
		ID:         id,
		Title:      r.Title,
		Year:       r.Year,
		Rating:     r.Rating,
		Genres:     r.Genres,
		Overview:   r.Overview,
		PosterLink: r.PosterLink,
	}
}

func ConvertFromMovieMeta(meta model.MovieMeta) MovieResponse {
	return MovieResponse{
		ID:         meta.ID,
		Title:      meta.Title,
		Year:       meta.Year,
		Rating:     meta.Rating,
		Genres:     meta.Genres,
		Overview:   meta.Overview,
		PosterLink: meta.PosterLink,
	}
}

func ConvertFromMovieMetaList(metas []*model.MovieMeta) []MovieResponse {
	movies := make([]MovieResponse, len(metas))
	for i, meta := range metas {
		movies[i] = ConvertFromMovieMeta(*meta)
	}
	return movies
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

type Controller struct {
	uc  *usecase_movie.Usecase
	hub *ws_room.Hub

	logger *slog.Logger
}

type ControllerOption func(*Controller)

func WithLogger(logger *slog.Logger) ControllerOption {
	return func(c *Controller) {
		c.logger = logger
	}
}

func New(uc *usecase_movie.Usecase,
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
	movies := router.Group("/movies")
	movies.POST("", c.createMovie)
	movies.GET("", c.getMovies)
	movies.DELETE("/:id", c.deleteMovie)
	movies.PUT("/:id", c.updateMovie)

	movies.GET("rooms/:room_id/voting/:voting_id/movies", c.getMoviesForVoting)
}

func (c *Controller) createMovie(ctx *gin.Context) {
	var req CreateMovieRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn("invalid request body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Code:  http.StatusBadRequest,
		})
		return
	}

	movieMeta := req.ConvertToMovieMeta()

	if err := c.uc.Upload(ctx.Request.Context(), movieMeta); err != nil {
		c.logger.Error("failed to create movie",
			slog.String("error", err.Error()),
			slog.String("title", req.Title),
		)
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to create movie",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.Status(http.StatusCreated)
}

func (c *Controller) getMovies(ctx *gin.Context) {
	movies, err := c.uc.Load(ctx.Request.Context())
	if err != nil {
		c.logger.Error("failed to load movies", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to load movies",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	response := MoviesListResponse{
		Movies: ConvertFromMovieMetaList(movies),
		Total:  len(movies),
	}

	ctx.JSON(http.StatusOK, response)
}

func (c *Controller) deleteMovie(ctx *gin.Context) {
	idParam := ctx.Param("id")
	movieID, err := uuid.Parse(idParam)
	if err != nil {
		c.logger.Warn("invalid movie ID",
			slog.String("id", idParam),
			slog.String("error", err.Error()),
		)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid movie ID",
			Code:  http.StatusBadRequest,
		})
		return
	}

	if err := c.uc.DeleteMovie(ctx.Request.Context(), movieID); err != nil {
		c.logger.Error("failed to delete movie",
			slog.String("error", err.Error()),
			slog.String("movie_id", movieID.String()),
		)

		if err.Error() == usecase_movie.ErrFailedToLoadMeta.Error() {
			ctx.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Movie not found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to delete movie",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}

func (c *Controller) updateMovie(ctx *gin.Context) {
	idParam := ctx.Param("id")
	movieID, err := uuid.Parse(idParam)
	if err != nil {
		c.logger.Warn("invalid movie ID",
			slog.String("id", idParam),
			slog.String("error", err.Error()),
		)
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid movie ID",
			Code:  http.StatusBadRequest,
		})
		return
	}

	var req UpdateMovieRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn("invalid request body", slog.String("error", err.Error()))
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error: "Invalid request body",
			Code:  http.StatusBadRequest,
		})
		return
	}

	movieMeta := req.ConvertToMovieMeta(movieID)

	if err := c.uc.UpdateMovie(ctx.Request.Context(), movieMeta); err != nil {
		c.logger.Error("failed to update movie",
			slog.String("error", err.Error()),
			slog.String("movie_id", movieID.String()),
		)

		if err.Error() == usecase_movie.ErrFailedToLoadMeta.Error() {
			ctx.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Movie not found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		ctx.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Failed to update movie",
			Message: err.Error(),
			Code:    http.StatusInternalServerError,
		})
		return
	}

	ctx.Status(http.StatusOK)
}

func (c *Controller) getMoviesForVoting(ctx *gin.Context) {

}
