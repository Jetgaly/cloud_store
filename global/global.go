package global

import (
	"cloud_store/config"
	"cloud_store/utils"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	Engine           *gin.Engine
	Config           *config.Config
	Logger           *zap.Logger
	DB               *gorm.DB
	EmailSender      *utils.EmailSender
	RDB              *redis.Client
	SnowFlakeCreater *utils.SafeSnowFlakeCreater
)

const (
	EmailCodePrefix string = "cs:"
	FileMetaPrefix  string = "cs:meta:"
	FileSetPrefix   string = "cs:set:"
)
