package core

import (
	"cloud_store/global"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

func InitOSSCli() {
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion(global.Config.OSS.Region)
	global.OSSCli = oss.NewClient(cfg)
}
