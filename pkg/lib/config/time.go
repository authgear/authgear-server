package config

import "time"

var durationSecondsSchema = `{ "type": "integer" }`
var _ = Schema.Add("DurationSeconds", durationSecondsSchema)
var _ = FeatureConfigSchema.Add("DurationSeconds", durationSecondsSchema)

type DurationSeconds int

func (d DurationSeconds) Duration() time.Duration {
	return time.Duration(d) * time.Second
}

var durationDaysSchema = `{ "type": "integer" }`
var _ = Schema.Add("DurationDays", durationDaysSchema)
var _ = FeatureConfigSchema.Add("DurationDays", durationDaysSchema)

type DurationDays int

func (d DurationDays) Duration() time.Duration {
	return time.Duration(d) * (24 * time.Hour)
}

var durationStringSchema = `{ "type": "string", "format": "x_duration_string" }`
var _ = Schema.Add("DurationString", durationStringSchema)
var _ = FeatureConfigSchema.Add("DurationString", durationStringSchema)

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
