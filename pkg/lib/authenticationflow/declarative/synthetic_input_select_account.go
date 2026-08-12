package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

// SyntheticInputSelectAccount carries the already-resolved-and-verified user
// ID directly, rather than the client-supplied option index. The index is
// only meaningful within the flow whose identify step it was issued for; a
// signup_login flow's select_account switch replays into a different
// login_flow whose own select_account option sits at its own,
// independently-computed position, so replaying an index across that switch
// would be meaningless. Carrying the user ID sidesteps that entirely,
// mirroring how SyntheticInputPasskey/SyntheticInputOAuth carry the actual
// resolved value (not a position) across a flow switch.
type SyntheticInputSelectAccount struct {
	Identification model.AuthenticationFlowIdentification `json:"identification,omitempty"`
	UserID         string                                 `json:"user_id,omitempty"`
	BotProtection  *InputTakeBotProtectionBody            `json:"bot_protection,omitempty"`
}

var _ authflow.Input = &SyntheticInputSelectAccount{}
var _ inputTakeIdentificationMethod = &SyntheticInputSelectAccount{}
var _ inputTakeSelectAccountUserID = &SyntheticInputSelectAccount{}
var _ inputTakeBotProtection = &SyntheticInputSelectAccount{}

func (*SyntheticInputSelectAccount) Input() {}

func (i *SyntheticInputSelectAccount) GetIdentificationMethod() model.AuthenticationFlowIdentification {
	return i.Identification
}

func (i *SyntheticInputSelectAccount) GetSelectAccountUserID() string {
	return i.UserID
}

func (i *SyntheticInputSelectAccount) GetBotProtectionProvider() *InputTakeBotProtectionBody {
	return i.BotProtection
}

func (i *SyntheticInputSelectAccount) GetBotProtectionProviderType() config.BotProtectionProviderType {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Type
}

func (i *SyntheticInputSelectAccount) GetBotProtectionProviderResponse() string {
	if i.BotProtection == nil {
		return ""
	}
	return i.BotProtection.Response
}
