package file

import (
	"cloud_store/global"
	"cloud_store/utils"
	"context"

	"time"

	"io"
	"os"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
)

var uploadLua string = `
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
`

func (*FileApi) UploadLogic(ctx *gin.Context) {
	form, err := ctx.MultipartForm()
	if err != nil {
		utils.ResponseWithMsg(err.Error(), ctx)
		return
	}
	dataList, ok := form.File["data"]
	if !ok || len(dataList) != 1 {
		utils.ResponseWithMsg("filedata错误", ctx)
		return
	}

	indexList, ok1 := form.Value["index"]
	if !ok1 || len(indexList) != 1 {
		utils.ResponseWithMsg("index错误", ctx)
		return
	}
	upIdList, ok2 := form.Value["upid"]
	if !ok2 || len(upIdList) != 1 {
		utils.ResponseWithMsg("upid错误", ctx)
		return
	}
	if dataList[0].Size > global.Config.Upload.ChunkSize {
		utils.ResponseWithMsg("filedata size over", ctx)
		return
	}

	_claims, _ := ctx.Get("claims")
	claims := _claims.(*utils.CustomClaims)
	userIdStr := strconv.Itoa(claims.UserId)
	//验证meta是否在redis
	key := global.FileMetaPrefix + userIdStr + ":" + upIdList[0]
	sKey := global.FileSetPrefix + upIdList[0]
	ret, rerr := global.RDB.Eval(context.TODO(), uploadLua, []string{key, sKey}, time.Now().Unix(), indexList[0]).Result()
	if rerr != nil {
		global.Logger.Error("redis Eval err: " + rerr.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	rCode, ok := ret.(int64)
	if !ok {
		global.Logger.Error("rCode, ok := ret.(int64) err")
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	switch rCode {
	case 0:
		//meta 不存在
		utils.ResponseWithMsg("please init upload", ctx)
		return
	case 1:
		utils.ResponseWithCode("1007", ctx)
		return
	case 2:
		//检查分片是否已经上传过
		utils.ResponseWithCode("1006", ctx)
		return
	}

	//rCode == 3
	//上传
	tmpDirPath := path.Join(global.Config.Upload.TempPath, upIdList[0])
	if e2 := utils.CreateDir(tmpDirPath); e2 != nil {
		global.Logger.Error("CreateDir err: " + e2.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	tmpFilePath := path.Join(global.Config.Upload.TempPath, upIdList[0], indexList[0])
	file, e3 := os.OpenFile(tmpFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if e3 != nil {
		global.Logger.Error("OpenFile err: " + e3.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	defer file.Close()
	src, e4 := dataList[0].Open()
	if e4 != nil {
		global.Logger.Error("Open uploaded file err: " + e4.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	defer src.Close()

	//copy自动处理缓冲区
	_, err = io.Copy(file, src)
	if err != nil {
		global.Logger.Error("Copy file err: " + err.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	//将分片信息加入redis set

	err = global.RDB.SAdd(context.TODO(), sKey, indexList[0]).Err()
	if err != nil {
		global.Logger.Error("SAdd err: " + err.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	hKey := global.FileMetaPrefix + upIdList[0]
	_, err = global.RDB.HSet(context.TODO(), hKey, "act", time.Now().Unix()).Result()
	if err != nil {
		global.Logger.Error("HSet err: " + err.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}

	utils.ResponseWithData(nil, ctx)
}
