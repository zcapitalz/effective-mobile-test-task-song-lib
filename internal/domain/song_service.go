package domain

import (
	"context"
	slogutils "song-lib/internal/utils/slog-utils"
	"strings"

	"github.com/pkg/errors"
	"github.com/segmentio/ksuid"
)

type SongService struct {
	songRepository      SongRepository
	SongInfoIntegration SongInfoIntegration
}

type SongRepository interface {
	SaveSong(ctx context.Context, song *Song) (*Song, error)

	GetSongsFilteredPaginated(
		ctx context.Context, filters *SongFilters,
		pagination Pagination,
	) ([]Song, error)

	GetSongCoupletsPaginated(
		ctx context.Context, songID ksuid.KSUID,
		pagination Pagination,
	) ([]string, error)

	SongExistsByID(
		ctx context.Context, songID ksuid.KSUID,
	) (bool, error)

	SongExistsByNameAndMusicGroupName(
		ctx context.Context, songName string,
		musicGroupName string,
	) (bool, error)

	UpdateSong(
		ctx context.Context, songID ksuid.KSUID,
		songUpdate *SongUpdate,
	) (*Song, error)
	DeleteSong(ctx context.Context, songID ksuid.KSUID) error
}

type SongInfoIntegration interface {
	GetSongInfo(
		songName, musicGroupName string,
	) (*IntegrationSongInfo, error)
}

func NewSongService(
	songRepository SongRepository,
	songInfoIntegration SongInfoIntegration,
) *SongService {

	return &SongService{
		songRepository:      songRepository,
		SongInfoIntegration: songInfoIntegration,
	}
}

func (s *SongService) CreateSong(
	ctx context.Context, dto *CreateSongDTO,
) (*Song, error) {

	exists, err := s.songRepository.
		SongExistsByNameAndMusicGroupName(
			ctx, dto.SongName,
			dto.MusicGroupName)
	switch {
	case err != nil:
		slogutils.Error(ctx, "create song:",
			errors.Wrap(err, "check song exists"))
		return nil, ErrInternal
	case exists:
		return nil, ErrSongAlreadyExists
	}

	additionalSongInfo, err := s.SongInfoIntegration.
		GetSongInfo(dto.SongName, dto.MusicGroupName)
	switch {
	case errors.As(err, new(SongInfoIntegrationError)):
		slogutils.Error(
			ctx, "create song:",
			errors.Wrap(err, "get song info"))
		return nil, ErrIntegration
	case err != nil:
		slogutils.Error(
			ctx, "create song:",
			errors.Wrap(err, "get song info"))
		return nil, ErrInternal
	}

	songCouplets := strings.Split(additionalSongInfo.Text, "\n\n")
	song, err := s.songRepository.SaveSong(
		ctx,
		&Song{
			Name:        dto.SongName,
			MusicGroup:  MusicGroup{Name: dto.MusicGroupName},
			Couplets:    songCouplets,
			ReleaseDate: additionalSongInfo.ReleaseDate,
			Link:        additionalSongInfo.Link,
		})
	if err != nil {
		slogutils.Error(ctx, "create song:",
			errors.Wrap(err, "save song"))
		return nil, ErrInternal
	}

	return song, nil
}

func (s *SongService) GetSongCoupletsPaginated(
	ctx context.Context, songID ksuid.KSUID,
	pagination Pagination,
) ([]string, error) {

	exists, err := s.songRepository.SongExistsByID(ctx, songID)
	switch {
	case err != nil:
		slogutils.Error(ctx, "get song couplets:",
			errors.Wrap(err, "check song exists"))
		return nil, ErrInternal
	case !exists:
		return nil, ErrSongNotFound
	}

	songCouplets, err := s.songRepository.
		GetSongCoupletsPaginated(ctx, songID, pagination)
	if err != nil {
		slogutils.Error(ctx, "get song couplets:", err)
		return nil, ErrInternal
	}

	return songCouplets, nil
}

func (s *SongService) GetSongsFilteredPaginated(
	ctx context.Context, filters *SongFilters,
	pagination Pagination,
) ([]Song, error) {

	songs, err := s.songRepository.
		GetSongsFilteredPaginated(ctx, filters, pagination)
	if err != nil {
		slogutils.Error(ctx, "get songs:", err)
		return nil, ErrInternal
	}

	return songs, nil
}

func (s *SongService) UpdateSong(
	ctx context.Context, songID ksuid.KSUID,
	songUpdate *SongUpdate,
) (*Song, error) {

	song, err := s.songRepository.
		UpdateSong(ctx, songID, songUpdate)
	if err != nil {
		slogutils.Error(ctx, "update song:", err)
		return nil, ErrInternal
	}

	return song, nil
}

func (s *SongService) DeleteSong(
	ctx context.Context, songID ksuid.KSUID,
) error {

	err := s.songRepository.DeleteSong(ctx, songID)
	if err != nil {
		slogutils.Error(ctx, "delete song:", err)
		return ErrInternal
	}

	return nil
}
