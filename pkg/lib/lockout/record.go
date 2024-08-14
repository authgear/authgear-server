package lockout

import (
	"context"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type attemptResult struct {
	IsSuccess   bool
	LockedUntil *time.Time
}

var constants = `
local GLOBAL_TOTAL_KEY = "total"
`

var makeAttemptsLuaScript = goredis.NewScript(constants + `
redis.replicate_commands()

local record_key = KEYS[1]
local history_duration = tonumber(ARGV[1])
local max_attempts = tonumber(ARGV[2])
local min_duration = tonumber(ARGV[3])
local max_duration = tonumber(ARGV[4])
local backoff_factor = tonumber(ARGV[5])
local is_global = ARGV[6] == "1"
local contributor = ARGV[7]
local new_attempts = tonumber(ARGV[8])

local lock_key = string.format("%s:lock:%s", record_key, is_global and "global" or contributor)

local now = redis.call("TIME")
local now_timestamp = tonumber(now[1])

local global_total = 0
local contributor_total = 0
local locked_until_epoch = nil


local function read_existing_lock ()
	local existing_lock = redis.pcall("GET", lock_key)
	if existing_lock and (not existing_lock["err"]) then
		locked_until_epoch = tonumber(existing_lock)
	end
end

local function read_existing_global_total ()
	local existing_total = redis.pcall("HGET", record_key, GLOBAL_TOTAL_KEY)
	if existing_total and (not existing_total["err"]) then
		global_total = tonumber(existing_total)
	end
end

local function read_existing_contributor_total ()
	local existing_total = redis.pcall("HGET", record_key, contributor)
	if existing_total and (not existing_total["err"]) then
		contributor_total = tonumber(existing_total)
	end
end

pcall(read_existing_lock)
pcall(read_existing_global_total)
pcall(read_existing_contributor_total)

local is_blocked = not (locked_until_epoch == nil) and (locked_until_epoch > now_timestamp)
local is_success = (not is_blocked) and 1 or 0

if new_attempts < 1 or is_blocked then
	return {is_success, locked_until_epoch}
end

global_total = global_total + new_attempts
contributor_total = contributor_total + new_attempts

local total = is_global and global_total or contributor_total

if total >= max_attempts then
	local exponent = total - max_attempts
	local lock_duration = min_duration * math.pow(backoff_factor, exponent)
	lock_duration = math.min(lock_duration, max_duration)
	locked_until_epoch = now_timestamp + lock_duration
end

local expire_at = now_timestamp + history_duration

redis.call("HSET", record_key, GLOBAL_TOTAL_KEY, global_total)
redis.call("HSET", record_key, contributor, contributor_total)
redis.call("EXPIREAT", record_key, expire_at)

if locked_until_epoch then
	redis.call("SET", lock_key, locked_until_epoch, "EXAT", locked_until_epoch)
end

return {is_success, locked_until_epoch}
`)

var clearAttemptsLuaScript = goredis.NewScript(constants + `
redis.replicate_commands()

local record_key = KEYS[1]
local history_duration = tonumber(ARGV[1])
local contributor = ARGV[2]

local now = redis.call("TIME")
local now_timestamp = tonumber(now[1])

local global_total = 0
local contributor_total = 0

local function read_existing_global_total ()
	local existing_total = redis.pcall("HGET", record_key, GLOBAL_TOTAL_KEY)
	if existing_total and (not existing_total["err"]) then
		global_total = tonumber(existing_total)
	end
end

local function read_existing_contributor_total ()
	local existing_total = redis.pcall("HGET", record_key, contributor)
	if existing_total and (not existing_total["err"]) then
		contributor_total = tonumber(existing_total)
	end
end

pcall(read_existing_global_total)
pcall(read_existing_contributor_total)

global_total = math.max(global_total - contributor_total, 0)

local expire_at = now_timestamp + history_duration

redis.call("HSET", record_key, GLOBAL_TOTAL_KEY, global_total)
redis.call("HDEL", record_key, contributor)

-- Redis 6.x does not support NX.
local original_ttl = redis.call("TTL", record_key)
if original_ttl < 0 then
	redis.call("EXPIREAT", record_key, expire_at)
end

return 1
`)

func makeAttempts(
	ctx context.Context, conn *goredis.Conn,
	key string,
	historyDuration time.Duration,
	maxAttempts int,
	minDuration time.Duration,
	maxDuration time.Duration,
	backoffFactor float64,
	isGlobal bool,
	contributor string,
	attempts int) (r *attemptResult, err error) {
	isGlobalStr := "0"
	if isGlobal {
		isGlobalStr = "1"
	}
	result, err := makeAttemptsLuaScript.Run(ctx, conn,
		[]string{key},
		int(historyDuration.Seconds()),
		maxAttempts,
		int(minDuration.Seconds()),
		int(maxDuration.Seconds()),
		backoffFactor,
		isGlobalStr,
		contributor,
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

func clearAttempts(
	ctx context.Context, conn *goredis.Conn,
	key string,
	historyDuration time.Duration,
	contributor string) error {
	_, err := clearAttemptsLuaScript.Run(ctx, conn,
		[]string{key},
		int(historyDuration.Seconds()),
		contributor,
	).Bool()
	return err
}
