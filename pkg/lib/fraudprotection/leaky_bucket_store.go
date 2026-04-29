package fraudprotection

import (
	"context"
	"fmt"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

const (
	bucketWindowHourly     = 3600
	bucketWindowDaily      = 86400
	ipCountriesThreshold   = 3
	bucketDimensionCountry = "country"
	bucketDimensionIP      = "ip"
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

// LeakyBucketLevels holds the current level of each bucket after the fill operation,
// as well as the distinct-country count for the IP. Used for structured logging.
type LeakyBucketLevels struct {
	CountryHourly    float64
	CountryDaily     float64
	IPHourly         float64
	IPDaily          float64
	IPCountriesCount int64
}

// leakyBucketScript is the unified Lua script for fill/drain/read operations.
// KEYS[1] = bucket key
// ARGV[1] = now (unix seconds, float)
// ARGV[2] = threshold (current adaptive threshold for this bucket)
// ARGV[3] = period (window_seconds: 3600 or 86400)
// ARGV[4] = n (positive=fill, negative=drain, 0=read)
// ARGV[5] = ttl_seconds (2 * window_seconds; only applied when n>0)
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
    if n > 0 then
        redis.call('EXPIRE', KEYS[1], ttl)
    end
end

return {new_level, (new_level > threshold) and 1 or 0}
`

// ipCountriesScript tracks distinct countries seen from a given IP in the past 24h
// using a sorted set keyed by country code with the last-seen timestamp as the score.
// KEYS[1] = sent-countries sorted set key
// KEYS[2] = verified-countries sorted set key
// ARGV[1] = alpha2 country code
// ARGV[2] = now (unix timestamp)
// ARGV[3] = threshold (fixed = 3)
// ARGV[4] = ttl_seconds (2 * 86400)
// Returns {filtered_count, triggered_int}.
var ipCountriesScript = `
local now    = tonumber(ARGV[2])
local cutoff = now - 86400

-- 1. Use ZADD to record a send event in sent-countries sorted set key
redis.call('ZADD', KEYS[1], now, ARGV[1])
-- 2. use ZREMRANGEBYSCORE to drop records older than cutoff in both sets before processing
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', cutoff)
redis.call('ZREMRANGEBYSCORE', KEYS[2], '-inf', cutoff)
-- 3. Update the expiry of both set to ensure they are not cleaned up when we still need them
redis.call('EXPIRE', KEYS[1], ARGV[4])
redis.call('EXPIRE', KEYS[2], ARGV[4])

-- 4. Derive counties without at least one verified otp
local sent_countries = redis.call('ZRANGE', KEYS[1], 0, -1)
local verified_countries = redis.call('ZRANGE', KEYS[2], 0, -1)
local verified_lookup = {}

for _, country in ipairs(verified_countries) do
    verified_lookup[country] = true
end

local count = 0
for _, country in ipairs(sent_countries) do
    if not verified_lookup[country] then
        count = count + 1
    end
end

return {count, (count > tonumber(ARGV[3])) and 1 or 0}
`

type LeakyBucketStore struct {
	Redis *appredis.Handle
	AppID config.AppID
	Clock clock.Clock
}

// RecordSMSOTPSent atomically fills all 4 leaky buckets and updates the ip_countries ZSET.
// Returns LeakyBucketTriggered indicating which warning conditions are now exceeded,
// and LeakyBucketLevels with the current level of each bucket after the fill.
// Called BEFORE the SMS send; blocked requests are still counted as attack signals.
func (s *LeakyBucketStore) RecordSMSOTPSent(ctx context.Context, ip, phoneCountry string, thresholds LeakyBucketThresholds) (LeakyBucketTriggered, LeakyBucketLevels, error) {
	now := float64(s.Clock.NowUTC().Unix())
	var triggered LeakyBucketTriggered
	var levels LeakyBucketLevels

	err := s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		evalBucket := func(key string, threshold float64, period, ttl int) (float64, bool, error) {
			res, err := conn.Eval(ctx, leakyBucketScript,
				[]string{key},
				now, threshold, period, 1, ttl,
			).Slice()
			if err != nil {
				return 0, false, err
			}
			if len(res) < 2 {
				return 0, false, nil
			}
			level, _ := res[0].(int64)
			triggeredInt, _ := res[1].(int64)
			return float64(level), triggeredInt == 1, nil
		}

		var err error

		levels.CountryHourly, triggered.CountryHourly, err = evalBucket(
			s.bucketKey(bucketWindowHourly, bucketDimensionCountry, phoneCountry),
			thresholds.CountryHourly, bucketWindowHourly, 2*bucketWindowHourly,
		)
		if err != nil {
			return err
		}

		levels.CountryDaily, triggered.CountryDaily, err = evalBucket(
			s.bucketKey(bucketWindowDaily, bucketDimensionCountry, phoneCountry),
			thresholds.CountryDaily, bucketWindowDaily, 2*bucketWindowDaily,
		)
		if err != nil {
			return err
		}

		levels.IPHourly, triggered.IPHourly, err = evalBucket(
			s.bucketKey(bucketWindowHourly, bucketDimensionIP, ip),
			thresholds.IPHourly, bucketWindowHourly, 2*bucketWindowHourly,
		)
		if err != nil {
			return err
		}

		levels.IPDaily, triggered.IPDaily, err = evalBucket(
			s.bucketKey(bucketWindowDaily, bucketDimensionIP, ip),
			thresholds.IPDaily, bucketWindowDaily, 2*bucketWindowDaily,
		)
		if err != nil {
			return err
		}

		// Update ip_countries ZSET and exclude countries with a verified OTP in the same window.
		ipCountriesKey := s.ipCountriesKey(ip)
		ipVerifiedCountriesKey := s.ipVerifiedCountriesKey(ip)
		res, err := conn.Eval(ctx, ipCountriesScript,
			[]string{ipCountriesKey, ipVerifiedCountriesKey},
			phoneCountry, now, ipCountriesThreshold, 2*bucketWindowDaily,
		).Slice()
		if err != nil {
			return err
		}
		if len(res) >= 2 {
			count, _ := res[0].(int64)
			triggeredInt, _ := res[1].(int64)
			levels.IPCountriesCount = count
			triggered.IPCountriesDaily = triggeredInt == 1
		}

		return nil
	})
	if err != nil {
		return LeakyBucketTriggered{}, LeakyBucketLevels{}, err
	}
	return triggered, levels, nil
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

		if err := drain(s.bucketKey(bucketWindowHourly, bucketDimensionCountry, phoneCountry), thresholds.CountryHourly, bucketWindowHourly); err != nil {
			return err
		}
		if err := drain(s.bucketKey(bucketWindowDaily, bucketDimensionCountry, phoneCountry), thresholds.CountryDaily, bucketWindowDaily); err != nil {
			return err
		}
		if err := drain(s.bucketKey(bucketWindowHourly, bucketDimensionIP, ip), thresholds.IPHourly, bucketWindowHourly); err != nil {
			return err
		}
		if err := drain(s.bucketKey(bucketWindowDaily, bucketDimensionIP, ip), thresholds.IPDaily, bucketWindowDaily); err != nil {
			return err
		}

		return nil
	})
}

func (s *LeakyBucketStore) RecordSMSOTPVerifiedCountry(ctx context.Context, ip, phoneCountry string) error {
	now := float64(s.Clock.NowUTC().Unix())

	return s.Redis.WithConnContext(ctx, func(ctx context.Context, conn redis.Redis_6_0_Cmdable) error {
		return conn.Eval(ctx, `
-- Update the last verified otp timestamp of the country code in the sorted set
redis.call('ZADD', KEYS[1], ARGV[1], ARGV[2])
redis.call('EXPIRE', KEYS[1], ARGV[3])
return 1
`,
			[]string{s.ipVerifiedCountriesKey(ip)},
			now, phoneCountry, 2*bucketWindowDaily,
		).Err()
	})
}

func (s *LeakyBucketStore) bucketKey(period int, dimension, value string) string {
	return fmt.Sprintf("app:%s:fraud_protection:leaky_bucket:%d:%s:%s", string(s.AppID), period, dimension, value)
}

func (s *LeakyBucketStore) ipCountriesKey(ip string) string {
	return fmt.Sprintf("app:%s:fraud_protection:ip_countries:%s", string(s.AppID), ip)
}

func (s *LeakyBucketStore) ipVerifiedCountriesKey(ip string) string {
	return fmt.Sprintf("app:%s:fraud_protection:ip_verified_countries:%s", string(s.AppID), ip)
}
