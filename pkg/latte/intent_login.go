package latte

import (
	"context"
	"sort"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentLogin{})
}

type VerifiedAuthenticatorGetter interface {
	GetVerifiedAuthenticator() (*authenticator.Info, bool)
}

var IntentLoginSchema = validation.NewSimpleSchema(`{}`)

type IntentLogin struct {
	CaptchaProtectedIntent
	Identity *identity.Info `json:"identity,omitempty"`
}

func (*IntentLogin) Kind() string {
	return "latte.IntentLogin"
}

func (*IntentLogin) JSONSchema() *validation.SimpleSchema {
	return IntentLoginSchema
}

func (*IntentLogin) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		return nil, nil
	case 1:
		return []workflow.Input{
			&InputSelectAuthenticatorType{},
		}, nil
	case 2:
		// Create a session, if needed.
		return nil, nil
	}

	return nil, workflow.ErrEOF
}

func (i *IntentLogin) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		// 1st step: authenticate oob otp phone
		phoneAuthenticator, err := i.getAuthenticator(deps,
			authenticator.KeepPrimaryAuthenticatorOfIdentity(i.Identity),
			authenticator.KeepType(model.AuthenticatorTypeOOBSMS),
		)
		if err != nil {
			return nil, err
		}
		intent := &IntentAuthenticateOOBOTPPhone{
			Authenticator: phoneAuthenticator,
		}
		intent.IsCaptchaProtected = i.IsCaptchaProtected
		return workflow.NewSubWorkflow(intent), nil
	case 1:
		// 2nd step: authenticate email login link / password
		var inputSelectAuthenticatorType inputSelectAuthenticatorType
		switch {
		case workflow.AsInput(input, &inputSelectAuthenticatorType):
			typ := inputSelectAuthenticatorType.GetAuthenticatorType()
			switch typ {
			case model.AuthenticatorTypeOOBEmail:
				emailAuthenticator, err := i.getAuthenticator(deps,
					authenticator.KeepKind(authenticator.KindPrimary),
					authenticator.KeepType(model.AuthenticatorTypeOOBEmail),
				)
				if err != nil {
					return nil, err
				}
				return workflow.NewSubWorkflow(&IntentAuthenticateEmailLoginLink{
					Authenticator: emailAuthenticator,
				}), nil
			case model.AuthenticatorTypePassword:
				return workflow.NewSubWorkflow(&IntentAuthenticatePassword{
					UserID:            i.userID(),
					AuthenticatorKind: authenticator.KindPrimary,
				}), nil
			default:
				return nil, workflow.ErrIncompatibleInput
			}
		}
	case 2:
		return workflow.NewSubWorkflow(&IntentCreateSession{
			UserID:       i.userID(),
			CreateReason: session.CreateReasonLogin,
			AMR:          i.GetAMR(workflows.Nearest),
			SkipCreate:   workflow.GetSuppressIDPSessionCookie(ctx),
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentLogin) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			createSession, workflow := workflow.MustFindSubWorkflow[*IntentCreateSession](workflows.Nearest)
			session := createSession.GetSession(workflow)
			if session == nil {
				return nil
			}

			// ref: https://github.com/authgear/authgear-server/issues/2930
			userRef := model.UserRef{
				Meta: model.Meta{
					ID: i.userID(),
				},
			}
			err := deps.Events.DispatchEvent(&nonblocking.UserAuthenticatedEventPayload{
				UserRef:  userRef,
				Session:  *session.ToAPIModel(),
				AdminAPI: false,
			})
			if err != nil {
				return err
			}

			return nil
		}),
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			authenticators, err := i.getVerifiedAuthenticators(workflows.Nearest)
			if err != nil {
				return err
			}
			err = deps.Authenticators.ClearLockoutAttempts(authenticators)
			if err != nil {
				return err
			}
			return nil
		}),
	}, nil
}

func (i *IntentLogin) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (i *IntentLogin) getAuthenticator(deps *workflow.Dependencies, filters ...authenticator.Filter) (*authenticator.Info, error) {
	ais, err := deps.Authenticators.List(i.Identity.UserID, filters...)
	if err != nil {
		return nil, err
	}

	if len(ais) == 0 {
		return nil, api.ErrNoAuthenticator
	}

	return ais[0], nil
}

func (i *IntentLogin) getVerifiedAuthenticators(w *workflow.Workflow) ([]*authenticator.Info, error) {
	result := []*authenticator.Info{}
	err := w.Traverse(workflow.WorkflowTraverser{
		NodeSimple: func(nodeSimple workflow.NodeSimple, w *workflow.Workflow) error {
			if n, ok := nodeSimple.(VerifiedAuthenticatorGetter); ok {
				info, ok := n.GetVerifiedAuthenticator()
				if ok && info != nil {
					result = append(result, info)
				}
			}

			return nil
		},
		Intent: func(intent workflow.Intent, w *workflow.Workflow) error {
			return nil
		},
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (i *IntentLogin) GetAMR(w *workflow.Workflow) []string {
	amrSet := map[string]interface{}{}
	workflows := workflow.FindSubWorkflows[AMRGetter](w)

	authCount := 0
	for _, perWorkflow := range workflows {
		if amrs := perWorkflow.Intent.(AMRGetter).GetAMR(perWorkflow); len(amrs) > 0 {
			authCount++
			for _, value := range amrs {
				amrSet[value] = struct{}{}
			}
		}
	}

	if authCount >= 2 {
		amrSet[model.AMRMFA] = struct{}{}
	}

	amr := make([]string, 0, len(amrSet))
	for k := range amrSet {
		amr = append(amr, k)
	}
	sort.Strings(amr)

	return amr
}

func (i *IntentLogin) userID() string {
	return i.Identity.UserID
}

type AMRGetter interface {
	workflow.Intent
	GetAMR(w *workflow.Workflow) []string
}
