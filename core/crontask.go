package core

import (
	"cloud_store/cron"
	"cloud_store/global"
	gcrontask "cloud_store/global/gCronTask"
)

func InitCronTask() {
	//todo 优雅关闭
	//清理任务
	RConn, err := global.RMQ.RmqPool.Pool.Get()
	if err != nil {
		global.Logger.Fatal("CronTask init err:" + err.Error())
	}

	cleanTask := cron.Consumers{
		Conn:    RConn,
		Queue:   "cs.clean.timeout",
		Handler: cron.CleanHandler,
		Count:   5,
	}
	gcrontask.CleanTask = &cleanTask
	cleanTask.Start()

	//uploadOSS任务

}
