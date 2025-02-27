package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zechao/faceit-user-svc/errors"
)

// handlerError handles the error response for the http handlers
func handlerError(ctx *gin.Context, err error) {
	svcErr := new(errors.Error)
	if errors.As(err, &svcErr) {
		ctx.JSON(svcErr.Code, svcErr)
		return
	}
	ctx.JSON(http.StatusInternalServerError, errors.NewInternal(
		err.Error(),
	))
}

