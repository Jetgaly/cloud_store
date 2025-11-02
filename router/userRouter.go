package router

import (
	"cloud_store/api"

	"github.com/gin-gonic/gin"
)

func InitUserRouter(e *gin.Engine) {
	r := e.Group("/api")

	r.POST("/user", api.Handler.UserApi.CreateUserLogic)
	r.POST("/code", api.Handler.UserApi.GetEmailCodeLogic)
	r.POST("/login", api.Handler.UserApi.LoginLogic)
}
