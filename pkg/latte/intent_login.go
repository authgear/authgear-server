package latte

import (
	"context"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/api"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
	"github.com/authgear/authgear-server/pkg/lib/authn/identity"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/mail"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/workflow"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
	"github.com/authgear/authgear-server/pkg/util/validation"
)

func init() {
	workflow.RegisterPrivateIntent(&IntentLogin{})
}

var latteLoginLogger = slogutil.NewLogger("latte-intent-login")

type VerifiedAuthenticationLockoutMethodGetter interface {
	GetVerifiedAuthenticationLockoutMethod() (config.AuthenticationLockoutMethod, bool)
}

var IntentLoginSchema = validation.NewSimpleSchema(`{}`)

type IntentLogin struct {
	CaptchaProtectedIntent
	Identity         *identity.Info `json:"identity,omitempty"`
	IdentityVerified bool           `json:"identity_verified"`
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
		phoneAuthenticator, err := i.getAuthenticator(ctx, deps,
			authenticator.KeepPrimaryAuthenticatorOfIdentity(i.Identity),
			authenticator.KeepType(model.AuthenticatorTypeOOBSMS),
		)
		if err != nil {
			return nil, err
		}
		intent := &IntentAuthenticateOOBOTPPhone{
			Authenticator:         phoneAuthenticator,
			AuthenticatorVerified: i.IdentityVerified,
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
				emailAuthenticator, err := i.getAuthenticator(ctx, deps,
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
		mode := EnsureSessionModeCreate
		if workflow.GetSuppressIDPSessionCookie(ctx) {
			mode = EnsureSessionModeNoop
		}
		return workflow.NewSubWorkflow(&IntentEnsureSession{
			UserID:       i.userID(),
			CreateReason: session.CreateReasonLogin,
			AMR:          GetAMR(workflows.Nearest),
			Mode:         mode,
		}), nil
	}
	return nil, workflow.ErrIncompatibleInput
}

func (i *IntentLogin) GetEffects(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (effs []workflow.Effect, err error) {
	logger := latteLoginLogger.GetLogger(ctx)
	return []workflow.Effect{
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			// NOTE(DEV-2982): This is for debugging the session lost problem
			userID := i.userID()
			now := deps.Clock.NowUTC()
			logger.WithSkipLogging().Error(ctx, "updated last login",
				slog.String("user_id", userID),
				slog.Bool("refresh_token_log", true))
			return deps.Users.UpdateLoginTime(ctx, userID, now)
		}),
		workflow.OnCommitEffect(func(ctx context.Context, deps *workflow.Dependencies) error {
			createSession, workflow := workflow.MustFindSubWorkflow[*IntentEnsureSession](workflows.Nearest)
			session := createSession.GetSession(workflow)
			if session == nil {
				// NOTE(DEV-2982): This is for debugging the session lost problem
				userID := i.userID()
				logger.WithSkipLogging().Error(ctx, "user.authenticated event skipped because session is nil",
					slog.String("user_id", userID),
					slog.Bool("refresh_token_log", true))
				return nil
			}

			// ref: https://github.com/authgear/authgear-server/issues/2930
			userRef := model.UserRef{
				Meta: model.Meta{
					ID: i.userID(),
				},
			}
			err := deps.Events.DispatchEventOnCommit(ctx, &nonblocking.UserAuthenticatedEventPayload{
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
			userID := i.userID()
			methods, err := i.getVerifiedAuthenticationLockoutMethods(workflows.Nearest)
			if err != nil {
				return err
			}
			err = deps.Authenticators.ClearLockoutAttempts(ctx, userID, methods)
			if err != nil {
				return err
			}
			return nil
		}),
	}, nil
}

func (i *IntentLogin) OutputData(ctx context.Context, deps *workflow.Dependencies, workflows workflow.Workflows) (interface{}, error) {
	methods, err := i.getVerifiedAuthenticationLockoutMethods(workflows.Nearest)
	if err != nil {
		return nil, err
	}
	if len(methods) == 0 {
		return nil, nil
	}
	userID := i.userID()
	identities, err := deps.Identities.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	maskedEmails := []string{}
	for _, identity := range identities {
		if identity.Type != model.IdentityTypeLoginID || identity.LoginID.LoginIDType != model.LoginIDKeyTypeEmail {
			continue
		}
		maskedEmails = append(maskedEmails, mail.MaskAddress(identity.LoginID.LoginID))
	}

	mfaTypes := []model.AuthenticatorType{}
	mfas, err := deps.Authenticators.List(ctx, userID, authenticator.KeepType(model.AuthenticatorTypeOOBEmail, model.AuthenticatorTypePassword))
	if err != nil {
		return nil, err
	}
	for _, mfa := range mfas {
		mfaTypes = append(mfaTypes, mfa.Type)
	}

	type IntentLoginOutput struct {
		MaskedEmails       []string                  `json:"masked_emails"`
		AuthenticatorTypes []model.AuthenticatorType `json:"authenticator_types"`
	}

	return &IntentLoginOutput{
		MaskedEmails:       maskedEmails,
		AuthenticatorTypes: mfaTypes,
	}, nil
}

func (i *IntentLogin) getAuthenticator(ctx context.Context, deps *workflow.Dependencies, filters ...authenticator.Filter) (*authenticator.Info, error) {
	ais, err := deps.Authenticators.List(ctx, i.Identity.UserID, filters...)
	if err != nil {
		return nil, err
	}

	if len(ais) == 0 {
		return nil, api.ErrNoAuthenticator
	}

	return ais[0], nil
}

func (i *IntentLogin) getVerifiedAuthenticationLockoutMethods(w *workflow.Workflow) ([]config.AuthenticationLockoutMethod, error) {
	result := []config.AuthenticationLockoutMethod{}
	err := w.Traverse(workflow.WorkflowTraverser{
		NodeSimple: func(nodeSimple workflow.NodeSimple, w *workflow.Workflow) error {
			if n, ok := nodeSimple.(VerifiedAuthenticationLockoutMethodGetter); ok {
				m, ok := n.GetVerifiedAuthenticationLockoutMethod()
				if ok {
					result = append(result, m)
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

func (i *IntentLogin) userID() string {
	return i.Identity.UserID
}
