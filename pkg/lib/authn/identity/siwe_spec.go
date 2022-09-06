package identity

import "github.com/authgear/authgear-server/pkg/api/model"

type SIWESpec struct {
	VerifiedData model.SIWEVerifiedData `json:"data"`
}
