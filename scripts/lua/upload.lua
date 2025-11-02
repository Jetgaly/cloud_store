--k1: mkey
--k2: skey
--a1: time
--a2: index

if redis.call('exists', KEYS[1]) == 0 then
    -- 不存在
    return 0
else
    local r = redis.call('hmget', KEYS[1], 'status', 'act')
    local status = tonumber(r[1])
    local l_act = tonumber(r[2])
    local n_act = tonumber(ARGV[1])

    local f = redis.call('sismember', KEYS[2], ARGV[2])
    local fno = tonumber(f)

    if l_act <= n_act then
        redis.call('hset', KEYS[1], 'act', ARGV[1])
    end
    if status ~= 1 and status ~= 0 then
        return 1
    end
    redis.call('hset', KEYS[1], 'status', '1')
    if fno == 1 then
        return 2
    end
    return 3
end
