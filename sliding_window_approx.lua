local key = KEYS[1]
local rate = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local old_window = now - window
local default_expiry = window * 5

