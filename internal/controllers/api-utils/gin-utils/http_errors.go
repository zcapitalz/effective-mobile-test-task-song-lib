package ginutils

import (
	"net/http"
	apiutils "song-lib/internal/controllers/api-utils"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

func Error(ctx *gin.Context, status int, err error) {
	ctx.JSON(status, apiutils.HTTPError{Message: err.Error()})
}

func BadRequest(ctx *gin.Context, err error) {
	Error(ctx, http.StatusBadRequest, err)
}

func BindJSONError(ctx *gin.Context, err error) {
	BadRequest(ctx, errors.Wrap(err, "parse and validate JSON body"))
}

func BindURIError(ctx *gin.Context, err error) {
	BadRequest(ctx, errors.Wrap(err, "parse and validate URI params"))
}

func BindQueryError(ctx *gin.Context, err error) {
	BadRequest(ctx, errors.Wrap(err, "parse and validate query params"))
}

func InternalError(ctx *gin.Context) {
	Error(ctx, http.StatusInternalServerError, errors.New(""))
}

func NotFoundError(ctx *gin.Context, err error) {
	Error(ctx, http.StatusNotFound, err)
}

func ConflictError(ctx *gin.Context, err error) {
	Error(ctx, http.StatusConflict, err)
}

func BadGateway(ctx *gin.Context) {
	Error(ctx, http.StatusBadGateway, errors.New(""))
}
