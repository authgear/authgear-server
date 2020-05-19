package loginid

type Identity struct {
	ID              string
	UserID          string
	LoginIDKey      string
	LoginID         string
	OriginalLoginID string
	UniqueKey       string
	Claims          map[string]string
}
