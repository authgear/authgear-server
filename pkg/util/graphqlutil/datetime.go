package graphqlutil

import (
	"fmt"
	"time"
)

// GetDateTimeInUTCFromInput gets a *time.Time from input.
// It is expected that the graphql library accepts RFC3339 timestamp only.
// As of graphql v0.8.1, https://github.com/graphql-go/graphql/blob/v0.8.1/scalars.go#L557
// The deserialization of DateTime delegates to time.Time.UnmarshalText,
// and the doc of UnmarshalText says it only accept RFC3339.
// That is, the input timestamp also has an explicit timezone.
// The job of this function is to ensure the returned time is always in UTC.
func GetDateTimeInUTCFromInput(input map[string]interface{}, key string) *time.Time {
	val, ok := input[key]
	if !ok {
		return nil
	}
	switch v := val.(type) {
	case time.Time:
		inUTC := v.In(time.UTC)
		return &inUTC
	case *time.Time:
		inUTC := v.In(time.UTC)
		return &inUTC
	case nil:
		return nil
	default:
		panic(fmt.Errorf("expected time.Time or *time.Time but found %T", val))
	}
}
