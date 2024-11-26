package domain

import "time"

type CreateSongDTO struct {
	SongName       string
	MusicGroupName string
}

type IntegrationSongInfo struct {
	ReleaseDate time.Time
	Text        string
	Link        string
}

type SongFilters struct {
	SongName             *string
	MusicGroupName       *string
	SongLink             *string
	SongCoupletContains  *string
	SongReleaseDateRange *TimeRange
}

type SongUpdate struct {
	Name        *string
	ReleaseDate *string
	Couplets    *[]string
	Link        *string
}
