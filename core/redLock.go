package core

import (
	"cloud_store/global"
	"cloud_store/utils"
	"context"
	"fmt"

	gr "github.com/redis/go-redis/v9"
)

func InitRedLock() {
	var clis []*gr.Client
	for i, v := range global.Config.RedLock.Hosts {
		r := gr.NewClient(&gr.Options{
			Addr:     v,
			Password: global.Config.RedLock.Passes[i],
		})
		t := r.Ping(context.Background())
		if t.Err() != nil {
			global.Logger.Fatal(fmt.Sprintf("[RedLock]初始化失败,%s", t.Err().Error()))
		}
		clis = append(clis, r)
	}
	rlc, e3 := utils.NewRedLockCreater(clis)
	if e3 != nil {
		global.Logger.Fatal(fmt.Sprintf("[RedLock]初始化失败,%s", e3.Error()))
	}
	global.RedLockCreater = rlc
}
