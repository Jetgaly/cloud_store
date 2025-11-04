package cron

import (
	"cloud_store/api/file"
	"cloud_store/global"
	"cloud_store/model"
	"context"
	"encoding/json"
	"fmt"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/minio/minio-go/v7"
)

func UploadOSSHandler(msg []byte) error {
	var msgModel file.OSSMqMsg
	err := json.Unmarshal(msg, &msgModel)
	if err != nil {
		global.Logger.Error(fmt.Sprintf("UploadOSSHandler json.Unmarshal err:%s", err.Error()))
		return fmt.Errorf("UploadOSSHandler json.Unmarshal err:%s", err.Error())
	}
	var obj *minio.Object
	obj, err = global.MinioCli.GetObject(context.TODO(), global.Config.MinIO.UploadBucket, msgModel.MinIOPath, minio.GetObjectOptions{})
	if err != nil {
		global.Logger.Error(fmt.Sprintf("UploadOSSHandler MinioCli.GetObject err:%s,fileId:%s", err.Error(), msgModel.Id))
		return fmt.Errorf("UploadOSSHandler MinioCli.GetObject err:%s,fileId:%s", err.Error(), msgModel.Id)
	}
	uploader := oss.NewUploader(global.OSSCli)
	_, err = uploader.UploadFrom(context.TODO(), &oss.PutObjectRequest{
		Bucket: oss.Ptr(global.Config.OSS.BucketName),
		Key:    oss.Ptr(msgModel.MinIOPath)},
		obj,
	)
	if err != nil {
		global.Logger.Error(fmt.Sprintf("UploadOSSHandler ossupload err:%s,fileId:%s", err.Error(), msgModel.Id))
		return fmt.Errorf("UploadOSSHandler ossupload err:%s,fileId:%s", err.Error(), msgModel.Id)
	}
	err = global.DB.Model(&model.File{}).Where("id=?", msgModel.Id).Update("position", 1).Error
	if err != nil {
		global.Logger.Error(fmt.Sprintf("UploadOSSHandler gorm update err:%s,fileId:%s", err.Error(), msgModel.Id))
		return fmt.Errorf("UploadOSSHandler gorm update err:%s,fileId:%s", err.Error(), msgModel.Id)
	}
	return nil
}
