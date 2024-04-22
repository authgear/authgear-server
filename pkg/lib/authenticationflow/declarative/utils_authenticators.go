package declarative

import (
	"github.com/authgear/authgear-server/pkg/api/model"
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
