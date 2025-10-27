package main

import (
	_ "cloud_store/core"
	"cloud_store/global"
	"cloud_store/router"
	//"time"
	//ginzap "github.com/gin-contrib/zap"
)

func main() {
	
	router.InitRouter()

	global.Engine.Run(":8080")
}
