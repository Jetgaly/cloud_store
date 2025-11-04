package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

func main() {
	accessKey := os.Getenv("OSS_ACCESS_KEY_ID")
	secretKey := os.Getenv("OSS_ACCESS_KEY_SECRET")

	if accessKey == "" || secretKey == "" {
		fmt.Println("请设置环境变量")
		return
	}
	//fmt.Println("-"+accessKey+"-", "-"+secretKey+"-l1")
	cfg := oss.LoadDefaultConfig().
		WithCredentialsProvider(credentials.NewEnvironmentVariableCredentialsProvider()).
		WithRegion("cn-shenzhen") // 填写Bucket所在地域，以华东1（杭州）为例，Region填写为cn-hangzhou，SDK会根据Region自动构造HTTPS访问域名
	client := oss.NewClient(cfg)
	bucketName := "cs-bucket-1104"

	// r, err := client.ListBuckets(context.TODO(), &oss.ListBucketsRequest{})
	// if err != nil {
	// 	fmt.Printf("列出 Buckets 失败: %v", err)
	// 	return
	// }
	// fmt.Println(r.Buckets)
	// return

	f, e := os.Open("/home/jjet/projects/GoProjects/cloud_store/test/file/bigImg.jpg")
	if e != nil {
		fmt.Println("open file err", e.Error())
		return
	}
	defer f.Close()
	uploader := oss.NewUploader(client)
	r, ee := uploader.UploadFrom(context.TODO(), &oss.PutObjectRequest{
		Bucket: oss.Ptr(bucketName),
		Key:    oss.Ptr("test.jpg")},
		f)
	if ee != nil {
		fmt.Println("upload file err", ee.Error())
		return
	}
	fmt.Println("r", r)
}
