package core

import (
	"cloud_store/global"

	"go.uber.org/zap"
)

func InitLogger() {
	var cfg zap.Config
	if global.Config.Logger.Dev {
		//dev
		cfg = zap.NewDevelopmentConfig()
		cfg.OutputPaths = global.Config.Logger.OutputPaths
		cfg.ErrorOutputPaths = global.Config.Logger.ErrorOutputPaths
		logger := zap.Must(cfg.Build())
		global.Logger = logger
	} else {
		//pro
		cfg = zap.NewProductionConfig()
		cfg.OutputPaths = global.Config.Logger.OutputPaths
		cfg.ErrorOutputPaths = global.Config.Logger.ErrorOutputPaths
		logger := zap.Must(cfg.Build())
		global.Logger = logger
	}
}
