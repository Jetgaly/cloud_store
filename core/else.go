package core

import (
	"cloud_store/global"
	"cloud_store/utils"
	"time"
)

func InitUploadTempDir(path string) {
	if err := utils.CreateDir(path); err != nil {
		global.Logger.DPanic(path + "create err: " + err.Error())
	}
}

func InitSnowFlake() {

	//snowflakes
	nodeID, e1 := global.RDB.Incr(global.RDB.Context(), "cs:nodeId").Result()
	if e1 != nil {
		global.Logger.Fatal("[Redis]cs:nodeId获取失败,err:" + e1.Error())
		return
	}
	node, e2 := utils.NewSafeSnowFlakeCreater(nodeID, time.Duration(500)*time.Millisecond)
	if e2 != nil {
		global.Logger.Fatal("[SnowFlake]初始化失败,err:" + e2.Error())
	}
	global.SnowFlakeCreater = node
}
