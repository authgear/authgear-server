package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/facade"
)

func init() {
	authflow.RegisterIntent(&IntentUseAuthenticatorFaceRecognition{})
}

type IntentUseAuthenticatorFaceRecognition struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentUseAuthenticatorFaceRecognition{}
var _ authflow.Milestone = &IntentUseAuthenticatorFaceRecognition{}
var _ MilestoneFlowSelectAuthenticationMethod = &IntentUseAuthenticatorFaceRecognition{}
var _ MilestoneDidSelectAuthenticationMethod = &IntentUseAuthenticatorFaceRecognition{}
var _ MilestoneFlowAuthenticate = &IntentUseAuthenticatorFaceRecognition{}
var _ authflow.InputReactor = &IntentUseAuthenticatorFaceRecognition{}

func (*IntentUseAuthenticatorFaceRecognition) Kind() string {
	return "IntentUseAuthenticatorFaceRecognition"
}

func (*IntentUseAuthenticatorFaceRecognition) Milestone() {}
func (n *IntentUseAuthenticatorFaceRecognition) MilestoneFlowSelectAuthenticationMethod(flows authflow.Flows) (MilestoneDidSelectAuthenticationMethod, authflow.Flows, bool) {
	return n, flows, true
}
func (n *IntentUseAuthenticatorFaceRecognition) MilestoneDidSelectAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (*IntentUseAuthenticatorFaceRecognition) MilestoneFlowAuthenticate(flows authflow.Flows) (MilestoneDidAuthenticate, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)
}

func (n *IntentUseAuthenticatorFaceRecognition) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, authenticated := authflow.FindMilestoneInCurrentFlow[MilestoneDidAuthenticate](flows)
	if authenticated {
		return nil, authflow.ErrEOF
	}
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeFaceRecognition{
		JSONPointer:    n.JSONPointer,
		FlowRootObject: flowRootObject,
	}, nil
}

func (n *IntentUseAuthenticatorFaceRecognition) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeFaceRecognition inputTakeFaceRecognition
	if authflow.AsInput(input, &inputTakeFaceRecognition) {
		as, err := deps.Authenticators.List(
			n.UserID,
			authenticator.KeepKind(n.Authentication.AuthenticatorKind()),
			authenticator.KeepType(model.AuthenticatorTypeFaceRecognition),
		)
		if err != nil {
			return nil, err
		}

		b64Image := inputTakeFaceRecognition.GetB64Image()
		spec := &authenticator.Spec{
			FaceRecognition: &authenticator.FaceRecognitionSpec{
				B64ImageString: b64Image,
			},
		}

		info, _, err := deps.Authenticators.VerifyOneWithSpec(
			n.UserID,
			model.AuthenticatorTypeFaceRecognition,
			as,
			spec,
			&facade.VerifyOptions{
				AuthenticationDetails: facade.NewAuthenticationDetails(
					n.UserID,
					authn.AuthenticationStageFromAuthenticationMethod(n.Authentication),
					authn.AuthenticationTypeFaceRecognition,
				),
			},
		)
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoUseAuthenticatorSimple{
			Authenticator: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
