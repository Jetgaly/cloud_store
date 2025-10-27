package main

import (
	"encoding/json"

	"go.uber.org/zap"
)

// https://pkg.go.dev/go.uber.org/zap@v1.24.0#hdr-Configuring_Zap
func main() {
	// 表示 zap.Config 的 json 原始编码
	// outputPath: 设置日志输出路径，日志内容输出到标准输出和文件 logs.log
	// errorOutputPaths：设置错误日志输出路径
	rawJSON := []byte(`{
      "level": "debug",
      "encoding": "json",
      "outputPaths": ["stdout", "./logs.log"],
      "errorOutputPaths": ["stderr"],
      "initialFields": {"foo": "bar"},
      "encoderConfig": {
        "messageKey": "message-customer",
        "levelKey": "level",
        "levelEncoder": "lowercase"
      }
    }`)

	// 把 json 格式数据解析到 zap.Config struct
	var cfg zap.Config
	if err := json.Unmarshal(rawJSON, &cfg); err != nil {
		panic(err)
	}
	// cfg.Build() 为配置对象创建一个 Logger
	// zap.Must() 封装了 Logger，Must()函数如果返回值不是 nil，就会报 panic。也就是检查Build是否错误
	logger := zap.Must(cfg.Build())
	defer logger.Sync()

	logger.Info("logger construction succeeded")
}

/*
Must() 函数
//  var logger = zap.Must(zap.NewProduction())
func Must(logger *Logger, err error) *Logger {
    if err != nil {
        panic(err)
    }

    return logger
}
*/