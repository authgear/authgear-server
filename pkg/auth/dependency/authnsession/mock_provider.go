package authnsession

import (
	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	authTesting "github.com/skygeario/skygear-server/pkg/core/auth/testing"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func NewMockProvider(
	timeProvider time.Provider,
	authInfoStore authinfo.Store,
	sessionProvider session.Provider,
	sessionWriter session.Writer,
	identityProvider principal.IdentityProvider,
	hookProvider hook.Provider,
	userProfileStore userprofile.Store,
) Provider {
	authContext := authTesting.NewMockContext()
	mfaConfiguration := config.MFAConfiguration{
		Enforcement: config.MFAEnforcementOff,
	}
	authenticationSessionConfiguration :=
		config.AuthenticationSessionConfiguration{
			Secret: "authnsessionsecret",
		}
	return NewProvider(
		authContext,
		mfaConfiguration,
		authenticationSessionConfiguration,
		timeProvider,
		authInfoStore,
		sessionProvider,
		sessionWriter,
		identityProvider,
		hookProvider,
		userProfileStore,
	)
}
