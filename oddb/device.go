package oddb

// Device represents a device owned by a user and ready to receive notification.
type Device struct {
	ID         string
	Type       string
	Token      string
	UserInfoID string
}
