package identity

type AnonymousSpec struct {
	KeyID              string `json:"key_id,omitempty"`
	Key                string `json:"key,omitempty"`
	ExistingUserID     string `json:"existing_user_id,omitempty"`
	ExistingIdentityID string `json:"existing_identity_id,omitempty"`
}
