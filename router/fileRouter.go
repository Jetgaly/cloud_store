package router

import (
	"cloud_store/api"
	"cloud_store/middleware"

	"github.com/gin-gonic/gin"
)

func InitFileRouter(e *gin.Engine) {
	r := e.Group("/api")
	r.POST("/file", middleware.JwtAuth(), api.Handler.FileApi.UploadInitLogic)
	r.POST("/file/chunk", middleware.JwtAuth(), api.Handler.FileApi.UploadLogic)
	r.POST("/file/finish", middleware.JwtAuth(), api.Handler.FileApi.UploadFinishLogic)
	r.POST("/file/cancel", middleware.JwtAuth(), api.Handler.FileApi.UploadCancelLogic)
	r.GET("/file/:id", middleware.JwtAuth(), api.Handler.FileApi.DownloadLogic)
}
