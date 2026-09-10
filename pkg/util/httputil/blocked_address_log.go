package httputil

import (
	"context"
	"errors"
	"log/slog"

	"github.com/authgear/authgear-server/pkg/util/slogutil"
)

// BlockedAddressLogKey is the field to alert on. It appears on exactly one log
// message, so filtering for it yields every outbound fetch this deployment
// refused on address grounds and nothing else.
const BlockedAddressLogKey = "blocked_by_ssrf_policy"

// InsecureFetchAddressAllowedFlag names the setting that lifts the
// restriction, so whoever reads the log does not have to go looking for it.
const InsecureFetchAddressAllowedFlag = "http.insecure_fetch_address_allowed"

var BlockedAddressLogger = slogutil.NewLogger("ssrf-address-policy")

// logIfBlockedAddress writes one ERROR when err is SafeDialer refusing the
// destination. Anything else -- a DNS failure, a refused connection, a
// timeout -- is not this log's business and passes through silently.
//
// A blocked fetch would otherwise be indistinguishable from an unreachable
// host in the delivery-failure log the caller already writes, which is the
// difference between "their server is down" and "we are refusing to call their
// server" -- an operator needs to know which.
func logIfBlockedAddress(ctx context.Context, err error, sink string, host string) {
	if !errors.Is(err, ErrBlockedAddress) {
		return
	}

	logger := BlockedAddressLogger.GetLogger(ctx)
	logger.WithError(err).Error(ctx, "outbound fetch blocked by SSRF address policy",
		slog.Bool(BlockedAddressLogKey, true),
		slog.String("sink", sink),
		slog.String("host", host),
		slog.String("flag", InsecureFetchAddressAllowedFlag),
	)
}
