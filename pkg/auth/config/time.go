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
