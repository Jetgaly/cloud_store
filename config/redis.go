package config

import (
	"strconv"
)

type Redis struct {
	// 	 ip: 127.0.0.1
	//   port: 6379
	//   password: ""
	//   pool_size: 100
	IP       string `json:"ip" yaml:"ip"`
	Port     int    `json:"port" yaml:"port"`
	PassWord string `json:"password" yaml:"password"`
	PoolSize int    `json:"pool_size" yaml:"pool_size"`
}

func (r *Redis) Addr() string {
	return r.IP + ":" + strconv.Itoa(r.Port)
}
