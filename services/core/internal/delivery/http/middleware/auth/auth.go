package http_auth_middleware

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	http_common "github.com/humanbelnik/kinoswap/core/internal/delivery/http/common"
	auth_client "github.com/humanbelnik/kinoswap/core/internal/infra/auth"
)

type Middleware struct {
	client auth_client.AuthClient
	logger *slog.Logger
}

func New(
	client auth_client.AuthClient,
) *Middleware {
	return &Middleware{
		client: client,
		logger: slog.Default(),
	}
}

func (m *Middleware) AuthRequired() gin.HandlerFunc {
	const header = "X-admin-token"
	return func(ctx *gin.Context) {
		t := ctx.GetHeader(header)
		m.logger.Info("auth middleware", slog.String("token", t))
		if t == "" {
			m.logger.Error(fmt.Sprintf("no %s header\n", header))
			ctx.JSON(http.StatusUnauthorized, http_common.ErrorResponse{
				Message: fmt.Sprintf("no %s header\n", header),
			})
			ctx.Abort()
			return
		}

		valid, err := m.client.ValidateToken(t)
		if err != nil {
			m.logger.Error("interna; error", slog.String("error", err.Error()))
			ctx.JSON(http.StatusInternalServerError, http_common.ErrorResponse{
				Message: "internal error",
			})
			ctx.Abort()
			return
		}
		if !valid {
			m.logger.Error("invalid token", slog.String("provided token", t))
			ctx.JSON(http.StatusUnauthorized, http_common.ErrorResponse{
				Message: "invalid token",
			})
			ctx.Abort()
			return
		}
		ctx.Next()
	}
}
