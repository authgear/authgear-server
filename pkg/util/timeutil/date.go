package timeutil

import (
	"fmt"
	"time"
)

// Date type is for date serialization
type Date time.Time

func (date Date) IsZero() bool {
	return date == Date{}
}

func (date Date) MarshalJSON() ([]byte, error) {
	t := time.Time(date)
	str := fmt.Sprintf(`"%s"`, t.Format(LayoutISODate))
	return []byte(str), nil
}

func (date *Date) Decode(value string) error {
	t, err := time.Parse(LayoutISODate, value)
	if err != nil {
		return err
	}
	*date = Date(TruncateToDate(t))
	return nil
}
