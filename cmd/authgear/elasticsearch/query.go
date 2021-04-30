package elasticsearch

import (
	identityloginid "github.com/authgear/authgear-server/pkg/lib/authn/identity/loginid"
	identityoauth "github.com/authgear/authgear-server/pkg/lib/authn/identity/oauth"
	"github.com/authgear/authgear-server/pkg/lib/authn/user"
)

type Query struct {
	AppID   config.AppID
	Users   *user.Store
	OAuth   *identityoauth.Store
	LoginID *identityloginid.Store
}
