package core

import (
	"cloud_store/global"
	"context"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func InitMinIO() {
	ctx := context.Background()
	endpoint := global.Config.MinIO.Endpoint
	accessKeyID := global.Config.MinIO.AccessKeyID
	secretAccessKey := global.Config.MinIO.SecretAccessKey
	useSSL := global.Config.MinIO.UseSSL

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		global.Logger.Fatal("minio init err: " + err.Error())
	}

	// Make a new bucket
	bucketName := global.Config.MinIO.UploadBucket
	location := "any"

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	if err != nil {
		// Check to see if we already own this bucket
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			global.Logger.Info(fmt.Sprintf("We already own %s", bucketName))
		} else {
			global.Logger.Fatal("minio init err: " + err.Error())
		}
	} 
	
}
