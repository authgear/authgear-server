package authnsession

import (
	"context"

	"github.com/skygeario/skygear-server/pkg/auth/dependency/hook"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/mfa"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/principal"
	"github.com/skygeario/skygear-server/pkg/auth/dependency/userprofile"
	"github.com/skygeario/skygear-server/pkg/core/auth/authinfo"
	"github.com/skygeario/skygear-server/pkg/core/auth/session"
	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/core/time"
)

func NewMockProvider(
	mfaConfiguration *config.MFAConfiguration,
	timeProvider time.Provider,
	mfaProvider mfa.Provider,
	authInfoStore authinfo.Store,
	sessionProvider session.Provider,
	sessionWriter session.Writer,
	identityProvider principal.IdentityProvider,
	hookProvider hook.Provider,
	userProfileStore userprofile.Store,
) Provider {
	authenticationSessionConfiguration :=
		&config.AuthenticationSessionConfiguration{
			Secret: "authnsessionsecret",
		}
	return NewProvider(
		context.Background(),
		mfaConfiguration,
		authenticationSessionConfiguration,
		timeProvider,
		mfaProvider,
		authInfoStore,
		sessionProvider,
		sessionWriter,
		identityProvider,
		hookProvider,
		userProfileStore,
	)
}
