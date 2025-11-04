package file

import (
	"cloud_store/global"
	"cloud_store/model"
	"cloud_store/utils"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DelReq struct {
	Id string
}

func (*FileApi) DelFileLogic(ctx *gin.Context) {
	var Req DelReq
	Req.Id = ctx.Param("id")
	if Req.Id == "" {
		utils.ResponseWithMsg("[input data err]: id ", ctx)
		return
	}
	var id int
	var err error
	if id, err = strconv.Atoi(Req.Id); err != nil {
		global.Logger.Error(fmt.Sprintf("strconv.Atoi(Req.Id) err: %s", err.Error()))
		utils.ResponseWithMsg("[input data err]: id ", ctx)
		return
	}
	_claims, _ := ctx.Get("claims")
	claims := _claims.(*utils.CustomClaims)
	var relation model.UserFile
	var fileModel model.File
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		e := tx.Take(&relation, id).Error
		if e != nil {
			return e
		}
		if claims.UserId != id {
			return nil
		}
		e = global.DB.Select("size").Find(&fileModel).Error
		if e != nil {
			return e
		}
		e = global.DB.Model(&model.User{}).Where("id=?", claims.UserId).Update("available_volume", gorm.Expr("available_volume + ?", fileModel.Size)).Error
		if e != nil {
			return e
		}
		e = global.DB.Delete(&relation).Error
		if e != nil {
			return e
		}
		return nil
	})
	if err != nil {
		global.Logger.Error(fmt.Sprintf("gorm trans err: %s", err.Error()))
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	utils.ResponseWithData(nil, ctx)
}
