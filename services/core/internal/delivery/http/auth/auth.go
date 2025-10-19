package http_auth

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	http_common "github.com/humanbelnik/kinoswap/core/internal/delivery/http/common"
	servie_simple_auth "github.com/humanbelnik/kinoswap/core/internal/service/auth/simple"
)

type Controller struct {
	service *servie_simple_auth.Service
	logger  *slog.Logger
}

func New(
	service *servie_simple_auth.Service,
) *Controller {
	return &Controller{
		service: service,
		logger:  slog.Default(),
	}
}

func (c *Controller) RegisterRoutes(router *gin.RouterGroup) {
	auth := router.Group("/auth")
	auth.POST("", c.auth)
}

// AuthRequestDTO DTO для запроса аутентификации
type AuthRequestDTO struct {
	Code string `json:"code" binding:"required" example:"secret123"`
}

// Auth выполняет аутентификацию пользователя
// @Summary Аутентификация пользователя
// @Description Проверяет код и возвращает токен в заголовке X-admin-token для доступа к защищенным endpoint
// @Tags Auth operations
// @Accept json
// @Produce json
// @Param request body AuthRequestDTO true "Данные для аутентификации"
// @Success 202
// @Header 202 {string} X-admin-token "Токен для доступа к операцями с фильмами"
// @Failure 400 {object} http_common.ErrorResponse "Неверный формат запроса"
// @Failure 403 {object} http_common.ErrorResponse "Неверный код аутентификации"
// @Failure 500 {object} http_common.ErrorResponse "Внутренняя ошибка сервера"
// @Router /auth [post]
func (c *Controller) auth(ctx *gin.Context) {
	var req AuthRequestDTO

	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.logger.Warn("invalid request format", "error", err)
		ctx.JSON(http.StatusBadRequest, http_common.ErrorResponse{
			Message: "Invalid request format",
		})
		return
	}

	token, err := c.service.Auth(req.Code)
	if err != nil {
		switch err {
		case servie_simple_auth.ErrWrongCode:
			c.logger.Warn("wrong code", slog.String("code", token))
			ctx.JSON(http.StatusForbidden, http_common.ErrorResponse{
				Message: "forbidden",
			})
		default:
			c.logger.Error("internal auth error", slog.String("error", err.Error()))
			ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
				Message: "internal error",
			})
		}
		return
	}

	ctx.Header("X-admin-token", token)

	ctx.Status(http.StatusAccepted)

}
