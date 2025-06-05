package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticator"
)

func findPassword(in []*authenticator.Info, kind model.AuthenticatorKind) *authenticator.Info {
	for _, authn := range in {
		authn := authn
		if authn.Type != model.AuthenticatorTypePassword {
			continue
		}
		if authn.Kind == kind {
			return authn
		}
	}
	return nil
}

func findPrimaryPasskey(in []*authenticator.Info, kind model.AuthenticatorKind) *authenticator.Info {
	for _, authn := range in {
		authn := authn
		if authn.Type != model.AuthenticatorTypePasskey {
			continue
		}
		if authn.Kind == kind {
			return authn
		}
	}
	return nil
}

func findEmailOOB(in []*authenticator.Info, kind model.AuthenticatorKind, target string) *authenticator.Info {
	for _, authn := range in {
		authn := authn
		if authn.Type != model.AuthenticatorTypeOOBEmail {
			continue
		}
		if authn.Kind == kind && authn.OOBOTP.Email == target {
			return authn
		}
	}
	return nil
}

func findSMSOOB(in []*authenticator.Info, kind model.AuthenticatorKind, target string) *authenticator.Info {
	for _, authn := range in {
		authn := authn
		if authn.Type != model.AuthenticatorTypeOOBSMS {
			continue
		}
		if authn.Kind == kind && authn.OOBOTP.Phone == target {
			return authn
		}
	}
	return nil
}

func findTOTP(in []*authenticator.Info, kind model.AuthenticatorKind) *authenticator.Info {
	for _, authn := range in {
		authn := authn
		if authn.Type != model.AuthenticatorTypeTOTP {
			continue
		}
		if authn.Kind == kind {
			return authn
		}
	}
	return nil
}

func collectAssertedAuthenticators(flows authenticationflow.Flows) (authenticators []*authenticator.Info, err error) {
	err = authenticationflow.TraverseFlow(authenticationflow.Traverser{
		NodeSimple: func(nodeSimple authenticationflow.NodeSimple, w *authenticationflow.Flow) error {
			if n, ok := nodeSimple.(MilestoneDidAuthenticate); ok {
				if info, ok := n.MilestoneDidAuthenticateAuthenticator(); ok {
					authenticators = append(authenticators, info)
				}
			}
			if n, ok := nodeSimple.(MilestoneDoCreateAuthenticator); ok {
				authenticators = append(authenticators, n.MilestoneDoCreateAuthenticator())
			}
			return nil
		},
		Intent: func(intent authenticationflow.Intent, w *authenticationflow.Flow) error {
			if i, ok := intent.(MilestoneDidAuthenticate); ok {
				if info, ok := i.MilestoneDidAuthenticateAuthenticator(); ok {
					authenticators = append(authenticators, info)
				}
			}
			if i, ok := intent.(MilestoneDoCreateAuthenticator); ok {
				authenticators = append(authenticators, i.MilestoneDoCreateAuthenticator())
			}
			return nil
		},
	}, flows.Root)

	if err != nil {
		return
	}

	return
}
