package middleware

import (
	"cloud_store/global"
	"context"
	"fmt"
	"strconv"
	"time"
)

var limitLua string = `

local rate = 10 

if redis.call('exists', KEYS[1]) == 0 then
    redis.call('hmset', KEYS[1], 'token', 9, 'act', ARGV[1])
    redis.call('expire', KEYS[1], 1800) -- 30min
    return 1
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
        return 1
    else
        return 0
    end
end

`

func RedisLimitBucket(id int) (bool, error) {
	idStr := strconv.Itoa(id)
	key := global.LimitKeyPrefix + idStr
	ret, err := global.RDB.Eval(context.TODO(), limitLua, []string{key}, time.Now().Unix()).Result()
	if err != nil {
		global.Logger.Error("[Redis eval err]" + err.Error())
		return false, err
	}
	rCode, ok := ret.(int64)
	if !ok {
		global.Logger.Error("rCode, ok := ret.(int64)")
		return false, fmt.Errorf("rCode, ok := ret.(int64)")
	}
	if rCode == 1 {
		return true, nil
	}
	return false, nil
}
