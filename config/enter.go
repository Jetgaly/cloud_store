package config

type Config struct {
	Mysql Mysql `yaml:"mysql"`
	Email Email `yaml:"email"`
	Logger Logger `yaml:"logger"`
	Redis Redis `yaml:"redis"`
}
