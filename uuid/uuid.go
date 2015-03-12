package uuid

import "github.com/twinj/uuid"

func init() {
	uuid.SwitchFormat(uuid.CleanHyphen)
}

// New returns a new uuid4 string
func New() string {
	return uuid.NewV4().String()
}
