package cimd

import (
	"context"
	"fmt"
	"time"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/crypto"
)

// fetchLockTTL must exceed FetchTimeout so a lock is never released while
// its holder is still fetching, and must be short enough that a holder that
// dies mid-fetch does not block refetches for long.
const fetchLockTTL = 10 * time.Second

// singleFlightPurposeDocument namespaces cimd.Service's document-fetch
// single-flight lock.
const singleFlightPurposeDocument = "cimd-fetch"

// FetchSingleFlight prevents N concurrent authorization requests for the
// same stale client_id from producing N simultaneous fetches of the same
// document -- a self-inflicted amplification attack on the client's own
// server, and N racing upserts.
type FetchSingleFlight struct {
	Redis *appredis.Handle
	AppID config.AppID
}

// Acquire reports whether the caller may perform the fetch. It is a plain
// SET NX with a TTL -- the established pattern in this repo (see
// pkg/lib/analytic/first_auth_sink.go's markFirstAuth) -- and deliberately
// NOT a blocking lock: a caller that loses the race serves the stale record
// (or fails, if there is none) immediately rather than queueing behind a 5s
// network call.
//
// purpose namespaces the lock key so the document fetch's single-flight and
// the logo fetch's single-flight (LogoService.Get) never collide, despite
// sharing this one implementation and TTL constant.
//
// There is no Release. Letting the key expire costs at most fetchLockTTL of
// extra staleness on a 1 hour interval, and an explicit release would have
// to be careful not to delete a lock a later holder now owns. Not worth the
// complexity here.
func (f *FetchSingleFlight) Acquire(ctx context.Context, purpose string, clientID string) (bool, error) {
	// clientID is hashed for the same reason ClientCache hashes it: it is an
	// attacker-influenced URL, and a ':' in it would otherwise let one
	// client_id collide with another's key namespace.
	key := fmt.Sprintf("app:%s:%s:%s", f.AppID, purpose, crypto.SHA256String(clientID))
	var acquired bool
	err := f.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		var err error
		acquired, err = conn.SetNX(ctx, key, "1", fetchLockTTL).Result()
		return err
	})
	return acquired, err
}
