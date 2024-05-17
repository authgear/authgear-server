package authenticator

import "time"

type PasswordSpec struct {
	PlainPassword string     `json:"-"`
	PasswordHash  string     `json:"-"`
	ExpireAfter   *time.Time `json:"-"`
}
