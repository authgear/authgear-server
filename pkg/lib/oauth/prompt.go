package oauth

import (
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/slice"
)

type PromptResolver struct {
	Clock clock.Clock
}

func (r *PromptResolver) ResolvePrompt(req protocol.AuthorizationRequest, sidSession session.ListableSession) (prompt []string) {
	prompt = req.Prompt()
	if maxAge, ok := req.MaxAge(); ok {
		impliesPromptLogin := false
		// When there is no session, the presence of max_age implies prompt=login.
		if sidSession == nil {
			impliesPromptLogin = true
		} else {
			// max_age=0 implies prompt=login
			if maxAge == 0 {
				impliesPromptLogin = true
			} else {
				// max_age=n implies prompt=login if elapsed time is greater than max_age.
				// In extreme rare case, elapsed time can be negative.
				elapsedTime := r.Clock.NowUTC().Sub(sidSession.GetAuthenticationInfo().AuthenticatedAt)
				if elapsedTime < 0 || elapsedTime > maxAge {
					impliesPromptLogin = true
				}
			}
		}
		if impliesPromptLogin {
			prompt = slice.AppendIfUniqueStrings(prompt, "login")
		}
	}

	return
}
