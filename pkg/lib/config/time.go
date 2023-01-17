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

func (d DurationString) Duration() time.Duration {
	t, err := time.ParseDuration(string(d))
	if err != nil {
		panic(err)
	}
	return t
}
