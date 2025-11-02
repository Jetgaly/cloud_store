if redis.call('exists', KEYS[1]) == 0 then
    -- 不存在
    return 0
else
    redis.call('hset', KEYS[1], 'status', '3')
    return 1
end
