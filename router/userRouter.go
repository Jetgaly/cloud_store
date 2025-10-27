package router

import (
	"cloud_store/api"

	"github.com/gin-gonic/gin"
)

func InitUserRouter(e *gin.Engine) {
	r := e.Group("/api")

	r.POST("/user", api.Handler.UserApi.CreateUserLogic)
}
