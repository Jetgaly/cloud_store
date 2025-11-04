package core

import (
	"cloud_store/cron"
	"cloud_store/global"
	gcrontask "cloud_store/global/gCronTask"
)

func InitCronTask() {
	//todo 优雅关闭
	//清理任务
	RConn1, err := global.RMQ.RmqPool.Pool.Get()
	if err != nil {
		global.Logger.Fatal("CronTask cleanTask init err:" + err.Error())
	}

	cleanTask := cron.Consumers{
		Conn:    RConn1,
		Queue:   "cs.clean.timeout",
		Handler: cron.CleanHandler,
		Count:   5,
	}
	gcrontask.CleanTask = &cleanTask
	cleanTask.Start()

	//uploadOSS任务
	RConn2, err := global.RMQ.RmqPool.Pool.Get()
	if err != nil {
		global.Logger.Fatal("CronTask OSSTask init err:" + err.Error())
	}

	OSSTask := cron.Consumers{
		Conn:    RConn2,
		Queue:   "cs.oss.queue",
		Handler: cron.UploadOSSHandler,
		Count:   5,
	}
	gcrontask.OSSUploadTask = &OSSTask
	OSSTask.Start()
}
