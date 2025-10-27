package config

import "strconv"

type Mysql struct {
	Host     string `yaml:"host"`
	Port     int `yaml:"port"`
	DB       string `yaml:"db"`
	Config   string `yaml:"config"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	LogLevel string `yaml:"log_level"` //日志等级，debug就是输出全部sql，dev，release
}

func (m *Mysql) Dsn() string {
	return m.User + ":" + m.Password + "@tcp(" + m.Host + ":" + strconv.Itoa(m.Port) + ")/" + m.DB + "?" + m.Config
}
