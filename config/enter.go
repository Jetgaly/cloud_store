package config

type Config struct {
	Mysql  Mysql  `yaml:"mysql"`
	Email  Email  `yaml:"Email"`
	Logger Logger `yaml:"logger"`
	Redis  Redis  `yaml:"redis"`
	Upload Upload `yaml:"upload"`
}
