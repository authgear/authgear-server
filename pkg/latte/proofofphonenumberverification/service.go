package proofofphonenumberverification

import (
	"context"
	"fmt"
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Service struct {
	Config  *config.ProofOfPhoneNumberVerificationHookConfig
	WebHook *ProofOfPhoneNumberVerificationWebHook
}

func (s *Service) Verify(ctx context.Context, proofOfPhoneNumberVerificationString string) (*HookResponse, error) {
	if s.Config.URL == "" {
		return nil, InvalidConfiguration.New("missing proof of phonenumber verification hook config")
	}

	u, err := url.Parse(s.Config.URL)
	if err != nil {
		return nil, err
	}

	req := &HookRequest{
		ProofOfPhoneNumberVerification: proofOfPhoneNumberVerificationString,
	}

	switch {
	case s.WebHook.SupportURL(u):
		return s.WebHook.Call(ctx, u, req)
	default:
		return nil, fmt.Errorf("unsupported hook URL: %v", u)
	}
}
