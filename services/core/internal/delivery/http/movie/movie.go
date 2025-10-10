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

// CreateMovieRequestDTO представляет запрос на создание фильма
type CreateMovieRequestDTO struct {
	Title      string   `json:"title" binding:"required" example:"Интерстеллар"`
	Year       int      `json:"year" binding:"required" example:"2014"`
	Rating     float64  `json:"rating" binding:"required" example:"8.6"`
	Genres     []string `json:"genres" binding:"required" example:"фантастика,драма,приключения"`
	Overview   string   `json:"overview" binding:"required" example:"Захватывающая история о путешествии через червоточину..."`
	PosterLink string   `json:"poster_link" binding:"required" example:"https://example.com/poster.jpg"`
}

// UpdateMovieRequestDTO представляет запрос на обновление фильма
type UpdateMovieRequestDTO struct {
	Title      string   `json:"title" example:"Интерстеллар (обновленное)"`
	Year       int      `json:"year" example:"2014"`
	Rating     float64  `json:"rating" example:"8.7"`
	Genres     []string `json:"genres" example:"фантастика,драма,приключения"`
	Overview   string   `json:"overview" example:"Обновленное описание фильма..."`
	PosterLink string   `json:"poster_link" example:"https://example.com/new-poster.jpg"`
}

// MovieResponseDTO представляет ответ с данными фильма
type MovieResponseDTO struct {
	ID         uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Title      string    `json:"title" example:"Интерстеллар"`
	Year       int       `json:"year" example:"2014"`
	Rating     float64   `json:"rating" example:"8.6"`
	Genres     []string  `json:"genres" example:"фантастика,драма,приключения"`
	Overview   string    `json:"overview" example:"Захватывающая история о путешествии через червоточину..."`
	PosterLink string    `json:"poster_link" example:"https://example.com/poster.jpg"`
}

// MoviesListResponseDTO DTO для списка фильмов
type MoviesListResponseDTO struct {
	Movies []MovieResponseDTO `json:"movies"`
	Total  int                `json:"total"`
}

func (r *CreateMovieRequestDTO) ConvertToMovieMeta() model.MovieMeta {
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

func (r *UpdateMovieRequestDTO) ConvertToMovieMeta(id uuid.UUID) model.MovieMeta {
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

func ConvertFromMovieMeta(meta model.MovieMeta) MovieResponseDTO {
	return MovieResponseDTO{
		ID:         meta.ID,
		Title:      meta.Title,
		Year:       meta.Year,
		Rating:     meta.Rating,
		Genres:     meta.Genres,
		Overview:   meta.Overview,
		PosterLink: meta.PosterLink,
	}
}

func ConvertFromMovieMetaList(metas []*model.MovieMeta) []MovieResponseDTO {
	movies := make([]MovieResponseDTO, len(metas))
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
	movies.DELETE("/:movie_id", c.deleteMovie)
	movies.PUT("/:movie_id", c.updateMovie)

	movies.GET("rooms/:room_id/voting/movies", c.getMoviesForVoting)
}

// @Summary Создание фильма
// @Description Создает новый фильм в системе
// @Tags Movies operations
// @Accept json
// @Produce json
// @Param request body CreateMovieRequestDTO true "Данные для создания фильма"
// @Success 201 "Фильм успешно создан"
// @Failure 400 {object} http_common.ErrorResponse "Некорректные данные запроса"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /movies [post]
func (c *Controller) createMovie(ctx *gin.Context) {
	var req CreateMovieRequestDTO
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

// GetMovies возвращает список фильмов
// @Summary Получение списка фильмов
// @Description Возвращает список всех фильмов в системе
// @Tags Movies operations
// @Produce json
// @Success 200 {object} MoviesListResponseDTO "Список фильмов успешно получен"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /movies [get]
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

	response := MoviesListResponseDTO{
		Movies: ConvertFromMovieMetaList(movies),
		Total:  len(movies),
	}

	ctx.JSON(http.StatusOK, response)
}

// DeleteMovie удаляет фильм
// @Summary Удаление фильма
// @Description Удаляет фильм по идентификатору
// @Tags Movies operations
// @Produce json
// @Param movie_id path string true "UUID фильма" example("550e8400-e29b-41d4-a716-446655440000")
// @Success 204 "Фильм успешно удален"
// @Failure 400 {object} http_common.ErrorResponse "Некорректный UUID фильма"
// @Failure 404 {object} http_common.ErrorResponse "Фильм не найден"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /movies/{id} [delete]
func (c *Controller) deleteMovie(ctx *gin.Context) {
	idParam := ctx.Param("movie_id")
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

// @Summary Обновление фильма
// @Description Обновляет данные фильма по идентификатору
// @Tags Movies operations
// @Accept json
// @Produce json
// @Param movie_id path string true "UUID фильма" example("550e8400-e29b-41d4-a716-446655440000")
// @Param request body UpdateMovieRequestDTO true "Данные для обновления фильма"
// @Success 200 "Фильм успешно обновлен"
// @Failure 400 {object} http_common.ErrorResponse "Некорректные данные запроса"
// @Failure 404 {object} http_common.ErrorResponse "Фильм не найден"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /movies/{id} [put]
func (c *Controller) updateMovie(ctx *gin.Context) {
	idParam := ctx.Param("movie_id")
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

	var req UpdateMovieRequestDTO
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

// @Summary Получение фильмов для голосования
// @Description Возвращает список фильмов доступных для голосования в комнате
// @Tags Movies operations
// @Produce json
// @Param room_id path string true "Идентификатор комнаты" example("550e8400-e29b-41d4-a716-446655440000")
// @Success 200 {object} MoviesListResponseDTO "Список фильмов для голосования"
// @Failure 404 {object} http_common.ErrorResponse "Комната не найдена"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /rooms/{room_id}/voting/movies [get]
func (c *Controller) getMoviesForVoting(ctx *gin.Context) {

}
