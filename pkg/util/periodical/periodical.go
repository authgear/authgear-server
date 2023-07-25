package periodical

import (
	"errors"
	"regexp"
	"strconv"
	"time"

	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/timeutil"
)

type Type string

const (
	Hourly  Type = "hourly"
	Daily   Type = "daily"
	Weekly  Type = "weekly"
	Monthly Type = "monthly"
)

var ErrInvalidPeriodical = errors.New("Invalid periodical format")

var iso8601WeekRegex = regexp.MustCompile(`^(\d{4})-W(\d{2})$`)

type ArgumentParser struct {
	Clock clock.Clock
}

// ParseAnalyticCollectCountPeriodicalArgument parse the argument input and
// returns periodical and the start date of the periodical
// if periodical is hourly, t is the start of the hour.
// if periodical is daily, t is the start of the day.
// if periodical is monthly, t is the start of the day on the first day of the month.
// if periodical is weekly, t is the start of the day on the monday of the week.
// Supported input format:
// - this-hour
// - today
// - this-week
// - this-month
// - last-hour
// - yesterday
// - last-week
// - last-month
// - 2016-01-02T15
// - 2016-01-02
// - 2016-01
// - 2016-W37
func (p *ArgumentParser) Parse(input string) (Type, *time.Time, error) {
	now := p.Clock.NowUTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	switch input {
	case "this-hour":
		thisHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)
		return Hourly, &thisHour, nil
	case "today":
		return Daily, &today, nil
	case "this-week":
		monday := today
		for monday.Weekday() != time.Monday {
			monday = monday.AddDate(0, 0, -1)
		}
		return Weekly, &monday, nil
	case "this-month":
		fistDateOfMonth := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, time.UTC)
		return Monthly, &fistDateOfMonth, nil
	case "last-hour":
		thisHour := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, time.UTC)
		lastHour := thisHour.Add(-time.Hour)
		return Hourly, &lastHour, nil
	case "yesterday":
		yesterday := today.AddDate(0, 0, -1)
		return Daily, &yesterday, nil
	case "last-week":
		lastWeek := today.AddDate(0, 0, -7)
		for lastWeek.Weekday() != time.Monday {
			lastWeek = lastWeek.AddDate(0, 0, -1)
		}
		return Weekly, &lastWeek, nil
	case "last-month":
		lastMonth := today.AddDate(0, -1, 0)
		fistDateOfMonth := time.Date(lastMonth.Year(), lastMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
		return Monthly, &fistDateOfMonth, nil
	}

	// match format "2006-01-02T15"
	t, err := time.Parse("2006-01-02T15", input)
	if err == nil {
		t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, time.UTC)
		return Hourly, &t, nil
	}

	// match format "2006-01-02"
	t, err = time.Parse("2006-01-02", input)
	if err == nil {
		t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		return Daily, &t, nil
	}

	// match format "2006-01"
	t, err = time.Parse("2006-01", input)
	if err == nil {
		fistDateOfMonth := time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, time.UTC)
		return Monthly, &fistDateOfMonth, nil
	}

	// match format "2006-W37"
	matches := iso8601WeekRegex.FindStringSubmatch(input)
	if len(matches) == 3 {
		year, err := strconv.Atoi(matches[1])
		if err != nil {
			return "", nil, ErrInvalidPeriodical
		}

		week, err := strconv.Atoi(matches[2])
		if err != nil {
			return "", nil, ErrInvalidPeriodical
		}

		firstDayOfISOWeek, err := timeutil.FirstDayOfISOWeek(year, week, time.UTC)
		if err != nil {
			return "", nil, err
		}

		return Weekly, firstDayOfISOWeek, nil
	}

	return "", nil, ErrInvalidPeriodical
}
