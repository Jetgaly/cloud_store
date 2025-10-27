package core

import (
	"cloud_store/config"
	"cloud_store/global"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
)

func InitConf() {
	const ConfFilePath = "conf.yml"
	c := &config.Config{}
	ConfFile, err := os.ReadFile(ConfFilePath)
	if err != nil {
		panic(fmt.Sprintf("conf.yml err:%s", err.Error()))
	}
	err = yaml.Unmarshal(ConfFile, c)
	if err != nil {
		panic(fmt.Sprintf("conf.yml Unmarshal err:%s", err.Error()))
	}
	global.Config = c
}
