package domain

import (
	"time"

	"github.com/segmentio/ksuid"
)

type Song struct {
	ID          ksuid.KSUID
	Name        string
	MusicGroup  MusicGroup
	Couplets    []string
	ReleaseDate time.Time
	Link        string
}

type MusicGroup struct {
	ID   ksuid.KSUID
	Name string
}
