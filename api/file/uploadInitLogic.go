package file

import (
	"cloud_store/global"
	"cloud_store/model"
	"cloud_store/utils"
	"context"
	"errors"

	"time"

	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type FileMetaInfo struct {
	Hash       string `json:"hash"`
	Size       string `json:"-"`
	FileName   string `json:"-"`
	UserId     string `json:"-"`
	Time       int64  `json:"-"`
	ChunkSize  string `json:"c_size"`
	ChunkCount string `json:"c_count"`
	UploadId   string `json:"upid"`
}
type UploadInitReq struct {
	Hash     string `json:"hash" binding:"required,len=64,hexadecimal"`
	Size     string `json:"size" binding:"required,numeric,min=1"`
	FileName string `json:"name" binding:"required,min=1,max=512"`
}

var (
	ErrVolumeNotEnough error = errors.New("ErrVolumeNotEnough")
)

func (*FileApi) UploadInitLogic(ctx *gin.Context) {
	var Req UploadInitReq
	if err := ctx.ShouldBindJSON(&Req); err != nil {
		utils.ResponseWithMsg("[input data err]: "+err.Error(), ctx)
		return
	}
	size, err := strconv.ParseInt(Req.Size, 10, 64)

	if err != nil {
		switch t := err.(type) {
		case *strconv.NumError:
			if t.Err == strconv.ErrRange {
				// 处理溢出错误
				utils.ResponseWithMsg("文件过大: "+err.Error(), ctx)
			} else {
				// 处理其他解析错误
				global.Logger.Error("parseInt err: " + err.Error())
			}
		default:
			// 其他错误
			global.Logger.Error("parseInt err: " + err.Error())
		}
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	//判断容量
	_claims, _ := ctx.Get("claims")
	claims := _claims.(*utils.CustomClaims)
	var upId int64
	if upId, err = global.SnowFlakeCreater.Generate(); err != nil {
		global.Logger.Error("SnowFlakeCreater err: " + err.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}

	//是否秒传
	isSecTran := false
	err = global.DB.Transaction(func(tx *gorm.DB) error {
		// 直接原子更新，通过 WHERE 条件保证容量足够
		result := tx.Model(&model.User{}).
			Where("id = ? AND available_volume >= ?", claims.UserId, size).
			Update("available_volume", gorm.Expr("available_volume - ?", size))

		if result.Error != nil {
			return result.Error
		}
		// 判断是否扣减成功
		if result.RowsAffected == 0 {
			// 扣减失败，可容量不足
			return ErrVolumeNotEnough
		}

		//sec trans 秒传
		var fileModel model.File
		if e2 := tx.Take(&fileModel, "hash=?", Req.Hash).Error; e2 != nil {
			if errors.Is(e2, gorm.ErrRecordNotFound) {
				isSecTran = false
			} else {
				return e2
			}
		} else {
			isSecTran = true
			userFileModel := model.UserFile{
				UserId:   int64(claims.UserId),
				FileId:   int64(fileModel.ID),
				FileName: Req.FileName,
			}
			if e3 := tx.Create(&userFileModel).Error; e3 != nil {
				return e3
			}
		}
		LogModel := model.VolumeOpLog{
			UserID:   int64(claims.UserId),
			FileHash: Req.Hash,
			FileSize: uint64(size),
			Status:   0,
			UploadId: upId,
			FileName: Req.FileName,
		}
		if e1 := tx.Create(&LogModel).Error; e1 != nil {
			return e1
		}
		return nil
	})
	if err != nil && !errors.Is(err, ErrVolumeNotEnough) {
		global.Logger.Error("gorm trans err: " + err.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	if errors.Is(err, ErrVolumeNotEnough) {
		utils.ResponseWithCode("1003", ctx)
		return
	}
	if isSecTran {
		utils.ResponseWithCode("1005", ctx) //秒传成功
		return
	}

	// 设置分片大小
	chunkSize := global.Config.Upload.ChunkSize

	// 计算分片个数
	var chunkCount int64 = 0
	// 使用向上取整计算分片个数
	chunkCount = size / chunkSize
	if size%chunkSize != 0 {
		chunkCount++
	}
	upIdStr := strconv.Itoa(int(upId))
	Meta := FileMetaInfo{
		Hash:       Req.Hash,
		FileName:   Req.FileName,
		Size:       Req.Size,
		ChunkSize:  strconv.Itoa(int(chunkSize)),
		ChunkCount: strconv.Itoa(int(chunkCount)),
		UserId:     strconv.Itoa(claims.UserId),
		Time:       time.Now().Unix(),
		UploadId:   upIdStr,
	}
	//redis cs:meta:{userId}:{upId}
	hkey := global.FileMetaPrefix + Meta.UserId + ":" + upIdStr

	// 执行所有命令
	_, err = global.RDB.HSet(context.TODO(), hkey, map[string]interface{}{
		"fn":     Req.FileName,
		"size":   Req.Size,
		"ccnt":   chunkCount,
		"act":    Meta.Time, //logic active time
		"hash":   Req.Hash,
		"status": "0",
		/*
			status:
			  0: init
			  1: upload
			  2: finish
			  3: cancel
		*/
	}).Result()
	if err != nil {
		global.Logger.Error("redis hset err: " + err.Error())
		utils.ResponseWithMsg("[internal server err]", ctx)
		return
	}
	utils.ResponseWithData(Meta, ctx)
}
