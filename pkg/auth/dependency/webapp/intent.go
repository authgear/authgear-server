package webapp

import (
	"github.com/authgear/authgear-server/pkg/auth/dependency/newinteraction"
)

type Intent struct {
	RedirectURI      string
	ErrorRedirectURI string
	Intent           newinteraction.Intent
}
