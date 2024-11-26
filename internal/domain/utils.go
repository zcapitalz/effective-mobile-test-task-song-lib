package domain

import "time"

type TimeRange struct {
	StartTime time.Time
	EndTime   time.Time
}

type Pagination struct {
	Page    int
	PerPage int
}
