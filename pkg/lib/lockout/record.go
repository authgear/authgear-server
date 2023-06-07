package lockout

import (
	"context"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type record struct {
	Attempts         int  `json:"attempts"`
	LockedUntilEpoch *int `json:"locked_until_epoch,omitempty"`
}

type attemptResult struct {
	IsSuccess   bool
	LockedUntil *time.Time
}

var makeAttemptLuaScript = goredis.NewScript(`
redis.replicate_commands()

local record_key = KEYS[1]
local history_duration = tonumber(ARGV[1])
local max_attempts = tonumber(ARGV[2])
local min_duration = tonumber(ARGV[3])
local max_duration = tonumber(ARGV[4])
local backoff_factor = tonumber(ARGV[5])
local new_attempts = tonumber(ARGV[6])



local now = redis.call("TIME")
local now_timestamp = tonumber(now[1])

local record = { attempts=0 }

local function read_existing_record ()
	local existing_record = redis.pcall("GET", record_key)
	if existing_record and (not existing_record["err"]) then
		record = cjson.decode(existing_record)
	end
end

local read_success = pcall(read_existing_record)

local is_blocked = not (record.locked_until_epoch == nil) and (record.locked_until_epoch > now_timestamp)
local is_success = (not is_blocked) and 1 or 0

if new_attempts < 1 or is_blocked then
	return {is_success, record.locked_until_epoch}
end

record.attempts = record.attempts + new_attempts

if record.attempts >= max_attempts then
	local exponent = record.attempts - max_attempts
	local lock_duration = min_duration * math.pow(backoff_factor, exponent)
	lock_duration = math.min(lock_duration, max_duration)
	record.locked_until_epoch = now_timestamp + lock_duration
end

local expire_at = now_timestamp + history_duration

redis.call("SET", record_key, cjson.encode(record), "EXAT", expire_at)

return {is_success, record.locked_until_epoch}
`)

func makeAttempt(
	ctx context.Context, conn *goredis.Conn,
	key string,
	historyDuration time.Duration,
	maxAttempts int,
	minDuration time.Duration,
	maxDuration time.Duration,
	backoffFactor float64,
	attempts int) (r *attemptResult, err error) {
	result, err := makeAttemptLuaScript.Run(ctx, conn,
		[]string{key},
		int(historyDuration.Seconds()),
		maxAttempts,
		int(minDuration.Seconds()),
		int(maxDuration.Seconds()),
		backoffFactor,
		attempts,
	).Slice()
	if err != nil {
		return nil, err
	}

	var isSuccess = result[0].(int64) == 1
	var lockedUntil *time.Time = nil

	if len(result) > 1 {
		lockedUntilT := time.Unix(result[1].(int64), 0)
		lockedUntil = &lockedUntilT
	}

	return &attemptResult{
		IsSuccess:   isSuccess,
		LockedUntil: lockedUntil,
	}, nil
}
