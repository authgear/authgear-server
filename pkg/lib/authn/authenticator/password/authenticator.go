package password

type Authenticator struct {
	ID           string
	UserID       string
	PasswordHash []byte
	Tag          []string
}
