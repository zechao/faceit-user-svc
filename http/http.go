package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zechao/faceit-user-svc/errors"
	"github.com/zechao/faceit-user-svc/log"
)

// handlerError handles the error response for the http handlers
func handlerError(ctx *gin.Context, err error) {
	svcErr := new(errors.Error)
	if errors.As(err, &svcErr) {
		log.Warn(ctx.Request.Context(), "service error", slog.Any("error", svcErr))
		ctx.JSON(svcErr.Code, svcErr)
		return
	}

	ctx.JSON(http.StatusInternalServerError, errors.NewInternal(
		"unexpected internal server error",
	))
	log.Error(ctx.Request.Context(), "internal server error", slog.Any("error", err))
}
