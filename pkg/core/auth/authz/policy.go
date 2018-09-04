package authz

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/skygeario/skygear-server/pkg/server/skyerr"

	"github.com/skygeario/skygear-server/pkg/core/handler"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

type Policy interface {
	IsAllowed(r *http.Request, ctx handler.AuthenticationContext) error
}

type PolicyFunc func(r *http.Request, ctx handler.AuthenticationContext) error

func (f PolicyFunc) IsAllowed(r *http.Request, ctx handler.AuthenticationContext) error {
	return f(r, ctx)
}

type DefaultPolicy struct {
	policies []Policy
}

func NewDefaultPolicy(allow []string, deny []string) Policy {
	policies := []Policy{}

	for _, name := range allow {
		policies = append(policies, resolvePolicy(name, true))
	}

	for _, name := range deny {
		policies = append(policies, resolvePolicy(name, false))
	}

	return DefaultPolicy{policies: policies}
}

func resolvePolicy(name string, allow bool) Policy {
	if name == "everybody" {
		return EverybodyPolicy{allow: allow}
	}

	if name == "apikey" {
		if !allow {
			panic("deny:apikey is not allowed")
		}

		return PolicyFunc(RequireAPIKey)
	}

	if name == "masterkey" {
		if !allow {
			panic("deny:masterkey is not allowed")
		}

		return PolicyFunc(RequireMasterKey)
	}

	if name == "authenticated" {
		return AuthenticatedPolicy{allow: allow}
	}

	if name == "disabled" {
		return UserDisabledPolicy{allow: allow}
	}

	if strings.HasPrefix(name, "role/") {
		return RolePolicy{role: name[5:], allow: allow}
	}

	panic("Unexpected authorisation policy, name: " + name + ", allow: " + strconv.FormatBool(allow))
}

func (p DefaultPolicy) IsAllowed(r *http.Request, ctx handler.AuthenticationContext) error {
	for _, policy := range p.policies {
		if err := policy.IsAllowed(r, ctx); err != nil {
			return err
		}
	}

	return nil
}

type EverybodyPolicy struct {
	allow bool
}

func (p EverybodyPolicy) IsAllowed(r *http.Request, ctx handler.AuthenticationContext) error {
	if !p.allow {
		// TODO:
		// return proper error code
		return skyerr.NewError(skyerr.UnexpectedError, "everybody is denied")
	}

	return nil
}

func RequireAPIKey(r *http.Request, ctx handler.AuthenticationContext) error {
	keyType := model.GetAccessKeyType(r)
	if keyType == model.NoAccessKey {
		return skyerr.NewError(skyerr.AccessKeyNotAccepted, "api key required")
	}

	return nil
}

func RequireMasterKey(r *http.Request, ctx handler.AuthenticationContext) error {
	keyType := model.GetAccessKeyType(r)
	if keyType != model.MasterAccessKey {
		return skyerr.NewError(skyerr.AccessKeyNotAccepted, "master key required")
	}

	return nil
}

type AuthenticatedPolicy struct {
	allow bool
}

func (p AuthenticatedPolicy) IsAllowed(r *http.Request, ctx handler.AuthenticationContext) error {
	if p.allow && ctx.AuthInfo == nil {
		return skyerr.NewError(skyerr.NotAuthenticated, "require authenticated user")
	}

	if !p.allow && ctx.AuthInfo != nil {
		// TODO:
		// deny authenticated user?
		return nil
	}

	return nil
}

type UserDisabledPolicy struct {
	allow bool
}

func (p UserDisabledPolicy) IsAllowed(r *http.Request, ctx handler.AuthenticationContext) error {
	if ctx.AuthInfo == nil {
		return skyerr.NewError(skyerr.UnexpectedAuthInfoNotFound, "user authentication info not found")
	}

	if !p.allow && ctx.AuthInfo.Disabled {
		// TODO:
		// return proper error code
		return skyerr.NewError(skyerr.UnexpectedError, "user disabled")
	}

	if !p.allow && ctx.AuthInfo != nil {
		// TODO:
		// deny authenticated user?
		return nil
	}

	return nil
}

type RolePolicy struct {
	role  string
	allow bool
}

func (p RolePolicy) IsAllowed(r *http.Request, ctx handler.AuthenticationContext) error {
	if ctx.AuthInfo == nil {
		return skyerr.NewError(skyerr.UnexpectedAuthInfoNotFound, "user authentication info not found")
	}

	containsRole := false
	for _, role := range ctx.AuthInfo.Roles {
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
