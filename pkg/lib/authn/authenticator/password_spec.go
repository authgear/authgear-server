package authenticator

type PasswordSpec struct {
	PlainPassword string `json:"-"`
	PasswordHash  string `json:"-"`
}
