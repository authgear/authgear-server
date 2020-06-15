package config

import "time"

type DurationSeconds int

func (d DurationSeconds) Duration() time.Duration {
	return time.Duration(d) * time.Second
}

type DurationDays int

func (d DurationDays) Duration() time.Duration {
	return time.Duration(d) * (24 * time.Hour)
}
