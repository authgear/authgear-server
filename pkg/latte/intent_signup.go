package latte

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/ratelimit"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/uuid"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

var latteSignupLogger = slogutil.NewLogger("latte-intent-signup")

func init() {
	workflow.RegisterPrivateIntent(&IntentSignup{})
}

var IntentSignupSchema = validation.NewSimpleSchema(`
	{
		"type": "object",
		"additionalProperties": false
	}
`)

type IntentSignup struct {
	CaptchaProtectedIntent
	PhoneNumberHint string `json:"phone_number_hint"`
}

func (*IntentSignup) Kind() string {
	return "latte.IntentSignup"
}

func (*IntentSignup) JSONSchema() *validation.SimpleSchema {
	return IntentSignupSchema
}

func (*IntentSignup) CanReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) ([]workflow.Input, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		// Generate a new user ID.
		return nil, nil
	case 1:
		// Create a phone login ID.
		return nil, nil
	case 2:
		// Create a email login ID.
		// We assume the project is set to skip verify email on sign up.
		return nil, nil
	case 3:
		// Create a primary password.
		return nil, nil
	case 4:
		// Create a session, if needed.
		return nil, nil
	default:
		return nil, workflow.ErrEOF
	}
}

func (i *IntentSignup) ReactTo(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows, input workflow.Input) (*workflow.Node, error) {
	switch len(workflows.Nearest.Nodes) {
	case 0:
		return workflow.NewNodeSimple(&NodeDoCreateUser{
			UserID: uuid.New(),
		}), nil
	case 1:
		intent := &IntentCreateLoginID{
			// LoginID key and LoginID type are fixed here.
			UserID:          i.userID(workflows.Nearest),
			LoginIDType:     model.LoginIDKeyTypePhone,
			LoginIDKey:      string(model.LoginIDKeyTypePhone),
			PhoneNumberHint: i.PhoneNumberHint,
		}
		intent.IsCaptchaProtected = i.IsCaptchaProtected
		return workflow.NewSubWorkflow(intent), nil
	case 2:
		return workflow.NewSubWorkflow(&IntentCreateLoginID{
			// LoginID key and LoginID type are fixed here.
			UserID:          i.userID(workflows.Nearest),
			LoginIDType:     model.LoginIDKeyTypeEmail,
			LoginIDKey:      string(model.LoginIDKeyTypeEmail),
			PhoneNumberHint: i.PhoneNumberHint,
		}), nil
	case 3:
		// The type, kind is fixed here.
		return workflow.NewSubWorkflow(&IntentCreatePassword{
			UserID:                 i.userID(workflows.Nearest),
			AuthenticatorKind:      authenticator.KindPrimary,
			AuthenticatorIsDefault: false,
		}), nil
	case 4:
		mode := EnsureSessionModeCreate
		if workflow.GetSuppressIDPSessionCookie(ctx) {
			mode = EnsureSessionModeNoop
		}
		return workflow.NewSubWorkflow(&IntentEnsureSession{
			UserID:       i.userID(workflows.Nearest),
			CreateReason: session.CreateReasonSignup,
			// AMR is NOT populated because
			// 1. Strictly speaking this is NOT an authentication. It is a sign up.
			// 2. 3 authenticators were created. Should we report all 3?
			AMR:  nil,
			Mode: mode,
		}), nil
	}

	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentSignup) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			// Apply ratelimit on sign up.
			specs := ratelimit.RateLimitAuthenticationSignup.ResolveBucketSpecs(deps.Config, deps.FeatureConfig, deps.RateLimitsEnvConfig, &ratelimit.ResolveBucketSpecOptions{
				IPAddress: string(deps.RemoteIP),
			})
			for _, spec := range specs {
				spec := *spec
				failed, err := deps.RateLimiter.Allow(ctx, spec)
				if err != nil {
					return err
				}
				if err := failed.Error(); err != nil {
					return err
				}
			}
			return nil
		}),
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			// NOTE(DEV-2982): This is for debugging the session lost problem
			userID := i.userID(workflows.Nearest)
			now := deps.Clock.NowUTC()
			logger := latteSignupLogger.GetLogger(ctx)
			logger.WithSkipLogging().WithSkipStackTrace().Error(ctx, "updated last login",
				slog.String("user_id", userID),
				slog.Bool("refresh_token_log", true))
			return deps.Users.UpdateLoginTime(ctx, userID, now)
		}),
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			var identities []*identity.Info
			identityWorkflows := workflow.FindSubWorkflows[NewIdentityGetter](workflows.Nearest)
			for _, subWorkflow := range identityWorkflows {
				if iden, ok := subWorkflow.Intent.(NewIdentityGetter).GetNewIdentities(subWorkflow); ok {
					identities = append(identities, iden...)
				}
			}

			var authenticators []*authenticator.Info
			authenticatorWorkflows := workflow.FindSubWorkflows[NewAuthenticatorGetter](workflows.Nearest)
			for _, subWorkflow := range authenticatorWorkflows {
				if a, ok := subWorkflow.Intent.(NewAuthenticatorGetter).GetNewAuthenticators(subWorkflow); ok {
					authenticators = append(authenticators, a...)
				}
			}

			userID := i.userID(workflows.Nearest)
			isAdminAPI := false

			u, err := deps.Users.GetRaw(ctx, userID)
			if err != nil {
				return err
			}

			err = deps.Users.AfterCreate(ctx,
				u,
				identities,
				authenticators,
				isAdminAPI,
			)
			if err != nil {
				return err
			}

			return nil
		}),
	}, nil
}

func (*IntentSignup) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	return nil, nil
}

func (i *IntentSignup) userID(w *workflow.Workflow) string {
	node, ok := workflow.FindSingleNode[*NodeDoCreateUser](w)
	if !ok {
		panic(fmt.Errorf("expected userID"))
	}
	return node.UserID
}

type NewIdentityGetter interface {
	workflow.Intent
	GetNewIdentities(w *workflow.Workflow) ([]*identity.Info, bool)
}

type NewAuthenticatorGetter interface {
	workflow.Intent
	GetNewAuthenticators(w *workflow.Workflow) ([]*authenticator.Info, bool)
}
