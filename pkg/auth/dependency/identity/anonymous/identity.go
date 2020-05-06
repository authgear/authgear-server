package anonymous

import "regexp"

var KeyIDFormat = regexp.MustCompile(`^[-\w]{8,64}$`)

type Identity struct {
	ID     string
	UserID string
	KeyID  string
	Key    []byte
}
