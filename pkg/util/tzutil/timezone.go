package tzutil

import (
	"time"
)

type Timezone struct {
	Name            string         `json:"name"`
	Ref             time.Time      `json:"ref"`
	Offset          int            `json:"offset"`
	FormattedOffset string         `json:"formattedOffset"`
	Location        *time.Location `json:"location"`
	DisplayLabel    string         `json:"displayLabel"`
}
