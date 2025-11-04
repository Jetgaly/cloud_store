package file

import (
	"cloud_store/global"
	"cloud_store/model"
	"cloud_store/utils"
	RMQUtils "cloud_store/utils/RabbitMQ"
	"context"
	"encoding/json"
	"errors"

	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redsync/redsync/v4"
	"github.com/rabbitmq/amqp091-go"

	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type UploadFinishReq struct {
	UploadId string `json:"upid" binding:"required"`
}

var finishLua string = `
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

`
var finishRecoverLua = `
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

`

type OSSMqMsg struct {
	MinIOPath string `json:"path"`
	Id        string `json:"id"`
}

func (*FileApi) UploadFinishLogic(ctx *gin.Context) {
	var Req UploadFinishReq
	if err := ctx.ShouldBindJSON(&Req); err != nil {
		utils.ResponseWithMsg("[input data err]: "+err.Error(), ctx)
		return
	}

	_claims, _ := ctx.Get("claims")
	claims := _claims.(*utils.CustomClaims)
	userIdStr := strconv.Itoa(claims.UserId)
	hKey := global.FileMetaPrefix + userIdStr + ":" + Req.UploadId
	sKey := global.FileSetPrefix + Req.UploadId

	ret, err := global.RDB.Eval(context.TODO(), finishLua, []string{hKey, sKey}, time.Now().Unix()).Result()

	if err != nil {
		global.Logger.Error("redis eval err:" + err.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	var rCode int64
	var rStr, rCcnt, rHash string
	var ok bool
	switch v := ret.(type) {
	case int64:
		rCode = v
	case []interface{}:
		rCode, ok = v[0].(int64)
		if !ok {
			global.Logger.Error("v[0].(int64) err")
			utils.ResponseWithMsg("[internal server err]", ctx)
			return
		}
		if len(v) == 2 {
			rStr, ok = v[1].(string) //missStr
			if !ok {
				global.Logger.Error("v[1].(string) err")
				utils.ResponseWithMsg("[internal server err]", ctx)
				return
			}
		} else if len(v) == 4 {
			rStr, ok = v[1].(string) //fn
			if !ok {
				global.Logger.Error("v[1].(string) err")
				utils.ResponseWithMsg("[internal server err]", ctx)
				return
			}
			rCcnt, ok = v[2].(string) //ccnt
			if !ok {
				global.Logger.Error("v[2].(string) err")
				utils.ResponseWithMsg("[internal server err]", ctx)
				return
			}

			rHash, ok = v[3].(string) //hash
			if !ok {
				global.Logger.Error("v[3].(string) err")
				utils.ResponseWithMsg("[internal server err]", ctx)
				return
			}
		}
	}

	var finalFilePath string

	switch rCode {
	case 0:
		utils.ResponseWithMsg("没有上传信息", ctx)
		return
	case 1:
		utils.ResponseWithCode("1008", ctx)
		return
	case 2:
		//分布式锁去重 hash锁
		cancelCtx, cancel := context.WithCancel(context.Background())
		var lock *redsync.Mutex
		var rlerr error
		for {
			lock, rlerr = global.RedLockCreater.GetLock(cancelCtx, global.RedLockPrefix+rHash, redsync.WithTries(3))
			if rlerr == nil {
				//获取成功
				break
			}
		}
		//释放锁
		defer global.RedLockCreater.ReleaseLock(lock, cancel)
		//检查上一个锁得者是否上传完毕
		e0 := global.DB.Transaction(func(tx *gorm.DB) error {
			var fModel model.File
			txerr := tx.Take(&fModel, "hash=?", rHash).Error
			if errors.Is(txerr, gorm.ErrRecordNotFound) {
				return nil
			}
			if txerr != nil {
				return txerr
			}
			var fn string
			var re error
			if fn, re = global.RDB.HGet(context.Background(), hKey, "fn").Result(); re != nil {
				return re
			}
			//找到记录
			relation := model.UserFile{
				UserId:   int64(claims.UserId),
				FileId:   int64(fModel.ID),
				FileName: fn,
			}
			if e := tx.Create(&relation).Error; e != nil {
				return e
			}
			//修改log
			if e := tx.Model(&model.VolumeOpLog{}).Where("upload_id = ?", Req.UploadId).Update("status", 1).Error; e != nil {
				return e
			}
			return nil
		})

		if e0 != nil {
			if _, re1 := global.RDB.Eval(context.TODO(), finishRecoverLua, []string{hKey}, time.Now().Unix()).Result(); re1 != nil {
				global.Logger.Error("finishRecoverLua err:" + re1.Error())
			}
			global.Logger.Error("global.DB.Transaction err:" + e0.Error())
			utils.ResponseWithMsg("[internal server err]", ctx)
			return
		}

		var fileName string
		fileName = rStr
		list := strings.Split(fileName, ".")
		fileSuffix := list[len(list)-1]
		finalFilePath = path.Join(global.Config.Upload.TempPath, Req.UploadId, Req.UploadId+"."+fileSuffix)
		fmt.Println("finalFilePath", finalFilePath)

		ccnt, e1 := strconv.Atoi(rCcnt)
		if e1 != nil {
			if _, re1 := global.RDB.Eval(context.TODO(), finishRecoverLua, []string{hKey}, time.Now().Unix()).Result(); re1 != nil {
				global.Logger.Error("finishRecoverLua err:" + re1.Error())
			}
			global.Logger.Error("trconv.Atoi(rCcnt) err:" + e1.Error())
			utils.ResponseWithMsg("[internal server err]", ctx)
			return
		}
		var finalFile *os.File
		if finalFile, err = os.OpenFile(finalFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644); err != nil {
			if _, re1 := global.RDB.Eval(context.TODO(), finishRecoverLua, []string{hKey}, time.Now().Unix()).Result(); re1 != nil {
				global.Logger.Error("finishRecoverLua err:" + re1.Error())
			}
			global.Logger.Error("OpenFile err: " + err.Error())
			utils.ResponseWithMsg("[internal server err]", ctx)
			return
		}
		defer finalFile.Close()

		for i := 0; i < ccnt; i++ {
			chunkPath := path.Join(global.Config.Upload.TempPath, Req.UploadId, strconv.Itoa(i))
			// 打开分片文件
			chunkFile, e1 := os.Open(chunkPath)
			if e1 != nil {
				if _, re1 := global.RDB.Eval(context.TODO(), finishRecoverLua, []string{hKey}, time.Now().Unix()).Result(); re1 != nil {
					global.Logger.Error("finishRecoverLua err:" + re1.Error())
				}
				global.Logger.Error("os.Open(chunkPath) err: " + e1.Error())
				utils.ResponseWithMsg("[internal server err]", ctx)
				return
			}
			// 复制分片内容到最终文件
			_, e1 = io.Copy(finalFile, chunkFile)
			chunkFile.Close()
			if e1 != nil {
				if _, re1 := global.RDB.Eval(context.TODO(), finishRecoverLua, []string{hKey}, time.Now().Unix()).Result(); re1 != nil {
					global.Logger.Error("finishRecoverLua err:" + re1.Error())
				}
				global.Logger.Error("io.Copy(finalFile, chunkFile) err: " + e1.Error())
				utils.ResponseWithMsg("[internal server err]", ctx)
				return
			}
		}
		//文件hash校验
		global.Logger.Info("hash:" + rHash)
		hash, e2 := utils.CalculateSHA256Stream(finalFilePath)
		if e2 != nil {
			if _, re1 := global.RDB.Eval(context.TODO(), finishRecoverLua, []string{hKey}, time.Now().Unix()).Result(); re1 != nil {
				global.Logger.Error("finishRecoverLua err:" + re1.Error())
			}
			global.Logger.Error("utils.CalculateSHA256Stream(finalFilePath) err: " + e2.Error())
			utils.ResponseWithMsg("[internal server err]", ctx)
			return
		}
		if hash != rHash {
			utils.ResponseWithCode("1010", ctx)
			return
		}
		//todo: minio ,mysql
		//minio
		objectName := Req.UploadId + "." + fileSuffix
		contentType := utils.GetContentType(fileSuffix)
		bucketName := global.Config.MinIO.UploadBucket
		// Upload
		info, e3 := global.MinioCli.FPutObject(ctx, bucketName, objectName, finalFilePath, minio.PutObjectOptions{ContentType: contentType})
		if e3 != nil {
			if _, re1 := global.RDB.Eval(context.TODO(), finishRecoverLua, []string{hKey}, time.Now().Unix()).Result(); re1 != nil {
				global.Logger.Error("finishRecoverLua err:" + re1.Error())
			}
			global.Logger.Error("minio upload err: " + e3.Error())
			utils.ResponseWithMsg("[internal server err]", ctx)
			return
		}
		//fmt.Println("infosize",info.Size)
		//fmt.Println("infosha",info.ChecksumSHA256)//显式声明才能返回hash
		//mysql
		fileModel := model.File{
			Name: objectName,
			Hash: hash,
			Path: objectName,
			Size: uint64(info.Size),
		}
		e4 := global.DB.Transaction(func(tx *gorm.DB) error {
			if e := tx.Create(&fileModel).Error; e != nil {
				return e
			}
			relation := model.UserFile{
				UserId:   int64(claims.UserId),
				FileId:   int64(fileModel.ID),
				FileName: fileName,
			}
			if e := tx.Create(&relation).Error; e != nil {
				return e
			}
			//修改log
			if e := tx.Model(&model.VolumeOpLog{}).Where("upload_id = ?", Req.UploadId).Update("status", 1).Error; e != nil {
				return e
			}
			return nil
		})
		if e4 != nil {
			global.Logger.Error(fmt.Sprintf("mysql err: %s,upid: %s"+e4.Error(), Req.UploadId))
			utils.ResponseWithMsg("[internal server err]", ctx)
			return
		}

		//清理垃圾
		//设置状态为cancel，留给定时任务清理
		_, e5 := global.RDB.HSet(context.TODO(), hKey, []string{
			"status",
			"3", //cancel
		}).Result()
		if e5 != nil {
			global.Logger.Error(fmt.Sprintf("redis set cancel err: %s,upid: %s", e5.Error(), Req.UploadId))
			utils.ResponseWithMsg("[internal server err]", ctx)
			return
		}

		//cs.oss.queue
		var cwc *RMQUtils.ChannelWithConfirm
		for {
			cwc, err = global.RMQ.Get()
			if errors.Is(err, RMQUtils.ErrTimeout) {
				global.Logger.Error("[RMQ] get channel timeout")
				continue
			} else if err != nil {
				global.Logger.Error(fmt.Sprintf("[RMQ] get channel err:%s,upid:%s", err.Error(), Req.UploadId))
				utils.ResponseWithMsg("[internal server err]", ctx)
				return
			}
			break
		}
		confirm := *(cwc.Confirm)
		var body []byte
		fileIdStr := strconv.Itoa(int(fileModel.ID))
		mqMsg := OSSMqMsg{
			MinIOPath: objectName,
			Id:        fileIdStr,
		}
		body, err = json.Marshal(mqMsg)
		if err != nil {
			global.Logger.Error(fmt.Sprintf("[json]oss marshal err:%s,upid:%s", err.Error(), Req.UploadId))
			utils.ResponseWithCode("1014", ctx)
			return
		}
		err = cwc.Channel.PublishWithContext(
			context.TODO(),
			"cs.oss.exc",   // exchange
			"cs.oss.queue", // routing key
			false,          // mandatory
			false,          // immediate
			amqp091.Publishing{
				ContentType:  "application/json",
				Body:         body,
				DeliveryMode: amqp091.Persistent, // 消息持久化
				Timestamp:    time.Now(),
			})
		if err != nil {
			global.Logger.Error(fmt.Sprintf("[RMQ]oss send err:%s,upid:%s", err.Error(), Req.UploadId))
			utils.ResponseWithCode("1014", ctx)
			return
		}
		select {
		case cf := <-confirm:
			if cf.Ack {
				global.Logger.Info("Message confirmed")
			} else {
				global.Logger.Error(fmt.Sprintf("[RMQ]oss send fail,upid:%s", Req.UploadId))
				cwc.Channel.Close()
				utils.ResponseWithCode("1014", ctx)
				return
			}
		case <-time.After(5 * time.Second): //超时时间
			global.Logger.Error(fmt.Sprintf("[RMQ]oss confirm timeout,upid:%s", Req.UploadId))
			//超时直接关闭，不要放回channel池
			cwc.Channel.Close()
			utils.ResponseWithCode("1014", ctx)
			return
		}
		global.RMQ.Put(cwc)
	case 3:
		//missing
		utils.ResponseWithCodeAndData("1009", rStr, ctx)
		return
	}

	utils.ResponseWithData(nil, ctx)

}
