package core

import "cloud_store/global"

func init() {
	InitConf()
	InitLogger()
	InitEmail()
	InitGorm()
	CreateTables()
	InitRedis()
	InitUploadTempDir(global.Config.Upload.TempPath)
	InitSnowFlake()
	InitRMQ()
	InitMinIO()
	InitRedLock()
	InitCronTask()
}
