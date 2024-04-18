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
	authflow.RegisterNode(&NodeCreateAuthenticatorTOTP{})
}

type NodeCreateAuthenticatorTOTPData struct {
	TypedData
	Secret     string `json:"secret"`
	OTPAuthURI string `json:"otpauth_uri"`
}

func NewNodeCreateAuthenticatorTOTPData(d NodeCreateAuthenticatorTOTPData) NodeCreateAuthenticatorTOTPData {
	d.Type = DataTypeCreateTOTPData
	return d
}

var _ authflow.Data = NodeCreateAuthenticatorTOTPData{}

func (m NodeCreateAuthenticatorTOTPData) Data() {}

type NodeCreateAuthenticatorTOTP struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
	Authenticator  *authenticator.Info                     `json:"authenticator,omitempty"`
}

var _ authflow.NodeSimple = &NodeCreateAuthenticatorTOTP{}
var _ authflow.Milestone = &NodeCreateAuthenticatorTOTP{}
var _ MilestoneAuthenticationMethod = &NodeCreateAuthenticatorTOTP{}
var _ MilestoneSwitchToExistingUser = &NodeCreateAuthenticatorTOTP{}
var _ authflow.InputReactor = &NodeCreateAuthenticatorTOTP{}
var _ authflow.DataOutputer = &NodeCreateAuthenticatorTOTP{}

func NewNodeCreateAuthenticatorTOTP(deps *authflow.Dependencies, n *NodeCreateAuthenticatorTOTP) (*NodeCreateAuthenticatorTOTP, error) {
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

func (*NodeCreateAuthenticatorTOTP) Kind() string {
	return "NodeCreateAuthenticatorTOTP"
}

func (*NodeCreateAuthenticatorTOTP) Milestone() {}
func (n *NodeCreateAuthenticatorTOTP) MilestoneAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}
func (i *NodeCreateAuthenticatorTOTP) MilestoneSwitchToExistingUser(newUserID string) {
	// TODO(tung): Skip creation if already have one
	i.UserID = newUserID
}

func (n *NodeCreateAuthenticatorTOTP) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	return &InputSchemaSetupTOTP{
		JSONPointer:    n.JSONPointer,
		FlowRootObject: flowRootObject,
	}, nil
}

func (n *NodeCreateAuthenticatorTOTP) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputSetupTOTP inputSetupTOTP
	if authflow.AsInput(input, &inputSetupTOTP) {
		_, err := deps.Authenticators.VerifyWithSpec(n.Authenticator, &authenticator.Spec{
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
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *NodeCreateAuthenticatorTOTP) OutputData(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.Data, error) {
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

	return NewNodeCreateAuthenticatorTOTPData(NodeCreateAuthenticatorTOTPData{
		Secret:     secret,
		OTPAuthURI: otpauthURI,
	}), nil
}

func (n *NodeCreateAuthenticatorTOTP) authenticatorKind() model.AuthenticatorKind {
	switch n.Authentication {
	case config.AuthenticationFlowAuthenticationSecondaryTOTP:
		return model.AuthenticatorKindSecondary
	default:
		panic(fmt.Errorf("unexpected authentication method: %v", n.Authentication))
	}
}
