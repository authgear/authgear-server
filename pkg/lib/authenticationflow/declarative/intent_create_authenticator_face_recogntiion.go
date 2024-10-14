package declarative

import (
	"context"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/uuid"
)

func init() {
	authflow.RegisterIntent(&IntentCreateAuthenticatorFaceRecognition{})
}

type IntentCreateAuthenticatorFaceRecognition struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Authentication config.AuthenticationFlowAuthentication `json:"authentication,omitempty"`
}

var _ authflow.Intent = &IntentCreateAuthenticatorFaceRecognition{}
var _ authflow.InputReactor = &IntentCreateAuthenticatorFaceRecognition{}
var _ authflow.Milestone = &IntentCreateAuthenticatorFaceRecognition{}
var _ MilestoneFlowCreateAuthenticator = &IntentCreateAuthenticatorFaceRecognition{}
var _ MilestoneFlowSelectAuthenticationMethod = &IntentCreateAuthenticatorFaceRecognition{}
var _ MilestoneDidSelectAuthenticationMethod = &IntentCreateAuthenticatorFaceRecognition{}

func (*IntentCreateAuthenticatorFaceRecognition) Kind() string {
	return "IntentCreateAuthenticatorFaceRecognition"
}

func (*IntentCreateAuthenticatorFaceRecognition) Milestone() {}
func (*IntentCreateAuthenticatorFaceRecognition) MilestoneFlowCreateAuthenticator(flows authflow.Flows) (MilestoneDoCreateAuthenticator, authflow.Flows, bool) {
	return authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateAuthenticator](flows)
}
func (n *IntentCreateAuthenticatorFaceRecognition) MilestoneFlowSelectAuthenticationMethod(flows authflow.Flows) (MilestoneDidSelectAuthenticationMethod, authflow.Flows, bool) {
	return n, flows, true
}
func (n *IntentCreateAuthenticatorFaceRecognition) MilestoneDidSelectAuthenticationMethod() config.AuthenticationFlowAuthentication {
	return n.Authentication
}

func (n *IntentCreateAuthenticatorFaceRecognition) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	_, _, created := authflow.FindMilestoneInCurrentFlow[MilestoneDoCreateAuthenticator](flows)
	if created {
		return nil, authflow.ErrEOF
	}
	flowRootObject, err := findFlowRootObjectInFlow(deps, flows)
	if err != nil {
		return nil, err
	}

	return &InputSchemaTakeFaceRecognition{
		FlowRootObject: flowRootObject,
		JSONPointer:    n.JSONPointer,
	}, nil
}

func (i *IntentCreateAuthenticatorFaceRecognition) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeNewFaceRecognition inputTakeFaceRecognition
	if authflow.AsInput(input, &inputTakeNewFaceRecognition) {
		authenticatorKind := i.Authentication.AuthenticatorKind()
		b64Image := inputTakeNewFaceRecognition.GetB64Image()
		isDefault, err := authenticatorIsDefault(deps, i.UserID, authenticatorKind)
		if err != nil {
			return nil, err
		}

		spec := &authenticator.Spec{
			UserID:    i.UserID,
			IsDefault: isDefault,
			Kind:      authenticatorKind,
			Type:      model.AuthenticatorTypeFaceRecognition,
			FaceRecognition: &authenticator.FaceRecognitionSpec{
				B64ImageString: b64Image,
			},
		}

		authenticatorID := uuid.New()
		info, err := deps.Authenticators.NewWithAuthenticatorID(authenticatorID, spec)
		if err != nil {
			return nil, err
		}

		return authflow.NewNodeSimple(&NodeDoCreateAuthenticator{
			Authenticator: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}
