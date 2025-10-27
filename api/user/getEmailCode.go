package user

import (
	"cloud_store/global"
	"cloud_store/utils"
	"context"
	"time"

	"github.com/gin-gonic/gin"
)

type getEmailCodeReq struct {
	Email string `json:"email" binding:"required,email"`
}

func (*UserApi) GetEmailCodeLogic(ctx *gin.Context) {
	var Req getEmailCodeReq
	if err := ctx.ShouldBindJSON(&Req); err != nil {
		utils.ResponseWithMsg("[input data err]: "+err.Error(), ctx)
		return
	}
	code := utils.GenerateCode()
	r := global.RDB.SetEX(context.TODO(), global.EmailCodePrefix+Req.Email, code, 5*time.Minute)
	if r.Err() != nil {
		global.Logger.Error("set email code err:" + r.Err().Error())
		utils.ResponseWithMsg("internal server err", ctx)
		return
	}
	err := global.EmailSender.SendEmail(Req.Email, code, "CloudStore Code")
	if err != nil {
		global.Logger.Error("send email code err:" + err.Error())
		utils.ResponseWithMsg("internal server err", ctx)
		return
	}
	utils.ResponseWithData(nil, ctx)
}
