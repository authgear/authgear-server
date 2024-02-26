package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

type OOBData struct {
	TypedData
	Channels         []model.AuthenticatorOOBChannel `json:"channels,omitempty"`
	MaskedClaimValue string                          `json:"masked_claim_value,omitempty"`
}

func NewOOBData(d OOBData) OOBData {
	d.Type = DataTypeOOBChannelsData
	return d
}

func (OOBData) Data() {}

var _ authflow.Data = OOBData{}
