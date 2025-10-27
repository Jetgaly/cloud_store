package user

import (
	"cloud_store/global"
	"cloud_store/model"
	"cloud_store/utils"
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

type createUserReq struct {
	Code     string `json:"code" binding:"required,len=5"`
	Nickname string `json:"nickname" binding:"required,min=1,max=16"`
	Password string `json:"password" binding:"required,min=8,max=16"`
	Email    string `json:"email" binding:"required,email"`
}

func (*UserApi) CreateUserLogic(ctx *gin.Context) {
	var Req createUserReq
	if err := ctx.ShouldBindJSON(&Req); err != nil {
		utils.ResponseWithMsg("[input data err]: "+err.Error(), ctx)
		return
	}
	//todo: check code
	r, err := global.RDB.Get(context.TODO(), global.EmailCodePrefix+Req.Email).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			utils.ResponseWithMsg("please click to get the code", ctx)
			return
		}
		global.Logger.Error("redis get key err: " + err.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	if r != Req.Code {
		utils.ResponseWithMsg("please input the right code", ctx)
		return
	}
	pwdHash, e1 := utils.HashPassword(Req.Password)
	if e1 != nil {
		global.Logger.Error("utils.HashPassword err: " + e1.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}

	UserModel := model.User{
		Nickname: Req.Nickname,
		Password: pwdHash,
		Email:    Req.Email,
	}

	//会回填id
	if err := global.DB.Create(&UserModel).Error; err != nil {
		global.Logger.Error("create user err: " + err.Error())
		if utils.IsDuplicateError(err) {
			utils.ResponseWithCode("1000", ctx)
			return
		}
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	utils.ResponseWithData(nil, ctx)
}
