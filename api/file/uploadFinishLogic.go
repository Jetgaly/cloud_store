package file

import (
	"cloud_store/global"
	"cloud_store/utils"
	"context"

	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
		fmt.Println("hash:", rHash)
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

	case 3:
		//missing
		utils.ResponseWithCodeAndData("1009", rStr, ctx)
		return
	}

	utils.ResponseWithData(nil, ctx)

}
