package file

import (
	"cloud_store/global"
	"cloud_store/model"
	"cloud_store/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type FileElem struct {
	Id       string    `json:"id"`
	FileName string    `json:"fn"`
	CreateAt time.Time `json:"ca"`
}
type GetFileListReq struct {
	Key   string `form:"key"`
	Limit int    `form:"limit" binding:"required,min=1"`
	Page  int    `form:"page" binding:"required,min=1"`
}

func (*FileApi) GetFileListLogic(ctx *gin.Context) {
	var Req GetFileListReq
	err := ctx.ShouldBindQuery(&Req)
	if err != nil {
		utils.ResponseWithMsg("[input data err]: "+err.Error(), ctx)
		return
	}
	_claims, _ := ctx.Get("claims")
	claims := _claims.(*utils.CustomClaims)
	var respList []FileElem
	if Req.Key == "" {
		//mysql
		offset := Req.Limit * (Req.Page - 1)
		if offset < 0 {
			offset = 0
		}
		if Req.Limit >= 100 {
			Req.Limit = 100
		}
		var List []model.UserFile
		err = global.DB.Limit(Req.Limit).Offset(offset).Where("user_id=?", claims.UserId).Find(&List).Error
		if err != nil {
			global.Logger.Error("gorm getlist err: " + err.Error())
			utils.ResponseWithMsg("[internal server err]", ctx)
			return
		}
		for _, v := range List {
			respList = append(respList, FileElem{
				Id:       strconv.Itoa(int(v.ID)),
				FileName: v.FileName,
				CreateAt: v.CreatedAt,
			})
		}
	} else {
		//es
		
	}
	utils.ResponseWithData(respList, ctx)
}
