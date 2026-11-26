package loginlink

import (
	"context"

	"github.com/authgear/authgear-server/pkg/api/model"
	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/otp"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

//go:generate go tool mockgen -source=service.go -destination=service_mock_test.go -package loginlink

type OTPCodeService interface {
	LookupCode(ctx context.Context, purpose otp.Purpose, code string) (target string, err error)
	VerifyOTP(ctx context.Context, kind otp.Kind, target string, otp string, opts *otp.VerifyOptions) error
	SetSubmittedCode(ctx context.Context, kind otp.Kind, target string, code string) (*otp.State, error)
}

type EventStore interface {
	Publish(ctx context.Context, websocketChannelName string, e authflow.Event) error
}

// Service approves login links. See docs/specs/login-link.md.
type Service struct {
	Config   *config.AppConfig
	OTPCodes OTPCodeService
	Events   EventStore
}

// Inspect returns nil when code can be approved.
func (s *Service) Inspect(ctx context.Context, code string) error {
	_, _, err := s.verify(ctx, code)
	return err
}

// Approve approves code and refreshes the Authentication Flow that sent it.
func (s *Service) Approve(ctx context.Context, code string) error {
	kind, target, err := s.verify(ctx, code)
	if err != nil {
		return err
	}

	state, err := s.OTPCodes.SetSubmittedCode(ctx, kind, target, code)
	if err != nil {
		return err
	}

	if state.AuthenticationFlowWebsocketChannelName != "" {
		err = s.Events.Publish(ctx, state.AuthenticationFlowWebsocketChannelName, authflow.NewEventRefresh())
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) verify(ctx context.Context, code string) (kind otp.Kind, target string, err error) {
	kind = otp.KindOOBOTPLink(s.Config, model.AuthenticatorOOBChannelEmail)

	target, err = s.OTPCodes.LookupCode(ctx, kind.Purpose(), code)
	if err != nil {
		return nil, "", err
	}

	// Rejects a consumed code, or an old code after a resend, without consuming.
	err = s.OTPCodes.VerifyOTP(ctx, kind, target, code, &otp.VerifyOptions{SkipConsume: true})
	if err != nil {
		return nil, "", err
	}

	return kind, target, nil
}
