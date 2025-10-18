package http_movie

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	http_common "github.com/humanbelnik/kinoswap/core/internal/delivery/http/common"
	ws_room "github.com/humanbelnik/kinoswap/core/internal/delivery/ws/room"
	"github.com/humanbelnik/kinoswap/core/internal/model"
	usecase_movie "github.com/humanbelnik/kinoswap/core/internal/usecase/movie"
)

// CreateMovieRequestDTO представляет запрос на создание фильма
type CreateMovieRequestDTO struct {
	Title    string   `json:"title" validate:"required" example:"Интерстеллар"`
	Genres   []string `json:"genres" validate:"required" example:"фантастика,драма,приключения"`
	Overview string   `json:"overview" validate:"required" example:"Захватывающая история о путешествии через червоточину..."`

	Year   int     `json:"year" example:"2014"`
	Rating float64 `json:"rating" example:"8.6"`
}

func (r *CreateMovieRequestDTO) Validate() error {
	return validator.New().Struct(r)
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
		ID:       uuid.New(),
		Title:    r.Title,
		Year:     r.Year,
		Rating:   r.Rating,
		Genres:   r.Genres,
		Overview: r.Overview,
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

type Controller[T model.FileObject] struct {
	uc  *usecase_movie.Usecase[T]
	hub *ws_room.Hub

	logger *slog.Logger
}

type ControllerOption[T model.FileObject] func(*Controller[T])

func WithLogger[T model.FileObject](logger *slog.Logger) ControllerOption[T] {
	return func(c *Controller[T]) {
		c.logger = logger
	}
}

func New[T model.FileObject](uc *usecase_movie.Usecase[T],
	hub *ws_room.Hub,
	opts ...ControllerOption[T]) *Controller[T] {
	c := &Controller[T]{
		uc:     uc,
		hub:    hub,
		logger: slog.Default(),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *Controller[T]) RegisterRoutes(router *gin.RouterGroup) {
	movies := router.Group("/movies")
	movies.Use(c.authMiddleware)

	movies.POST("", c.createMovie)
	movies.GET("", c.getMovies)
	movies.DELETE("/:movie_id", c.deleteMovie)
}

func (c *Controller[T]) authMiddleware(ctx *gin.Context) {
	t := ctx.GetHeader("X-admin-token")
	if t == "" {
		ctx.JSON(http.StatusUnauthorized, http_common.ErrorResponse{
			Message: "X-admin-token invalid",
		})
		ctx.Abort()
		return
	}

	valid, err := c.validateToken(ctx, t)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
			Message: "error on token validation",
		})
		ctx.Abort()
		return
	}
	if !valid {
		ctx.JSON(http.StatusUnauthorized, http_common.ErrorResponse{
			Message: "unauthorized",
		})
		ctx.Abort()
		return
	}
	ctx.Next()
}

func (c *Controller[T]) validateToken(ctx *gin.Context, t string) (bool, error) {
	masterToken := os.Getenv("ADMIN_TOKEN")
	return masterToken == t, nil
}

// @Summary Создание фильма
// @Description Создает новый фильм в системе. Принимает multipart/form-data с JSON данными о фильме и опциональным файлом постера
// @Tags Movies operations
// @Accept multipart/form-data
// @Produce json
// @Param body formData string true "Данные фильма в JSON формате" example({"title":"Inception","year":2010,"rating":8.8,"genres":["sci-fi","action"],"overview":"A thief who steals corporate secrets..."})
// @Param file formData file false "Файл постера "
// @Success 201 "Фильм успешно создан"
// @Failure 400 {object} http_common.ErrorResponse "Некорректные данные запроса: невалидный JSON, отсутствует поле body"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /movies [post]
func (c *Controller[T]) createMovie(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, http_common.ErrorResponse{
			Message: "bad request",
		})
		return
	}

	body := form.Value["body"]
	if len(body) == 0 {
		ctx.JSON(http.StatusBadRequest, http_common.ErrorResponse{
			Message: "empty body",
		})
		return
	}

	var req CreateMovieRequestDTO
	if err := json.Unmarshal([]byte(body[0]), &req); err != nil {
		ctx.JSON(http.StatusBadRequest, http_common.ErrorResponse{
			Message: "invalid body",
		})
		return
	}

	if err := req.Validate(); err != nil {
		ctx.JSON(http.StatusBadRequest, http_common.ErrorResponse{
			Message: "empty fields",
		})
		return
	}

	var posterData []byte
	var posterFilename string
	var poster *model.Poster

	if files := form.File["file"]; len(files) > 0 {
		fileHeader := files[0]
		file, err := fileHeader.Open()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{Message: "error on read file"})
			return
		}
		defer file.Close()

		posterFilename = fileHeader.Filename
		posterData, err = io.ReadAll(file)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{Message: "error on read file"})
			return
		}
		poster = &model.Poster{
			Filename: posterFilename,
			Content:  posterData,
		}
	}

	movieMeta := req.ConvertToMovieMeta()
	movie := model.Movie{
		MM:     &movieMeta,
		Poster: poster,
	}
	if err := c.uc.Upload(ctx.Request.Context(), movie); err != nil {
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
			Message: "Failed to create movie",
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
func (c *Controller[T]) getMovies(ctx *gin.Context) {
	movies, err := c.uc.Load(ctx.Request.Context())
	if err != nil {
		c.logger.Error("failed to load movies", slog.String("error", err.Error()))
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
			Message: "internal error",
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
func (c *Controller[T]) deleteMovie(ctx *gin.Context) {
	movieID, err := uuid.Parse(ctx.Param("movie_id"))
	if err != nil {
		ctx.JSON(http.StatusBadRequest, http_common.ErrorResponse{
			Message: "invalid resource id format",
		})
		return
	}

	if err := c.uc.DeleteMovie(ctx.Request.Context(), movieID); err != nil {
		ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
			Message: "failed to delete resource",
		})
		return
	}

	ctx.Status(http.StatusNoContent)
}
