package lockout

import (
	"context"
	"sort"
	"time"

	goredis "github.com/redis/go-redis/v9"

	apimodel "github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
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
	redis.call("SET", lock_key, locked_until_epoch)
	redis.call("EXPIREAT", lock_key, locked_until_epoch)
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
	ctx context.Context, conn redis.Redis_6_0_Cmdable,
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
		lockedUntilT := time.Unix(result[1].(int64), 0).UTC()
		lockedUntil = &lockedUntilT
	}

	return &attemptResult{
		IsSuccess:   isSuccess,
		LockedUntil: lockedUntil,
	}, nil
}

func clearAttempts(
	ctx context.Context, conn redis.Redis_6_0_Cmdable,
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

var getStatusLuaScript = goredis.NewScript(constants + `
-- KEYS[1]: record_key
-- ARGV[1]: is_global ("1" for per_user, "0" for per_user_per_ip)
--
-- Returns for per_user (is_global=1):
--   {0}                    -- not locked
--   {1, locked_until_epoch} -- locked
--
-- Returns for per_user_per_ip (is_global=0):
--   First element: is_any_locked (0 or 1)
--   Followed by pairs: ip_string, locked_until_epoch
--   Example: {1, "1.2.3.4", 1234567890, "5.6.7.8", 1234567999}
redis.replicate_commands()
local record_key = KEYS[1]
local is_global = ARGV[1] == "1"
local now_raw = redis.call("TIME")
local now = tonumber(now_raw[1])

if is_global then
    local lock_key = record_key .. ":lock:global"
    local v = redis.pcall("GET", lock_key)
    if v and not v["err"] and type(v) == "string" then
        local epoch = tonumber(v)
        if epoch and epoch > now then
            return {1, epoch}
        end
    end
    return {0}
else
    local result = {}
    local is_any_locked = 0
    local hash_data = redis.pcall("HGETALL", record_key)
    if hash_data and not hash_data["err"] then
        for i = 1, #hash_data, 2 do
            local field = hash_data[i]
            if field ~= GLOBAL_TOTAL_KEY then
                local lock_key = record_key .. ":lock:" .. field
                local v = redis.pcall("GET", lock_key)
                if v and not v["err"] and type(v) == "string" then
                    local epoch = tonumber(v)
                    if epoch and epoch > now then
                        is_any_locked = 1
                        table.insert(result, field)
                        table.insert(result, epoch)
                    end
                end
            end
        end
    end
    table.insert(result, 1, is_any_locked)
    return result
end
`)

var clearAllLuaScript = goredis.NewScript(constants + `
-- KEYS[1]: record_key
-- ARGV[1]: is_global ("1" or "0")
-- Returns: 1
redis.replicate_commands()
local record_key = KEYS[1]
local is_global = ARGV[1] == "1"

if is_global then
    redis.call("DEL", record_key, record_key .. ":lock:global")
else
    local hash_data = redis.pcall("HGETALL", record_key)
    if hash_data and not hash_data["err"] then
        local keys_to_del = {record_key}
        for i = 1, #hash_data, 2 do
            local field = hash_data[i]
            if field ~= GLOBAL_TOTAL_KEY then
                table.insert(keys_to_del, record_key .. ":lock:" .. field)
            end
        end
        redis.call("DEL", unpack(keys_to_del))
    else
        redis.call("DEL", record_key)
    end
end
return 1
`)

func getStatus(
	ctx context.Context, conn redis.Redis_6_0_Cmdable,
	key string,
	isGlobal bool,
) (*LockoutStatus, error) {
	isGlobalStr := "0"
	if isGlobal {
		isGlobalStr = "1"
	}
	result, err := getStatusLuaScript.Run(ctx, conn, []string{key}, isGlobalStr).Slice()
	if err != nil {
		return nil, err
	}

	if isGlobal {
		// {0} or {1, epoch}
		isLocked := result[0].(int64) == 1
		status := &LockoutStatus{IsLocked: isLocked}
		if isLocked && len(result) > 1 {
			t := time.Unix(result[1].(int64), 0).UTC()
			status.LockedUntil = &t
		}
		return status, nil
	}

	// {is_any_locked, ip1, epoch1, ip2, epoch2, ...}
	isLocked := result[0].(int64) == 1
	var lockedIPs []apimodel.LockedIP
	for i := 1; i+1 < len(result); i += 2 {
		ip := result[i].(string)
		t := time.Unix(result[i+1].(int64), 0).UTC()
		lockedIPs = append(lockedIPs, apimodel.LockedIP{IPAddress: ip, LockedUntil: t})
	}
	// Sort by LockedUntil in descending order (most recent first)
	sort.Slice(lockedIPs, func(i, j int) bool {
		return lockedIPs[i].LockedUntil.After(lockedIPs[j].LockedUntil)
	})
	return &LockoutStatus{IsLocked: isLocked, LockedIPs: lockedIPs}, nil
}

func clearAll(
	ctx context.Context, conn redis.Redis_6_0_Cmdable,
	key string,
	isGlobal bool,
) error {
	isGlobalStr := "0"
	if isGlobal {
		isGlobalStr = "1"
	}
	_, err := clearAllLuaScript.Run(ctx, conn, []string{key}, isGlobalStr).Bool()
	return err
}
