package tzutil

import (
	"sort"
	"time"
)

// List returns a list of embedded timezones.
func List(ref time.Time) ([]Timezone, error) {
	var out []Timezone
	for _, name := range timezoneNames {
		loc, err := time.LoadLocation(name)
		if err != nil {
			return nil, err
		}
		t := ref.In(loc)
		_, offset := t.Zone()
		formattedOffset := t.Format("-07:00")
		out = append(out, Timezone{
			Name:            name,
			Ref:             ref,
			Offset:          offset,
			FormattedOffset: formattedOffset,
			Location:        loc,
		})
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
