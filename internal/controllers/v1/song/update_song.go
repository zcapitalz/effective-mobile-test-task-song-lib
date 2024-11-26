package songcontroller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	apiutils "song-lib/internal/controllers/api-utils"
	ginutils "song-lib/internal/controllers/api-utils/gin-utils"
	"song-lib/internal/domain"
	"song-lib/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
)

//	@Summary		Update song
//	@Description	Update song by passing fields to be updated
//	@Tags			song
//	@Accept			json
//	@Param			songID		path		string					true	"Song ID"
//	@Param			update_info	body		updateSongRequestBody	true	"Song updating details"
//	@Success		200			{nil}		nil						"Success"
//	@Failure		404			{object}	apiutils.HTTPError		"Song not found"
//	@Failure		500			{object}	apiutils.HTTPError		"Internal server error"
//	@Router			/songs/{songID} [put]
func (ctr *SongController) updateSong(c *gin.Context) {
	songID := c.MustGet("songID").(ksuid.KSUID)
	var reqBody updateSongRequestBody
	if err := c.BindJSON(&reqBody); err != nil {
		ginutils.BindJSONError(c, err)
		return
	}
	if err := reqBody.validate(); err != nil {
		ginutils.BindJSONError(c, err)
		return
	}

	songUpdate := domain.SongUpdate(reqBody)
	ctx := utils.PassContextLogger(c, context.Background())
	song, err := ctr.songService.UpdateSong(ctx, songID, &songUpdate)
	switch {
	case errors.Is(err, domain.ErrSongNotFound):
		ginutils.NotFoundError(c, err)
		return
	case err != nil:
		ginutils.InternalError(c)
		return
	}

	c.JSON(http.StatusOK, newSongDTOFromEntity(song))
}

type updateSongRequestBody struct {
	Name        *string   `json:"name"`
	ReleaseDate *string   `json:"releaseDate"`
	Couplets    *[]string `json:"couplets"`
	Link        *string   `json:"link"`
}

func (b *updateSongRequestBody) validate() error {
	if b.Name == nil && b.ReleaseDate == nil &&
		b.Couplets == nil && b.Link == nil {
		return apiutils.ErrUpdateObjectEmpty
	}
	if b.Couplets != nil && len(*b.Couplets) == 0 {
		return fmt.Errorf("song should have at least one couplet")
	}

	return nil
}
