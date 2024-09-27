package ratelimit

import (
	"context"
	"time"

	goredis "github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
)

// ref: https://en.wikipedia.org/wiki/Generic_cell_rate_algorithm

var gcraLuaScript = goredis.NewScript(`
redis.replicate_commands()

local rate_limit_key = KEYS[1]
local period = tonumber(ARGV[1])
local burst = tonumber(ARGV[2])
local n = tonumber(ARGV[3])

local now = redis.call("TIME")
local now_timestamp = now[1] * 1000 + math.floor(now[2] / 1000)

local emission_interval = math.floor(period / burst)
local tolerance = burst

local tat = redis.pcall("GET", rate_limit_key)
if not tat then          -- key not found
	tat = now_timestamp
elseif tat["err"] then   -- old rate limit keys
	tat = now_timestamp
else
	tat = tonumber(tat)
end

local increment = emission_interval * n
local new_tat = math.max(tat, now_timestamp) + increment
local dvt = emission_interval * tolerance

local allow_at = new_tat - dvt
local is_conforming = now_timestamp >= allow_at
local time_to_act = allow_at
if is_conforming then
	redis.call("SET", rate_limit_key, new_tat)
	redis.call("EXPIREAT", rate_limit_key, new_tat)
	time_to_act = allow_at + math.max(1, n) * emission_interval
end

return {is_conforming and 1 or 0, time_to_act}
`)

type gcraResult struct {
	IsConforming bool
	TimeToAct    time.Time
}

func gcra(ctx context.Context, conn redis.Redis_6_0_Cmdable, key string, period time.Duration, burst int, n int) (*gcraResult, error) {
	result, err := gcraLuaScript.Run(ctx, conn,
		[]string{key},
		period.Milliseconds(), burst, n,
	).Slice()
	if err != nil {
		return nil, err
	}

	return &gcraResult{
		IsConforming: result[0].(int64) == 1,
		TimeToAct:    time.UnixMilli(result[1].(int64)),
	}, nil
}
