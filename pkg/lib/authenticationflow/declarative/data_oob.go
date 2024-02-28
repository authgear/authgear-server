package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

type SelectOOBOTPChannelsData struct {
	TypedData
	Channels         []model.AuthenticatorOOBChannel `json:"channels,omitempty"`
	MaskedClaimValue string                          `json:"masked_claim_value,omitempty"`
}

func NewOOBData(d SelectOOBOTPChannelsData) SelectOOBOTPChannelsData {
	d.Type = DataTypeSelectOOBOTPChannelsData
	return d
}

func (SelectOOBOTPChannelsData) Data() {}

var _ authflow.Data = SelectOOBOTPChannelsData{}
