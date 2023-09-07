package declarative

import (
	"context"
	"errors"
	"fmt"

	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

func init() {
	authflow.RegisterNode(&NodeCreateIdentityLoginID{})
}

type NodeCreateIdentityLoginID struct {
	JSONPointer    jsonpointer.T                           `json:"json_pointer,omitempty"`
	UserID         string                                  `json:"user_id,omitempty"`
	Identification config.AuthenticationFlowIdentification `json:"identification,omitempty"`
}

var _ authflow.NodeSimple = &NodeCreateIdentityLoginID{}
var _ authflow.Milestone = &NodeCreateIdentityLoginID{}
var _ MilestoneIdentificationMethod = &NodeCreateIdentityLoginID{}
var _ authflow.InputReactor = &NodeCreateIdentityLoginID{}

func (*NodeCreateIdentityLoginID) Kind() string {
	return "NodeCreateIdentityLoginID"
}

func (*NodeCreateIdentityLoginID) Milestone() {}
func (n *NodeCreateIdentityLoginID) MilestoneIdentificationMethod() config.AuthenticationFlowIdentification {
	return n.Identification
}

func (n *NodeCreateIdentityLoginID) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	return &InputSchemaTakeLoginID{
		JSONPointer: n.JSONPointer,
	}, nil
}

func (n *NodeCreateIdentityLoginID) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	var inputTakeLoginID inputTakeLoginID
	if authflow.AsInput(input, &inputTakeLoginID) {
		loginID := inputTakeLoginID.GetLoginID()
		spec := n.makeLoginIDSpec(loginID)

		// FIXME(authflow): allow bypassing email blocklist for Admin API.
		info, err := deps.Identities.New(n.UserID, spec, identity.NewIdentityOptions{})
		if err != nil {
			return nil, err
		}

		duplicate, err := deps.Identities.CheckDuplicated(info)
		if err != nil && !errors.Is(err, identity.ErrIdentityAlreadyExists) {
			return nil, err
		}

		if err != nil {
			spec := info.ToSpec()
			otherSpec := duplicate.ToSpec()
			return nil, identityFillDetails(api.ErrDuplicatedIdentity, &spec, &otherSpec)
		}

		return authflow.NewNodeSimple(&NodeDoCreateIdentity{
			Identity: info,
		}), nil
	}

	return nil, authflow.ErrIncompatibleInput
}

func (n *NodeCreateIdentityLoginID) makeLoginIDSpec(loginID string) *identity.Spec {
	spec := &identity.Spec{
		Type: model.IdentityTypeLoginID,
		LoginID: &identity.LoginIDSpec{
			Value: loginID,
		},
	}
	switch n.Identification {
	case config.AuthenticationFlowIdentificationEmail:
		spec.LoginID.Type = model.LoginIDKeyTypeEmail
		spec.LoginID.Key = string(spec.LoginID.Type)
	case config.AuthenticationFlowIdentificationPhone:
		spec.LoginID.Type = model.LoginIDKeyTypePhone
		spec.LoginID.Key = string(spec.LoginID.Type)
	case config.AuthenticationFlowIdentificationUsername:
		spec.LoginID.Type = model.LoginIDKeyTypeUsername
		spec.LoginID.Key = string(spec.LoginID.Type)
	default:
		panic(fmt.Errorf("unexpected identification method: %v", n.Identification))
	}

	return spec
}
