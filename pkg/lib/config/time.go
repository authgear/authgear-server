package config

import "time"

var _ = Schema.Add("DurationSeconds", `{ "type": "integer" }`)

type DurationSeconds int

func (d DurationSeconds) Duration() time.Duration {
	return time.Duration(d) * time.Second
}

var _ = Schema.Add("DurationDays", `{ "type": "integer" }`)

type DurationDays int

func (d DurationDays) Duration() time.Duration {
	return time.Duration(d) * (24 * time.Hour)
}

var _ = Schema.Add("DurationString", `{ "type": "string", "format": "x_duration_string" }`)

type DurationString string

func (d DurationString) duration() (time.Duration, error) {
	t, err := time.ParseDuration(string(d))
	if err != nil {
		return time.Duration(0), err
	}
	return t, err
}

func (d DurationString) MaybeDuration() (time.Duration, bool) {
	t, err := d.duration()
	if err != nil {
		return time.Duration(0), false
	}
	return t, true
}

func (d DurationString) Duration() time.Duration {
	t, err := d.duration()
	if err != nil {
		panic(err)
	}
	return t
}
