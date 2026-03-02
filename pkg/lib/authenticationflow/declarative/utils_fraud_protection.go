package declarative

import (
	"context"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
)

// revertUnverifiedSMSOTPs returns an OnCommitEffect that drains the leaky
// bucket for any SMS OTPs that were sent but never verified (alt-auth
// exclusion). It only runs on the root flow so that sub-flows (e.g. a login
// flow embedded inside a signup flow for account linking) do not double-drain.
func revertUnverifiedSMSOTPs(flows authflow.Flows) authflow.Effect {
	return authflow.OnCommitEffect(func(ctx context.Context, deps *authflow.Dependencies) error {
		if flows.Nearest != flows.Root {
			return nil
		}
		session := authflow.GetSession(ctx)
		for phone, sentCount := range session.SMSOTPSentCountByPhone {
			verifiedCount := session.SMSOTPVerifiedCountByPhone[phone]
			unverifiedCount := sentCount - verifiedCount
			if unverifiedCount > 0 {
				deps.FraudProtection.RevertSMSOTPSent(ctx, phone, unverifiedCount)
			}
		}
		return nil
	})
}
