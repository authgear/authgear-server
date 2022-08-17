package identity

import "github.com/authgear/authgear-server/pkg/api/model"

type SIWESpec struct {
	VerificationRequest model.SIWEVerificationRequest `json:"data"`
}
