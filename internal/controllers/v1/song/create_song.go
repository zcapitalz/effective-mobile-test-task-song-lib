package songcontroller

import (
	"context"
	"errors"
	"net/http"
	ginutils "song-lib/internal/controllers/api-utils/gin-utils"
	"song-lib/internal/domain"
	"song-lib/internal/utils"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

type createSongRequestBody struct {
	SongName       string `json:"song" binding:"required"`
	MusicGroupName string `json:"group" binding:"required"`
}

//	@Summary	Create a new song
//	@Tags		song
//	@Accept		json
//	@Produce	json
//	@Param		song_details	body		createSongRequestBody	yes	"Song details"
//	@Success	201				{object}	songDTO					"Success"
//	@Failure	409				{object}	apiutils.HTTPError		"Song already exists"
//	@Failure	502				{object}	apiutils.HTTPError		"Error from upstream service"
//	@Failure	500				{object}	apiutils.HTTPError		"Internal server error"
//	@Router		/songs [post]
func (ctr *SongController) createSong(c *gin.Context) {
	_, span := otel.Tracer("gin-server").Start(c.Request.Context(), "create song")
	defer span.End()

	var reqBody createSongRequestBody
	err := c.ShouldBindBodyWithJSON(&reqBody)
	if err != nil {
		ginutils.BindJSONError(c, err)
		return
	}

	createSongDTO := domain.CreateSongDTO(reqBody)
	ctx := utils.PassContextLogger(c, context.Background())
	song, err := ctr.songService.CreateSong(ctx, &createSongDTO)
	switch {
	case errors.Is(err, domain.ErrSongAlreadyExists):
		ginutils.ConflictError(c, err)
		return
	case errors.Is(err, domain.ErrIntegration):
		ginutils.BadGateway(c)
		return
	case err != nil:
		ginutils.InternalError(c)
		return
	}

	c.JSON(http.StatusCreated, newSongDTOFromEntity(song))

}
