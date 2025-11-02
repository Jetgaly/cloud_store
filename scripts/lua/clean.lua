-- k1 hKey
-- k2 skey
-- a1 unix
if redis.call('exists', KEYS[1]) == 0 then
    -- 不存在
    return 0
else
    local r = redis.call('hmget', KEYS[1], 'act', 'status')
    local l_act = tonumber(r[1])
    local n_act = tonumber(ARGV[1])
    local s = tonumber(r[2])

    if s == 2 then -- finish
        return 1
    end
    redis.call('hset', KEYS[1], 'status', '3') -- cancel
    if (n_act - l_act > 300) and s ~= 3 then -- 旧的status是init/upload 300s
        redis.call('hset', KEYS[1], 'act', ARGV[1])
        return 1
    end
    if (n_act - l_act > 600) then -- 旧的status是cancel
        redis.call('del', KEYS[1])
        redis.call('del', KEYS[2])
        return 2
    end
    return 1 -- 延迟清理时间没到
end
