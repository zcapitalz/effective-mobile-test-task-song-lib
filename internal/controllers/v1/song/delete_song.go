package songcontroller

import (
	"context"
	"errors"
	ginutils "song-lib/internal/controllers/api-utils/gin-utils"
	"song-lib/internal/domain"
	"song-lib/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
)

// @Summary	Delete song
// @Tags		song
// @Param		songID	path		string				true	"Song ID"
// @Success	200		{nil}		nil					"Success"
// @Failure	404		{object}	apiutils.HTTPError	"Song not found"
// @Failure	500		{object}	apiutils.HTTPError	"Internal server error"
// @Router		/songs/{songID} [delete]
func (ctr *SongController) deleteSong(c *gin.Context) {
	songID := c.MustGet("songID").(ksuid.KSUID)

	ctx := utils.PassContextLogger(c, context.Background())
	err := ctr.songService.DeleteSong(ctx, songID)
	switch {
	case errors.Is(err, domain.ErrSongNotFound):
		ginutils.NotFoundError(c, err)
		return
	case err != nil:
		ginutils.InternalError(c)
		return
	}
}
