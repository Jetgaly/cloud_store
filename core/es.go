package core

import (
	"cloud_store/global"
	"github.com/elastic/go-elasticsearch/v7"
)

func InitES() {
	cfg := elasticsearch.Config{
		Addresses: []string{
			global.Config.ES.Addr,
		},
	}
	cli, err := elasticsearch.NewClient(cfg)
	if err != nil {
		global.Logger.Fatal("es init err:" + err.Error())
	}
	global.ESCli = cli
}
