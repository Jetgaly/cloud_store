local r = redis.call('hmget', KEYS[1], 'status', 'act')
local s = tonumber(r[1])
local l_act = tonumber(r[2])
local n_act = tonumber(ARGV[1])

if l_act <= n_act then
    redis.call('hset', KEYS[1], 'act', ARGV[1])
end

if status == 2 then
    redis.call('hset', KEYS[1], 'status', '1')
end

return 0
