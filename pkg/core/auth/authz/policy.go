package authz

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"

	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

type PolicyProvider interface {
	ProvideAuthzPolicy() Policy
}

type Policy interface {
	IsAllowed(r *http.Request, authInfo auth.AuthInfo) error
}

type PolicyFunc func(r *http.Request, authInfo auth.AuthInfo) error

func (f PolicyFunc) IsAllowed(r *http.Request, authInfo auth.AuthInfo) error {
	return f(r, authInfo)
}

type AllOfPolicy struct {
	policies []Policy
}

func NewAllOfPolicy(policies ...Policy) Policy {
	return AllOfPolicy{policies: policies}
}

func (p AllOfPolicy) IsAllowed(r *http.Request, authInfo auth.AuthInfo) error {
	for _, policy := range p.policies {
		if err := policy.IsAllowed(r, authInfo); err != nil {
			return err
		}
	}

	return nil
}

type EverybodyPolicy struct {
	allow bool
}

func (p EverybodyPolicy) IsAllowed(r *http.Request, authInfo auth.AuthInfo) error {
	if !p.allow {
		// TODO:
		// return proper error code
		return skyerr.NewError(skyerr.UnexpectedError, "everybody is denied")
	}

	return nil
}

func DenyNoAccessKey(r *http.Request, authInfo auth.AuthInfo) error {
	keyType := authInfo.AccessKeyType
	if keyType == model.NoAccessKey {
		return skyerr.NewError(skyerr.AccessKeyNotAccepted, "api key required")
	}

	return nil
}

func RequireMasterKey(r *http.Request, authInfo auth.AuthInfo) error {
	keyType := authInfo.AccessKeyType
	if keyType != model.MasterAccessKey {
		return skyerr.NewError(skyerr.AccessKeyNotAccepted, "master key required")
	}

	return nil
}

func RequireAuthenticated(r *http.Request, authInfo auth.AuthInfo) error {
	if authInfo.AuthInfo == nil {
		return skyerr.NewError(skyerr.NotAuthenticated, "require authenticated user")
	}

	if authInfo.TokenValidSince != nil {
		tokenValidSince := *authInfo.TokenValidSince

		// Not all types of access token support this field. The token is
		// still considered if it does not have an issue time.
		//
		// Due to precision, the issue time of the token can be before
		// AuthInfo.TokenValidSince. We consider the token still valid
		// if the token is issued within 1 second before tokenValidSince.
		if !authInfo.Token.IssuedAt().IsZero() &&
			authInfo.Token.IssuedAt().After(tokenValidSince.Add(-1*time.Second)) {
			return skyerr.NewError(skyerr.NotAuthenticated, "require authenticated user")
		}
	}

	return nil
}

func DenyDisabledUser(r *http.Request, authInfo auth.AuthInfo) error {
	if authInfo.AuthInfo == nil {
		return skyerr.NewError(skyerr.UnexpectedAuthInfoNotFound, "user authentication info not found")
	}

	if authInfo.Disabled {
		// TODO:
		// return proper error code
		return skyerr.NewError(skyerr.UnexpectedError, "user disabled")
	}

	return nil
}

type RolePolicy struct {
	role  string
	allow bool
}

func (p RolePolicy) IsAllowed(r *http.Request, authInfo auth.AuthInfo) error {
	if authInfo.AuthInfo == nil {
		return skyerr.NewError(skyerr.UnexpectedAuthInfoNotFound, "user authentication info not found")
	}

	containsRole := false
	for _, role := range authInfo.Roles {
		if role == p.role {
			containsRole = true
			break
		}
	}

	if p.allow && !containsRole {
		// TODO:
		// return proper error code
		return skyerr.NewError(skyerr.UnexpectedError, "Role "+p.role+" is required")
	}

	if !p.allow && containsRole {
		// TODO:
		// return proper error code
		return skyerr.NewError(skyerr.UnexpectedError, "Unexpected role "+p.role)
	}

	return nil
}
