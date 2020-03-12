package authn

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/loginid"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/async"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
)

type Provider interface {
	// CreateUserWithLoginIDs is sign up.
	CreateUserWithLoginIDs(
		loginIDs []loginid.LoginID,
		plainPassword string,
		metadata map[string]interface{},
		onUserDuplicate model.OnUserDuplicate,
	) (authInfo *authinfo.AuthInfo, userprofile *userprofile.UserProfile, firstPrincipal principal.Principal, tasks []async.TaskSpec, err error)
	// AuthenticateWithLoginID is sign in.
	AuthenticateWithLoginID(loginID loginid.LoginID, plainPassword string) (authInfo *authinfo.AuthInfo, principal principal.Principal, err error)
}
