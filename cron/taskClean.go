package cron

import (
	"cloud_store/api/file"
	"cloud_store/global"
	"cloud_store/model"
	RMQUtils "cloud_store/utils/RabbitMQ"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

var CleanLua string = `
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

`

func CleanHandler(msg []byte) error {
	var msgModel file.CleanMsg
	err := json.Unmarshal(msg, &msgModel)
	if err != nil {
		global.Logger.Error(fmt.Sprintf("CleanHandler json.Unmarshal err:%s", err.Error()))
		return fmt.Errorf("CleanHandler json.Unmarshal err:%s", err.Error())
	}
	//redis cs:meta:{userId}:{upId}
	hkey := global.FileMetaPrefix + msgModel.UserId + ":" + msgModel.UploadId
	ret, rerr := global.RDB.Eval(context.TODO(), CleanLua, []string{hkey}).Result()
	if rerr != nil {
		global.Logger.Error(fmt.Sprintf("redis eval err:%s,upid:%s", rerr.Error(), msgModel.UploadId))
		return fmt.Errorf("redis eval err:%s,upid:%s", rerr.Error(), msgModel.UploadId)
	}
	rCode, ok := ret.(int64)
	if !ok {
		global.Logger.Error(fmt.Sprintf("rCode, ok := ret.(int64) err,upid:%s", msgModel.UploadId))
		return fmt.Errorf("rCode, ok := ret.(int64) err,upid:%s", msgModel.UploadId)
	}
	switch rCode {
	case 0:
		//不存在skey：已经被清理
		return nil
	case 1:
		//正在finish//刚刚设置cancel//清理时间没到
		//重新投放
		var cwc *RMQUtils.ChannelWithConfirm

		cwc, err = global.RMQ.Get()
		if err != nil {
			return err
		}

		confirm := *(cwc.Confirm)

		err = cwc.Channel.PublishWithContext(
			context.TODO(),
			"cs.clean.delayexc", // exchange
			"cs.clean.delay",    // routing key
			false,               // mandatory
			false,               // immediate
			amqp091.Publishing{
				ContentType:  "application/json",
				Body:         msg,
				DeliveryMode: amqp091.Persistent, // 消息持久化
				Timestamp:    time.Now(),
			})
		if err != nil {
			global.Logger.Error(fmt.Sprintf("[RMQ] send err:%s,upid:%s", err.Error(), msgModel.UploadId))
			return err
		}
		select {
		case cf := <-confirm:
			if cf.Ack {
				global.Logger.Info("Message confirmed")
			} else {
				global.Logger.Error(fmt.Sprintf("[RMQ] send fail,upid:%s", msgModel.UploadId))
				cwc.Channel.Close()
				return fmt.Errorf("[RMQ] send fail,upid:%s", msgModel.UploadId)
			}
		case <-time.After(5 * time.Second): //超时时间
			global.Logger.Error(fmt.Sprintf("[RMQ] confirm timeout,upid:%s", msgModel.UploadId))
			//超时直接关闭，不要放回channel池
			cwc.Channel.Close()
			return fmt.Errorf("[RMQ] confirm timeout,upid:%s", msgModel.UploadId)
		}
		global.RMQ.Put(cwc)
		return nil
	case 2:
		//clean
		//mysql
		if err = global.DB.Model(&model.VolumeOpLog{}).Where("upload_id=?", msgModel.UploadId).Update("status", 2).Error; err != nil {
			global.Logger.Error(fmt.Sprintf("[MYSQL]update oplog cancel err:%s,upid:%s", err.Error(), msgModel.UploadId))
			return err
		}
		//temp
		tmpPath := path.Join(global.Config.Upload.TempPath, msgModel.UploadId)
		err = os.RemoveAll(tmpPath)
		if err != nil {
			global.Logger.Error(fmt.Sprintf("os.RemoveAll(tmpPath) err:%s,upid:%s", err.Error(), msgModel.UploadId))
			return err
		}
		return nil
	}
	return nil
}
