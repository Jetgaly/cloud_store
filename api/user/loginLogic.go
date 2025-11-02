package user

import (
	"cloud_store/global"
	"cloud_store/model"
	"cloud_store/utils"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LoginReq struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"pwd" binding:"required,min=8,max=16"`
}

func (*UserApi) LoginLogic(ctx *gin.Context) {
	var Req LoginReq
	if err := ctx.ShouldBindJSON(&Req); err != nil {
		utils.ResponseWithMsg("[input data err]: "+err.Error(), ctx)
		return
	}
	var UserModel model.User
	if err := global.DB.Take(&UserModel, "email=?", Req.Email).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ResponseWithCode("1001", ctx)
			return
		} else {
			global.Logger.Error("gorm take err:" + err.Error())
			utils.ResponseWithMsg("internal server err", ctx)
			return
		}
	}
	if !utils.CheckPasswordHash(Req.Password, UserModel.Password) {
		utils.ResponseWithCode("1002", ctx)
		return
	}
	token, err := utils.GenerateToken(int(UserModel.ID), UserModel.Nickname)
	if err != nil {
		global.Logger.Error("token generate err:" + err.Error())
		utils.ResponseWithMsg("internal server err", ctx)
		return
	}
	utils.ResponseWithData(token, ctx)
}
