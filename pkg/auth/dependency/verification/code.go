package verification

type Code struct {
	UserID     string `json:"user_id"`
	IdentityID string `json:"identity_id"`
	Code       string `json:"code"`
}
