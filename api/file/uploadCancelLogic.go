package file

import (
	"cloud_store/global"
	"cloud_store/utils"
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UploadCancelReq struct {
	UploadId string `json:"upid" binding:"required"`
}

var cancelLua string = `
if redis.call('exists', KEYS[1]) == 0 then
    return 0
else
    redis.call('hset', KEYS[1], 'status', '3')
    return 1
end

`

func (*FileApi) UploadCancelLogic(ctx *gin.Context) {
	var Req UploadCancelReq
	if err := ctx.ShouldBindJSON(&Req); err != nil {
		utils.ResponseWithMsg("[input data err]: "+err.Error(), ctx)
		return
	}
	_claims, _ := ctx.Get("claims")
	claims := _claims.(*utils.CustomClaims)
	userIdStr := strconv.Itoa(claims.UserId)
	hKey := global.FileMetaPrefix + userIdStr + ":" + Req.UploadId
	ret, err := global.RDB.Eval(context.TODO(), cancelLua, []string{hKey}).Result()
	if err != nil {
		global.Logger.Error("redis Eval err: " + err.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	rCode, ok := ret.(int64)
	if !ok {
		global.Logger.Error("rCode, ok := ret.(int64) err")
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	if rCode == 0 {
		utils.ResponseWithMsg("upid not exists", ctx)
		return
	}
	utils.ResponseWithData(nil, ctx)
}
