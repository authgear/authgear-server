package ratelimit

import (
	"context"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

// ref: https://en.wikipedia.org/wiki/Generic_cell_rate_algorithm

var gcraLuaScript = goredis.NewScript(`
redis.replicate_commands()

local rate_limit_key = KEYS[1]
local period = ARGV[1]
local burst = ARGV[2]
local n = ARGV[3]

local now = redis.call("TIME")
local now_timestamp = now[1] * 1000 + math.floor(now[2] / 1000)

local emission_interval = math.floor(period / burst)
local tolerance = emission_interval * (burst - 1)

local tat = redis.pcall("GET", rate_limit_key)
if not tat then          -- key not found
	tat = now_timestamp
elseif tat["err"] then   -- old rate limit keys
	tat = now_timestamp
else
	tat = tonumber(tat)
end

local elapsed = emission_interval * (n - 1)
local new_tat = math.max(tat, now_timestamp) + elapsed

local time_to_act = new_tat - tolerance
local is_conforming = now_timestamp >= time_to_act
if is_conforming then
	new_tat = new_tat + emission_interval
	redis.call("SET", rate_limit_key, new_tat, "PXAT", new_tat)
	tat = new_tat
end

time_to_act = tat - tolerance
return {is_conforming and 1 or 0, time_to_act}
`)

type gcraResult struct {
	IsConforming bool
	TimeToAct    time.Time
}

func gcra(ctx context.Context, conn *goredis.Conn, key string, period time.Duration, burst int, n int) (*gcraResult, error) {
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
