package tzutil

import (
	"fmt"
	"sort"
	"time"
)

func AsTimezone(name string, ref time.Time) (tz *Timezone, err error) {
	loc, err := time.LoadLocation(name)
	if err != nil {
		return
	}
	t := ref.In(loc)
	_, offset := t.Zone()
	formattedOffset := t.Format("-07:00")
	displayLabel := fmt.Sprintf("[UTC %s] %s", formattedOffset, name)
	tz = &Timezone{
		Name:            name,
		Ref:             ref,
		Offset:          offset,
		FormattedOffset: formattedOffset,
		Location:        loc,
		DisplayLabel:    displayLabel,
	}
	return
}

// List returns a list of embedded timezones.
func List(ref time.Time) ([]Timezone, error) {
	var out []Timezone
	for _, name := range timezoneNames {
		tz, err := AsTimezone(name, ref)
		if err != nil {
			return nil, err
		}
		out = append(out, *tz)
	}

	sort.Slice(out, func(i, j int) bool {
		t1 := out[i]
		t2 := out[j]
		switch {
		case t1.Offset < t2.Offset:
			return true
		case t1.Offset > t2.Offset:
			return false
		default:
			return t1.Name < t2.Name
		}
	})

	return out, nil
}
