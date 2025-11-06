package file

import (
	"cloud_store/global"
	"cloud_store/model"
	"cloud_store/utils"
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"gorm.io/gorm"
)

type DownloadReq struct {
	Id string
}

func (*FileApi) DownloadLogic(ctx *gin.Context) {
	var Req DownloadReq
	Req.Id = ctx.Param("id")
	if Req.Id == "" {
		utils.ResponseWith400Msg("[input data err]: id ", ctx)
		return
	}
	var id int
	var err error
	if id, err = strconv.Atoi(Req.Id); err != nil {
		global.Logger.Error(fmt.Sprintf("strconv.Atoi(Req.Id) err: %s", err.Error()))
		ctx.Status(500)
		return
	}
	if id <= 0 {
		utils.ResponseWith400Msg("[input data err]: id  ", ctx)
		return
	}
	rangeHeader := ctx.GetHeader("Range")

	var userFileModel model.UserFile
	err = global.DB.Take(&userFileModel, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ResponseWith400Code("1011", ctx)
			return
		}
		global.Logger.Error(fmt.Sprintf("mysql gorm err: %s", err.Error()))
		ctx.Status(500)
		return
	}
	_claims, _ := ctx.Get("claims")
	claims := _claims.(*utils.CustomClaims)
	if userFileModel.UserId != int64(claims.UserId) {
		//判断文件是否属于该用户
		utils.ResponseWith400Code("1011", ctx)
		return
	}
	var fileModel model.File
	err = global.DB.Take(&fileModel, userFileModel.FileId).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.ResponseWith400Code("1011", ctx)
			return
		}
		global.Logger.Error(fmt.Sprintf("mysql gorm err: %s", err.Error()))
		ctx.Status(500)
		return
	}

	bucketName := global.Config.MinIO.UploadBucket
	if fileModel.Position == 0 {
		stat, merr := global.MinioCli.StatObject(context.Background(), bucketName, fileModel.Name, minio.GetObjectOptions{})
		if merr != nil {
			global.Logger.Error(fmt.Sprintf("minio get obj stat err: %s", merr.Error()))
			ctx.Status(500)
			return
		}
		totalSize := stat.Size
		var start, end int64
		if rangeHeader == "" {
			//直接返回整个文件
			start = 0
			end = totalSize - 1
			ctx.Status(206)
			ctx.Header("Content-Length", fmt.Sprintf("%d", totalSize))
			ctx.Header("Content-Type", "application/octet-stream")
		} else {
			if !strings.HasPrefix(rangeHeader, "bytes=") {
				utils.ResponseWith400Code("1012", ctx)
				return
			}
			ranges := strings.TrimPrefix(rangeHeader, "bytes=")
			parts := strings.Split(ranges, "-")
			if len(parts) != 2 {
				//"bytes=1-" => 解析也是len == 2
				utils.ResponseWith400Code("1012", ctx)
				return
			}
			if parts[0] == "" {
				n, perr := strconv.ParseInt(parts[1], 10, 64)
				if perr != nil {
					global.Logger.Error(fmt.Sprintf("parseInt err: %s", perr.Error()))
					ctx.Status(500)
					return
				}
				if n <= 0 {
					utils.ResponseWith400Code("1013", ctx)
					return
				}
				if n > totalSize {
					n = totalSize
				}
				start = totalSize - n
				end = totalSize - 1
			} else {
				//解析 start
				s, perr := strconv.ParseInt(parts[0], 10, 64)
				if perr != nil {
					global.Logger.Error(fmt.Sprintf("parseInt err: %s", perr.Error()))
					ctx.Status(500)
					return
				}
				if s < 0 || s >= totalSize {
					utils.ResponseWith400Code("1013", ctx)
					return
				}
				start = s

				// 情况 2: bytes=100- 不指定 end
				if parts[1] == "" {
					end = totalSize - 1
				} else {
					// 情况 3: bytes=100-200
					e, perr1 := strconv.ParseInt(parts[1], 10, 64)
					if perr1 != nil {
						global.Logger.Error(fmt.Sprintf("parseInt err: %s", perr1.Error()))
						ctx.Status(500)
						return
					}
					if e < start {
						utils.ResponseWith400Code("1013", ctx)
						return
					}
					if e >= totalSize {
						e = totalSize - 1
					}
					end = e
				}
			}
			ctx.Status(206)
			ctx.Header("Content-Type", "application/octet-stream")
			ctx.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, totalSize))
			ctx.Header("Content-Length", fmt.Sprintf("%d", end-start+1))
		}
		fmt.Println("start-end", start, end)
		opts := minio.GetObjectOptions{}
		opts.SetRange(start, end)
		var object *minio.Object
		object, err = global.MinioCli.GetObject(context.Background(), bucketName, fileModel.Name, opts)
		if err != nil {
			global.Logger.Error(fmt.Sprintf(" strconv.ParseInt(parts[0], 10, 64) err: %s", err.Error()))
			ctx.Status(500)
			return
		}
		defer object.Close()
		if _, err := io.Copy(ctx.Writer, object); err != nil {
			// 用户取消下载，err 可能是 broken pipe，不要 panic
			global.Logger.Info("download aborted:" + err.Error())
		}
		return
	} else {
		//oss
		result, oerr := global.OSSCli.Presign(context.TODO(), &oss.GetObjectRequest{
			Bucket: oss.Ptr(global.Config.OSS.BucketName),
			Key:    oss.Ptr(fileModel.Path),
		},
			oss.PresignExpires(10*time.Minute),
		)
		if oerr != nil {
			global.Logger.Error(fmt.Sprintf("OSS Presign err: %s", oerr.Error()))
			ctx.Status(500)
			return
		}
		OSSResp := struct {
			URL string    `json:"url"`
			Ex  time.Time `json:"ex"`
		}{
			URL: result.URL,
			Ex:  result.Expiration,
		}
		utils.ResponseWithData(OSSResp, ctx)
	}
}
