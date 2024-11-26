package domain

import "github.com/pkg/errors"

var (
	ErrInternal    = errors.New("internal error")
	ErrIntegration = errors.New("integration error")

	ErrSongNotFound      = errors.New("song not found")
	ErrSongAlreadyExists = errors.New("song already exists")
)

type SongInfoIntegrationError error
