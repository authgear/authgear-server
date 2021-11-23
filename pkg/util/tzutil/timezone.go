package tzutil

import (
	"time"
)

type Timezone struct {
	Name            string
	Ref             time.Time
	Offset          int
	FormattedOffset string
	Location        *time.Location
	DisplayLabel    string
}
