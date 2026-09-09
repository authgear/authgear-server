package dpop

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/duration"
)

// replayTTL is how long a proof identifier is remembered.
//
// A proof is rejected once it is older than duration.Short, so remembering an
// identifier for longer than that plus the accepted clock skew is enough: by
// the time the entry expires, the proof itself no longer validates.
const replayTTL = duration.Short + duration.ClockSkew

func proofReplayKey(appID config.AppID, jkt string, jti string) string {
	// Keyed on the thumbprint as well as the jti, so that a jti chosen by one
	// client cannot lock out an unrelated client that happens to pick the same
	// value.
	return fmt.Sprintf("app:%s:dpop-proof:%s:%s", appID, jkt, jti)
}

// StoreRedis remembers the proofs that have been seen, so that each one is
// accepted at most once.
//
// RFC 9449 section 11.1 requires this: without it an intercepted proof stays
// replayable for as long as it remains within its validity window.
type StoreRedis struct {
	Redis *appredis.Handle
	AppID config.AppID
}

// MarkProofUsed records the proof and reports whether it had already been seen.
func (s *StoreRedis) MarkProofUsed(ctx context.Context, proof *DPoPProof) (alreadyUsed bool, err error) {
	key := proofReplayKey(s.AppID, proof.JKT, proof.JTI)
	err = s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		ok, err := conn.SetNX(ctx, key, []byte("1"), replayTTL).Result()
		if err != nil {
			return err
		}
		alreadyUsed = !ok
		return nil
	})
	if err != nil {
		return false, err
	}
	return alreadyUsed, nil
}
