-- k1 keyname
-- a1 unix
local rate = 10 -- 令牌生成速率，每秒10个

if redis.call('exists', KEYS[1]) == 0 then
    redis.call('hmset', KEYS[1], 'token', 9, 'act', ARGV[1])
    redis.call('expire', KEYS[1], 1800) -- 30min
    return 1 -- 成功获取令牌
else
    local r = redis.call('hmget', KEYS[1], 'token', 'act')
    local t = tonumber(r[1])
    local act = tonumber(r[2])
    local now = tonumber(ARGV[1])

    local pass = now - act
    local newtoken = pass * rate
    if newtoken > 0 then
        t = math.min(10, t + newtoken)
        redis.call('hset', KEYS[1], 'act', now)
    end
    
    redis.call('expire', KEYS[1], 1800)
    if t >= 1 then
        t = t - 1
        redis.call('hset', KEYS[1], 'token', t)
        return 1 -- 成功获取令牌
    else
        return 0 -- 获取令牌失败
    end
end
