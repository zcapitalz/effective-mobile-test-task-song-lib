package songcontroller

import (
	"fmt"
	"regexp"
	"song-lib/internal/domain"
	"time"
)

const DateLayout = time.DateOnly

var dateRangeRegexp = regexp.MustCompile(`^\[(\d{4}-\d{2}-\d{2});(\d{4}-\d{2}-\d{2})\]$`)

func parseDateRange(dateRange string) (domain.TimeRange, error) {
	matches := dateRangeRegexp.FindStringSubmatch(dateRange)
	if len(matches) != 3 {
		return domain.TimeRange{}, fmt.Errorf("date range \"%s\" has invalid format", dateRange)
	}

	startDate, err := time.Parse(DateLayout, matches[1])
	if err != nil {
		return domain.TimeRange{}, fmt.Errorf("start date \"%s\" has invalid format", matches[0])
	}
	endDate, err := time.Parse(DateLayout, matches[2])
	if err != nil {
		return domain.TimeRange{}, fmt.Errorf("end date \"%s\" has invalid format", matches[1])
	}

	if startDate.After(endDate) {
		return domain.TimeRange{}, fmt.Errorf("date range start date %s is after end date %s", startDate.Format(DateLayout), endDate.Format(DateLayout))
	}

	return domain.TimeRange{StartTime: startDate, EndTime: endDate}, nil
}
