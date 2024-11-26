package songcontroller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	ginutils "song-lib/internal/controllers/api-utils/gin-utils"
	"song-lib/internal/domain"
	"song-lib/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
)

type getSongCoupletsRequestQuery struct {
	Page    int `form:"couplets_page" binding:"required"`
	PerPage int `form:"couplets_per_page" binding:"required"`
}

type getSongCoupletsResponseBody struct {
	SongCouplets []string `json:"songCouplets"`
}

//	@Summary		Get song text
//	@Description	Get song text with optional pagination by couplets
//	@Tags			song
//	@Produce		json
//	@Param			songID				path		string						true	"Song ID"
//	@Param			couplets_page		query		int							true	"Number of page with couplets to return"
//	@Param			couplets_per_page	query		int							true	"Number of couplets per page to return"
//	@Success		200					{object}	getSongCoupletsResponseBody	"Success"
//	@Failure		404					{object}	apiutils.HTTPError			"Song not found"
//	@Failure		500					{object}	apiutils.HTTPError			"Internal server error"
//	@Router			/songs/{songID}/couplets [get]
func (ctr *SongController) getSongCouplets(c *gin.Context) {
	songID := c.MustGet("songID").(ksuid.KSUID)
	var reqQuery getSongCoupletsRequestQuery
	err := c.ShouldBindQuery(&reqQuery)
	if err != nil {
		ginutils.BindQueryError(c, err)
		return
	}
	if err := reqQuery.validate(); err != nil {
		ginutils.BindQueryError(c, err)
		return
	}

	ctx := utils.PassContextLogger(c, context.Background())
	songCouplets, err := ctr.songService.GetSongCoupletsPaginated(
		ctx,
		songID,
		domain.Pagination{
			Page:    reqQuery.Page - 1,
			PerPage: reqQuery.PerPage})
	switch {
	case errors.Is(err, domain.ErrSongNotFound):
		ginutils.NotFoundError(c, err)
		return
	case err != nil:
		ginutils.InternalError(c)
		return
	}

	if songCouplets == nil {
		songCouplets = make([]string, 0)
	}
	c.JSON(http.StatusOK, getSongCoupletsResponseBody{
		SongCouplets: songCouplets,
	})
}

func (q *getSongCoupletsRequestQuery) validate() error {
	if q.Page < 1 {
		return fmt.Errorf("page value is less than 1")
	}
	if q.PerPage < 0 {
		return fmt.Errorf("per page value is less than 0")
	}

	return nil
}
