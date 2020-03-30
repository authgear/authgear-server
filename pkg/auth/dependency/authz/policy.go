package authz

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/auth/authz"
	"github.com/skygeario/skygear-server/pkg/core/auth/authz/policy"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
)

var errAuthAPIDisabled = skyerr.Forbidden.
	WithReason("AuthAPIDisabled").
	New("auth API is disabled")

func requireAuthAPIEnabled(r *http.Request) error {
	config := config.GetTenantConfig(r.Context())
	if !config.AppConfig.Auth.EnableAPI {
		return errAuthAPIDisabled
	}

	return nil
}

var AuthAPIRequireValidUser = policy.AllOf(
	authz.PolicyFunc(requireAuthAPIEnabled),
	policy.RequireValidUser,
)

var AuthAPIRequireClient = policy.AllOf(
	authz.PolicyFunc(requireAuthAPIEnabled),
	authz.PolicyFunc(policy.RequireClient),
)

var AuthAPIRequireValidUserOrMasterKey = policy.AnyOf(
	authz.PolicyFunc(policy.RequireMasterKey),
	AuthAPIRequireValidUser,
)
