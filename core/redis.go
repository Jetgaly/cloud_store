package core

import (
	"cloud_store/global"
	"context"

	"github.com/go-redis/redis/v8"
)

// 连接redis
func InitRedis() {

	ctx := context.Background()
	rdb := redis.NewClient(&redis.Options{
		Addr:     global.Config.Redis.Addr(),   
		Password: global.Config.Redis.PassWord, 
		DB:       0,                            
		PoolSize: global.Config.Redis.PoolSize,
	})

	// 使用 Ping() 方法测试是否成功连接到 Redis 服务器
	// _, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	// defer cancel()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		global.Logger.Fatal("failed to connect to Redis: " + err.Error())
		return
	}
	global.RDB = rdb
	global.Logger.Info("connected to Redis:" + pong)
}
