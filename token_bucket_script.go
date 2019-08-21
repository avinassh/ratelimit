package ratelimit

var tokenBucketScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local default_expiry = math.floor(window/1000 * 3)

-- https://stackoverflow.com/a/1252776
local function is_empty(table)
    local next = next
    if next(table) == nil then
        return true
    end
    return false
end

-- https://stackoverflow.com/a/34313599
local function hgetall(hash_key)
    local flat_map = redis.call("HGETALL", hash_key)
    local result = {}
    for i = 1, #flat_map, 2 do
        result[flat_map[i]] = flat_map[i + 1]
    end
    return result
end

-- we would have stored as {ts: <timestamp>, c: <counter>}
local value = hgetall(key)

local function set(ts, counter)
    redis.call("HMSET", key, "ts", now, "c", counter)
    redis.call("EXPIRE", key, default_expiry)
    return {"ts", ts, "c", counter, "s", 1}
end

local function existing_counter(last_refill, counter)
    redis.debug("checking if", counter, rate)
    if counter < rate then
        -- return redis.error_reply("counter rate")
        return set(last_refill, counter + 1)
    end
    -- current limit has exceeded, lets check if it can be refiled
    redis.debug("checking limit", last_refill+1000, now)
    if last_refill + 1000 <= now then
        -- return redis.error_reply("text")
        return set(now, 1)
    end
    -- current limit has exceeded, but not refill either. just return the values
    return {"ts", last_refill, "c", counter, "s", 0}
end

local function run()
    if is_empty(value) then
        return set(now, 1)
    else
        local last_refill = tonumber(value.ts) or 0
        local counter = tonumber(value.c) or 0
        return existing_counter(last_refill, counter)
    end
end

return run()
`

var slidingWindowScript = `
local key = KEYS[1]
local rate = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local old_window = now - window
local default_expiry = window * 5

local function set(ts, counter)
    redis.call("ZADD", key, ts, ts)
    redis.call("EXPIRE", key, default_expiry)
    return {"ts", ts, "c", counter, "s", 1}
end

local function run()
    -- remove all the old window scores
    redis.call("ZREMRANGEBYSCORE", key, "-inf", old_window)
    local counter = redis.call("ZCARD", key)
    if counter < rate then
        return set(now, counter+1)
    end
    -- the limit has reached, so we just return the counter values
    -- the oldest record in the set gives us the last refill time
    local last_refill = tonumber(redis.call("ZRANGE", key, 0, 0)[1])
    return {"ts", last_refill, "c", counter, "s", 0}
end

return run()
`
