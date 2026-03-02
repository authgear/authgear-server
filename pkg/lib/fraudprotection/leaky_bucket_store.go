package fraudprotection

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

// LeakyBucketThresholds holds per-bucket adaptive threshold values.
type LeakyBucketThresholds struct {
	CountryHourly float64 // used with period=3600
	CountryDaily  float64 // used with period=86400
	IPHourly      float64 // used with period=3600
	IPDaily       float64 // used with period=86400
}

// LeakyBucketTriggered holds the per-bucket warning trigger state.
// A bucket is triggered when new_level > threshold after the fill.
type LeakyBucketTriggered struct {
	CountryHourly    bool // SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__HOURLY_THRESHOLD_EXCEEDED
	CountryDaily     bool // SMS__UNVERIFIED_OTPS__BY_PHONE_COUNTRY__DAILY_THRESHOLD_EXCEEDED
	IPHourly         bool // SMS__UNVERIFIED_OTPS__BY_IP__HOURLY_THRESHOLD_EXCEEDED
	IPDaily          bool // SMS__UNVERIFIED_OTPS__BY_IP__DAILY_THRESHOLD_EXCEEDED
	IPCountriesDaily bool // SMS__PHONE_COUNTRIES__BY_IP__DAILY_THRESHOLD_EXCEEDED
}

// leakyBucketScript is the unified Lua script for fill/drain/read operations.
// KEYS[1] = bucket key
// ARGV[1] = now (unix seconds, float)
// ARGV[2] = threshold (current adaptive threshold for this bucket)
// ARGV[3] = period (window_seconds: 3600 or 86400)
// ARGV[4] = n (positive=fill, negative=drain, 0=read)
// ARGV[5] = ttl_seconds (2 * window_seconds; only applied when n=1)
// Returns {new_level (integer, truncated), triggered (0 or 1)}.
var leakyBucketScript = `
local now       = tonumber(ARGV[1])
local threshold = tonumber(ARGV[2])
local period    = tonumber(ARGV[3])
local n         = tonumber(ARGV[4])
local ttl       = tonumber(ARGV[5])

local drain_rate = threshold / period

local data = redis.call('HMGET', KEYS[1], 'level', 'last_updated')
local level        = tonumber(data[1]) or 0
local last_updated = tonumber(data[2]) or now

-- Cap level at the current threshold before any operation.
level = math.min(level, threshold)

local elapsed   = now - last_updated
local leaked    = elapsed * drain_rate
local new_level = math.max(0, level - leaked + n)

if n ~= 0 then
    redis.call('HMSET', KEYS[1], 'level', new_level, 'last_updated', now)
    if n == 1 then
        redis.call('EXPIRE', KEYS[1], ttl)
    end
end

return {new_level, (new_level > threshold) and 1 or 0}
`

// ipCountriesScript tracks distinct countries seen from a given IP in the past 24h.
// KEYS[1] = sorted set key
// ARGV[1] = alpha2 country code
// ARGV[2] = now (unix timestamp)
// ARGV[3] = threshold (fixed = 3)
// ARGV[4] = ttl_seconds (2 * 86400)
// Returns {count, triggered_int}.
var ipCountriesScript = `
local now    = tonumber(ARGV[2])
local cutoff = now - 86400

redis.call('ZADD', KEYS[1], 'GT', now, ARGV[1])
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', cutoff)
redis.call('EXPIRE', KEYS[1], ARGV[4])

local count = redis.call('ZCARD', KEYS[1])
return {count, (count > tonumber(ARGV[3])) and 1 or 0}
`

type LeakyBucketStore struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

// RecordSMSOTPSent atomically fills all 4 leaky buckets and updates the ip_countries ZSET.
// Returns LeakyBucketTriggered indicating which warning conditions are now exceeded.
// Called BEFORE the SMS send; blocked requests are still counted as attack signals.
func (s *LeakyBucketStore) RecordSMSOTPSent(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds) (LeakyBucketTriggered, error) {
	now := float64(s.Clock.NowUTC().Unix())
	var triggered LeakyBucketTriggered

	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		evalTriggered := func(key string, threshold float64, period, ttl int) (bool, error) {
			res, err := conn.Eval(ctx, leakyBucketScript,
				[]string{key},
				now, threshold, period, 1, ttl,
			).Slice()
			if err != nil {
				return false, err
			}
			if len(res) < 2 {
				return false, nil
			}
			triggeredInt, _ := res[1].(int64)
			return triggeredInt == 1, nil
		}

		var err error

		triggered.CountryHourly, err = evalTriggered(
			s.bucketKey("3600", "country", phoneCountry),
			thresholds.CountryHourly, 3600, 2*3600,
		)
		if err != nil {
			return err
		}

		triggered.CountryDaily, err = evalTriggered(
			s.bucketKey("86400", "country", phoneCountry),
			thresholds.CountryDaily, 86400, 2*86400,
		)
		if err != nil {
			return err
		}

		triggered.IPHourly, err = evalTriggered(
			s.bucketKey("3600", "ip", ip),
			thresholds.IPHourly, 3600, 2*3600,
		)
		if err != nil {
			return err
		}

		triggered.IPDaily, err = evalTriggered(
			s.bucketKey("86400", "ip", ip),
			thresholds.IPDaily, 86400, 2*86400,
		)
		if err != nil {
			return err
		}

		// Update ip_countries ZSET.
		ipCountriesKey := s.ipCountriesKey(ip)
		res, err := conn.Eval(ctx, ipCountriesScript,
			[]string{ipCountriesKey},
			phoneCountry, now, 3, 2*86400,
		).Slice()
		if err != nil {
			return err
		}
		if len(res) >= 2 {
			triggeredInt, _ := res[1].(int64)
			triggered.IPCountriesDaily = triggeredInt == 1
		}

		return nil
	})
	if err != nil {
		return LeakyBucketTriggered{}, err
	}
	return triggered, nil
}

// RecordSMSOTPVerified drains all 4 leaky buckets by count units (fire-and-forget).
// count is the number of unverified sends to cancel out.
func (s *LeakyBucketStore) RecordSMSOTPVerified(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds, count int) error {
	now := float64(s.Clock.NowUTC().Unix())
	n := -count // negative n drains the bucket

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		drain := func(key string, threshold float64, period int) error {
			return conn.Eval(ctx, leakyBucketScript,
				[]string{key},
				now, threshold, period, n, 0,
			).Err()
		}

		if err := drain(s.bucketKey("3600", "country", phoneCountry), thresholds.CountryHourly, 3600); err != nil {
			return err
		}
		if err := drain(s.bucketKey("86400", "country", phoneCountry), thresholds.CountryDaily, 86400); err != nil {
			return err
		}
		if err := drain(s.bucketKey("3600", "ip", ip), thresholds.IPHourly, 3600); err != nil {
			return err
		}
		if err := drain(s.bucketKey("86400", "ip", ip), thresholds.IPDaily, 86400); err != nil {
			return err
		}

		return nil
	})
}

func (s *LeakyBucketStore) bucketKey(period, dimension, value string) string {
	return fmt.Sprintf("%s:fraud_protection:leaky_bucket:%s:%s:%s", string(s.AppID), period, dimension, value)
}

func (s *LeakyBucketStore) ipCountriesKey(ip string) string {
	return fmt.Sprintf("%s:fraud_protection:ip_countries:%s", string(s.AppID), ip)
}
