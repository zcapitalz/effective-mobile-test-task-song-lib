package songcontroller

import (
	"context"
	"fmt"
	"net/http"
	ginutils "song-lib/internal/controllers/api-utils/gin-utils"
	"song-lib/internal/domain"
	"song-lib/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
)

type getSongsRequestQuery struct {
	Page                 *int    `form:"page" binding:"required"`
	PerPage              *int    `form:"per_page" binding:"required"`
	SongName             *string `form:"song"`
	MusicGroupName       *string `form:"group"`
	SongLink             *string `form:"link"`
	SongTextContains     *string `form:"text_contains"`
	SongReleaseDateRange *string `form:"release_date_range"`
}

type getSongsResponseBody struct {
	Songs []songDTO `json:"songs"`
}

type songDTO struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	ReleaseDate string        `json:"releaseDate"`
	Couplets    []string      `json:"couplets"`
	Link        string        `json:"link"`
	MusicGroup  musicGroupDTO `json:"group"`
}

type musicGroupDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

//	@Summary	Get songs
//	@Tags		song
//	@Produce	json
//	@Param		page				query		int						true	"Number of page to return"
//	@Param		per_page			query		int						true	"Number of items per returned page"
//	@Param		song				query		string					false	"Equality filter for name"
//	@Param		group				query		string					false	"Equality filter for music group name"
//	@Param		link				query		string					false	"Equality filter for link"
//	@Param		text_contains		query		string					false	"'in' filter for text"
//	@Param		release_date_range	query		string					false	"'in range' filter for release data e.g., [12-03-2001;21-11-2024]"
//	@Success	200					{object}	getSongsResponseBody	"Success"
//	@Failure	500					{object}	apiutils.HTTPError		"Internal server error"
//	@Router		/songs [get]
func (ctr *SongController) getSongs(c *gin.Context) {
	var reqQuery getSongsRequestQuery
	if err := c.BindQuery(&reqQuery); err != nil {
		ginutils.BindQueryError(c, err)
		return
	}
	if err := reqQuery.validate(); err != nil {
		ginutils.BindQueryError(c, err)
		return
	}

	songFilters, err := reqQuery.toSongFilters()
	if err != nil {
		ginutils.BindQueryError(c, err)
		return
	}
	ctx := utils.PassContextLogger(c, context.Background())
	songs, err := ctr.songService.GetSongsFilteredPaginated(
		ctx,
		songFilters,
		domain.Pagination{
			Page:    *reqQuery.Page - 1,
			PerPage: *reqQuery.PerPage})
	if err != nil {
		ginutils.InternalError(c)
		return
	}

	songDTOs := make([]songDTO, 0, len(songs))
	for _, song := range songs {
		songDTOs = append(songDTOs, songDTO{
			ID:          song.ID.String(),
			Name:        song.Name,
			ReleaseDate: song.ReleaseDate.Format(DateLayout),
			Couplets:    song.Couplets,
			Link:        song.Link,
			MusicGroup: musicGroupDTO{
				ID:   song.MusicGroup.ID.String(),
				Name: song.MusicGroup.Name,
			},
		})
	}
	c.JSON(http.StatusOK, &getSongsResponseBody{
		Songs: songDTOs,
	})
}

func (q *getSongsRequestQuery) validate() error {
	if *q.Page < 1 {
		return fmt.Errorf("page value is less than 1")
	}
	if *q.PerPage < 0 {
		return fmt.Errorf("per page value is less than 0")
	}

	return nil
}

func (q *getSongsRequestQuery) toSongFilters() (*domain.SongFilters, error) {
	var (
		releaseDateRange *domain.TimeRange
		err              error
	)
	if q.SongReleaseDateRange != nil {
		releaseDateRange = new(domain.TimeRange)
		*releaseDateRange, err = parseDateRange(*q.SongReleaseDateRange)
		if err != nil {
			return nil, errors.Wrap(err, "parse release date range")
		}
	}

	return &domain.SongFilters{
		SongName:             q.SongName,
		MusicGroupName:       q.MusicGroupName,
		SongLink:             q.SongLink,
		SongCoupletContains:  q.SongTextContains,
		SongReleaseDateRange: releaseDateRange,
	}, nil
}

func newSongDTOFromEntity(song *domain.Song) *songDTO {
	return &songDTO{
		ID:          song.ID.String(),
		Name:        song.Name,
		ReleaseDate: song.ReleaseDate.Format(DateLayout),
		Couplets:    song.Couplets,
		Link:        song.Link,
		MusicGroup: musicGroupDTO{
			ID:   song.MusicGroup.ID.String(),
			Name: song.MusicGroup.Name,
		}}
}
