package declarative

import (
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type AccountLinkingIdentificationOption struct {
	Identifcation     config.AuthenticationFlowIdentification `json:"identification"`
	MaskedDisplayName string                                  `json:"masked_display_name,omitempty"`

	// ProviderType is specific to OAuth.
	ProviderType config.OAuthSSOProviderType `json:"provider_type,omitempty"`
	// Alias is specific to OAuth.
	Alias string `json:"string,omitempty"`
}

type AccountLinkingIdentificationOptionInternal struct {
	AccountLinkingIdentificationOption
	Identity *identity.Info
}

type AccountLinkingIdentifyData struct {
	TypedData
	Options []AccountLinkingIdentificationOption `json:"options"`
}

var _ authflow.Data = AccountLinkingIdentifyData{}

func (AccountLinkingIdentifyData) Data() {}

func NewAccountLinkingIdentifyData(options []AccountLinkingIdentificationOptionInternal) AccountLinkingIdentifyData {
	return AccountLinkingIdentifyData{
		TypedData: TypedData{Type: DataTypeAccountLinkingIdentificationData},
		Options: slice.Map(options, func(o AccountLinkingIdentificationOptionInternal) AccountLinkingIdentificationOption {
			return o.AccountLinkingIdentificationOption
		}),
	}
}
