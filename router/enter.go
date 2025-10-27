package router

import (
	"cloud_store/global"

	"github.com/gin-gonic/gin"
)

func InitRouter() {
	gin.SetMode("release")
	global.Engine = gin.Default()
	// global.Engine = gin.New()
	// global.Engine.Use(ginzap.Ginzap(global.Logger, time.RFC3339, true))
	// // Logs all panic to error log
	// //   - stack means whether output the stack info.
	// global.Engine.Use(ginzap.RecoveryWithZap(global.Logger, true))

	InitUserRouter(global.Engine)
	InitFileRouter()
}
