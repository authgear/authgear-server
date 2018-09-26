package db

import "encoding/json"

// NullJSONStringSlice will reject empty member, since pq will give [null]
// array if we use `array_to_json` on null column. So the result slice will be
// []string{}, but not []string{""}
type NullJSONStringSlice struct {
	Slice []string
	Valid bool
}

func (njss *NullJSONStringSlice) Scan(value interface{}) error {
	data, ok := value.([]byte)
	if value == nil || !ok {
		njss.Slice = nil
		njss.Valid = false
		return nil
	}

	njss.Slice = []string{}
	allSlice := []string{}
	err := json.Unmarshal(data, &allSlice)
	for _, s := range allSlice {
		if s != "" {
			njss.Slice = append(njss.Slice, s)
		}
	}
	njss.Valid = err == nil
	return err
}
