package declarative

import (
	"context"
	"fmt"
	"time"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
	"github.com/authgear/authgear-server/pkg/util/accesscontrol"
	"github.com/authgear/authgear-server/pkg/util/secretcode"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	authflow.RegisterIntent(&IntentCreateAuthenticatorTOTP{})
}

type IntentCreateAuthenticatorTOTPData struct {
	TypedData
	Secret     string `json:"secret"`
	OTPAuthURI string `json:"otpauth_uri"`
}

func NewIntentCreateAuthenticatorTOTPData(d IntentCreateAuthenticatorTOTPData) IntentCreateAuthenticatorTOTPData {
	d.Type = DataTypeCreateTOTPData
	return d
}

var _ authflow.Data = IntentCreateAuthenticatorTOTPData{}

func (m IntentCreateAuthenticatorTOTPData) Data() {}

type IntentCreateAuthenticatorTOTP struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	Authenticator  *authenticator.Info                     `json:"authenticator,omitempty"`
}

var _ authflow.Intent = &IntentCreateAuthenticatorTOTP{}
var _ authflow.Milestone = &IntentCreateAuthenticatorTOTP{}
var _ MilestoneAuthenticationMethod = &IntentCreateAuthenticatorTOTP{}
var _ MilestoneFlowCreateAuthenticator = &IntentCreateAuthenticatorTOTP{}
var _ authflow.InputReactor = &IntentCreateAuthenticatorTOTP{}
var _ authflow.DataOutputer = &IntentCreateAuthenticatorTOTP{}

func NewIntentCreateAuthenticatorTOTP(deps *authflow.Dependencies, n *IntentCreateAuthenticatorTOTP) (*IntentCreateAuthenticatorTOTP, error) {
	authenticatorKind := n.authenticatorKind()

	isDefault, err := authenticatorIsDefault(deps, n.UserID, authenticatorKind)
	if err != nil {
		return nil, err
	}

	now := deps.Clock.NowUTC()
	displayName := fmt.Sprintf("TOTP @ %s", now.Format(time.RFC3339))

	spec := &authenticator.Spec{
		UserID:    n.UserID,
		IsDefault: isDefault,
		Kind:      authenticatorKind,
		Type:      model.AuthenticatorTypeTOTP,
		TOTP: &authenticator.TOTPSpec{
			DisplayName: displayName,
		},
	}

	id := uuid.New()
	info, err := deps.Authenticators.NewWithAuthenticatorID(id, spec)
	if err != nil {
		return nil, err
	}

	n.Authenticator = info
	return n, nil
}

func (*IntentCreateAuthenticatorTOTP) Kind() string {
	return "IntentCreateAuthenticatorTOTP"
}

func (*IntentCreateAuthenticatorTOTP) Milestone() {}
func (*IntentCreateAuthenticatorTOTP) MilestoneFlowCreateAuthenticator(flows authflow.Flows) (MilestoneDoCreateAuthenticator, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateAuthenticator](flows)
}
func (n *IntentCreateAuthenticatorTOTP) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (n *IntentCreateAuthenticatorTOTP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, created := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateAuthenticator](flows)
	if created {
		return nil, authflow.ErrEOF
	}
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	isBotProtectionRequired, err := IsBotProtectionRequired(ctx, flowRootObject, n.JSONPointer)
	if err != nil {
		return nil, err
	}
	return &InputSchemaSetupTOTP{
		JSONPointer:             n.JSONPointer,
		FlowRootObject:          flowRootObject,
		IsBotProtectionRequired: isBotProtectionRequired,
	}, nil
}

func (n *IntentCreateAuthenticatorTOTP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputSetupTOTP inputSetupTOTP
	if authflow.AsInput(input, &inputSetupTOTP) {
		var bpSpecialErr error
		bpRequired, err := IsNodeBotProtectionRequired(ctx, deps, flows, n.JSONPointer)
		if err != nil {
			return nil, err
		}
		if bpRequired {
			var inputTakeBotProtection inputTakeBotProtection
			if !authflow.AsInput(input, &inputTakeBotProtection) {
				return nil, authflow.ErrIncompatibleInput
			}

			token := inputTakeBotProtection.GetBotProtectionProviderResponse()
			bpSpecialErr, err = HandleBotProtection(ctx, deps, token)
			if err != nil {
				return nil, err
			}
		}
		_, err = deps.Authenticators.VerifyWithSpec(n.Authenticator, &authenticator.Spec{
			TOTP: &authenticator.TOTPSpec{
				Code: inputSetupTOTP.GetCode(),
			},
		}, &facade.VerifyOptions{
			AuthenticationDetails: facade.NewAuthenticationDetails(
				n.UserID,
				authn.AuthenticationStageFromAuthenticationMethod(n.Authentication),
				authn.AuthenticationTypeTOTP,
			),
		})
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoCreateAuthenticator{
			Authenticator: n.Authenticator,
		}), bpSpecialErr
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *IntentCreateAuthenticatorTOTP) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
	secret := n.Authenticator.TOTP.Secret

	issuer := deps.HTTPOrigin
	user, err := deps.Users.Get(n.UserID, accesscontrol.RoleGreatest)
	if err != nil {
		return nil, err
	}
	accountName := user.EndUserAccountID()
	opts := secretcode.URIOptions{
		Issuer:      string(issuer),
		AccountName: accountName,
	}
	totp, err := secretcode.NewTOTPFromSecret(secret)
	if err != nil {
		return nil, err
	}
	otpauthURI := totp.GetURI(opts).String()

	return NewIntentCreateAuthenticatorTOTPData(IntentCreateAuthenticatorTOTPData{
		Secret:     secret,
		OTPAuthURI: otpauthURI,
	}), nil
}

func (n *IntentCreateAuthenticatorTOTP) authenticatorKind() model.AuthenticatorKind {
	switch n.Authentication {
	case config.AuthenticationFlowAuthenticationSecondaryTOTP:
		return model.AuthenticatorKindSecondary
	default:
		panic(fmt.Errorf("unexpected authentication method: %v", n.Authentication))
	}
}
