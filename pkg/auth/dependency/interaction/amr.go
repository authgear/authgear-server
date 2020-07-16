package interaction

import (
	"sort"

	"github.com/authgear/authgear-server/pkg/auth/dependency/authenticator"
	"github.com/authgear/authgear-server/pkg/core/authn"
)

func DeriveAMR(primary *authenticator.Info, secondary *authenticator.Info) []string {
	seen := make(map[string]struct{})
	out := []string{}

	if primary != nil {
		for _, value := range primary.AMR() {
			_, ok := seen[value]
			if !ok {
				seen[value] = struct{}{}
				out = append(out, value)
			}
		}
	}

	if secondary != nil {
		if secondary.Type != authn.AuthenticatorTypeBearerToken {
			out = append(out, "mfa")
		}
		for _, value := range secondary.AMR() {
			_, ok := seen[value]
			if !ok {
				seen[value] = struct{}{}
				out = append(out, value)
			}
		}
	}

	sort.Strings(out)

	return out
}
