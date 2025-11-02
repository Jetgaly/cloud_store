if redis.call('exists', KEYS[1]) == 0 then
    -- 不存在
    return 0
else
    local r = redis.call('hmget', KEYS[1], 'status', 'act', 'ccnt', 'fn', 'hash')
    local status = tonumber(r[1])
    local l_act = tonumber(r[2])
    local n_act = tonumber(ARGV[1])
    local ccnt = tonumber(r[3])
    local r1 = redis.call('scard', KEYS[2])
    local tc = tonumber(r1)

    if l_act <= n_act then
        redis.call('hset', KEYS[1], 'act', ARGV[1])
    end

    -- 检查状态
    if status ~= 1 then
        return 1
    end

    if tc == ccnt then
        redis.call('hset', KEYS[1], 'status', '2')
        return {2, r[4], r[3], r[5]}
    end

    -- 根据ccnt查找不在set KEYS[2]里面的index
    -- 查找缺失的索引
    local missing = {}
    for i = 0, ccnt - 1 do
        if redis.call('sismember', KEYS[2], tostring(i)) == 0 then
            table.insert(missing, i)
        end
    end
    return {3, table.concat(missing, ",")} -- 返回缺失的索引列表string

end
