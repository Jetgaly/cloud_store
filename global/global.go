package global

import (
	"cloud_store/config"

	"cloud_store/utils"
	RMQUtils "cloud_store/utils/RabbitMQ"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/minio/minio-go/v7"
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
	RMQ              *RMQUtils.RMQ
	MinioCli         *minio.Client
	RedLockCreater   *utils.RedLockCreater
	OSSCli           *oss.Client
)

const (
	EmailCodePrefix string = "cs:"
	FileMetaPrefix  string = "cs:meta:"
	FileSetPrefix   string = "cs:set:"
	RedLockPrefix   string = "cs:lock:"
	LimitKeyPrefix string = "cs:limit:"
)
