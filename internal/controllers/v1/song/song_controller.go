package songcontroller

import (
	"context"
	controllers "song-lib/internal/controllers"
	ginutils "song-lib/internal/controllers/api-utils/gin-utils"
	"song-lib/internal/domain"

	"github.com/gin-gonic/gin"
	"github.com/segmentio/ksuid"
)

type SongController struct {
	songService SongService
}

type SongService interface {
	CreateSong(
		ctx context.Context,
		dto *domain.CreateSongDTO,
	) (*domain.Song, error)

	GetSongsFilteredPaginated(
		ctx context.Context,
		filters *domain.SongFilters,
		pagination domain.Pagination,
	) ([]domain.Song, error)

	GetSongCoupletsPaginated(
		ctx context.Context,
		songID ksuid.KSUID,
		pagination domain.Pagination,
	) ([]string, error)

	UpdateSong(
		ctx context.Context,
		songID ksuid.KSUID,
		songUpdate *domain.SongUpdate,
	) (*domain.Song, error)

	DeleteSong(
		ctx context.Context,
		songID ksuid.KSUID,
	) error
}

func NewSongController(accountStorage SongService) controllers.Controller {
	return &SongController{
		songService: accountStorage,
	}
}

func (c *SongController) RegisterRoutes(engine *gin.Engine) {
	songsGroup := engine.Group("api/v1/songs")
	songsGroup.POST("", c.createSong)
	songsGroup.GET("", c.getSongs)

	songIDParsingMiddleware := ginutils.CreateParamParsingMiddleware(
		"songID",
		"songID",
		func(param string) (any, error) { return ksuid.Parse(param) },
	)
	songGroup := songsGroup.Group("/:songID", songIDParsingMiddleware)
	songGroup.GET("/couplets", c.getSongCouplets)
	songGroup.DELETE("", c.deleteSong)
	songGroup.PUT("", c.updateSong)
}
