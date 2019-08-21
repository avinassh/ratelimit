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
