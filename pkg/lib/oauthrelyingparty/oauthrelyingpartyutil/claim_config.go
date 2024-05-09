package oauthrelyingpartyutil

import "github.com/authgear/authgear-server/pkg/api/oauthrelyingparty"

func Email_AssumeVerified_Required() oauthrelyingparty.ProviderClaimConfig {
	return oauthrelyingparty.ProviderClaimConfig{
		"assume_verified": true,
		"required":        true,
	}
}

func Email_AssumeVerified_NOT_Required() oauthrelyingparty.ProviderClaimConfig {
	return oauthrelyingparty.ProviderClaimConfig{
		"assume_verified": true,
		"required":        false,
	}
}
