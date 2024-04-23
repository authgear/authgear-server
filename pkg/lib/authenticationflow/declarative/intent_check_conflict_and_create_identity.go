package declarative

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/slice"
	"github.com/iawaknahc/jsonschema/pkg/jsonpointer"
)

func init() {
	authflow.RegisterIntent(&IntentCheckConflictAndCreateIdenity{})
}

type IntentCheckConflictAndCreateIdenity struct {
	JSONPointer jsonpointer.T          `json:"json_pointer,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Request     *CreateIdentityRequest `json:"request,omitempty"`
}

var _ authflow.Intent = &IntentCheckConflictAndCreateIdenity{}

func (*IntentCheckConflictAndCreateIdenity) Kind() string {
	return "IntentCheckConflictAndCreateIdenity"
}

func (*IntentCheckConflictAndCreateIdenity) CanReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows) (authflow.InputSchema, error) {
	switch len(flows.Nearest.Nodes) {
	case 0: // next node is NodeDoCreateIdentity, or account linking intent
		return nil, nil
	}
	return nil, authflow.ErrEOF
}

func (i *IntentCheckConflictAndCreateIdenity) ReactTo(ctx context.Context, deps *authflow.Dependencies, flows authflow.Flows, input authflow.Input) (*authflow.Node, error) {
	switch len(flows.Nearest.Nodes) {
	case 0: // next node is NodeDoCreateIdentity, or account linking intent
		cfg, conflicts, err := i.checkConflictByAccountLinkings(ctx, deps, flows)
		if err != nil {
			return nil, err
		}
		spec := i.getIdenitySpec()
		if len(conflicts) == 0 {
			info, err := newIdentityInfo(deps, i.UserID, spec)
			if err != nil {
				return nil, err
			}
			return authflow.NewNodeSimple(&NodeDoCreateIdentity{
				Identity: info,
			}), nil
		} else {
			// Currently skipLogin is always false
			// We may support always_link_without_login later
			var skipLogin bool = false
			var loginFlow string = ""
			switch cfg.GetAction() {
			case config.AuthenticationFlowAccountLinkingActionError:
				spec := spec
				conflictSpecs := slice.Map(conflicts, func(i *identity.Info) *identity.Spec {
					s := i.ToSpec()
					return &s
				})
				return nil, identityFillDetailsMany(api.ErrDuplicatedIdentity, spec, conflictSpecs)
			case config.AuthenticationFlowAccountLinkingActionLoginAndLink:
				loginFlow = cfg.GetLoginFlow()
				if loginFlow == "" {
					// Use the current flow name if it is not specified
					loginFlow = authflow.FindCurrentFlowReference(flows.Root).Name
				}
			default:
				panic(fmt.Errorf("unknown action %v", cfg.GetAction()))
			}
			return authflow.NewSubFlow(&IntentAccountLinking{
				JSONPointer:           i.JSONPointer,
				ConflictingIdentities: conflicts,
				OAuthIdentitySpec:     spec,
				SkipLogin:             skipLogin,
				LoginFlowName:         loginFlow,
			}), nil
		}
	}
	return nil, authflow.ErrIncompatibleInput
}

func (i *IntentCheckConflictAndCreateIdenity) checkConflictByAccountLinkings(
	ctx context.Context,
	deps *authflow.Dependencies,
	flows authflow.Flows) (config config.AccountLinkingConfigObject, conflicts []*identity.Info, err error) {
	switch i.Request.Type {
	case model.IdentityTypeOAuth:
		return linkByOAuthIncomingOAuthSpec(ctx, deps, flows, i.Request.OAuth)
	default:
		// Linking of other types are not supported at the moment
		return nil, []*identity.Info{}, nil
	}
}

func (i *IntentCheckConflictAndCreateIdenity) getIdenitySpec() *identity.Spec {
	var spec *identity.Spec
	switch i.Request.Type {
	case model.IdentityTypeLoginID:
		spec = i.Request.LoginID.Spec
	case model.IdentityTypeOAuth:
		spec = i.Request.OAuth.Spec
	default:
		panic(fmt.Errorf("unexpected identity type %v", i.Request.Type))
	}
	return spec
}
